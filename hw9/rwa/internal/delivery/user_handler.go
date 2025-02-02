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
	errCreateUser  error = errors.New("Create user error.")
	errLoginUser   error = errors.New("Login user error.")
	errRequestBody error = errors.New("Unmarshal body error.")
	errHTTPMethod  error = errors.New("Bad HTTP method.")
)

// интерфейс для обращения в usecase
type UserUseCase interface {
	Register(ctx context.Context, userCreate *dto.UserCreate) (*dto.UserCreateResp, error)
	Login(ctx context.Context, userCreate *dto.UserCreate) (*dto.UserCreateResp, error)
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
		if err := json.NewDecoder(r.Body).Decode(userCreate); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf(errRequestBody.Error(), err.Error()))
			return
		}
		//валидация

		//сохранение
		userResp, err := u.useCase.Register(ctx, userCreate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errCreateUser.Error(), err.Error()))
			return
		}
		resp, err := json.Marshal(userResp)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errCreateUser.Error(), err.Error()))
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
		return
	}
	writeError(w, http.StatusBadRequest, errHTTPMethod)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ctx := context.TODO()
		user := &dto.UserCreate{}
		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf(errRequestBody.Error(), err.Error()))
			return
		}
		userResp, err := u.useCase.Login(ctx, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errLoginUser.Error(), err.Error()))
			return
		}
		resp, err := json.Marshal(userResp)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errLoginUser.Error(), err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Println("string(resp)", string(resp))
		w.Write(resp)
		return
	}
	writeError(w, http.StatusBadRequest, errHTTPMethod)
}
