package dto

import "time"

type ArticleInfo struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

type Article struct {
	Info *ArticleInfo `json:"article"`
}

type Author struct {
	Bio      string `json:"bio"`
	Username string `json:"username"`
}
type ArticleRespInfo struct {
	Author      Author    `json:"author"`
	Body        string    `json:"body"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TagList     []string  `json:"tagList"`
}

type ArticleResp struct {
	ArticleRespInfo `json:"article"`
}

type ArticleAllResp struct {
	Articles      []*ArticleRespInfo `json:"articles"`
	ArticlesCount int                `json:"articlesCount"`
}
