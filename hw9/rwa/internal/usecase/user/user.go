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

type Repository interface {
	Register(ctx context.Context, user *dm.User) error
}

type UserUseCase struct {
	db Repository
}

func NewUserUseCase(db Repository) *UserUseCase {
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
func (uc *UserUseCase) passwordIsValid(password string) bool {
	return true
}

// Register бизнес логика регистрации пользователя
func (uc *UserUseCase) Register(ctx context.Context, userCreate *dto.UserCreate) error {
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
	if err := uc.db.Register(ctx, user); err != nil {
		return fmt.Errorf("Ошибка при сохранении пользователя в БД: %s", err.Error())
	}

	// инициализируем поля присланной структуры для ответа
	userCreate.Info.CreatedAt = user.CreatedAt
	userCreate.Info.UpdatedAt = user.UpdatedAt

	return nil
}
