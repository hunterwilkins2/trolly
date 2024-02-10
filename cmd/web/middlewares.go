package main

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hunterwilkins2/trolly/components"
)

func (app *application) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		dur := time.Since(start)
		app.logger.Info(r.Method+" "+r.URL.Path, "remote", r.RemoteAddr, "duration", dur)
	})
}

func (app *application) Authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := app.sessionManager.Get(r.Context(), "userId").(uuid.UUID)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		user, err := app.users.GetUser(r.Context(), id)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), components.UserKey, user.Name))
		next.ServeHTTP(w, r)
	})
}

func (app *application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)
				app.logger.Error("recovered panic", "error", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
