package dto

import "time"

type Info struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCreate struct {
	Info Info `json:"user"`
}
