package main

import (
	"fmt"
	"net/http"
	"rwa/internal/db"
	"rwa/internal/delivery/http/route"
	"rwa/internal/delivery/http/route/handler"
	"rwa/internal/usecase"
	"rwa/internal/usecase/article"
	"rwa/internal/usecase/session"
	"rwa/internal/usecase/user"
)

func main() {
	dbLocal := db.NewDBStorage()
	sessCase := session.NewSessionUseCase(dbLocal.Session)
	userCase := user.NewUserUseCase(dbLocal.User, sessCase)
	articleCase := article.NewArticleUseCase(dbLocal.Article, userCase)
	useCase := usecase.NewUseCase(userCase, articleCase)
	userHandler := handler.NewUserHandler(useCase, sessCase)
	sessionHandler := handler.NewSessionHandler(sessCase)
	articleHandler := handler.NewArticleHandler(articleCase)
	mainHandler := route.NewMainHandler(userHandler, sessionHandler, articleHandler)
	mux := route.NewServerMUX(mainHandler)
	http.ListenAndServe(":8080", mux)
}


