package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"shortik/internal/core/app/model"
	coreModel "shortik/internal/core/model"
	randgenModel "shortik/internal/core/service/randgen/model"
	dbModel "shortik/internal/infra/store/db/model"
)

type RandGen interface {
	GenerateRandomBytes(
		req randgenModel.GenerateRandomBytesRequest,
	) (randgenModel.GenerateRandomBytesResponse, error)
}

type DB interface {
	StoreURL(ctx context.Context, req dbModel.StoreURLRequest) (dbModel.StoreURLResponse, error)
	GetURL(ctx context.Context, req dbModel.GetURLRequest) (dbModel.GetURLResponse, error)
}

type App struct {
	randGen RandGen
	db      DB

	params ConfigParams
}

type Config struct {
	RandGen RandGen
	DB      DB
	ConfigParams
}

type ConfigParams struct {
	SlugsAlphabet   string `yaml:"slugsAlphabet" validate:"required,alphanum"`
	SlugsMinLen     int    `yaml:"slugsMinLen" validate:"required,gt=0"`
	SlugsMaxLen     int    `yaml:"slugsMaxLen" validate:"required,gtefield=SlugsMinLen"`
	SlugsBatchCount int    `yaml:"slugsBatchCount" validate:"required,gt=0"`
}

func GetDefaultConfigParams() ConfigParams {
	return ConfigParams{
		SlugsAlphabet:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		SlugsMinLen:     6,
		SlugsMaxLen:     20,
		SlugsBatchCount: 5,
	}
}

func NewApp(cfg *Config) *App {
	return &App{
		randGen: cfg.RandGen,
		db:      cfg.DB,

		params: cfg.ConfigParams,
	}
}

func newURLNotValidError(u coreModel.URL, err error) error {
	return fmt.Errorf("problem with URL %s: %w: %w", string(u), model.ErrURLNotValid, err)
}

func (a *App) ShortenURL(ctx context.Context, req model.ShortenURLRequest) (model.ShortenURLResponse, error) {
	var resp model.ShortenURLResponse

	if err := validateURL(req.URL); err != nil {
		return resp, newURLNotValidError(req.URL, err)
	}

	var shortened bool
TryStoreLoop:
	for i := a.params.SlugsMinLen; i <= a.params.SlugsMaxLen; i++ {
		randGenResp, err := a.randGen.GenerateRandomBytes(randgenModel.GenerateRandomBytesRequest{
			BufsCount: a.params.SlugsBatchCount,
			Alphabet:  []byte(a.params.SlugsAlphabet),
			Len:       i,
		})
		if err != nil {
			return resp, fmt.Errorf("failed to generate a URL slug: %w", err)
		}
		for _, sb := range randGenResp.Bufs {
			slug, err := validateSlug(sb)
			if err != nil {
				return resp, err
			}
			storeURLRes, err := a.db.StoreURL(ctx, dbModel.StoreURLRequest{
				URL:  req.URL,
				Slug: coreModel.Slug(slug),
			})
			if err != nil {
				if errors.Is(err, dbModel.ErrSlugAlreadyExists) {
					continue
				}
				return resp, fmt.Errorf("failed to save the URL: %w", err)
			}
			resp.URL = storeURLRes.URL
			resp.Slug = storeURLRes.Slug
			shortened = true
			break TryStoreLoop
		}
	}
	if !shortened {
		return resp, errors.New("failed to generate a unique slug")
	}
	return resp, nil
}

func validateURL(u coreModel.URL) error {
	rawURL := string(u)
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	if len(parsedURL.Scheme) == 0 {
		return errors.New("URL does not contain a scheme")
	}
	if len(parsedURL.Host) == 0 {
		return errors.New("URL does not contain a host")
	}
	return nil
}

func validateSlug(slug []byte) (string, error) {
	s := string(slug)
	if url.PathEscape(s) == s {
		return s, nil
	}
	return "", fmt.Errorf("string %s cannot be used as a slug", s)
}

func newURLNotFoundErr() error {
	return fmt.Errorf("failed to get a URL from store: %w", model.ErrURLNotFound)
}

func (a *App) GetFullURL(ctx context.Context, req model.GetFullURLRequest) (model.GetFullURLResponse, error) {
	var resp model.GetFullURLResponse
	getURLRes, err := a.db.GetURL(ctx, dbModel.GetURLRequest{
		Slug: req.Slug,
	})
	if err != nil {
		if errors.Is(err, dbModel.ErrSlugNotFound) {
			return resp, newURLNotFoundErr()
		}
		return resp, fmt.Errorf("failed to get a URL from store: %w", err)
	}
	resp.URL = string(getURLRes.FullURL)
	return resp, nil
}
