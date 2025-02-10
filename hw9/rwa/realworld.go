package main

import (
	"net/http"
	"rwa/internal/db"
	http2 "rwa/internal/delivery/http"
	"rwa/internal/delivery/http/handler"
	use_case "rwa/internal/usecase"
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
	useCase := use_case.NewUseCase(userCase, articleCase)
	userHandler := handler.NewUserHandler(useCase, sessCase)
	sessionHandler := handler.NewSessionHandler(sessCase)
	articleHandler := handler.NewArticleHandler(articleCase)
	mainHandler := http2.NewMainHandler(userHandler, sessionHandler, articleHandler)
	mux := http2.NewServerMUX(mainHandler)
	return mux
}
