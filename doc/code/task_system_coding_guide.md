# 任务系统编码指南

本文档是 `task_system_design.md` 的编码落地指南。

目标不是直接给出完整代码，而是告诉你：

- 先写哪个文件
- 每个文件负责什么
- 每个文件大概要有哪些结构和方法
- 写完后如何验证

建议你按顺序自己敲，不要一口气写完整个系统。

## 1. 最终目标

第一阶段任务系统只做一个最小闭环：

```text
创建文章
    ↓
创建 AI 优化任务
    ↓
worker 执行任务
    ↓
更新任务状态
    ↓
更新文章内容
    ↓
前端或接口查询任务结果
```

先不要做 Redis Queue、复杂 Workflow、多 worker、任务取消、任务重试。

第一阶段只要让 `article_optimize` 任务跑通。

## 2. 推荐编码顺序

按这个顺序写：

```text
1. internal/model/task.go
2. migrations/000002_create_tasks_table.up.sql
3. migrations/000002_create_tasks_table.down.sql
4. internal/repository/task.go
5. internal/ai/article_optimizer.go
6. internal/ai/fake_optimizer.go
7. internal/service/task_service.go
8. internal/api/task.go
9. internal/api/article.go
10. internal/api/router.go
11. internal/worker/task_worker.go
12. cmd/main.go
```

这个顺序的好处是：

- 先定义数据
- 再定义数据库
- 再写底层查询
- 再写业务逻辑
- 再暴露 API
- 最后装配 worker

## 3. 修改 `internal/model/task.go`

这个文件只负责定义任务模型。

你要写三部分：

### 3.1 定义任务类型

需要定义：

```go
type TaskType string
```

然后定义任务类型常量：

```go
TaskTypeArticleOptimize
TaskTypeArticlePublish
```

第一阶段真正用到的是：

```text
article_optimize
```

`article_publish` 只是先预留。

### 3.2 定义任务状态

需要定义：

```go
type TaskStatus string
```

然后定义状态常量：

```text
pending
processing
completed
failed
canceled
```

第一阶段重点理解这几个状态：

```text
pending -> processing -> completed
pending -> processing -> failed
```

### 3.3 扩展 Task 结构体

当前 `Task` 已经有：

```text
ID
Type
ArticleID
Status
Result
Error
CreatedAt
UpdatedAt
Article
```

你要把它扩展成更完整的任务模型。

建议增加：

```text
Payload
RetryCount
MaxRetries
StartedAt
FinishedAt
```

字段职责：

| 字段 | 作用 |
| --- | --- |
| `Type` | 任务类型 |
| `Status` | 任务状态 |
| `ArticleID` | 关联文章 |
| `Payload` | 创建任务时的参数 |
| `Result` | 任务执行结果 |
| `Error` | 失败原因 |
| `RetryCount` | 已重试次数 |
| `MaxRetries` | 最大重试次数 |
| `StartedAt` | 开始执行时间 |
| `FinishedAt` | 结束时间 |

注意：

- `Payload` 和 `Result` 第一阶段可以用 `string`
- GORM tag 可以设置为 `type:jsonb`
- `StartedAt` 和 `FinishedAt` 用 `*time.Time`

## 4. 新增 migrations

### 4.1 新增 `migrations/000002_create_tasks_table.up.sql`

这个文件负责创建 `tasks` 表。

你需要写这些字段：

```text
id
type
status
article_id
payload
result
error
retry_count
max_retries
started_at
finished_at
created_at
updated_at
```

还要加外键：

```text
tasks.article_id -> articles.id
```

建议设置：

```text
ON DELETE CASCADE
```

意思是文章删除后，关联任务也一起删除。

### 4.2 新增索引

建议加：

```text
idx_tasks_type
idx_tasks_status
idx_tasks_article_id
idx_tasks_created_at
```

原因：

- worker 会按 `status` 查 pending 任务
- 文章详情可能按 `article_id` 查任务
- 后台列表会按 `created_at` 排序

### 4.3 新增 `migrations/000002_create_tasks_table.down.sql`

这个文件负责回滚。

