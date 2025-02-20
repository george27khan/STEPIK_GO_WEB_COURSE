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

func main1() {
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

func main() {
	slice := make([]int, 3, 4)

	prikol(slice)
	fmt.Println(slice, len(slice), cap(slice)) // вопрос: Изменился ли слайс?
	// Ответ: выведется [0,0,0] - т.е. 1 мы не увидим. Но 1 занесется в подкапотный массив
	// просто len будет 3, поскольку на 12 строке слайс копируется - соответственно и лен тоже, как и все в go.
	// Т.е. len не изменился, а массив под капотом изменился. Нужно просто увеличить len, сделав так slice[:4]

	slice = append(slice, 1)
	prikol(slice)
	fmt.Println(slice, len(slice), cap(slice)) // вопрос: Изменился ли слайс?
	// Ответ: выведется [0,0,0,1] - нет, во-первых len не изменился, во-вторых, даже в подкапотный массив у нас ничего не занеслось
	// Почему? Потому что у нас тут происходит аллокация, поскольку мы добавляем пятый элемент, а у нас capacity = 4

	prikol(slice[:1])
	fmt.Println(slice) // вопрос: Изменился ли слайс?
	// Ответ: выведется [0,1,0,1] - т.к. у нас передается тот же массив в скопированный слайс, но с измененным леном равным 1
	// поэтому при аппенде с таким леном он аппендит однерку и перезатирает ноль, стоящий под индексом 1 - потому что не знает,
	// что в подкапотном массиве уже что-то есть на этом индексе
}

func prikol(slice []int) {
	slice = append(slice, 1)
}
