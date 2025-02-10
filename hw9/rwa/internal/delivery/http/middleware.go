package http

import (
	"context"
	"net/http"
)

var pathNoMiddleware = map[string]map[string]struct{}{
	"/api/users":       map[string]struct{}{}, // все методы
	"/api/users/login": map[string]struct{}{}, // все методы
	"/api/articles":    map[string]struct{}{"GET": struct{}{}},
}

func AuthMiddleware(main *mainHandler, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//тут можно исключить часть путей из обработки
		if methods, ok := pathNoMiddleware[r.URL.String()]; ok {
			if _, ok := methods[r.Method]; ok || len(methods) == 0 {
				next.ServeHTTP(w, r)
				return
			}
		}
		// вызов проверки сессии
		session, err := main.sessionHandler.Get(r)

		if err != nil {
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}
		// добавляем сессию в контекст всех запросов
		ctx := context.WithValue(r.Context(), "session", session)
		r1 := r.WithContext(ctx)
		next.ServeHTTP(w, r1)
	})
}
