package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"shopql/graph"
	"shopql/graph/directives"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
)

const defaultPort = "8080"

// GQLHandler основная структура для gql сервера с резолвером
type GQLHandler struct {
	db *graph.Resolver
}

// NewGQLHandler конструктор для GQLHandler, создаем через структуру, чтобы прокидывать resolver в наши ручки
func NewGQLHandler(resolver *graph.Resolver) *GQLHandler {
	return &GQLHandler{resolver}
}

// regHandler хэндлер для регистрации пользователя с добавлением юзера в контекст
func (srv *GQLHandler) regHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		body := graph.Reg{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("ошибка декодирования тела запроса"))
			return
		}
		if _, ok := srv.db.Data.Users[body.User.Email]; !ok {
			srv.db.Data.Users[body.User.Email] = body.User
		}
		token := uuid.NewString()
		srv.db.Data.Session[token] = body.User
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{\"body\":{\"token\":\"%s\"}}", token)))
	} else {
		http.Error(w, "Некорректный метод вызова", http.StatusBadRequest)
	}
}

// getRequestUser функция для получения юзера из токена в контексте
func getRequestUser(resolver *graph.Resolver, r *http.Request) (graph.User, error) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Token ")
	if ok && token != "" {
		if usr, ok := resolver.Data.Session[token]; ok {
			return usr, nil
		}
		return graph.User{}, fmt.Errorf("Для переданного токена не найден пользователь")
	}
	return graph.User{}, fmt.Errorf("В запросе отсутствует токен авторизации")
}

// AuthMiddleware функция добавления юзера в контекст по токену авторизации
func AuthMiddleware(resolver *graph.Resolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := getRequestUser(resolver, r)
		// если токена авторизации нет, то вызываем следующий хэндлер без доп контекста
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		//добавляем в контест информацию о пользователе
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	resolver := graph.NewResolver()
	auth := graph.DirectiveRoot{directives.Authorized}
	cfg := graph.Config{Resolvers: resolver,
		Directives: auth, // добавление функции авторизации для директивы @authorized
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))
	handler := NewGQLHandler(resolver)
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	mux.HandleFunc("/register", handler.regHandler) // ручка авторизации
	muxAuth := AuthMiddleware(resolver, mux)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

	log.Fatal(http.ListenAndServe(":"+port, muxAuth))
}

func GetApp() http.Handler {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	resolver := graph.NewResolver()
	auth := graph.DirectiveRoot{directives.Authorized}
	cfg := graph.Config{Resolvers: resolver,
		Directives: auth,
	}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(cfg))
	h := NewGQLHandler(resolver)
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	mux.HandleFunc("/register", h.regHandler)

	muxAuth := AuthMiddleware(resolver, mux)
	return muxAuth
}
