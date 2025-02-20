package service

import (
	"context"
	"taskbot/internal/domain"
)

type taskStorage interface {
	GetAll(ctx context.Context) []*domain.Task
	Create(ctx context.Context, task domain.Task) (uint64, error)
	AssignTask(ctx context.Context, executor string, executorChatID int64, taskID uint64) (domain.Task, error)
	GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error)
}

type TaskService struct {
	storage taskStorage
}

func NewTaskService(storage taskStorage) *TaskService {
	return &TaskService{storage}
}

func (t *TaskService) GetAll(ctx context.Context) ([]*domain.Task, error) {
	res := t.storage.GetAll(ctx)
	return res, nil
}

func (t *TaskService) Create(ctx context.Context, task domain.Task) (uint64, error) {

	if id, err := t.storage.Create(ctx, task); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (t *TaskService) AssignTask(ctx context.Context, executor string, executorChatID int64, taskID uint64) (domain.Task, error) {
	if task, err := t.storage.AssignTask(ctx, executor, executorChatID, taskID); err != nil {
		return domain.Task{}, err
	} else {
		return task, nil
	}
}

func (t *TaskService) GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error) {
	if chatID, err := t.storage.GetExecutorChatID(ctx, taskID); err != nil {
		return 0, err
	} else {
		return chatID, nil
	}
}
