package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"trolly.hunterwilkins.dev/internal/models"
	"trolly.hunterwilkins.dev/ui"
)

type templateData struct {
	CurrentYear     int
	Items           []*models.Item
	ItemID          int
	Total           float32
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
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

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		patterns := []string{
			"html/base.html",
			"html/partials/*html",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
