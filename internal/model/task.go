package model

import "time"

type TaskType string

const (
	TaskTypeArticleOptimize TaskType = "article_optimize"
	TaskTypeArticlePublish  TaskType = "article_publish"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCanceled   TaskStatus = "canceled"
)

type Task struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	Type       TaskType   `json:"type" gorm:"size:50;not null;index"`          // 优化任务、发布任务
	Status     TaskStatus `json:"status" gorm:"size:20;default:pending;index"` // 待处理、处理中、已完成、失败
	ArticleID  uint       `json:"article_id" gorm:"not null;index"`
	Payload    *string    `json:"payload,omitempty" gorm:"type:jsonb"` // 任务的输入数据，存储为 JSONB 类型
	Result     *string    `json:"result,omitempty" gorm:"type:jsonb"`  // 任务的输出数据，存储为 JSONB 类型
	Error      string     `json:"error" gorm:"type:text"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	MaxRetries int        `json:"max_retries"`

	Article *Article `json:"article" gorm:"foreignKey:ArticleID"`
}

func (Task) TableName() string {
	return "tasks"
}
