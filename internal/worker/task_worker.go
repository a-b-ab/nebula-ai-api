package worker

import (
	"context"
	"log"
	"nebula-ai-api/internal/service"
	"time"
)

type TaskWorker struct {
	taskService *service.TaskService
	interval    time.Duration
}

func NewTaskWorker(taskService *service.TaskService, interval time.Duration) *TaskWorker {
	return &TaskWorker{
		taskService: taskService,
		interval:    interval,
	}
}

func (w *TaskWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("任务 worker 已启动")

	for {
		select {
		case <-ctx.Done():
			log.Println("任务 worker 已停止")
			return
		case <-ticker.C:
			if err := w.taskService.ExecuteNext(ctx); err != nil {
				log.Println("任务执行失败:", err)
			}
		}
	}
}
