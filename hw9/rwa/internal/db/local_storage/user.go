package local_storage

import (
	"context"
	"fmt"
	dm "rwa/internal/domain"
	"sync"
	"time"
)

type User struct {
	ID        string
	Email     string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Bio       string
	Image     string
	Token     string
	Password  []byte
	Following bool
}

type UserStorage struct {
	UsersEmail    map[string]*User // хранение для быстрого поиcка по email
	UsersID       map[string]*User // хранение для быстрого поиcка по ID
	UsersUsername map[string]*User // хранение для быстрого поиcка по username
	Mu            *sync.RWMutex
}

// NewDBStorage создание хранилища данных
func NewUserStorage() *UserStorage {
	return &UserStorage{
		make(map[string]*User),
		make(map[string]*User),
		make(map[string]*User),
		&sync.RWMutex{},
	}
}

// Create создание пользователя в хранилище, дополняет переданную модель
func (ls *UserStorage) Create(ctx context.Context, userDM *dm.User) error {
	userDB := &User{
		ID:        userDM.ID,
		Username:  userDM.Username,
		Email:     userDM.Email,
		Password:  userDM.Password,
		CreatedAt: userDM.CreatedAt,
		UpdatedAt: userDM.UpdatedAt,
	}

	ls.Mu.Lock()
	defer ls.Mu.Unlock()
	//ключем для поиска и создания будет почта
	if _, ok := ls.UsersEmail[userDB.Email]; ok {
		return fmt.Errorf("Пользователь с email %s уже существует", userDB.Email)
	}
	if _, ok := ls.UsersUsername[userDB.Username]; ok {
		return fmt.Errorf("Пользователь с username %s уже существует", userDB.Username)
	}
	ls.UsersEmail[userDB.Email] = userDB
	ls.UsersID[userDB.ID] = userDB
	ls.UsersUsername[userDB.Username] = userDB
	return nil
}

// GetByEmail поиск пользователя по email
func (ls *UserStorage) GetByEmail(ctx context.Context, email string) (*dm.User, error) {
	ls.Mu.RLock()
	defer ls.Mu.RUnlock()
	if user, ok := ls.UsersEmail[email]; ok {
		userDm := &dm.User{
			user.ID,
			user.Email,
			user.Username,
			user.Password,
			user.CreatedAt,
			user.UpdatedAt,
			user.Bio,
		}
		return userDm, nil
	} else {
		return nil, fmt.Errorf("User not found")
	}
}

// GetByID поиск пользователя по ID
func (ls *UserStorage) GetByID(ctx context.Context, ID string) (*dm.User, error) {
	ls.Mu.RLock()
	defer ls.Mu.RUnlock()
	if user, ok := ls.UsersID[ID]; ok {
		userDm := &dm.User{
			user.ID,
			user.Email,
			user.Username,
			user.Password,
			user.CreatedAt,
			user.UpdatedAt,
			user.Bio,
		}
		return userDm, nil
	} else {
		return nil, fmt.Errorf("User not found")
	}
}

// GetByUsername поиск пользователя по username
func (ls *UserStorage) GetByUsername(ctx context.Context, username string) (*dm.User, error) {
	ls.Mu.RLock()
	defer ls.Mu.RUnlock()
	if user, ok := ls.UsersUsername[username]; ok {
		userDm := &dm.User{
			user.ID,
			user.Email,
			user.Username,
			user.Password,
			user.CreatedAt,
			user.UpdatedAt,
			user.Bio,
		}
		return userDm, nil
	} else {
		return nil, fmt.Errorf("User not found")
	}
}

// Update обновление данных пользователя
func (ls *UserStorage) Update(ctx context.Context, userNew *dm.User) (*dm.User, error) {
	ls.Mu.RLock()
	defer ls.Mu.RUnlock()
	if user, ok := ls.UsersID[userNew.ID]; ok {
		user.UpdatedAt = time.Now()
		user.Bio = userNew.BIO
		if user.Email != userNew.Email { // если изменили email, то нужно удалить в мапу по email запись и добавить новую
			delete(ls.UsersEmail, user.Email) // если изменили почту, то удаляем старую запись
		}
		user.Email = userNew.Email
		ls.UsersID[userNew.ID] = user
		ls.UsersEmail[userNew.Email] = user
		return &dm.User{
			user.ID,
			user.Email,
			user.Username,
			[]byte{},
			user.CreatedAt,
			user.UpdatedAt,
			user.Bio}, nil
	} else {
		return nil, fmt.Errorf("User for update not found")
	}
}
