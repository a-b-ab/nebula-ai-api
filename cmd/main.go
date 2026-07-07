package main

import (
	"log"
	"nebula-ai-api/internal/api"
	"nebula-ai-api/internal/config"
	"nebula-ai-api/internal/model"
	"nebula-ai-api/internal/repository"
	"nebula-ai-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("警告: 无法加载 .env 文件，使用默认配置")
	}

	cfg := config.Load()

	// 初始化 GORM 数据库
	dsn := "host=" + cfg.DBHost + " port=" + cfg.DBPort + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&model.Article{}, &model.Task{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化 Repository
	articleRepo := repository.NewArticleRepository(db)

	// 初始化 Service
	articleService := service.NewArticleService(articleRepo)
	markdownService := service.NewMarkdownService()

	// 初始化 Gin
	r := gin.Default()

	// 设置路由
	api.SetupRoutes(r, articleService, markdownService)

	// 启动服务器
	log.Println("服务器启动，端口: " + cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
