package dto

import (
	"time"
)

type UserReqInfo struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username,omitempty"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	BIO       string    `json:"bio"`
	Token     string    `json:"token"`
}

type UserReq struct {
	Info *UserReqInfo `json:"user"`
}

type Errors struct {
	Body []string `json:"body"`
}
type Error struct {
	Errors `json:"errors"`
}
