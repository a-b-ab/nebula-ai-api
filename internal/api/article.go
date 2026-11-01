package api

import (
	"encoding/json"
	"errors"
	"io"
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/service"
	"nebula-ai-api/pkg/query"
	"nebula-ai-api/pkg/response"
	"path/filepath"
	"strconv"
	"strings"

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

// 文章格式(导出导入)
type articleMetadata struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Tags    string `json:"tags"`
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

// Markdown 解析函数
func parseArticleMarkdown(content []byte) (articleMetadata, string, error) {
	const delimiter = "---"
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	if !strings.HasPrefix(text, delimiter+"\n") {
		return articleMetadata{}, "", errors.New("Markdown 文件缺少 Front Matter")
	}

	rest := strings.TrimPrefix(text, delimiter+"\n")
	end := strings.Index(rest, "\n"+delimiter+"\n")
	if end < 0 {
		return articleMetadata{}, "", errors.New("Markdown Front Matter 格式错误")
	}

	var metadata articleMetadata
	if err := json.Unmarshal([]byte(rest[:end]), &metadata); err != nil {
		return articleMetadata{}, "", errors.New("Markdown JSON 元数据解析失败: " + err.Error())
	}

	body := strings.TrimPrefix(rest[end+len("\n"+delimiter+"\n"):], "\n")
	return metadata, body, nil
}

// 导出文章
func (h *ArticleHandler) Export(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	article, err := h.articleService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "文章不存在")
		return
	}

	metadata, err := json.Marshal(articleMetadata{
		Title: article.Title, Summary: article.Summary, Tags: article.Tags,
	})
	if err != nil {
		response.InternalError(c, "生成 Mardown 失败")
		return
	}

	markdown := "---\n" + string(metadata) + "\n--\n\n" + article.Content
	response.Success(c, markdown)
}

// 导入文章
func (h *ArticleHandler) Import(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择 Markdown 文件")
		return
	}
	if !strings.EqualFold(filepath.Ext(header.Filename), ".md") {
		response.BadRequest(c, "只支持 .md 文件")
		return
	}
	if header.Size > 2*1024*1024 {
		response.BadRequest(c, "文件不能超过 2MB")
		return
	}

	file, err := header.Open()
	if err != nil {
		response.BadRequest(c, "无法读取导入文件")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		response.BadRequest(c, "无法读取导入文件")
		return
	}
	metadata, body, err := parseArticleMarkdown(content)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if strings.TrimSpace(metadata.Title) == "" {
		response.BadRequest(c, "文章标题不能为空")
		return
	}
	if strings.TrimSpace(body) == "" {
		response.BadRequest(c, "文章正文不能为空")
		return
	}

	article, err := h.articleService.Create(&service.CreateArticleInput{
		Title: metadata.Title, Content: body, Summary: metadata.Summary, Tags: metadata.Tags,
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, article)
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
