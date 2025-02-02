package local_storage

import (
	"context"
	"fmt"
	dm "rwa/internal/domain"
	"strconv"
	"sync"
	"time"
)

type user struct {
	ID        string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string
	Bio       string
	Image     string
	Token     string
	Password  []byte
	Following bool
}

type UserStorage struct {
	Users  map[string]*user
	UserID int
	Mu     *sync.RWMutex
}

// NewDBStorage создание хранилища данных
func NewUserStorage() *UserStorage {
	users := make(map[string]*user)
	return &UserStorage{users,
		0,
		&sync.RWMutex{},
	}
}

// CreateUser создание пользователя в хранилище, дополняет переданную модель
func (ls *UserStorage) Create(ctx context.Context, userDM *dm.User) error {
	userDB := &user{Username: userDM.Username,
		Email:     userDM.Email,
		Password:  userDM.Password,
		CreatedAt: userDM.CreatedAt,
		UpdatedAt: userDM.UpdatedAt,
	}

	ls.Mu.Lock()
	defer ls.Mu.Unlock()
	//ключем для поиска и создания будет почта
	if _, ok := ls.Users[userDB.Email]; ok {
		return fmt.Errorf("Пользователь с email %s уже существует", userDB.Email)
	}
	userDB.ID = strconv.Itoa(ls.UserID)
	ls.UserID++
	ls.Users[userDB.Email] = userDB
	userDM.ID = userDB.ID
	return nil
}

//func (ls *UserStorage) Get() {
//
//}
