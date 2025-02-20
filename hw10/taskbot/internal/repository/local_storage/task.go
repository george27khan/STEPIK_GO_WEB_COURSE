package local_storage

import (
	"context"
	"errors"
	"sync"
	"taskbot/internal/domain"
)

type TaskStorage struct {
	mu      sync.RWMutex
	storage []*domain.Task
	id      uint64
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{sync.RWMutex{}, []*domain.Task{}, 1}
}

func (t *TaskStorage) GetAll(ctx context.Context) []*domain.Task {
	res := make([]*domain.Task, len(t.storage))
	t.mu.RLock()
	copy(res, t.storage)
	t.mu.RUnlock()
	return t.storage //возвращаем копию, чтобы не давать доступ к данным
}

func (t *TaskStorage) Create(ctx context.Context, task domain.Task) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	task.ID = t.id
	t.id++
	t.storage = append(t.storage, &task)
	return task.ID, nil
}

func (t *TaskStorage) AssignTask(ctx context.Context, executor string, executorChatID int64, taskID uint64) (domain.Task, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, task := range t.storage {
		if task.ID == taskID {
			task.Executor = executor
			task.ExecutorChatID = executorChatID
			return *task, nil
		}
	}
	return domain.Task{}, errors.New("таск не найден в хранилище")
}

func (t *TaskStorage) GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, task := range t.storage {
		if task.ID == taskID {
			return task.ExecutorChatID, nil
		}
	}
	return 0, errors.New("таск не найден в хранилище")
}
