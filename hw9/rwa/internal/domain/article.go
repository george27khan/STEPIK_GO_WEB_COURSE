package domain

type Article struct {
	Author         User
	Body           string
	Description    string
	Favorited      bool
	FavoritesCount int
	Slug           string
	TagList        []string
	Title          string
	UpdatedAt      bool
}