只需要删除 `tasks` 表。

注意：

- down 文件要和 up 文件成对出现
- 不要在 down 文件里删除 `articles`

## 5. 新增 `internal/repository/task.go`

Repository 层只写数据库操作，不写复杂业务判断。

你需要定义：

```go
type TaskRepository struct {
    db *gorm.DB
}
```

然后写构造函数：

```go
func NewTaskRepository(db *gorm.DB) *TaskRepository
```

建议实现这些方法：

```text
Create
GetByID
List
ListByArticleID
FindRunningTask
FindNextPending
MarkProcessing
MarkCompleted
MarkFailed
```

### 5.1 `Create`

职责：

```text
插入一条任务记录
```

不要在这里判断文章是否存在。

文章是否存在属于 Service 层逻辑。

### 5.2 `GetByID`

职责：

```text
根据任务 ID 查询任务
```

可以考虑 `Preload("Article")`，但第一阶段不是必须。

### 5.3 `List`

职责：

```text
分页查询任务列表
```

排序建议：

```text
created_at DESC
```

### 5.4 `ListByArticleID`

职责：

```text
查询某篇文章下面的任务
```

前端文章详情页后续会用到。

### 5.5 `FindRunningTask`

职责：

```text
查某篇文章是否已有 pending 或 processing 的同类型任务
```

它用于防止重复点击 AI 优化按钮后创建多个任务。

查询条件大概是：

```text
article_id = ?
type = ?
status IN ('pending', 'processing')
```

### 5.6 `FindNextPending`

职责：

```text
给 worker 找一个待执行任务
```

第一阶段可以简单查：

```text
status = pending
ORDER BY created_at ASC
LIMIT 1
```

以后如果有多个 worker，再考虑事务和行锁。

### 5.7 `MarkProcessing`

职责：

```text
把任务状态改成 processing
设置 started_at
清空 error
```

### 5.8 `MarkCompleted`

职责：

```text
把任务状态改成 completed
写入 result
设置 finished_at
```

### 5.9 `MarkFailed`

职责：

```text
把任务状态改成 failed
写入 error
设置 finished_at
```

## 6. 新增 `internal/ai/article_optimizer.go`

这个文件只定义 AI 优化能力的接口。

你要定义一个返回结果结构：

```text
ArticleOptimizeResult
```

字段建议：

```text
Title
Summary
Tags
Content
```

然后定义接口：

```text
ArticleOptimizer
```

它应该有一个方法：

```text
OptimizeArticle
```

这个接口的意义是：

- `TaskService` 不关心底层用 DeepSeek 还是 OpenAI
- 测试和开发阶段可以先用 fake 实现
- 后续真实模型接入时不用大改任务系统

## 7. 新增 `internal/ai/fake_optimizer.go`

这个文件写假的 AI 优化器。

第一阶段不要急着接真实 DeepSeek。

Fake 实现的作用是：

```text
不请求外部模型
直接返回模拟优化结果
先验证任务系统是否跑通
```

你要写：

```text
FakeArticleOptimizer
```

它实现 `ArticleOptimizer` 接口。

模拟返回可以包含：

```text
原文章标题
模拟摘要
模拟标签
原内容 + 一段模拟优化说明
```

当 Fake AI 跑通后，再接真实 AI。

## 8. 新增 `internal/service/task_service.go`

Service 层负责业务逻辑。

你需要定义：

```go
type TaskService struct {
    taskRepo    *repository.TaskRepository
    articleRepo *repository.ArticleRepository
    optimizer   ai.ArticleOptimizer
}
```

然后写构造函数：

```text
NewTaskService
```

建议实现这些方法：

```text
CreateArticleOptimizeTask
GetByID
List
ListByArticleID
ExecuteNext
Execute
```

### 8.1 `CreateArticleOptimizeTask`

职责：

```text
为某篇文章创建 AI 优化任务
```

逻辑顺序：

```text
1. 根据 articleID 查询文章
2. 如果文章不存在，返回错误
3. 查询是否已有运行中的 article_optimize 任务
4. 如果有，直接返回已有任务
5. 如果没有，创建 pending 任务
6. 返回任务
```

