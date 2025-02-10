package local_storage

import (
	"context"
	"fmt"
	dm "rwa/internal/domain"
	"sync"
	"time"
)

type Article struct {
	ID             string
	UserID         string
	Body           string
	Description    string
	Favorited      bool
	FavoritesCount int
	Slug           string
	TagList        []string
	Title          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ArticleStorage struct {
	ArticlesByID     map[string]*Article   // храним статьи по ид  для быстрого ответа
	ArticlesByUserID map[string][]*Article // храним статьи по ид юзеру для быстрого ответа
	ArticlesSlice    []*Article            // храним статьи по порядку создания для ответа, т.к. для теста нужен порядок в ответе
	Mu               *sync.RWMutex
}

// NewArticleStorage создание хранилища данных
func NewArticleStorage() *ArticleStorage {
	return &ArticleStorage{
		make(map[string]*Article),
		make(map[string][]*Article),
		make([]*Article, 0),
		&sync.RWMutex{},
	}
}

// Create создание статьи в хранилище
func (a *ArticleStorage) Create(ctx context.Context, article *dm.Article) error {
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()
	articleDB := &Article{
		ID:             article.ID,
		UserID:         article.Author.ID,
		Body:           article.Body,
		Description:    article.Description,
		Favorited:      article.Favorited,
		FavoritesCount: article.FavoritesCount,
		Slug:           article.Slug,
		TagList:        article.TagList,
		Title:          article.Title,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	a.Mu.Lock()
	defer a.Mu.Unlock()
	//ключем для поиска и создания будет почта
	if _, ok := a.ArticlesByID[articleDB.ID]; ok {
		return fmt.Errorf("Статья уже существует")
	}
	a.ArticlesByID[articleDB.ID] = articleDB
	a.ArticlesSlice = append(a.ArticlesSlice, articleDB)
	if articles, ok := a.ArticlesByUserID[articleDB.UserID]; ok {
		articles = append(articles, articleDB)
		a.ArticlesByUserID[articleDB.UserID] = articles

	} else {
		a.ArticlesByUserID[articleDB.UserID] = []*Article{articleDB}
	}
	return nil
}

// GetAll получение всех статей пользователя
func (as *ArticleStorage) GetAll(ctx context.Context) ([]*dm.Article, error) {
	as.Mu.RLock()
	defer as.Mu.RUnlock()
	res := make([]*dm.Article, 0, len(as.ArticlesSlice))
	for _, article := range as.ArticlesSlice {
		res = append(res, as.mapToDomain(article))
	}
	return res, nil
}

// GetByUserID получение всех статей пользователя по автору
func (as *ArticleStorage) GetByUserID(ctx context.Context, userID string) ([]*dm.Article, error) {
	as.Mu.RLock()
	defer as.Mu.RUnlock()
	if articles, ok := as.ArticlesByUserID[userID]; ok {
		articlesRes := make([]*dm.Article, len(articles))
		for i, article := range articles {
			articlesRes[i] = as.mapToDomain(article)
		}
		return articlesRes, nil
	} else {
		return nil, fmt.Errorf("No articles for user")
	}
}

// GetByTag получение всех статей по тэгу
func (as *ArticleStorage) GetByTag(ctx context.Context, tag string) ([]*dm.Article, error) {
	as.Mu.RLock()
	defer as.Mu.RUnlock()
	articlesRes := make([]*dm.Article, 0)
	for _, article := range as.ArticlesSlice {
		for _, articleTag := range article.TagList {
			if articleTag == tag {
				articlesRes = append(articlesRes, as.mapToDomain(article))
				break
			}
		}
	}
	if len(articlesRes) > 0 {
		return articlesRes, nil
	} else {
		return nil, fmt.Errorf("No articles for tag")
	}
}

// mapToDomain маппинг структуры БД в структуру домена
func (as *ArticleStorage) mapToDomain(article *Article) *dm.Article {
	articleDM := &dm.Article{
		ID:             article.ID,
		Author:         &dm.User{ID: article.UserID},
		Body:           article.Body,
		Description:    article.Description,
		Favorited:      article.Favorited,
		FavoritesCount: article.FavoritesCount,
		Slug:           article.Slug,
		TagList:        article.TagList,
		Title:          article.Title,
		UpdatedAt:      article.UpdatedAt,
		CreatedAt:      article.CreatedAt,
	}
	return articleDM
}
