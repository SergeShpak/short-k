package app

import (
	"context"
	"shortik/internal/core/model"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) ShortenURL(ctx context.Context, req model.ShortenURLRequest) (model.ShortenURLResponse, error) {
	panic("NYI")
}

func (a *App) GetFullURL(ctx context.Context, req model.GetFullURLRequest) (model.GetFullURLResponse, error) {
	panic("NYI")
}
