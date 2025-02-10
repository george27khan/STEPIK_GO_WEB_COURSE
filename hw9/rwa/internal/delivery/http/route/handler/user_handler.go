package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	dm "rwa/internal/domain"
	"rwa/internal/dto"
)

// интерфейс для обращения в usecase
type UserUseCase interface {
	Register(ctx context.Context, userCreate *dto.UserReq) (*dm.User, error)
	Login(ctx context.Context, userCreate *dto.UserReq) (*dm.User, error)
	Get(ctx context.Context, ID string) (*dm.User, error)
	Update(ctx context.Context, user *dto.UserReqInfo) (*dm.User, error)
	Logout(ctx context.Context, sessionID string) error
}

func NewUserHandler(userCase UserUseCase, sessionCase SessionUseCase) *UserHandler {
	return &UserHandler{userCase, sessionCase}
}

type UserHandler struct {
	userCase    UserUseCase
	sessionCase SessionUseCase
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ctx := r.Context()
		userCreate := &dto.UserReq{}
		if err := json.NewDecoder(r.Body).Decode(userCreate); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf(errRequestBody.Error(), err.Error()))
			return
		}
		//валидация

		//сохранение пользователя
		userDB, err := u.userCase.Register(ctx, userCreate)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errCreateUser.Error(), err.Error()))
			return
		}
		//создание сессии
		session, err := u.sessionCase.Create(ctx, userDB.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errCreateSession.Error(), err.Error()))
			return
		}
		// инициализируем поля структуры для ответа
		userResp := &dto.UserReq{&dto.UserReqInfo{
			Email:     userDB.Email,
			Username:  userDB.Username,
			CreatedAt: userDB.CreatedAt,
			UpdatedAt: userDB.UpdatedAt,
			Token:     session.ID,
		}}
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
		ctx := r.Context()
		user := &dto.UserReq{}
		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf(errRequestBody.Error(), err.Error()))
			return
		}
		userDB, err := u.userCase.Login(ctx, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errLoginUser.Error(), err.Error()))
			return
		}

		session, err := u.sessionCase.Create(ctx, userDB.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errCreateSession.Error(), err.Error()))
			return
		}
		userResp := &dto.UserReq{
			&dto.UserReqInfo{
				Email:     userDB.Email,
				Username:  userDB.Username,
				CreatedAt: userDB.CreatedAt,
				UpdatedAt: userDB.UpdatedAt,
				Token:     session.ID,
			}}

		resp, err := json.Marshal(userResp)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return
	}
	writeError(w, http.StatusBadRequest, errHTTPMethod)
}

func (u *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value("session").(*dm.Session)
	if userDM, err := u.userCase.Get(ctx, session.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errGetUser.Error(), err.Error()))
		return
	} else {
		user := &dto.UserReq{
			&dto.UserReqInfo{
				Email:     userDM.Email,
				Username:  userDM.Username,
				CreatedAt: userDM.CreatedAt,
				UpdatedAt: userDM.UpdatedAt,
				BIO:       userDM.BIO,
				Token:     session.ID,
			}}
		resp, err := json.Marshal(user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return
	}
}

// postUser хендлер обновления данных пользователя
func (u *UserHandler) postUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := ctx.Value("session").(*dm.Session)

	// получаем тело запроса
	user := &dto.UserReq{&dto.UserReqInfo{}}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf(errRequestBody.Error(), err.Error()))
		return
	}

	// обновление клиентских данных
	user.Info.ID = session.UserID
	userUpd, err := u.userCase.Update(ctx, user.Info)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf(errUserUpd.Error(), err.Error()))
		return
	}
	userResp := &dto.UserReq{
		&dto.UserReqInfo{
			Email:     userUpd.Email,
			Username:  userUpd.Username,
			CreatedAt: userUpd.CreatedAt,
			UpdatedAt: userUpd.UpdatedAt,
			BIO:       userUpd.BIO,
			Token:     session.ID,
		}}
	resp, err := json.Marshal(userResp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf(errMarshalResp.Error(), err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}

func (u *UserHandler) User(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		u.getUser(w, r)
		return
	} else if r.Method == http.MethodPut {
		u.postUser(w, r)
		return
	}
	writeError(w, http.StatusBadRequest, errHTTPMethod)
}

func (u *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ctx := r.Context()
		session := ctx.Value("session").(*dm.Session)
		if err := u.userCase.Logout(ctx, session.ID); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf(errUserUpd.Error(), err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	writeError(w, http.StatusBadRequest, errHTTPMethod)
}
