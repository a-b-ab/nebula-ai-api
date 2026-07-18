# Nebula AI 任务系统设计文档

## 1. 背景

当前后端已经有：

- `Article` 模型和文章 CRUD
- `Task` 模型雏形
- `api -> service -> repository -> model` 分层
- PostgreSQL 和 Redis 的配置项
- 前端已经预留 `/tasks` 接口调用

但后端目前还没有完整任务系统：

- 没有 `TaskRepository`
- 没有 `TaskService`
- 没有 `TaskHandler`
- 没有 `/api/tasks` 路由
- 没有 worker 执行任务
- migrations 里还没有 `tasks` 表

所以任务系统的第一阶段目标不是一次性做复杂调度平台，而是先把 Nebula AI 的核心链路跑通：

```text
用户触发动作
    ↓
创建任务
    ↓
后台执行
    ↓
更新任务状态
    ↓
写回文章或发布结果
    ↓
前端查询任务进度
```

## 2. 设计目标

任务系统要解决三个问题：

1. 长耗时操作不能阻塞 HTTP 请求
2. AI 优化、平台发布等操作需要状态追踪
3. 失败任务需要能记录错误，并为后续重试打基础

第一阶段只做 MVP：

- 支持创建任务
- 支持查询任务列表
- 支持查询任务详情
- 支持任务状态流转
- 支持后台 worker 执行任务
- 支持 AI 优化任务作为第一个真实任务类型

暂时不做：

- 分布式 worker
- 复杂优先级队列
- Cron 定时任务
- 工作流 DAG
- 多步骤任务编排
- Redis Stream / MQ 强依赖

这些可以在后续演进。

## 3. 任务类型

第一阶段建议定义两个任务类型：

```go
type TaskType string

const (
    TaskTypeArticleOptimize TaskType = "article_optimize"
    TaskTypeArticlePublish  TaskType = "article_publish"
)
```

MVP 先实现 `article_optimize`。

`article_publish` 可以先建类型和数据结构，但不急着接真实发布器。

## 4. 任务状态

建议把当前 `Task.Status string` 收敛成枚举：

```go
type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusProcessing TaskStatus = "processing"
    TaskStatusCompleted  TaskStatus = "completed"
    TaskStatusFailed     TaskStatus = "failed"
    TaskStatusCanceled   TaskStatus = "canceled"
)
```

状态流转：

```text
pending
    ↓
processing
    ↓
completed

pending
    ↓
processing
    ↓
failed

pending
    ↓
canceled
```

第一阶段可以只支持：

- `pending -> processing`
- `processing -> completed`
- `processing -> failed`

`canceled` 先预留。

## 5. 数据模型设计

当前 `Task` 模型太简单：

