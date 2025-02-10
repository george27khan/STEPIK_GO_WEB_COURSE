package user

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"math/rand"
	dm "rwa/internal/domain"
	"rwa/internal/dto"
	"rwa/internal/usecase/session"
	"time"
)

const (
	saltLen = 8
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type UserRepository interface {
	Create(ctx context.Context, user *dm.User) error
	GetByEmail(ctx context.Context, email string) (*dm.User, error)
	GetByID(ctx context.Context, ID string) (*dm.User, error)
	GetByUsername(ctx context.Context, username string) (*dm.User, error)
	Update(ctx context.Context, user *dm.User) (*dm.User, error)
}

type UserUseCase struct {
	DBUser         UserRepository
	SessionUseCase *session.SessionUseCase
}

func NewUserUseCase(dbUser UserRepository, SessionUseCase *session.SessionUseCase) *UserUseCase {
	return &UserUseCase{dbUser, SessionUseCase}
}

// randStringRune генерация случайно строки заданной длинны в байтах
func randStringRune(bytes int) string {
	randRunes := make([]rune, bytes)
	for i := range randRunes {
		randRunes[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(randRunes)
}

// hashPassword хэширование пароля с солью для безопасного хранения
func (uc *UserUseCase) hashPassword(password string, salt string) []byte {
	hashPwd := argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)
	return append([]byte(salt), hashPwd...)
}

// passwordIsValid валидация пароля при входе
func (uc *UserUseCase) passwordIsValid(password string, passwordDB []byte) bool {
	return true
}

// Register регистрация пользователя
func (uc *UserUseCase) Register(ctx context.Context, UserReq *dto.UserReq) (*dm.User, error) {
	//бизнес логика
	user := &dm.User{
		ID:        uuid.NewString(),
		Email:     UserReq.Info.Email,
		Username:  UserReq.Info.Username,
		Password:  uc.hashPassword(UserReq.Info.Password, randStringRune(saltLen)), //шифрование пароля
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// отправка в хранилище
	if err := uc.DBUser.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("Ошибка при сохранении пользователя в БД: %s", err.Error())
	}
	return user, nil
}

// Login аутентификация пользователя
func (uc *UserUseCase) Login(ctx context.Context, userCreate *dto.UserReq) (*dm.User, error) {
	userDB, err := uc.DBUser.GetByEmail(ctx, userCreate.Info.Email)
	if err != nil {
		return nil, err
	}
	if !uc.passwordIsValid(userCreate.Info.Password, userDB.Password) {
		return nil, fmt.Errorf("Invalid password for user %s", userCreate.Info.Email)
	}
	return userDB, nil
}

// Logout выход пользователя из системы
func (uc *UserUseCase) Logout(ctx context.Context, sessionID string) error {
	if err := uc.SessionUseCase.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("Ошибка при удалении сесии: %s", err.Error())
	}
	return nil
}

// Get получение пользователя по ID
func (uc *UserUseCase) Get(ctx context.Context, ID string) (*dm.User, error) {
	userDB, err := uc.DBUser.GetByID(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("Get user error: %s", err.Error())
	}
	return userDB, nil
}

// Update обновление данных пользователя
func (uc *UserUseCase) Update(ctx context.Context, userNew *dto.UserReqInfo) (*dm.User, error) {
	userDM := &dm.User{
		userNew.ID,
		userNew.Email,
		userNew.Username,
		[]byte{},
		userNew.CreatedAt,
		userNew.UpdatedAt,
		userNew.BIO}
	userDM, err := uc.DBUser.Update(ctx, userDM)
	if err != nil {
		return nil, fmt.Errorf("Update user error: %s", err.Error())
	}
	return userDM, nil
}
