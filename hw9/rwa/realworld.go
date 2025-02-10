package main

import (
	"net/http"
	"rwa/internal/db"
	"rwa/internal/delivery/http/route"
	"rwa/internal/delivery/http/route/handler"
	"rwa/internal/usecase"
	article "rwa/internal/usecase/article"
	session "rwa/internal/usecase/session"
	user "rwa/internal/usecase/user"
)

// сюда писать код

func GetApp() http.Handler {
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
	return mux
}
