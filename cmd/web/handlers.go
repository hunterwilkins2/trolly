package main

import (
	"net/http"

	"github.com/hunterwilkins2/trolly/components/pages"
)

func (app *application) GroceryListPage(w http.ResponseWriter, r *http.Request) {
	pages.GroceryList().Render(r.Context(), w)
}

func (app *application) PantryPage(w http.ResponseWriter, r *http.Request) {
	pages.Pantry().Render(r.Context(), w)
}

func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.Login().Render(r.Context(), w)
}

func (app *application) RegisterPage(w http.ResponseWriter, r *http.Request) {
	pages.Register().Render(r.Context(), w)
}
