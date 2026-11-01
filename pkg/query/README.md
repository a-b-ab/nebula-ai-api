# 列表查询组件使用指南

## 功能

封装通用的列表查询参数解析、验证和响应格式化。

---

## 快速使用

### 1. 基础用法

```go
import "nebula-ai-api/pkg/query"

func (h *Handler) List(c *gin.Context) {
    // 解析并验证查询参数
    listQuery := query.NewListQuery(c)

    // 使用查询参数
    data, total, err := h.service.List(listQuery.Page, listQuery.PageSize)

    // 构造响应
    listResp := query.NewListResponse(data, total, listQuery)
    response.Success(c, listResp)
}
```

### 2. 支持的查询参数

| 参数 | 默认值 | 说明 | 示例 |
|------|--------|------|------|
| `page` | 1 | 当前页码 | `?page=2` |
| `page_size` | 10 | 每页数量（最大 100） | `?page_size=20` |
| `status` | - | 状态筛选（可选） | `?status=1` |
| `keyword` | - | 关键词搜索（可选） | `?keyword=Go` |
| `sort_by` | created_at | 排序字段 | `?sort_by=title` |
| `sort_order` | DESC | 排序方向（ASC/DESC） | `?sort_order=ASC` |

---

## 完整示例

### API Handler

```go
package api

import (
    "nebula-ai-api/internal/model"
    "nebula-ai-api/internal/service"
    "nebula-ai-api/pkg/query"
    "nebula-ai-api/pkg/response"
    "github.com/gin-gonic/gin"
)

func (h *ArticleHandler) List(c *gin.Context) {
    // 1. 解析查询参数（自动验证）
    listQuery := query.NewListQuery(c)

    var articles []model.Article
    var total int64
    var err error

    // 2. 根据条件查询
    if listQuery.HasStatus() {
        // 按状态筛选
        articles, total, err = h.service.ListByStatus(
            model.ArticleStatus(listQuery.GetStatus()),
            listQuery.Page,
            listQuery.PageSize,
        )
    } else {
        // 普通列表
        articles, total, err = h.service.List(
            listQuery.Page,
            listQuery.PageSize,
        )
    }

    if err != nil {
        response.InternalError(c, err.Error())
        return
    }

    // 3. 构造响应
    listResp := query.NewListResponse(articles, total, listQuery)
    response.Success(c, listResp)
}
```

### 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data": [...],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 高级用法

### 1. 关键词搜索

```go
func (h *Handler) List(c *gin.Context) {
    listQuery := query.NewListQuery(c)

    if listQuery.HasKeyword() {
        // 使用关键词搜索
        data, total := h.service.Search(listQuery.Keyword, listQuery.Page, listQuery.PageSize)
    } else {
        // 普通查询
        data, total := h.service.List(listQuery.Page, listQuery.PageSize)
    }
}
```

### 2. 自定义排序

```go
func (h *Handler) List(c *gin.Context) {
    listQuery := query.NewListQuery(c)

    // 获取排序字符串
    orderBy := listQuery.GetOrderBy() // "created_at DESC" 或 "title ASC"

    data, total := h.service.ListWithOrder(orderBy, listQuery.Page, listQuery.PageSize)
}
```

### 3. 响应扩展信息

```go
listResp := query.NewListResponse(articles, total, listQuery)

// 添加额外信息
response.Success(c, gin.H{
    "data": listResp.Data,
    "total": listResp.Total,
    "page": listResp.Page,
    "page_size": listResp.PageSize,
    "total_pages": listResp.GetTotalPages(),  // 总页数
    "has_next": listResp.HasNext(),           // 是否有下一页
    "has_prev": listResp.HasPrev(),           // 是否有上一页
})
```

---

## 参数验证规则

### Page（页码）
- 最小值：1
- 如果 `< 1`，自动设为 `1`

### PageSize（每页数量）
- 最小值：1
- 最大值：100
- 如果 `< 1`，自动设为 `10`
- 如果 `> 100`，自动设为 `100`

### Status（状态）
- 可选参数
- 不传时为 `nil`
- 需要通过 `HasStatus()` 检查是否存在

---

## API 使用示例

### 基础查询
```bash
GET /api/articles
```

### 分页查询
```bash
GET /api/articles?page=2&page_size=20
```

### 状态筛选
```bash
GET /api/articles?status=1
```

### 关键词搜索
```bash
GET /api/articles?keyword=Go
```

### 排序
```bash
GET /api/articles?sort_by=title&sort_order=ASC
```

### 组合查询
```bash
GET /api/articles?page=1&page_size=10&status=1&keyword=AI&sort_by=created_at&sort_order=DESC
```

---

## 扩展自定义字段

如果需要更多查询字段，可以扩展 `ListQuery`：

```go
type ListQuery struct {
    Page      int
    PageSize  int
    Offset    int
    Status    *int
    Keyword   string
    SortBy    string
    SortOrder string

    // 自定义字段
    Category  string   // 分类筛选
    Tags      []string  // 标签筛选
    DateFrom  string   // 日期范围-开始
    DateTo    string   // 日期范围-结束
}

func NewListQuery(c *gin.Context) *ListQuery {
    q := &ListQuery{}
    // ... 原有解析逻辑

    // 解析自定义字段
    q.Category = c.Query("category")
    q.Tags = c.QueryArray("tags")
    q.DateFrom = c.Query("date_from")
    q.DateTo = c.Query("date_to")

    return q
}
```

---

## Repository 层使用

```go
func (r *ArticleRepository) List(offset, limit int) ([]model.Article, int64, error) {
    var articles []model.Article
    var total int64

    if err := r.db.Model(&model.Article{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error
    return articles, total, err
}

// Service 层调用
func (s *ArticleService) List(page, pageSize int) ([]model.Article, int64, error) {
    offset := (page - 1) * pageSize
    return s.repo.List(offset, pageSize)
}
```

---

## 常见问题

### Q: 为什么 page_size 最大限制为 100？
A: 防止一次查询过多数据导致性能问题。

### Q: 如何实现"无限滚动"？
A: 前端可以根据 `has_next` 判断是否还有更多数据，自动加载下一页。

### Q: 如何支持多个状态筛选？
A: 修改 `Status` 字段类型为 `[]int`，并在 Repository 层处理 `IN` 查询。

### Q: 如何实现游标分页？
A: 创建新的 `CursorQuery` 组件，使用 `WHERE id > last_id LIMIT 10` 方式。

---

## 最佳实践

1. ✅ **始终使用 `ListQuery` 组件**：保持接口一致性
2. ✅ **参数验证在组件内完成**：Handler 不需要额外验证
3. ✅ **使用 `HasStatus()` 等方法检查**：避免直接判断 nil
4. ✅ **响应格式统一**：使用 `ListResponse` 封装
5. ✅ **默认值合理**：page=1, page_size=10
6. ✅ **限制最大值**：防止恶意请求大页码
