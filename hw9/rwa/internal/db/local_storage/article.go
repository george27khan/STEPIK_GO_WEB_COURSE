package local_storage

import (
	"sync"
	"time"
)

type article struct {
	Author         user
	Body           string
	CreatedAt      time.Time
	Description    string
	Favorited      bool
	FavoritesCount int
	Slug           string
	TagList        []string
	Title          string
	UpdatedAt      time.Time
}
type ArticleStorage struct {
	Article   map[string]article
	ArticleID int
	Mu        *sync.RWMutex
}

// NewDBStorage создание хранилища данных
func NewArticleStorage() *ArticleStorage {
	article := make(map[string]article)
	return &ArticleStorage{
		article,
		0,
		&sync.RWMutex{},
	}
}
