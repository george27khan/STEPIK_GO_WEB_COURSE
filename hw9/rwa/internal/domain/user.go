package domain

import "time"

type User struct {
	ID        string
	Email     string
	Username  string
	Password  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
	BIO       string
}
