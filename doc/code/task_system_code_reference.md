# 任务系统参考代码

这份文档只作为你手敲代码时的参考，不是让你一次性复制进项目。

建议你按 `task_system_coding_guide.md` 的顺序敲：

```text
model -> migration -> repository -> ai -> service -> api -> worker -> main
```

第一阶段先用 fake AI，目标是跑通：

```text
POST /api/articles/:id/optimize
    ↓
创建 article_optimize 任务
    ↓
worker 执行任务
    ↓
更新文章
    ↓
任务 completed
```

## 1. `internal/model/task.go`

```go
package model

import "time"

type TaskType string

const (
	TaskTypeArticleOptimize TaskType = "article_optimize"
	TaskTypeArticlePublish  TaskType = "article_publish"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCanceled   TaskStatus = "canceled"
)

type Task struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Type      TaskType   `json:"type" gorm:"size:50;not null;index"`
	Status    TaskStatus `json:"status" gorm:"size:20;not null;default:pending;index"`
	ArticleID uint       `json:"article_id" gorm:"not null;index"`

	Payload *string `json:"payload,omitempty" gorm:"type:jsonb"`
	Result  *string `json:"result,omitempty" gorm:"type:jsonb"`
	Error   string  `json:"error" gorm:"type:text"`

	RetryCount int `json:"retry_count" gorm:"default:0"`
	MaxRetries int `json:"max_retries" gorm:"default:3"`

	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	Article *Article `json:"article,omitempty" gorm:"foreignKey:ArticleID"`
}

func (Task) TableName() string {
	return "tasks"
}
```

说明：

- `Payload` 和 `Result` 用 `*string`，是为了避免空字符串写入 PostgreSQL `jsonb` 报错。
- 写入 `Result` 时，要保证字符串是合法 JSON。

## 2. `migrations/000002_create_tasks_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    article_id INTEGER NOT NULL,

    payload JSONB,
    result JSONB,
    error TEXT,

    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,

    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_tasks_article
        FOREIGN KEY(article_id)
        REFERENCES articles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_article_id ON tasks(article_id);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
```

## 3. `migrations/000002_create_tasks_table.down.sql`

```sql
DROP TABLE IF EXISTS tasks;
```

## 4. `internal/repository/task.go`

```go
package repository

import (
	"errors"
	"time"

	"nebula-ai-api/internal/model"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) List(offset, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	if err := r.db.Model(&model.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRepository) ListByArticleID(articleID uint, offset, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Model(&model.Task{}).Where("article_id = ?", articleID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRepository) FindRunningTask(articleID uint, taskType model.TaskType) (*model.Task, error) {
	var task model.Task
	err := r.db.Where(
		"article_id = ? AND type = ? AND status IN ?",
		articleID,
		taskType,
		[]model.TaskStatus{model.TaskStatusPending, model.TaskStatusProcessing},
	).Order("created_at DESC").First(&task).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) FindNextPending() (*model.Task, error) {
	var task model.Task
	err := r.db.Where("status = ?", model.TaskStatusPending).
		Order("created_at ASC").
		First(&task).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) MarkProcessing(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.TaskStatusProcessing,
		"started_at": now,
		"error":      "",
	}).Error
}

func (r *TaskRepository) MarkCompleted(id uint, result string) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      model.TaskStatusCompleted,
		"result":      result,
		"finished_at": now,
	}).Error
}

func (r *TaskRepository) MarkFailed(id uint, errMsg string) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      model.TaskStatusFailed,
		"error":       errMsg,
		"finished_at": now,
	}).Error
}
```

## 5. `internal/ai/article_optimizer.go`

```go
package ai

import (
	"context"

	"nebula-ai-api/internal/model"
)

type ArticleOptimizeResult struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}

type ArticleOptimizer interface {
	OptimizeArticle(ctx context.Context, article *model.Article) (*ArticleOptimizeResult, error)
}
```

## 6. `internal/ai/fake_optimizer.go`

```go
package ai

import (
	"context"

	"nebula-ai-api/internal/model"
)

type FakeArticleOptimizer struct{}

func NewFakeArticleOptimizer() *FakeArticleOptimizer {
	return &FakeArticleOptimizer{}
}

func (o *FakeArticleOptimizer) OptimizeArticle(ctx context.Context, article *model.Article) (*ArticleOptimizeResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return &ArticleOptimizeResult{
		Title:   article.Title,
		Summary: "这是 Fake AI 生成的文章摘要，用于验证任务系统流程。",
		Tags:    []string{"Go", "AI", "Nebula"},
		Content: article.Content + "\n\n## AI 优化建议\n\n这是 Fake AI 生成的优化内容，用于测试任务执行链路。",
	}, nil
}
```

## 7. `internal/service/task_service.go`

```go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"nebula-ai-api/internal/ai"
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/repository"
)

