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
			articles.POST("", articleHandler.Create)
			articles.GET("", articleHandler.List)
			articles.GET("/:id", articleHandler.GetByID)
			articles.PUT("/:id", articleHandler.Update)
			articles.DELETE("/:id", articleHandler.Delete)
		}
	}
}
