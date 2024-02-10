package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/hunterwilkins2/trolly/components"
	"github.com/hunterwilkins2/trolly/components/pages"
	"github.com/hunterwilkins2/trolly/internal/models"
	"github.com/hunterwilkins2/trolly/internal/service"
	"github.com/hunterwilkins2/trolly/internal/validator"
)

func (app *application) GroceryListPage(w http.ResponseWriter, r *http.Request) {
	pages.GroceryList().Render(r.Context(), w)
}

func (app *application) PantryPage(w http.ResponseWriter, r *http.Request) {
	pages.Pantry().Render(r.Context(), w)
}

func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.Login(nil, nil).Render(r.Context(), w)
}

func (app *application) RegisterPage(w http.ResponseWriter, r *http.Request) {
	pages.Register(nil, nil).Render(r.Context(), w)
}

func (app *application) ValidateName(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	v := validator.New()
	models.ValidateName(v, name)
	if v.HasErrors() {
		w.Write([]byte(v.GetError("name").Error()))
		return
	}

	w.Write([]byte(""))
}

func (app *application) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	v := validator.New()
	models.ValidateEmail(v, email)
	if v.HasErrors() {
		w.Write([]byte(v.GetError("email").Error()))
		return
	}

	w.Write([]byte(""))
}

func (app *application) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	v := validator.New()
	models.ValidatePassword(v, password)
	if v.HasErrors() {
		w.Write([]byte(v.GetError("password").Error()))
		return
	}

	w.Write([]byte(""))
}

func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := app.users.Register(r.Context(), name, email, password)
	if err != nil {
		app.logger.Error("Failed to create user", "error", err.Error(), "name", name, "email", email)
		var ee map[string]error
		var v *validator.Validator
		if errors.As(err, &v) {
			ee = v.FieldErrors
		} else if err == models.ErrDuplicateEmail {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "User with that email already exists"))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not create account. Please try again."))
		}

		pages.Register(map[string]string{"name": name, "email": email}, ee).Render(r.Context(), w)
		return
	}
	app.logger.Info("created new user", "name", user.Name, "email", email)

	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), "userId", user.ID)
	app.sessionManager.Put(r.Context(), "userName", user.Name)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := app.users.Login(r.Context(), email, password)
	if err != nil {
		app.logger.Error("Failed to log user in", "error", err.Error(), "email", email)
		var ee map[string]error
		var v *validator.Validator
		if errors.As(err, &v) {
			ee = v.FieldErrors
		} else if err == service.ErrInvalidCredentials {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Email or password is incorrect"))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not log in. Please try again."))
		}
		pages.Login(map[string]string{"email": email}, ee).Render(r.Context(), w)
		return
	}

	app.logger.Info("login from", "name", user.Name, "email", email)

	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), "userId", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Remove(r.Context(), "userId")

	w.Header().Add("HX-Redirect", "/login")
}
