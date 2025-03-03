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
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
)

const defaultPort = "8080"

type GQLHandler struct {
	db *graph.Resolver
}

func NewGQLHandler(resolver *graph.Resolver) *GQLHandler {
	return &GQLHandler{resolver}

}
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Некорректный метод вызова"))
	}
}

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

func AuthMiddleware(resolver *graph.Resolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := getRequestUser(resolver, r)
		//if err == nil {
		//	w.WriteHeader(http.StatusUnauthorized)
		//	w.Write([]byte(fmt.Sprintf("Ошибка авторизации: %s", err.Error())))
		//}
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
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	handler := NewGQLHandler(resolver)
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	mux.HandleFunc("/register", handler.regHandler)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func GetApp() http.Handler {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	resolver := graph.NewResolver()
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	handler := NewGQLHandler(resolver)
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)
	mux.HandleFunc("/register", handler.regHandler)

	muxAuth := AuthMiddleware(resolver, mux)
	return muxAuth
}
