package model

import "time"

type Article struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"size:200;not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Summary   string    `json:"summary" gorm:"type:text"`
	Tags      string    `json:"tags" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20;default:draft"` // 草稿、已优化、已发布
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Article) TableName() string {
	return "articles"
}
