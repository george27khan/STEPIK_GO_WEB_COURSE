package user

import (
	"context"
	"fmt"
	"golang.org/x/crypto/argon2"
	"math/rand"
	dm "rwa/internal/domain"
	"rwa/internal/dto"
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
	Get(ctx context.Context, email string) (*dm.User, error)
}

type UserUseCase struct {
	db UserRepository
}

func NewUserUseCase(db UserRepository) *UserUseCase {
	return &UserUseCase{db}
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

// Register бизнес логика регистрации пользователя
func (uc *UserUseCase) Register(ctx context.Context, userCreate *dto.UserCreate) (*dto.UserCreateResp, error) {
	//бизнес логика
	user := &dm.User{
		ID:        "",
		Email:     userCreate.Info.Email,
		Username:  userCreate.Info.Username,
		Password:  uc.hashPassword(userCreate.Info.Password, randStringRune(saltLen)), //шифрование пароля
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// отправка в хранилище
	if err := uc.db.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("Ошибка при сохранении пользователя в БД: %s", err.Error())
	}

	// инициализируем поля присланной структуры для ответа

	return &dto.UserCreateResp{dto.InfoCreateResp{
		user.Email,
		user.Username,
		user.CreatedAt,
		user.UpdatedAt}}, nil
}

func (uc *UserUseCase) Login(ctx context.Context, userCreate *dto.UserCreate) (*dto.UserCreateResp, error) {
	userDB, err := uc.db.Get(ctx, userCreate.Info.Email)
	if err != nil {
		return nil, err
	}
	if !uc.passwordIsValid(userCreate.Info.Password, userDB.Password) {
		return nil, fmt.Errorf("Invalid password for user %s", userCreate.Info.Email)
	}
	userResp := &dto.UserCreateResp{
		dto.InfoCreateResp{
			userDB.Email,
			userDB.Username,
			userDB.CreatedAt,
			userDB.UpdatedAt,
		}}
	return userResp, nil

}
