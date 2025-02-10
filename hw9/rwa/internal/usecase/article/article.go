package article

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	h "rwa/internal/delivery/http/handler"
	dm "rwa/internal/domain"
	"rwa/internal/usecase/user"
)

var _ h.ArticleUseCase = (*ArticleUseCase)(nil)

// ArticleRepository интерфейс репозитория статей
type ArticleRepository interface {
	Create(ctx context.Context, article *dm.Article) error
	GetAll(ctx context.Context) ([]*dm.Article, error)
	GetByUserID(ctx context.Context, userID string) ([]*dm.Article, error)
	GetByTag(ctx context.Context, tag string) ([]*dm.Article, error)
}

// ArticleUseCase структура юзкейсов со статьями
type ArticleUseCase struct {
	DBArticle ArticleRepository
	UserCase  *user.UserUseCase
}

// NewArticleUseCase конструктор юзкейсов со статьями
func NewArticleUseCase(DBArticle ArticleRepository, UserCase *user.UserUseCase) *ArticleUseCase {
	return &ArticleUseCase{DBArticle, UserCase}
}

// Create создание статьи
func (a *ArticleUseCase) Create(ctx context.Context, userID string, article *dm.Article) (*dm.Article, error) {
	// получаем автора статьи
	user, err := a.UserCase.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Create article error: %s", err.Error())
	}
	article.ID = uuid.NewString()
	article.Author = user
	if err := a.DBArticle.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("Create article error: %s", err.Error())
	}
	article.Author = user
	return article, nil
}

// GetAll получение всех статей
func (a *ArticleUseCase) GetAll(ctx context.Context) ([]*dm.Article, error) {
	//получаем статьи их репозитория
	articles, err := a.DBArticle.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("Get articles error: %s", err.Error())
	}
	// дополняем информацией о пользователе
	for _, article := range articles {
		user, err := a.UserCase.Get(ctx, article.Author.ID)
		if err != nil {
			return nil, fmt.Errorf("Create article error: %s", err.Error())
		}
		article.Author = user
	}
	return articles, nil
}

// GetAll получение статей по автору
func (a *ArticleUseCase) GetByAuthor(ctx context.Context, username string) ([]*dm.Article, error) {
	//получаем статьи их репозитория
	userDM, err := a.UserCase.DBUser.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("Author not found: %s", err.Error())
	}
	articles, err := a.DBArticle.GetByUserID(ctx, userDM.ID)
	if err != nil {
		return nil, fmt.Errorf("Get articles error: %s", err.Error())
	}
	// дополняем информацией о пользователе
	for _, article := range articles {
		article.Author = userDM
	}
	return articles, nil
}

// GetAll получение статей по тэгу
func (a *ArticleUseCase) GetByTag(ctx context.Context, tag string) ([]*dm.Article, error) {
	//получаем статьи из репозитория
	articles, err := a.DBArticle.GetByTag(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("Get articles error: %s", err.Error())
	}
	// дополняем информацией о пользователе
	for _, article := range articles {
		userDM, err := a.UserCase.DBUser.GetByID(ctx, article.Author.ID)
		if err != nil {
			return nil, fmt.Errorf("Author not found: %s", err.Error())
		}
		article.Author = userDM
	}
	return articles, nil
}
