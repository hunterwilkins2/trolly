package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"trolly.hunterwilkins.dev/internal/models"
)

type templateData struct {
	CurrentYear     int
	Items           []*models.Item
	ItemID          int
	Total           float32
	Form            any
	Flash           string
	IsAuthenticated bool
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func formatMoney(m float32) string {
	return fmt.Sprintf("$%.2f", m)
}

var functions = template.FuncMap{
	"humanData":   humanDate,
	"formatMoney": formatMoney,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
