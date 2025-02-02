package main

import (
	"net/http"
	"rwa/internal/db"
	"rwa/internal/delivery"
)

func main() {
	dbLocal := db.NewDBStorage()
	mux := delivery.NewServerMUX()
	http.ListenAndServe(":8080", mux)
}
