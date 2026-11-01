package repository

import (
	"nebula-ai-api/internal/model"

	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// Create 创建文章
func (r *ArticleRepository) Create(article *model.Article) error {
	return r.db.Create(article).Error
}

// GetByID 根据ID获取文章
func (r *ArticleRepository) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	err := r.db.First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// List 获取文章列表（分页）
func (r *ArticleRepository) List(offset, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	if err := r.db.Model(&model.Article{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error
	return articles, total, err
}

// ListByStatus 根据状态获取文章列表
func (r *ArticleRepository) ListByStatus(status model.ArticleStatus, offset, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error
	return articles, total, err
}

// Update 更新文章
func (r *ArticleRepository) Update(article *model.Article) error {
	return r.db.Save(article).Error
}

// Delete 删除文章
func (r *ArticleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Article{}, id).Error
}

// UpdateStatus 更新文章状态
func (r *ArticleRepository) UpdateStatus(id uint, status model.ArticleStatus) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateAIOptimized 更新AI优化状态
func (r *ArticleRepository) UpdateAIOptimized(id uint, optimized bool) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).Update("ai_optimized", optimized).Error
}
