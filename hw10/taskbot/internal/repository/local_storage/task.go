package local_storage

import (
	"context"
	"errors"
	"sync"
	"taskbot/internal/domain"
)

type TaskStorage struct {
	mu      sync.RWMutex
	storage []domain.Task
	id      uint64
}

func NewTaskStorage() *TaskStorage {
	return &TaskStorage{sync.RWMutex{}, []domain.Task{}, 1}
}

func (t *TaskStorage) GetAll(ctx context.Context) ([]*domain.Task, error) {
	res := make([]*domain.Task, len(t.storage))
	t.mu.RLock()
	defer t.mu.RUnlock()
	for i, task := range t.storage {
		task := task // переопределяем для создания копий
		res[i] = &task
	}
	return res, nil //возвращаем копию, чтобы не давать доступ к данным
}

func (t *TaskStorage) GetTask(ctx context.Context, taskID uint64) (domain.Task, error) {
	t.mu.RLock()
	t.mu.RUnlock()
	for _, task := range t.storage {
		if task.ID == taskID {
			return task, nil
		}
	}
	return domain.Task{}, errors.New("таск не найден в хранилище")
}

func (t *TaskStorage) Create(ctx context.Context, task domain.Task) (uint64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	task.ID = t.id
	t.id++
	t.storage = append(t.storage, task)
	return task.ID, nil
}

func (t *TaskStorage) AssignTask(ctx context.Context, executor string, executorChatID int64, taskID uint64) (domain.Task, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, task := range t.storage {
		if task.ID == taskID {
			task.Executor = executor
			task.ExecutorChatID = executorChatID
			t.storage[i] = task
			return task, nil
		}
	}
	return domain.Task{}, errors.New("таск не найден в хранилище")
}

func (t *TaskStorage) UnassignTask(ctx context.Context, taskID uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, task := range t.storage {
		if task.ID == taskID {
			task.Executor = ""
			task.ExecutorChatID = 0
			t.storage[i] = task
			return nil
		}
	}
	return errors.New("таск не найден в хранилище")
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

func (t *TaskStorage) DeleteTask(ctx context.Context, taskID uint64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, task := range t.storage {
		if task.ID == taskID {
			t.storage = append(t.storage[:i], t.storage[i+1:]...)
			return nil
		}
	}
	return errors.New("таск не найден в хранилище")
}

func (t *TaskStorage) GetUserTask(ctx context.Context, username string, role string) ([]*domain.Task, error) {
	userTasks := make([]*domain.Task, 0)
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, task := range t.storage {
		if role == "executor" {
			if task.Executor == username {
				task := task // переопределяем для создания копий
				userTasks = append(userTasks, &task)
			}
		} else if role == "author" {
			if task.Author == username {
				task := task // переопределяем для создания копий
				userTasks = append(userTasks, &task)
			}
		}
	}
	return userTasks, nil
}
