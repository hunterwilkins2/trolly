package main

import (
	"context"
	"database/sql"
	"encoding/gob"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/flow"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/hunterwilkins2/trolly/internal/models"
	"github.com/hunterwilkins2/trolly/internal/service"
)

type application struct {
	items  *service.ItemService
	users  *service.UserService
	basket *service.BasketService

	sessionManager *scs.SessionManager
	logger         *slog.Logger
}

func main() {
	gob.Register(uuid.New())
	port := flag.Int("port", 4000, "Port to serve server on")
	hotReload := flag.Bool("hot-reload", false, "Hot-reload web browser on save")
	dbHost := flag.String("db-host", "0.0.0.0:3306", "MySQL hostname")
	dbUser := flag.String("db-user", "trolly", "MySQL username")
	dbPass := flag.String("db-pass", "pa55word", "MySQL password")
	dbName := flag.String("db-name", "trolly", "MySQL database name")
	flag.Parse()

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(logHandler)

	db, err := openDb(*dbHost, *dbUser, *dbPass, *dbName)
	if err != nil {
		logger.Error("could not create connection to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)

	userRepo := models.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	itemRepo := models.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo)

	basketRepo := models.NewBasketRepository(db)
	basketService := service.NewBasketService(basketRepo)

	app := &application{
		users:          userService,
		items:          itemService,
		basket:         basketService,
		sessionManager: sessionManager,
		logger:         logger,
	}

	mux := flow.New()
	if *hotReload {
		mux.HandleFunc("/hot-reload", HotReload)
		mux.HandleFunc("/hot-reload/ready", Ready, http.MethodGet)
	}

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/...", http.StripPrefix("/static/", fs))

	mux.Use(app.RecoverPanic, UseHotReload(*hotReload), app.LogRequest, sessionManager.LoadAndSave)
	mux.HandleFunc("/signup", app.RegisterPage, http.MethodGet)
	mux.HandleFunc("/register", app.Register, http.MethodPost)
	mux.HandleFunc("/user/validate/name", app.ValidateName, http.MethodPost)
	mux.HandleFunc("/user/validate/email", app.ValidateEmail, http.MethodPost)
	mux.HandleFunc("/user/validate/password", app.ValidatePassword, http.MethodPost)

	mux.HandleFunc("/login", app.LoginPage, http.MethodGet)
	mux.HandleFunc("/login", app.Login, http.MethodPost)
	mux.HandleFunc("/logout", app.Logout, http.MethodPost)

	mux.Group(func(m *flow.Mux) {
		mux.Use(app.Authenticated)
		mux.HandleFunc("/", app.GroceryListPage, http.MethodGet)
		mux.HandleFunc("/pantry", app.PantryPage, http.MethodGet)

		mux.HandleFunc("/search", app.Search, http.MethodPost)
		mux.HandleFunc("/items", app.AddItem, http.MethodPost)
		mux.HandleFunc("/items/:id", app.DeleteItem, http.MethodDelete)
		mux.HandleFunc("/items/edit", app.EditItemPage, http.MethodGet)
		mux.HandleFunc("/items/:id", app.EditItem, http.MethodPatch)

		mux.HandleFunc("/suggestions", app.Suggest, http.MethodGet)
		mux.HandleFunc("/basket", app.CreateNewItemAndAddToBasket, http.MethodPost)
		mux.HandleFunc("/basket/:itemId", app.AddItemToBasket, http.MethodPost)
		mux.HandleFunc("/basket/:id", app.MarkPurchased, http.MethodPatch)
		mux.HandleFunc("/basket/:id", app.RemoveItemFromBasket, http.MethodDelete)
		mux.HandleFunc("/basket", app.RemoveAllItems, http.MethodDelete)
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", *port),
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           mux,
		ErrorLog:          slog.NewLogLogger(logHandler, slog.LevelError),
	}

	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		logger.Info("shutting down server", "signal", sig.String())

		timeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(timeout)
	}()

	logger.Info("starting server", "addr", "http://"+srv.Addr)
	err = srv.ListenAndServe()
	if err != nil {
		logger.Error("uncaught error occurred", "error", err)
		os.Exit(1)
	}

	if err = <-shutdownErr; err != nil {
		logger.Error("error shutting down server", "error", err)
		os.Exit(1)
	}

	logger.Info("stopped server")
}

func openDb(dbHost, dbUser, dbPass, dbName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbName))
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
