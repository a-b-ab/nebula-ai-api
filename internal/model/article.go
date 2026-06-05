package model

import "time"

// ArticleStatus 文章状态
type ArticleStatus int

const (
	StatusDraft     ArticleStatus = 0 // 草稿
	StatusOptimized ArticleStatus = 1 // 已优化
	StatusPublished ArticleStatus = 2 // 已发布
)

func (s ArticleStatus) String() string {
	switch s {
	case StatusDraft:
		return "draft"
	case StatusOptimized:
		return "optimized"
	case StatusPublished:
		return "published"
	default:
		return "unknown"
	}
}

type Article struct {
	ID              uint          `json:"id" gorm:"primaryKey"`
	Title           string        `json:"title" gorm:"size:500;not null"`
	Content         string        `json:"content" gorm:"type:text;not null"`
	Summary         string        `json:"summary" gorm:"size:1000"`
	Tags            string        `json:"tags" gorm:"size:1000"`
	AIOptimized     bool          `json:"ai_optimized" gorm:"default:false"`
	PublishPlatform string        `json:"publish_platform" gorm:"size:200"`
	Status          ArticleStatus `json:"status" gorm:"default:0"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// TableName 指定表名
func (Article) TableName() string {
	return "articles"
}

// GetPublishPlatforms 获取已发布的平台列表
func (a *Article) GetPublishPlatforms() []string {
	if a.PublishPlatform == "" {
		return []string{}
	}
	platforms := make([]string, 0)
	for _, p := range []string{"juejin", "csdn", "zhihu"} {
		if a.IsPublishedTo(p) {
			platforms = append(platforms, p)
		}
	}
	return platforms
}

// AddPublishPlatform 添加发布平台
func (a *Article) AddPublishPlatform(platform string) {
	if a.PublishPlatform == "" {
		a.PublishPlatform = platform
		return
	}

	// 检查是否已存在
	platforms := make([]string, 0)
	current := a.PublishPlatform
	start := 0
	for i := 0; i <= len(current); i++ {
		if i == len(current) || current[i] == ',' {
			if i > start {
				p := current[start:i]
				if p == platform {
					return // 已存在
				}
				platforms = append(platforms, p)
			}
			start = i + 1
		}
	}

	// 添加新平台
	a.PublishPlatform = a.PublishPlatform + "," + platform
}

// IsPublishedTo 检查是否已发布到某平台
func (a *Article) IsPublishedTo(platform string) bool {
	if a.PublishPlatform == "" {
		return false
	}

	// 简单实现：检查字符串包含
	start := 0
	for i := 0; i <= len(a.PublishPlatform); i++ {
		if i == len(a.PublishPlatform) || a.PublishPlatform[i] == ',' {
			if i > start && a.PublishPlatform[start:i] == platform {
				return true
			}
			start = i + 1
		}
	}
	return false
}