type TaskService struct {
	taskRepo    *repository.TaskRepository
	articleRepo *repository.ArticleRepository
	optimizer   ai.ArticleOptimizer
}

func NewTaskService(taskRepo *repository.TaskRepository, articleRepo *repository.ArticleRepository, optimizer ai.ArticleOptimizer) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		articleRepo: articleRepo,
		optimizer:   optimizer,
	}
}

func (s *TaskService) CreateArticleOptimizeTask(articleID uint) (*model.Task, error) {
	if _, err := s.articleRepo.GetByID(articleID); err != nil {
		return nil, errors.New("文章不存在")
	}

	runningTask, err := s.taskRepo.FindRunningTask(articleID, model.TaskTypeArticleOptimize)
	if err != nil {
		return nil, err
	}
	if runningTask != nil {
		return runningTask, nil
	}

	payload := fmt.Sprintf(`{"article_id":%d}`, articleID)
	task := &model.Task{
		Type:       model.TaskTypeArticleOptimize,
		Status:     model.TaskStatusPending,
		ArticleID:  articleID,
		Payload:    &payload,
		MaxRetries: 3,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) GetByID(id uint) (*model.Task, error) {
	return s.taskRepo.GetByID(id)
}

func (s *TaskService) List(page, pageSize int) ([]model.Task, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	return s.taskRepo.List(offset, pageSize)
}

func (s *TaskService) ListByArticleID(articleID uint, page, pageSize int) ([]model.Task, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	offset := (page - 1) * pageSize
	return s.taskRepo.ListByArticleID(articleID, offset, pageSize)
}

func (s *TaskService) ExecuteNext(ctx context.Context) error {
	task, err := s.taskRepo.FindNextPending()
	if err != nil {
		return err
	}
	if task == nil {
		return nil
	}
	return s.Execute(ctx, task)
}

func (s *TaskService) Execute(ctx context.Context, task *model.Task) error {
	if task == nil {
		return errors.New("任务不能为空")
	}

	if err := s.taskRepo.MarkProcessing(task.ID); err != nil {
		return err
	}

	var result string
	var err error

	switch task.Type {
	case model.TaskTypeArticleOptimize:
		result, err = s.executeArticleOptimize(ctx, task)
	default:
		err = fmt.Errorf("不支持的任务类型: %s", task.Type)
	}

	if err != nil {
		_ = s.taskRepo.MarkFailed(task.ID, err.Error())
		return err
	}

	return s.taskRepo.MarkCompleted(task.ID, result)
}

func (s *TaskService) executeArticleOptimize(ctx context.Context, task *model.Task) (string, error) {
	article, err := s.articleRepo.GetByID(task.ArticleID)
	if err != nil {
		return "", errors.New("文章不存在")
	}

	optimizeResult, err := s.optimizer.OptimizeArticle(ctx, article)
	if err != nil {
		return "", err
	}

	if optimizeResult.Title != "" {
		article.Title = optimizeResult.Title
	}
	if optimizeResult.Summary != "" {
		article.Summary = optimizeResult.Summary
	}
	if len(optimizeResult.Tags) > 0 {
		article.Tags = strings.Join(optimizeResult.Tags, ",")
	}
	if optimizeResult.Content != "" {
		article.Content = optimizeResult.Content
	}

	article.AIOptimized = true
	article.Status = model.StatusOptimized

	if err := s.articleRepo.Update(article); err != nil {
		return "", err
	}

	resultBytes, err := json.Marshal(optimizeResult)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
```

## 8. `internal/api/task.go`

```go
package api

import (
	"nebula-ai-api/internal/service"
	"nebula-ai-api/pkg/query"
	"nebula-ai-api/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) List(c *gin.Context) {
	listQuery := query.NewListQuery(c)

	var tasks interface{}
	var total int64
	var err error

	articleIDParam := c.Query("article_id")
	if articleIDParam != "" {
		articleID, parseErr := strconv.ParseUint(articleIDParam, 10, 32)
		if parseErr != nil {
			response.BadRequest(c, "无效的文章ID")
			return
		}
		tasks, total, err = h.taskService.ListByArticleID(uint(articleID), listQuery.Page, listQuery.PageSize)
	} else {
		tasks, total, err = h.taskService.List(listQuery.Page, listQuery.PageSize)
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, query.NewListResponse(tasks, total, listQuery))
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.taskService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	response.Success(c, task)
}
```

## 9. `internal/api/article.go` 需要改的地方

这个文件已经很长，不建议你整文件重写，只改三处。

### 9.1 修改 `ArticleHandler`

```go
type ArticleHandler struct {
	articleService *service.ArticleService
	taskService    *service.TaskService
}
```

### 9.2 修改构造函数

```go
func NewArticleHandler(articleService *service.ArticleService, taskService *service.TaskService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
		taskService:    taskService,
	}
}
```

### 9.3 在文件底部新增 `Optimize`

```go
func (h *ArticleHandler) Optimize(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	task, err := h.taskService.CreateArticleOptimizeTask(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, task)
}
```

注意：

- `article.go` 里已经导入了 `strconv`，所以新增方法一般不需要新增 import。
- `Optimize` 只创建任务，不直接调用 AI。

## 10. `internal/api/router.go`

```go
package api

import (
	"nebula-ai-api/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	articleService *service.ArticleService,
	markdownService *service.MarkdownService,
	taskService *service.TaskService,
) {
	articleHandler := NewArticleHandler(articleService, taskService)
	markdownHandler := NewMarkdownHandler(markdownService)
	taskHandler := NewTaskHandler(taskService)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		articles := api.Group("/articles")
		{
			articles.POST("", articleHandler.Create)
			articles.GET("", articleHandler.List)
			articles.GET("/:id", articleHandler.GetByID)
			articles.PUT("/:id", articleHandler.Update)
			articles.DELETE("/:id", articleHandler.Delete)
			articles.PUT("/:id/status", articleHandler.UpdateStatus)
			articles.POST("/:id/platforms", articleHandler.AddPublishPlatform)
			articles.POST("/:id/optimize", articleHandler.Optimize)
		}

		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.List)
			tasks.GET("/:id", taskHandler.GetByID)
		}

		markdown := api.Group("/markdown")
		{
			markdown.POST("/convert/csdn", markdownHandler.ConvertToCSDN)
		}
	}
}
```

## 11. `internal/worker/task_worker.go`

```go
package worker

