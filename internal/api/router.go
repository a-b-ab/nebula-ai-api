package api

import (
	"nebula-ai-api/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, articleService *service.ArticleService) {
	articleHandler := NewArticleHandler(articleService)

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		articles := api.Group("/articles")
		{
			articles.POST("", articleHandler.Create)                  // 创建文章
			articles.GET("", articleHandler.List)                     // 获取文章列表
			articles.GET("/:id", articleHandler.GetByID)              // 获取文章详情
			articles.PUT("/:id", articleHandler.Update)               // 更新文章
			articles.DELETE("/:id", articleHandler.Delete)            // 删除文章
			articles.PUT("/:id/status", articleHandler.UpdateStatus)   // 更新文章状态
			articles.POST("/:id/platforms", articleHandler.AddPublishPlatform) // 添加发布平台
		}
	}
}
