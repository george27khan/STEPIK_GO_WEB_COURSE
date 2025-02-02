package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rwa/internal/dto"
)

var (
	errCreateUser error = errors.New("Create user error.")
)

// интерфейс для обращения в usecase
type UserUseCase interface {
	Register(ctx context.Context, user *dto.UserCreate) error
}

func NewUserHandler(useCase UserUseCase) *UserHandler {
	return &UserHandler{useCase}
}

type UserHandler struct {
	useCase UserUseCase
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf("Handle error: %s", err.Error())))
}

//curl -X POST -H "Content-Type: application/json" -d '{"user":{"email":"kkk", "password":"kkk", "username":"kkk"}}' http://10.200.129.200:8080/users

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ctx := context.TODO()
		userCreate := &dto.UserCreate{}
		body := r.Body
		if err := json.NewDecoder(body).Decode(userCreate); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Ошибка при разборе JSON из тела запроса"))
		}
		//валидация

		//сохранение
		err := u.useCase.Register(ctx, userCreate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errCreateUser.Error(), err.Error()))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Пользователь создан!"))
	}

	//ser\":{\"email\":\"{{EMAIL}}\", \"password\":\"{{PASSWORD}}\", \"username\":\"{{USERNAME}}\"}}",

}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world!"))
}
