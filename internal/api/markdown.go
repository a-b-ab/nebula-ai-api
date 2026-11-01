package api

import (
	"nebula-ai-api/internal/service"
	"nebula-ai-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type MarkdownHandler struct {
	markdownService *service.MarkdownService
}

func NewMarkdownHandler(markdownService *service.MarkdownService) *MarkdownHandler {
	return &MarkdownHandler{
		markdownService: markdownService,
	}
}

type ConvertCSDNRequest struct {
	Title          string `json:"title"`
	Markdown       string `json:"markdown" binding:"required"`
	IncludeWrapper *bool  `json:"include_wrapper"`
}

func (h *MarkdownHandler) ConvertToCSDN(c *gin.Context) {
	var req ConvertCSDNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	includeWrapper := true
	if req.IncludeWrapper != nil {
		includeWrapper = *req.IncludeWrapper
	}

	result, err := h.markdownService.ConvertToCSDN(&service.ConvertToCSDNInput{
		Title:          req.Title,
		Markdown:       req.Markdown,
		IncludeWrapper: includeWrapper,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}
