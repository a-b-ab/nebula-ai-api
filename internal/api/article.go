package api

import (
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/service"
	"nebula-ai-api/pkg/query"
	"nebula-ai-api/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	articleService *service.ArticleService
	taskService    *service.TaskService
}

func NewArticleHandler(articleService *service.ArticleService, taskService *service.TaskService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
		taskService:    taskService,
	}
}

type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Summary string `json:"summary"`
	Tags    string `json:"tags"`
}

type UpdateArticleRequest struct {
	Title           *string              `json:"title"`
	Content         *string              `json:"content"`
	Summary         *string              `json:"summary"`
	Tags            *string              `json:"tags"`
	PublishPlatform *string              `json:"publish_platform"`
	Status          *model.ArticleStatus `json:"status"`
}

type UpdateStatusRequest struct {
	Status model.ArticleStatus `json:"status" binding:"required,min=0,max=2"`
}

// Create 创建文章
func (h *ArticleHandler) Create(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	input := &service.CreateArticleInput{
		Title:   req.Title,
		Content: req.Content,
		Summary: req.Summary,
		Tags:    req.Tags,
	}

	article, err := h.articleService.Create(input)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, article)
}

// GetByID 获取文章详情
func (h *ArticleHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	article, err := h.articleService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "文章不存在")
		return
	}

	response.Success(c, article)
}

// List 获取文章列表
func (h *ArticleHandler) List(c *gin.Context) {
	// 使用封装的 ListQuery 组件
	listQuery := query.NewListQuery(c)

	var articles []model.Article
	var total int64
	var err error

	// 如果传了 status 参数，按状态筛选
	if listQuery.HasStatus() {
		articles, total, err = h.articleService.ListByStatus(
			model.ArticleStatus(listQuery.GetStatus()),
			listQuery.Page,
			listQuery.PageSize,
		)
	} else {
		articles, total, err = h.articleService.List(listQuery.Page, listQuery.PageSize)
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 使用封装的 ListResponse
	listResp := query.NewListResponse(articles, total, listQuery)
	response.Success(c, listResp)
}

// Update 更新文章
func (h *ArticleHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	input := &service.UpdateArticleInput{
		Title:           req.Title,
		Content:         req.Content,
		Summary:         req.Summary,
		Tags:            req.Tags,
		PublishPlatform: req.PublishPlatform,
		Status:          req.Status,
	}

	article, err := h.articleService.Update(uint(id), input)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, article)
}

// Delete 删除文章
func (h *ArticleHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	if err := h.articleService.Delete(uint(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateStatus 更新文章状态
func (h *ArticleHandler) UpdateStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.articleService.UpdateStatus(uint(id), req.Status); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "状态更新成功",
	})
}

// AddPublishPlatform 添加发布平台
func (h *ArticleHandler) AddPublishPlatform(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	type AddPlatformRequest struct {
		Platform string `json:"platform" binding:"required"`
	}

	var req AddPlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.articleService.AddPublishPlatform(uint(id), req.Platform); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "平台添加成功",
	})
}

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
