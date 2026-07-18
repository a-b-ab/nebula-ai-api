package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nebula-ai-api/internal/ai"
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/repository"
)

type TaskService struct {
	taskRepo    *repository.TaskRespository
	articleRepo *repository.ArticleRepository
	optimizer   ai.ArticleOptimizer
}

func NewTaskService(taskRepo *repository.TaskRespository, articleRepo *repository.ArticleRepository, optimizer ai.ArticleOptimizer) *TaskService {
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
		Payload:    &payload,
		MaxRetries: 3,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}
	return task, err
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
		err = fmt.Errorf("不支持的任务类型：%s", task.Type)
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