注意：

- 这个方法只创建任务
- 不要在这里直接调用 AI
- AI 调用交给 worker 执行

### 8.2 `GetByID`

职责：

```text
查询任务详情
```

直接调用 repository。

### 8.3 `List`

职责：

```text
分页查询任务列表
```

分页逻辑和文章列表类似：

```text
offset = (page - 1) * pageSize
```

### 8.4 `ListByArticleID`

职责：

```text
分页查询某篇文章的任务列表
```

### 8.5 `ExecuteNext`

职责：

```text
找一个 pending 任务并执行
```

逻辑：

```text
1. 调 taskRepo.FindNextPending
2. 如果没有任务，直接返回 nil
3. 如果有任务，调用 Execute
```

注意：

- 没有任务不是错误
- worker 会频繁调用这个方法

### 8.6 `Execute`

职责：

```text
执行具体任务
```

逻辑：

```text
1. 标记任务为 processing
2. 根据 task.Type 分发
3. 如果是 article_optimize，执行文章优化
4. 成功则 MarkCompleted
5. 失败则 MarkFailed
```

你可以把文章优化逻辑拆成私有方法：

```text
executeArticleOptimize
```

这样 `Execute` 会更清晰。

### 8.7 `executeArticleOptimize`

职责：

```text
执行 AI 优化文章任务
```

逻辑：

```text
1. 根据 task.ArticleID 查询文章
2. 调用 optimizer.OptimizeArticle
3. 把返回结果写回 Article
4. 设置 article.AIOptimized = true
5. 设置 article.Status = StatusOptimized
6. 更新文章
7. 把优化结果序列化为 JSON 字符串
8. 返回 result 字符串
```

Tags 第一阶段可以继续存成逗号字符串。

## 9. 新增 `internal/api/task.go`

这个文件写任务查询接口。

你需要定义：

```go
type TaskHandler struct {
    taskService *service.TaskService
}
```

然后写构造函数：

```text
NewTaskHandler
```

需要实现两个接口：

```text
List
GetByID
```

### 9.1 `List`

对应接口：

```text
GET /api/tasks
```

支持：

```text
GET /api/tasks?page=1&page_size=10
GET /api/tasks?article_id=1
```

逻辑：

```text
1. 解析分页参数
2. 如果有 article_id，调用 ListByArticleID
3. 否则调用 List
4. 返回分页结果
```

可以复用已有：

```text
pkg/query.ListQuery
pkg/response
```

### 9.2 `GetByID`

对应接口：

```text
GET /api/tasks/:id
```

逻辑：

```text
1. 解析 id
2. 调 taskService.GetByID
3. 不存在返回 404
4. 存在返回任务
```

## 10. 修改 `internal/api/article.go`

这个文件已经有文章 CRUD。

你要给文章增加“创建优化任务”的能力。

### 10.1 修改 ArticleHandler 结构体

当前大概是：

```go
type ArticleHandler struct {
    articleService *service.ArticleService
}
```

你要增加：

```text
taskService
```

因为 `Optimize` 方法需要调用任务服务。

### 10.2 修改构造函数

当前构造函数只接收 `articleService`。

你要让它也接收 `taskService`。

注意：

- 改了构造函数后，`router.go` 也要跟着改

### 10.3 新增 `Optimize`

对应接口：

```text
POST /api/articles/:id/optimize
```

逻辑：

```text
1. 解析文章 id
2. 调 taskService.CreateArticleOptimizeTask
3. 返回 task
```

注意：

- 这里只创建任务
- 不直接调用 AI
- 不等待任务完成

## 11. 修改 `internal/api/router.go`

这个文件负责注册路由。

### 11.1 修改 `SetupRoutes` 参数

当前 `SetupRoutes` 接收：

```text
articleService
markdownService
```

你要增加：

```text
taskService
```

### 11.2 初始化 TaskHandler

新增：

```text
taskHandler := NewTaskHandler(taskService)
```

### 11.3 注册 tasks 路由

新增：

