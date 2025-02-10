package local_storage

import (
	"context"
	"fmt"
	dm "rwa/internal/domain"
	"sync"
)

type Session struct {
	ID     string
	UserID string
}

type SessionStorage struct {
	Session map[string]Session
	Mu      *sync.RWMutex
}

// NewDBStorage создание хранилища данных
func NewSessionStorage() *SessionStorage {
	session := make(map[string]Session)
	return &SessionStorage{
		session,
		&sync.RWMutex{},
	}
}

// Create создание сессии в хранилище
func (ss *SessionStorage) Create(ctx context.Context, session *dm.Session) error {
	ss.Mu.Lock()
	defer ss.Mu.Unlock()
	//
	if _, ok := ss.Session[session.ID]; ok {
		return fmt.Errorf("Для пользователя уже существует сессия")
	}
	ss.Session[session.ID] = Session{session.ID, session.UserID}
	return nil
}

// Get получить сессию юзера, подразумевается одна сессия на пользователя
func (ss *SessionStorage) Get(ctx context.Context, sessionID string) (*dm.Session, error) {
	ss.Mu.RLock()
	defer ss.Mu.RUnlock()
	if session, ok := ss.Session[sessionID]; ok {
		return &dm.Session{session.ID, session.UserID}, nil
	} else {
		return &dm.Session{}, fmt.Errorf("Session for %s not found")
	}
}

// Create создание сессии в хранилище
func (ss *SessionStorage) Delete(ctx context.Context, sessionID string) error {
	ss.Mu.Lock()
	defer ss.Mu.Unlock()
	if _, ok := ss.Session[sessionID]; ok {
		delete(ss.Session, sessionID)
		return nil
	} else{
		return fmt.Errorf("Сессия не существует")
	}
}

