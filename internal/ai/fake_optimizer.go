package ai

import (
	"context"
	"nebula-ai-api/internal/model"
)

type FakeArticleOptimizer struct{}

func NewFakeArticleOptimizer() *FakeArticleOptimizer {
	return &FakeArticleOptimizer{}
}

func (o *FakeArticleOptimizer) OptimizeArticle(ctx context.Context, article *model.Article) (*ArticleOptimizeResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return &ArticleOptimizeResult{
		Title:   article.Title,
		Summary: "这是 Fake AI 生成的文章摘要，用于验证任务系统流程。",
		Tags:    []string{"Go", "AI", "Nebula"},
		Content: article.Content + "\n\n## AI 优化建议\n\n这是 Fake AI 生成的优化内容，用于测试任务执行链路。",
	}, nil
}
