package dto

import "time"

type InfoCreate struct {
	Email     string `json:"email"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCreate struct {
	Info InfoCreate `json:"user"`
}

type InfoCreateResp struct {
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserCreateResp struct {
	Info InfoCreateResp `json:"user"`
}
