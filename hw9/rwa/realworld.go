package main

import (
	"net/http"
	"rwa/internal/db"
	"rwa/internal/delivery"
	use_case "rwa/internal/usecase"
	user_case "rwa/internal/usecase/user"
)

// сюда писать код

func GetApp() http.Handler {
	dbLocal := db.NewDBStorage()
	userCase := user_case.NewUserUseCase(dbLocal.User)
	useCase := use_case.NewUseCase(userCase)
	userHandler := delivery.NewUserHandler(useCase)
	mainHandler := delivery.NewMainHandler(userHandler)
	mux := delivery.NewServerMUX(mainHandler)
	return mux

}
