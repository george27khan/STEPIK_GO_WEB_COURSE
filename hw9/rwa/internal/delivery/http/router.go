package http

import (
	"net/http"
	"rwa/internal/delivery/http/handler"
)

type mainHandler struct {
	userHandler    *handler.UserHandler
	sessionHandler *handler.SessionHandler
	articleHandler *handler.ArticleHandler
}

func NewMainHandler(userHandler *handler.UserHandler, sessionHandler *handler.SessionHandler, articleHandler *handler.ArticleHandler) *mainHandler {
	return &mainHandler{userHandler, sessionHandler, articleHandler}
}
func NewServerMUX(mainHandler *mainHandler) http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", mainHandler.userHandler.Register)
	mux.HandleFunc("/api/users/login", mainHandler.userHandler.Login)
	mux.HandleFunc("/api/user", mainHandler.userHandler.User)
	mux.HandleFunc("/api/user/logout", mainHandler.userHandler.Logout)
	mux.HandleFunc("/api/articles", mainHandler.articleHandler.Articles)

	handler := AuthMiddleware(mainHandler, mux)

	return handler
}
