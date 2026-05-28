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

func (s *ArticleService) Create(title, content string) (*model.Article, error) {
	article := &model.Article{
		Title:   title,
		Content: content,
		Status:  "draft",
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, err
	}

	return article, nil
}

func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	return s.articleRepo.GetByID(id)
}

func (s *ArticleService) List(page, pageSize int) ([]model.Article, int64, error) {
	offset := (page - 1) * pageSize
	return s.articleRepo.List(offset, pageSize)
}

func (s *ArticleService) Update(id uint, title, content string) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("文章不存在")
	}

	article.Title = title
	article.Content = content

	if err := s.articleRepo.Update(article); err != nil {
		return nil, err
	}

	return article, nil
}

func (s *ArticleService) Delete(id uint) error {
	_, err := s.articleRepo.GetByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	return s.articleRepo.Delete(id)
}
