package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/flow"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hunterwilkins2/trolly/internal/service"
)

type application struct {
	items service.Item
	users service.User

	logger *slog.Logger
}

func main() {
	port := flag.Int("port", 4000, "Port to serve server on")
	hotReload := flag.Bool("hot-reload", false, "Hot-reload web browser on save")
	dbHost := flag.String("db-host", "127.0.0.1:3306", "MySQL hostname")
	dbUser := flag.String("db-user", "root", "MySQL username")
	dbPass := flag.String("db-pass", "admin", "MySQL password")
	dbName := flag.String("db-name", "trolly", "MySQL database name")
	flag.Parse()

	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(logHandler)

	db, err := openDb(*dbHost, *dbUser, *dbPass, *dbName)
	if err != nil {
		logger.Error("could not create connection to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	app := &application{
		logger: logger,
	}

	mux := flow.New()
	if *hotReload {
		mux.HandleFunc("/hot-reload", HotReload)
		mux.HandleFunc("/hot-reload/ready", Ready, http.MethodGet)
	}

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/...", http.StripPrefix("/static/", fs))

	mux.Use(UseHotReload(*hotReload))
	mux.HandleFunc("/", app.GroceryListPage, http.MethodGet)
	mux.HandleFunc("/pantry", app.PantryPage, http.MethodGet)
	mux.HandleFunc("/signup", app.RegisterPage, http.MethodGet)
	mux.HandleFunc("/login", app.LoginPage, http.MethodGet)

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
	// if err = db.Ping(); err != nil {
	// 	return nil, err
	// }
	return db, nil
}
