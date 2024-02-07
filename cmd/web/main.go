package main

import (
	"net/http"

	"github.com/alexedwards/flow"
	"github.com/hunterwilkins2/trolly/internal/service"
)

type application struct {
	items service.Item
	users service.User
}

func main() {
	app := &application{}

	mux := flow.New()

	mux.HandleFunc("/", app.Register, http.MethodGet)
	mux.HandleFunc("/pantry", app.Register, http.MethodGet)
	mux.HandleFunc("/signup", app.Register, http.MethodGet)
	mux.HandleFunc("/login", app.Register, http.MethodGet)
}
