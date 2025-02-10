package session

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	dm "rwa/internal/domain"
)

// SessionRepository интерфейс репозитория сессий
type SessionRepository interface {
	Create(ctx context.Context, session *dm.Session) error
	Get(ctx context.Context, sessionID string) (*dm.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

// SessionUseCase структура юзкейсов с сессиями
type SessionUseCase struct {
	db SessionRepository
}

// NewSessionUseCase конструктор юзкейсов с сессиями
func NewSessionUseCase(db SessionRepository) *SessionUseCase {
	return &SessionUseCase{db}
}

// Create создание сессии для пользователя
func (uc *SessionUseCase) Create(ctx context.Context, userID string) (*dm.Session, error) {
	sessionID := uuid.NewString()
	session := &dm.Session{sessionID, userID}
	if err := uc.db.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("Create session error: %s", err.Error())
	}
	return session, nil
}

// Get получение сессии пользователя
func (uc *SessionUseCase) Get(ctx context.Context, sessionID string) (*dm.Session, error) {
	if session, err := uc.db.Get(ctx, sessionID); err != nil {
		return nil, err
	} else {
		return session, nil
	}
}

// Delete удаление сессии пользователя
func (uc *SessionUseCase) Delete(ctx context.Context, sessionID string) error {
	if err := uc.db.Delete(ctx, sessionID); err != nil {
		return err
	}
	return nil

}
