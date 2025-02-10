package handler

import (
	"context"
	"fmt"
	"net/http"
	dm "rwa/internal/domain"
	"strings"
)

// интерфейс для обращения в usecase
type SessionUseCase interface {
	Create(ctx context.Context, userID string) (*dm.Session, error)
	Get(ctx context.Context, sessionID string) (*dm.Session, error)
}

func NewSessionHandler(useCase SessionUseCase) *SessionHandler {
	return &SessionHandler{useCase}
}

type SessionHandler struct {
	useCase SessionUseCase
}

// CheckSession проверка сессии в бд из куков запроса
func (sh *SessionHandler) Get(r *http.Request) (*dm.Session, error) {
	ctx := r.Context()
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("No session token")
	}
	sessionID, found := strings.CutPrefix(token, "Token ")
	if !found {
		return nil, fmt.Errorf("Parse session error")
	}
	if session, err := sh.useCase.Get(ctx, sessionID); err != nil {
		return nil, fmt.Errorf("Check session error: %s", err.Error())
	} else {
		return session, nil
	}
}
