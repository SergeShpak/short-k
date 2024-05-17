package rest

import (
	"context"
	"net/http"

	"shortik/internal/core/model"
)

type App interface {
	ShortenURL(ctx context.Context, req model.ShortenURLRequest) (model.ShortenURLResponse, error)
	GetFullURL(ctx context.Context, req model.GetFullURLRequest) (model.GetFullURLResponse, error)
}

func NewServer(cfg ServerConfig, svc App) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ErrorLog:          cfg.ErrorLog,
		Handler:           newHandler(svc),
	}
}

func newHandler(svc App) http.Handler {
	panic("NYI")
}
