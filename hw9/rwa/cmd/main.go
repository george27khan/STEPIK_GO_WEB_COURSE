package main

import (
	"net/http"
	"rwa/internal/db"
	http2 "rwa/internal/delivery/http"
	"rwa/internal/delivery/http/handler"
	use_case "rwa/internal/usecase"
	session_case "rwa/internal/usecase/session"
	user_case "rwa/internal/usecase/user"
)

func main() {
	dbLocal := db.NewDBStorage()
	userCase := user_case.NewUserUseCase(dbLocal.User)
	sessCase := session_case.NewSessionUseCase(dbLocal.Session)
	useCase := use_case.NewUseCase(userCase)
	userHandler := handler.NewUserHandler(useCase, sessCase)
	sessionHandler := handler.NewSessionHandler(sessCase)
	mainHandler := http2.NewMainHandler(userHandler, sessionHandler)
	mux := http2.NewServerMUX(mainHandler)
	http.ListenAndServe(":8080", mux)
}
