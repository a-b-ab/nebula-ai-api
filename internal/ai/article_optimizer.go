package ai

import (
	"context"
	"nebula-ai-api/internal/model"
)

type ArticleOptimizeResult struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}

type ArticleOptimizer interface {
	OptimizeArticle(ctx context.Context, article *model.Article) (*ArticleOptimizeResult, error)
}
