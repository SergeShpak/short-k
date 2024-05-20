package model

import (
	"errors"
	core "shortik/internal/core/model"
)

type ShortenURLRequest struct {
	URL core.URL
}

type ShortenURLResponse struct {
	URL  core.URL
	Slug core.Slug
}

type GetFullURLRequest struct {
	Slug core.Slug
}

type GetFullURLResponse struct {
	URL string
}

var (
	ErrURLNotValid = errors.New("URL not valid")
	ErrURLNotFound = errors.New("URL not found")
)
