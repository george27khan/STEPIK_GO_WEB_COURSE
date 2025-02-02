package repository

import (
	"context"
	dm "rwa/internal/domain"
)

//
//type User interface {
//	Create(ctx context.Context, user domain.User) error
//}
//
//type Article interface {
//	Create(ctx context.Context, article domain.Article) error
//	Get(ctx context.Context, user domain.User) ([]*domain.Article, error)
//}
//
//type Session interface {
//	Create(ctx context.Context, session domain.Session) error
//	Get(ctx context.Context, user domain.User) ([]*domain.Session, error)
//}

type UserStorage interface {
	Register(ctx context.Context, user *dm.User) error
}

type Repository struct {
	UserStorage
	//Article
	//Session
}

func NewRepository() {

}
