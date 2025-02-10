package domain

import "time"

type Article struct {
	ID             string
	Author         *User
	Body           string
	Description    string
	Favorited      bool
	FavoritesCount int
	Slug           string
	TagList        []string
	Title          string
	UpdatedAt      time.Time
	CreatedAt      time.Time
}
