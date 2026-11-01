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
		}
		tasks, total, err = h.taskService.ListByArticleID(uint(articleID), listQuery.Page, listQuery.PageSize)
	} else {
		tasks, total, err = h.taskService.List(listQuery.Page, listQuery.PageSize)
	}
	if err != nil {
		response.InternalError(c, err.Error())
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
