package delivery

import (
	"net/http"
	"rwa/internal/db"
)

func NewServerMUX() *http.ServeMux {
	db := db.NewDBStorage()
	user := NewUserHandler(db)
	mux := http.NewServeMux()
	mux.HandleFunc("/users", user.Register)
	mux.HandleFunc("/users/login", user.Login)
	return mux
}
