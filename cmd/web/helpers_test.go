package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"trolly.hunterwilkins.dev/internal/assert"
)

func TestNotFound(t *testing.T) {
	app := newTestApplication(t)

	recorder := httptest.NewRecorder()
	app.notFound(recorder)

	assert.Equal(t, recorder.Code, 404)
	assert.Equal(t, strings.TrimSpace(recorder.Body.String()), http.StatusText(http.StatusNotFound))
}

func TestServerError(t *testing.T) {
	app := newTestApplication(t)

	err := errors.New("Test error")

	recorder := httptest.NewRecorder()
	app.serverError(recorder, err)

	assert.Equal(t, recorder.Code, http.StatusInternalServerError)
	assert.Equal(t, strings.TrimSpace(recorder.Body.String()), http.StatusText(http.StatusInternalServerError))
}

func TestClientError(t *testing.T) {
	testCases := []struct {
		name     string
		err      int
		wantCode int
		wantBody string
	}{
		{
			name:     "Bad Request Error",
			err:      http.StatusBadRequest,
			wantCode: http.StatusBadRequest,
			wantBody: http.StatusText(http.StatusBadRequest),
		},
		{
			name:     "Unprocessable Entity Error",
			err:      http.StatusUnprocessableEntity,
			wantCode: http.StatusUnprocessableEntity,
			wantBody: http.StatusText(http.StatusUnprocessableEntity),
		},
	}
	app := newTestApplication(t)
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			app.clientError(recorder, tC.err)

			assert.Equal(t, recorder.Code, tC.wantCode)
			assert.Equal(t, strings.TrimSpace(recorder.Body.String()), tC.wantBody)
		})
	}
}

func TestRender(t *testing.T) {
	app := newTestApplication(t)
	r, err := http.NewRequest(http.MethodGet, "/user/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := app.sessionManager.Load(r.Context(), "flash")
	if err != nil {
		t.Fatal(err)
	}
	r = r.WithContext(ctx)

	data := app.newTemplateData(r)
	data.Form = userLoginForm{}

	testCases := []struct {
		name     string
		code     int
		page     string
		data     *templateData
		wantCode int
		wantBody string
	}{
		{
			name:     "Page not found",
			code:     http.StatusOK,
			page:     "doesntexist.html",
			data:     &templateData{},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "Error when template data is empty",
			code:     http.StatusOK,
			page:     "login.html",
			data:     &templateData{},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "Renders page",
			code:     http.StatusOK,
			page:     "login.html",
			data:     data,
			wantCode: http.StatusOK,
			wantBody: fmt.Sprintf("in %d", time.Now().Year()),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			app.render(recorder, tC.code, tC.page, tC.data)

			assert.Equal(t, recorder.Code, tC.wantCode)

			body := recorder.Body.String()

			if body != "" {
				assert.StringContains(t, body, tC.wantBody)
			}
		})
	}
}

func TestDecodePostForm(t *testing.T) {
	app := newTestApplication(t)

	t.Run("Parse form error", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodPost, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		err = app.decodePostForm(r, nil)
		assert.Equal(t, err.Error(), "missing form body")
	})

	t.Run("Form decode error", func(t *testing.T) {
		data := url.Values{}
		data.Set("email", "john.doe@gmail.com")
		data.Set("password", "pass")

		r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := recover(); err == nil {
				t.Error("Expected to panic with nil interface")
			}
		}()
		app.decodePostForm(r, nil)
	})
	t.Run("Decode form", func(t *testing.T) {
		data := url.Values{}
		data.Set("email", "john.doe@gmail.com")
		data.Set("password", "pass")

		r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(data.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		if err != nil {
			t.Fatal(err)
		}

		var form userLoginForm
		err = app.decodePostForm(r, &form)
		assert.NilError(t, err)
	})
}

func TestNewTemplateData(t *testing.T) {
	app := newTestApplication(t)

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx, err := app.sessionManager.Load(r.Context(), "flash")
	if err != nil {
		t.Fatal(err)
	}
	r = r.WithContext(ctx)
	app.sessionManager.Put(r.Context(), "flash", "This is my message")
	data := app.newTemplateData(r)

	assert.Equal(t, data.CurrentYear, time.Now().Year())
	assert.Equal(t, data.Flash, "This is my message")
}

func TestIsAuthenticated(t *testing.T) {
	app := newTestApplication(t)
	t.Run("No IsAuthenticatedKey", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		result := app.isAuthenticated(r)

		assert.Equal(t, result, false)
	})
	t.Run("IsAuthenticatedKey = false", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, false)
		r = r.WithContext(ctx)

		result := app.isAuthenticated(r)

		assert.Equal(t, result, false)
	})
	t.Run("IsAuthenticatedKey = true", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
		r = r.WithContext(ctx)

		result := app.isAuthenticated(r)

		assert.Equal(t, result, true)
	})
}
