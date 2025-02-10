package route

import (
	"net/http"
	handler2 "rwa/internal/delivery/http/route/handler"
)

type mainHandler struct {
	userHandler    *handler2.UserHandler
	sessionHandler *handler2.SessionHandler
	articleHandler *handler2.ArticleHandler
}

func NewMainHandler(userHandler *handler2.UserHandler, sessionHandler *handler2.SessionHandler, articleHandler *handler2.ArticleHandler) *mainHandler {
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