```text
GET /api/tasks
GET /api/tasks/:id
```

### 11.4 注册文章优化路由

在 articles group 里新增：

```text
POST /api/articles/:id/optimize
```

## 12. 新增 `internal/worker/task_worker.go`

这个文件写后台 worker。

第一版 worker 保持简单：

```text
每隔 2 秒调用一次 taskService.ExecuteNext
```

你需要定义：

```go
type TaskWorker struct {
    taskService *service.TaskService
    interval    time.Duration
}
```

建议实现：

```text
NewTaskWorker
Start
```

### 12.1 `Start`

逻辑：

```text
1. 创建 ticker
2. 循环监听 ticker
3. 每次触发调用 taskService.ExecuteNext
4. 如果执行出错，打印日志
5. 支持 context 退出
```

注意：

- 没有 pending 任务时，不应该打印错误
- 只有真正执行失败时才打印日志

## 13. 修改 `cmd/main.go`

最后在入口文件里装配依赖。

### 13.1 初始化 Repository

已有：

```text
articleRepo
```

新增：

```text
taskRepo
```

### 13.2 初始化 AI Optimizer

第一阶段使用：

```text
FakeArticleOptimizer
```

暂时不要接 DeepSeek。

### 13.3 初始化 Service

已有：

```text
articleService
markdownService
```

新增：

```text
taskService
```

### 13.4 修改路由注册

把 `taskService` 传给 `SetupRoutes`。

### 13.5 启动 worker

在启动 HTTP 服务前，启动 worker goroutine。

逻辑：

```text
创建 TaskWorker
go taskWorker.Start(...)
```

## 14. 验证顺序

不要等全部写完才测。

建议每完成一阶段就跑：

```bash
go test ./...
```

### 14.1 模型和 migration 写完后

检查：

```text
go test ./...
```

确保编译没问题。

### 14.2 Repository 写完后

检查：

```text
go test ./...
```

如果你愿意，可以给 repository 写集成测试，但第一阶段不强制。

### 14.3 Service 写完后

建议至少写一个单元测试：

```text
CreateArticleOptimizeTask
```

验证：

```text
文章不存在时返回错误
有运行中任务时返回已有任务
没有运行中任务时创建 pending 任务
```

### 14.4 API 和 worker 写完后

启动服务，手动验证：

```bash
curl -X POST http://localhost:8080/api/articles/1/optimize
curl http://localhost:8080/api/tasks/1
curl http://localhost:8080/api/articles/1
```

最终看到：

```text
task.status = completed
article.ai_optimized = true
article.status = 1
```

就说明第一阶段跑通。

## 15. 常见错误

### 15.1 在 API 里直接调用 AI

不要这样做。

API 只负责创建任务并立即返回。

AI 调用应该在 worker 里执行。

### 15.2 任务重复创建

如果用户连续点两次 AI 优化，可能创建两个任务。

所以 `CreateArticleOptimizeTask` 里要先查运行中任务。

### 15.3 Repository 写业务逻辑

Repository 只负责数据库。

比如“文章不存在能不能创建任务”这种判断应该放在 Service。

### 15.4 一开始就接真实 AI

不要一开始就接 DeepSeek。

先用 Fake AI 跑通任务流。

真实 AI 有网络、API key、返回格式、超时等问题，会干扰你理解任务系统本身。

### 15.5 一开始就上 Redis

当前项目阶段不需要。

先用 PostgreSQL 任务表和进程内 worker。

后面需要多 worker 或高并发时再升级。

## 16. 你应该重点理解的核心

任务系统不是简单多一张表。

它的本质是把长耗时业务拆成两段：

```text
HTTP 请求阶段：
    只创建任务，快速返回

后台执行阶段：
    慢慢执行任务，更新状态和结果
```

你写代码时要一直保持这个边界：

```text
Handler：解析请求，返回响应
Service：业务规则和状态流转
Repository：数据库读写
Worker：循环调度任务
AI Optimizer：具体 AI 能力
```

只要这个边界清楚，后面接 DeepSeek、Redis、发布系统都会更容易。
