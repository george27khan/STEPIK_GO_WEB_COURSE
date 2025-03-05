package directives

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
)

// Authorized проверяет, есть ли пользователь в контексте для директивы @authorized в schema.graphqls
func Authorized(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
	user := ctx.Value("user") // Достаем пользователя из контекста
	if user == nil {
		return nil, errors.New("User not authorized")
	}
	return next(ctx) // Если авторизован, выполняем запрос дальше
}
