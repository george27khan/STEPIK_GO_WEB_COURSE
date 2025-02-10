package usecase

import (
	a "rwa/internal/usecase/article"
	u "rwa/internal/usecase/user"
)

type UseCase struct {
	*u.UserUseCase
	*a.ArticleUseCase
}

func NewUseCase(user *u.UserUseCase, article *a.ArticleUseCase) *UseCase {
	return &UseCase{user, article}
}
