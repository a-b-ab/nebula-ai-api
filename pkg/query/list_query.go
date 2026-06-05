package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListQuery 列表查询参数
type ListQuery struct {
	Page      int    `json:"page"`       // 当前页码
	PageSize  int    `json:"page_size"`  // 每页数量
	Offset    int    `json:"-"`          // 计算出的 offset（不序列化到 JSON）
	Status    *int   `json:"status"`     // 状态筛选（可选）
	Keyword   string `json:"keyword"`    // 关键词搜索（可选）
	SortBy    string `json:"sort_by"`    // 排序字段（可选）
	SortOrder string `json:"sort_order"` // 排序方向（可选）
}

// DefaultListQuery 默认列表查询参数
var DefaultListQuery = ListQuery{
	Page:     1,
	PageSize: 10,
}

// NewListQuery 从 Gin Context 创建 ListQuery
func NewListQuery(c *gin.Context) *ListQuery {
	q := &ListQuery{}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 解析状态筛选
	statusStr := c.Query("status")
	if statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		q.Status = &status
	}

	// 解析其他参数
	q.Keyword = c.Query("keyword")
	q.SortBy = c.DefaultQuery("sort_by", "created_at")
	q.SortOrder = c.DefaultQuery("sort_order", "DESC")

	// 设置分页参数并验证
	q.SetPage(page)
	q.SetPageSize(pageSize)

	// 计算 offset
	q.CalculateOffset()

	return q
}

// SetPage 设置页码并验证
func (q *ListQuery) SetPage(page int) {
	if page < 1 {
		q.Page = 1
	} else {
		q.Page = page
	}
}

// SetPageSize 设置每页数量并验证
func (q *ListQuery) SetPageSize(pageSize int) {
	if pageSize < 1 {
		q.PageSize = 10
	} else if pageSize > 100 {
		q.PageSize = 100
	} else {
		q.PageSize = pageSize
	}
}

// CalculateOffset 计算 offset
func (q *ListQuery) CalculateOffset() {
	q.Offset = (q.Page - 1) * q.PageSize
}

// GetLimit 获取 limit（用于 SQL LIMIT）
func (q *ListQuery) GetLimit() int {
	return q.PageSize
}

// GetOffset 获取 offset（用于 SQL OFFSET）
func (q *ListQuery) GetOffset() int {
	return q.Offset
}

// HasStatus 检查是否有状态筛选
func (q *ListQuery) HasStatus() bool {
	return q.Status != nil
}

// GetStatus 获取状态值
func (q *ListQuery) GetStatus() int {
	if q.Status != nil {
		return *q.Status
	}
	return -1
}

// HasKeyword 检查是否有关键词搜索
func (q *ListQuery) HasKeyword() bool {
	return q.Keyword != ""
}

// GetOrderBy 获取排序字符串（用于 SQL ORDER BY）
func (q *ListQuery) GetOrderBy() string {
	if q.SortBy == "" {
		return "created_at DESC"
	}
	return q.SortBy + " " + q.SortOrder
}

// ListResponse 列表响应结构
type ListResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// NewListResponse 创建列表响应
func NewListResponse(data interface{}, total int64, query *ListQuery) *ListResponse {
	return &ListResponse{
		Data:     data,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}
}

// GetTotalPages 获取总页数
func (r *ListResponse) GetTotalPages() int {
	if r.PageSize == 0 {
		return 0
	}
	return int(int(r.Total) + r.PageSize - 1) / r.PageSize
}

// HasNext 是否有下一页
func (r *ListResponse) HasNext() bool {
	return r.Page < r.GetTotalPages()
}

// HasPrev 是否有上一页
func (r *ListResponse) HasPrev() bool {
	return r.Page > 1
}
