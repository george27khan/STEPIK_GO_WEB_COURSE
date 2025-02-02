package usecase

import (
	u "rwa/internal/usecase/user"
)

type UseCase struct {
	*u.UserUseCase
}

func NewUseCase(user *u.UserUseCase) *UseCase {
	return &UseCase{user}
}
