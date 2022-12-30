package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"trolly.hunterwilkins.dev/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/api/healthcheck", app.healthcheckHandler)

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodPost, "/user/logout", dynamic.ThenFunc(app.userLogoutPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/", protected.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/pantry", protected.ThenFunc(app.pantry))
	router.Handler(http.MethodGet, "/search/:query", protected.ThenFunc(app.search))
	router.Handler(http.MethodPost, "/items/add/home", protected.ThenFunc(app.addHomeItem))
	router.Handler(http.MethodPost, "/items/add/pantry", protected.ThenFunc(app.addPantryItem))
	router.Handler(http.MethodPost, "/items/delete/:id", protected.ThenFunc(app.deleteItem))
	router.Handler(http.MethodGet, "/items/update/:id", protected.ThenFunc(app.updateItem))
	router.Handler(http.MethodPost, "/items/home/:id", protected.ThenFunc(app.updateHomeItems))
	router.Handler(http.MethodPost, "/items/pantry/:id", protected.ThenFunc(app.updatePantryItems))
	router.Handler(http.MethodPost, "/items/update/:id", protected.ThenFunc(app.updateItemForm))

	standard := alice.New(app.logRequest)
	return standard.Then(router)
}
