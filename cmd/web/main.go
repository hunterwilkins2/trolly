package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"

	"trolly.hunterwilkins.dev/internal/models"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	dsn  string
}

type application struct {
	config         config
	formDecoder    *form.Decoder
	items          models.ItemModelInterface
	logger         *log.Logger
	sessionManager *scs.SessionManager
	templateCache  map[string]*template.Template
	users          models.UserModelInterface
	etag           string
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Trolly server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.dsn, "dsn", "web:pass@/trolly?parseTime=true", "MySQL data source name")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Lshortfile)

	db, err := openDB(cfg.dsn)
	if err != nil {
		logger.Fatal(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 24 * time.Hour

	app := &application{
		config:         cfg,
		formDecoder:    formDecoder,
		logger:         logger,
		items:          &models.ItemModel{DB: db},
		sessionManager: sessionManager,
		templateCache:  templateCache,
		users:          &models.UserModel{DB: db},
		etag:           time.Now().UTC().String(),
	}

	err = app.serve()
	logger.Fatal(err)
}

func openDB(dns string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dns)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