import (
	"context"
	"log"
	"time"

	"nebula-ai-api/internal/service"
)

type TaskWorker struct {
	taskService *service.TaskService
	interval    time.Duration
}

func NewTaskWorker(taskService *service.TaskService, interval time.Duration) *TaskWorker {
	return &TaskWorker{
		taskService: taskService,
		interval:    interval,
	}
}

func (w *TaskWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("任务 worker 已启动")

	for {
		select {
		case <-ctx.Done():
			log.Println("任务 worker 已停止")
			return
		case <-ticker.C:
			if err := w.taskService.ExecuteNext(ctx); err != nil {
				log.Println("任务执行失败:", err)
			}
		}
	}
}
```

## 12. `cmd/main.go`

```go
package main

import (
	"context"
	"log"
	"time"

	"nebula-ai-api/internal/ai"
	"nebula-ai-api/internal/api"
	"nebula-ai-api/internal/config"
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/repository"
	"nebula-ai-api/internal/service"
	"nebula-ai-api/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("警告: 无法加载 .env 文件，使用默认配置")
	}

	cfg := config.Load()

	dsn := "host=" + cfg.DBHost +
		" port=" + cfg.DBPort +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" sslmode=disable TimeZone=Asia/Shanghai"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	if err := db.AutoMigrate(&model.Article{}, &model.Task{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	articleRepo := repository.NewArticleRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	articleOptimizer := ai.NewFakeArticleOptimizer()

	articleService := service.NewArticleService(articleRepo)
	markdownService := service.NewMarkdownService()
	taskService := service.NewTaskService(taskRepo, articleRepo, articleOptimizer)

	r := gin.Default()

	api.SetupRoutes(r, articleService, markdownService, taskService)

	taskWorker := worker.NewTaskWorker(taskService, 2*time.Second)
	go taskWorker.Start(context.Background())

	log.Println("服务器启动，端口: " + cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
```

说明：

- 这里仍然保留了 `AutoMigrate`，方便开发阶段。
- 如果你决定完全使用 migrations，后续可以删除 `AutoMigrate`。
- 第一阶段先用 `FakeArticleOptimizer`，不要急着接 DeepSeek。

## 13. 验证命令

写完后先跑：

```bash
go test ./...
```

如果本机 Go cache 被权限拦住，可以临时指定：

```bash
GOCACHE=/private/tmp/nebula-go-build-cache go test ./...
```

## 14. 手动验证流程

先启动数据库：

```bash
docker-compose up -d postgres
```

启动后端：

```bash
go run ./cmd
```

假设已经有文章 ID 为 `1`，创建优化任务：

```bash
curl -X POST http://localhost:8080/api/articles/1/optimize
```

查询任务：

```bash
curl http://localhost:8080/api/tasks/1
```

查询文章：

```bash
curl http://localhost:8080/api/articles/1
```

如果看到：

```text
task.status = completed
article.ai_optimized = true
article.status = 1
```

说明第一阶段任务系统已经跑通。

## 15. 下一步

Fake AI 跑通以后，再做真实 AI：

```text
internal/ai/deepseek_optimizer.go
```

那一步再接：

- API key
- HTTP client
- Prompt
- JSON 格式解析
- 超时控制
- AI 调用日志

不要在任务系统还没跑通前接真实模型。
