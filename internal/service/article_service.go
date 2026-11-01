package service

import (
	"errors"
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/repository"
)

type ArticleService struct {
	articleRepo *repository.ArticleRepository
}

func NewArticleService(articleRepo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
	}
}

// CreateArticleInput 创建文章输入
type CreateArticleInput struct {
	Title   string
	Content string
	Summary string
	Tags    string
}

// UpdateArticleInput 更新文章输入
type UpdateArticleInput struct {
	Title           *string
	Content         *string
	Summary         *string
	Tags            *string
	PublishPlatform *string
	Status          *model.ArticleStatus
}

// Create 创建文章
func (s *ArticleService) Create(input *CreateArticleInput) (*model.Article, error) {
	article := &model.Article{
		Title:           input.Title,
		Content:         input.Content,
		Summary:         input.Summary,
		Tags:            input.Tags,
		Status:          model.StatusDraft,
		AIOptimized:     false,
		PublishPlatform: "",
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, err
	}

	return article, nil
}

// GetByID 根据ID获取文章
func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	return s.articleRepo.GetByID(id)
}

// List 获取文章列表（分页）
func (s *ArticleService) List(page, pageSize int) ([]model.Article, int64, error) {
	offset := (page - 1) * pageSize
	return s.articleRepo.List(offset, pageSize)
}

// ListByStatus 根据状态获取文章列表
func (s *ArticleService) ListByStatus(status model.ArticleStatus, page, pageSize int) ([]model.Article, int64, error) {
	offset := (page - 1) * pageSize
	return s.articleRepo.ListByStatus(status, offset, pageSize)
}

// Update 更新文章
func (s *ArticleService) Update(id uint, input *UpdateArticleInput) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("文章不存在")
	}

	// 只更新提供的字段
	if input.Title != nil {
		article.Title = *input.Title
	}
	if input.Content != nil {
		article.Content = *input.Content
	}
	if input.Summary != nil {
		article.Summary = *input.Summary
	}
	if input.Tags != nil {
		article.Tags = *input.Tags
	}
	if input.PublishPlatform != nil {
		article.PublishPlatform = *input.PublishPlatform
	}
	if input.Status != nil {
		article.Status = *input.Status
	}

	if err := s.articleRepo.Update(article); err != nil {
		return nil, err
	}

	return article, nil
}

// Delete 删除文章
func (s *ArticleService) Delete(id uint) error {
	_, err := s.articleRepo.GetByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	return s.articleRepo.Delete(id)
}

// UpdateStatus 更新文章状态
func (s *ArticleService) UpdateStatus(id uint, status model.ArticleStatus) error {
	return s.articleRepo.UpdateStatus(id, status)
}

// UpdateAIOptimized 更新AI优化状态
func (s *ArticleService) UpdateAIOptimized(id uint, optimized bool) error {
	return s.articleRepo.UpdateAIOptimized(id, optimized)
}

// AddPublishPlatform 添加发布平台
func (s *ArticleService) AddPublishPlatform(id uint, platform string) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	article.AddPublishPlatform(platform)

	return s.articleRepo.Update(article)
}
