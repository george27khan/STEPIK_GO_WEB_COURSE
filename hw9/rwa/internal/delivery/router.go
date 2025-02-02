package delivery

import (
	"net/http"
)

type mainHandler struct {
	userHandler *UserHandler
}

func NewMainHandler(userHandler *UserHandler) *mainHandler {
	return &mainHandler{userHandler}
}
func NewServerMUX(mainHAndler *mainHandler) *http.ServeMux {

	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", mainHAndler.userHandler.Register)
	mux.HandleFunc("/api/users/login", mainHAndler.userHandler.Login)
	return mux
}
