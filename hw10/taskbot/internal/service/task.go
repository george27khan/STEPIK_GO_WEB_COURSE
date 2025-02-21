package service

import (
	"context"
	"taskbot/internal/domain"
)

type taskStorage interface {
	GetAll(ctx context.Context) ([]*domain.Task, error)
	GetTask(ctx context.Context, taskID uint64) (domain.Task, error)
	Create(ctx context.Context, task domain.Task) (uint64, error)
	AssignTask(ctx context.Context, executor string, executorChatID int64, taskID uint64) (domain.Task, error)
	UnassignTask(ctx context.Context, taskID uint64) error
	GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error)
	DeleteTask(ctx context.Context, taskID uint64) error
	GetUserTask(ctx context.Context, username string, role string) ([]*domain.Task, error)
}

type TaskService struct {
	storage taskStorage
}

func NewTaskService(storage taskStorage) *TaskService {
	return &TaskService{storage}
}

func (t *TaskService) GetAll(ctx context.Context) ([]*domain.Task, error) {
	if res, err := t.storage.GetAll(ctx); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func (t *TaskService) GetTask(ctx context.Context, taskID uint64) (*domain.Task, error) {
	task, err := t.storage.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return &task, nil
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

func (t *TaskService) UnassignTask(ctx context.Context, taskID uint64) error {
	if err := t.storage.UnassignTask(ctx, taskID); err != nil {
		return err
	} else {
		return nil
	}
}

func (t *TaskService) GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error) {
	if chatID, err := t.storage.GetExecutorChatID(ctx, taskID); err != nil {
		return 0, err
	} else {
		return chatID, nil
	}
}

func (t *TaskService) ResolveTask(ctx context.Context, taskID uint64) error {
	if err := t.storage.DeleteTask(ctx, taskID); err != nil {
		return err
	}
	return nil
}

func (t *TaskService) GetUserTask(ctx context.Context, username string, role string) ([]*domain.Task, error) {
	if tasks, err := t.storage.GetUserTask(ctx, username, role); err != nil {
		return nil, err
	} else {
		return tasks, nil
	}
}
