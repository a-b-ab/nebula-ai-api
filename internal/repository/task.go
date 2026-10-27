package repository

import (
	"errors"
	"nebula-ai-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type TaskRespository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRespository {
	return &TaskRespository{db: db}
}

func (r *TaskRespository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRespository) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRespository) List(offset, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	if err := r.db.Model(&model.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRespository) ListByArticleID(articlID uint, offset, limit int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Model(&model.Task{}).Where("article_id=?", articlID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tasks).Error
	return tasks, total, err
}

func (r *TaskRespository) FindRunningTask(articleID uint, taskType model.TaskType) (*model.Task, error) {
	var task model.Task
	err := r.db.Where(
		"article_id = ? AND type = ? AND status IN ?",
		articleID,
		taskType,
		[]model.TaskStatus{model.TaskStatusPending, model.TaskStatusProcessing},
	).Order("created_at DESC").First(&task).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRespository) FindNextPending() (*model.Task, error) {
	var task model.Task
	err := r.db.Where("status = ?", model.TaskStatusPending).
		Order("created_at ASC").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRespository) MarkProcessing(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     model.TaskStatusProcessing,
		"started_at": now,
		"error":      "",
	}).Error
}

func (r *TaskRespository) MarkCompleted(id uint, result string) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      model.TaskStatusCompleted,
		"result":      result,
		"finished_at": now,
	}).Error
}

func (r *TaskRespository) MarkFailed(id uint, errMsg string) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ? ", id).Updates(map[string]interface{}{
		"status":      model.TaskStatusFailed,
		"error":       errMsg,
		"finished_at": now,
	}).Error
}

// FailProcessingTasks 将遗留的处理中的任务标记为失败
func (r *TaskRespository) FailProcessingTasks(reason string) (int64, error) {
	now := time.Now()
	result := r.db.Model(&model.Task{}).
		Where("status = ? ", model.TaskStatusProcessing).
		Updates(map[string]interface{}{
			"status":      model.TaskStatusFailed,
			"error":       reason,
			"finished_at": now,
			"updated_at":  now,
		})
	return result.RowsAffected, result.Error
}