```go
type Task struct {
    ID        uint
    Type      string
    ArticleID uint
    Status    string
    Result    string
    Error     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

建议扩展为：

```go
type Task struct {
    ID        uint       `json:"id" gorm:"primaryKey"`
    Type      TaskType   `json:"type" gorm:"size:50;not null;index"`
    Status    TaskStatus `json:"status" gorm:"size:20;not null;default:pending;index"`
    ArticleID uint       `json:"article_id" gorm:"not null;index"`

    Payload string `json:"payload" gorm:"type:jsonb"`
    Result  string `json:"result" gorm:"type:jsonb"`
    Error   string `json:"error" gorm:"type:text"`

    RetryCount int `json:"retry_count" gorm:"default:0"`
    MaxRetries int `json:"max_retries" gorm:"default:3"`

    StartedAt  *time.Time `json:"started_at"`
    FinishedAt *time.Time `json:"finished_at"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`

    Article *Article `json:"article,omitempty" gorm:"foreignKey:ArticleID"`
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `type` | 任务类型，例如 `article_optimize` |
| `status` | 当前任务状态 |
| `article_id` | 任务关联的文章 |
| `payload` | 创建任务时的参数，JSON 格式 |
| `result` | 任务执行结果，JSON 格式 |
| `error` | 失败原因 |
| `retry_count` | 已重试次数 |
| `max_retries` | 最大重试次数 |
| `started_at` | 开始执行时间 |
| `finished_at` | 结束时间 |

为什么 `Payload` 和 `Result` 先用 `string`？

因为项目目前还比较小，直接存 JSON 字符串最简单。后续如果 JSON 查询变多，可以引入 `datatypes.JSON` 或自定义 JSON 类型。

## 6. 数据库迁移

新增 migration：

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

注意：

- 当前 `main.go` 里有 `AutoMigrate`
- migrations 里只有 `articles` 表
- 建议后续逐步把表结构变更都放进 migrations
- `AutoMigrate` 可以开发期保留，但正式环境不应依赖它

## 7. 后端分层设计

按当前项目风格，新增四类文件。

```text
internal/model/task.go
internal/repository/task.go
internal/service/task_service.go
internal/api/task.go
```

### 7.1 Repository

职责：

- 创建任务
- 查询任务详情
- 查询任务列表
- 获取下一个待处理任务
- 更新任务状态
- 写入任务结果
- 写入任务错误

建议接口：

```go
type TaskRepository struct {
    db *gorm.DB
}

func (r *TaskRepository) Create(task *model.Task) error
func (r *TaskRepository) GetByID(id uint) (*model.Task, error)
func (r *TaskRepository) List(offset, limit int) ([]model.Task, int64, error)
func (r *TaskRepository) ListByArticleID(articleID uint, offset, limit int) ([]model.Task, int64, error)
func (r *TaskRepository) FindNextPending() (*model.Task, error)
func (r *TaskRepository) MarkProcessing(id uint) error
func (r *TaskRepository) MarkCompleted(id uint, result string) error
func (r *TaskRepository) MarkFailed(id uint, errMsg string) error
```

`FindNextPending` 第一阶段可以简单查询：

```go
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT 1
```

如果以后多 worker 并发，再升级为事务 + 行锁：

```sql
FOR UPDATE SKIP LOCKED
```

### 7.2 Service

职责：

- 校验业务参数
- 创建具体类型任务
- 组织任务执行逻辑
- 根据任务类型分发到对应 handler

建议接口：

```go
type TaskService struct {
    taskRepo    *repository.TaskRepository
    articleRepo *repository.ArticleRepository
}

func (s *TaskService) CreateArticleOptimizeTask(articleID uint) (*model.Task, error)
func (s *TaskService) GetByID(id uint) (*model.Task, error)
func (s *TaskService) List(page, pageSize int) ([]model.Task, int64, error)
func (s *TaskService) ListByArticleID(articleID uint, page, pageSize int) ([]model.Task, int64, error)
func (s *TaskService) ExecuteNext() error
func (s *TaskService) Execute(task *model.Task) error
```

`CreateArticleOptimizeTask` 需要先确认文章存在：

```text
查询文章
    ↓
不存在则返回错误
    ↓
创建 pending 任务
    ↓
返回任务
```

### 7.3 API Handler

建议新增接口：

```text
GET  /api/tasks
GET  /api/tasks/:id
POST /api/articles/:id/optimize
```

第一阶段不建议开放通用 `POST /api/tasks`。

原因：

- 任务创建应该和具体业务动作绑定
- 文章优化任务需要校验文章是否存在
- 后续发布任务也需要校验平台、文章状态、账号配置等

所以 API 更适合这样设计：

```text
POST /api/articles/:id/optimize
```

语义是“对这篇文章发起 AI 优化”，不是“创建一个抽象任务”。

响应示例：

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 1,
    "type": "article_optimize",
    "status": "pending",
    "article_id": 12
  }
}
```

前端随后轮询：

```text
GET /api/tasks/1
```

### 7.4 Router

在 `SetupRoutes` 中新增：

```go
tasks := api.Group("/tasks")
{
    tasks.GET("", taskHandler.List)
    tasks.GET("/:id", taskHandler.GetByID)
}

