package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func (app *application) serve() error {
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist("hunterwilkins.dev"),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	app.logger.Printf("starting %s server on port :%d", app.config.env, app.config.port)
	go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	err := srv.ListenAndServeTLS("", "")
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
