package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"shortik/internal/core/app"
	"shortik/internal/infra/api/rest"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	app := app.NewApp()
	srv := rest.NewServer(rest.ServerConfig{}, app)

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server has failed: %w", err)
		}
	}
	return nil
}