articles.POST("/:id/optimize", articleHandler.Optimize)
```

注意：`articleHandler.Optimize` 需要依赖 `TaskService`。

可以有两种做法：

1. 给 `ArticleHandler` 注入 `TaskService`
2. 新建 `ArticleTaskHandler`

第一阶段建议直接给 `ArticleHandler` 注入 `TaskService`，简单直接。

## 8. Worker 设计

第一阶段建议先做进程内 worker，不急着接 Redis。

启动方式：

```text
main.go 启动 HTTP Server
    ↓
goroutine 启动 TaskWorker
    ↓
定时扫描 pending 任务
    ↓
执行任务
```

文件建议：

```text
internal/worker/task_worker.go
```

结构：

```go
type TaskWorker struct {
    taskService *service.TaskService
    interval    time.Duration
    stopCh      chan struct{}
}

func NewTaskWorker(taskService *service.TaskService, interval time.Duration) *TaskWorker
func (w *TaskWorker) Start()
func (w *TaskWorker) Stop()
```

执行逻辑：

```text
每隔 2 秒
    ↓
查找一个 pending 任务
    ↓
标记 processing
    ↓
按任务类型执行
    ↓
成功则 completed
    ↓
失败则 failed
```

为什么先不用 Redis？

因为你现在的核心问题是业务闭环，不是高并发队列。PostgreSQL 轮询足够支撑开发阶段，而且更容易调试。

后续如果需要多实例 worker，再引入 Redis Queue 或 PostgreSQL 行锁。

## 9. AI 优化任务执行设计

`article_optimize` 的执行流程：

```text
获取 task.article_id
    ↓
读取文章
    ↓
调用 AI Provider
    ↓
生成优化结果
    ↓
更新文章 summary/tags/content/ai_optimized/status
    ↓
写入 task.result
    ↓
task 状态改为 completed
```

第一阶段 AI 返回结构建议：

```json
{
  "title": "优化后的标题",
  "summary": "文章摘要",
  "tags": ["Go", "Gin", "AI"],
  "content": "优化后的 Markdown 内容"
}
```

落库时可以：

- `article.title` 可选更新
- `article.summary` 更新
- `article.tags` 存成逗号字符串
- `article.content` 更新为优化后内容
- `article.ai_optimized = true`
- `article.status = optimized`

如果你担心覆盖原文，后续可以新增文章版本表：

```text
article_versions
```

但 MVP 先不做，避免过早复杂化。

## 10. AI Provider 预留接口

为了后续接 DeepSeek、OpenAI、Qwen，建议先定义一个小接口：

```go
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

第一阶段可以先做一个 fake 实现：

```go
type FakeArticleOptimizer struct{}
```

它不调用真实 AI，只返回模拟结果。

好处：

- 先把任务系统跑通
- 不被 API key、网络、模型格式卡住
- 后续替换成 DeepSeek 实现时，只换 optimizer 实现

真实 AI 接入建议放在：

```text
internal/ai
```

例如：

```text
internal/ai/article_optimizer.go
internal/ai/deepseek_optimizer.go
internal/ai/fake_optimizer.go
```

## 11. 前端交互设计

文章详情页可以增加一个按钮：

```text
AI 优化
```

点击后：

```text
POST /api/articles/:id/optimize
    ↓
拿到 task_id
    ↓
每 2 秒 GET /api/tasks/:id
    ↓
pending / processing 显示进度中
    ↓
completed 后刷新文章详情
    ↓
failed 显示错误信息
```

任务列表页可以后续再做。

第一阶段前端只需要在文章详情页看到当前任务状态即可。

## 12. 错误处理

任务失败时不能只返回 HTTP 错误，因为任务是在后台执行的。

失败信息要写入：

```text
tasks.error
tasks.status = failed
tasks.finished_at
```

常见失败场景：

