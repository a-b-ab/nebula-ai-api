package model

import "time"

type Task struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type" gorm:"size:50;not null"`     // 优化任务、发布任务
	ArticleID uint      `json:"article_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"size:20;default:pending"` // 待处理、处理中、已完成、失败
	Result    string    `json:"result" gorm:"type:text"`
	Error     string    `json:"error" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Article *Article `json:"article" gorm:"foreignKey:ArticleID"`
}

func (Task) TableName() string {
	return "tasks"
}
