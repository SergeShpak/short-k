package model

import (
	"errors"
	"shortik/internal/core/model"
)

type StoreURLRequest struct {
	URL  model.URL
	Slug model.Slug
}

type StoreURLResponse struct {
	URL               model.URL
	Slug              model.Slug
	IsNewSlugInserted bool
}

type GetURLRequest struct {
	Slug model.Slug
}

type GetURLResponse struct {
	FullURL model.URL
}

var (
	ErrSlugAlreadyExists = errors.New("slug already exists")
	ErrSlugNotFound      = errors.New("slug not found")
)