| 场景 | 处理 |
| --- | --- |
| 文章不存在 | 创建任务前直接返回错误 |
| AI API key 缺失 | 任务失败，写入 error |
| AI 请求超时 | 任务失败，写入 error |
| AI 返回格式不合法 | 任务失败，写入 error |
| 数据库更新失败 | 任务失败，写入 error |

第一阶段不要自动重试所有失败。

建议先记录 `retry_count/max_retries`，后续再加：

```text
POST /api/tasks/:id/retry
```

## 13. 幂等控制

同一篇文章如果重复点击 AI 优化，可能会创建多个任务。

第一阶段建议做一个简单限制：

```text
如果某篇文章已有 pending 或 processing 的 article_optimize 任务
则直接返回已有任务
```

这样可以避免前端重复点击造成多个 AI 请求。

Repository 可增加：

```go
func (r *TaskRepository) FindRunningTask(articleID uint, taskType model.TaskType) (*model.Task, error)
```

查询条件：

```sql
WHERE article_id = ?
  AND type = ?
  AND status IN ('pending', 'processing')
ORDER BY created_at DESC
LIMIT 1
```

## 14. 推荐落地顺序

### 第一步：补全任务基础 CRUD

- 扩展 `Task` 模型
- 新增 `tasks` migration
- 新增 `TaskRepository`
- 新增 `TaskService`
- 新增 `TaskHandler`
- 注册 `/api/tasks`

完成后，前端已有 `taskApi.getList/getDetail` 就能连上。

### 第二步：新增文章优化任务入口

- 新增 `POST /api/articles/:id/optimize`
- 创建 `article_optimize` 任务
- 同一文章运行中任务做幂等返回

### 第三步：新增进程内 worker

- 新增 `internal/worker/task_worker.go`
- 启动后定时扫描 pending 任务
- 支持任务状态流转

### 第四步：先接 fake AI optimizer

- 定义 `ArticleOptimizer` 接口
- 实现 `FakeArticleOptimizer`
- worker 执行后更新文章

### 第五步：接真实 DeepSeek/OpenAI-compatible API

- 增加 `DeepSeekArticleOptimizer`
- 从 `config` 读取 API key
- 替换 fake 实现

## 15. 目录结构建议

```text
internal/
  ai/
    article_optimizer.go
    fake_optimizer.go
    deepseek_optimizer.go
  api/
    article.go
    markdown.go
    router.go
    task.go
  model/
    article.go
    task.go
  repository/
    article.go
    task.go
  service/
    article_service.go
    markdown_service.go
    task_service.go
  worker/
    task_worker.go
```

## 16. 第一阶段完成标准

任务系统第一阶段完成后，应该能做到：

1. 创建文章
2. 点击 AI 优化
3. 后端创建 `article_optimize` 任务
4. worker 自动执行任务
5. 前端能查询任务状态
6. 任务完成后文章变成 `optimized`
7. 文章写入摘要、标签、优化内容
8. 任务失败时能看到失败原因

验证命令：

```bash
go test ./...
```

手动验证接口：

```bash
curl -X POST http://localhost:8080/api/articles/1/optimize

curl http://localhost:8080/api/tasks/1

curl http://localhost:8080/api/articles/1
```

## 17. 后续演进

当第一阶段跑通后，再考虑：

- Redis Queue
- 多 worker 并发
- 任务重试接口
- 任务取消接口
- 任务日志表
- 任务进度百分比
- SSE 推送任务状态
- 文章版本表
- 发布任务和发布记录表
- Workflow Pipeline

推荐演进路线：

```text
DB 轮询任务
    ↓
PostgreSQL 行锁防并发
    ↓
Redis Queue
    ↓
任务日志 / 重试 / 取消
    ↓
Workflow Pipeline
```

不要一开始就做 Workflow Pipeline。当前最重要的是先让一篇文章完成：

```text
创建 -> AI 优化 -> 任务完成 -> 结果落库
```
