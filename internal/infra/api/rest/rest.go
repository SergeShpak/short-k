package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	appModel "shortik/internal/core/app/model"
	"shortik/internal/core/model"
)

type App interface {
	ShortenURL(ctx context.Context, req appModel.ShortenURLRequest) (appModel.ShortenURLResponse, error)
	GetFullURL(ctx context.Context, req appModel.GetFullURLRequest) (appModel.GetFullURLResponse, error)
}

func NewServer(cfg *ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Host,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ErrorLog:          cfg.ErrorLog,
		Handler:           newRouter(cfg.Handler),
	}
}

func newRouter(cfg HandlerConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := newHandler(cfg)
	r.Route("/v1", func(r chi.Router) {
		r.Post("/", h.shortenURL)
		r.Get("/{slug}", h.getURL)
	})

	return r
}

type handler struct {
	cfg HandlerConfig
}

func newHandler(cfg HandlerConfig) *handler {
	return &handler{
		cfg: cfg,
	}
}

type shortenURLRequest struct {
	URL string `json:"url"`
}

type shortenURLResponse struct {
	URL          string `json:"url"`
	ShortenedURL string `json:"shortened_url"`
}

const (
	slogErrName = "err"
)

func (h *handler) shortenURL(w http.ResponseWriter, r *http.Request) {
	limitedReader := &io.LimitedReader{R: r.Body, N: h.cfg.MaxRequestBodySize + 1}
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		h.cfg.Logger.ErrorContext(r.Context(), "failed to read client's request", slog.Any(slogErrName, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(data) > int(h.cfg.MaxRequestBodySize) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req shortenURLRequest
	if err := json.Unmarshal(data, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.cfg.App.ShortenURL(r.Context(), appModel.ShortenURLRequest{
		URL: model.URL(req.URL),
	})
	if err != nil {
		if errors.Is(err, appModel.ErrURLNotValid) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.cfg.Logger.ErrorContext(
			r.Context(),
			"failed to shorten URL",
			slog.String("url", req.URL),
			slog.Any(slogErrName, err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortenedURL, err := url.JoinPath(h.cfg.BaseAddr, string(res.Slug))
	if err != nil {
		h.cfg.Logger.ErrorContext(r.Context(), "failed to compose the shortened URL", slog.Any(slogErrName, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := shortenURLResponse{
		URL:          req.URL,
		ShortenedURL: shortenedURL,
	}

	respBody, err := json.Marshal(resp)
	if err != nil {
		h.cfg.Logger.ErrorContext(r.Context(), "failed to marshal the shorten URL response", slog.Any(slogErrName, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(respBody); err != nil {
		h.cfg.Logger.ErrorContext(r.Context(), "failed to write the shorten URL response body", slog.Any(slogErrName, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) getURL(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	resp, err := h.cfg.App.GetFullURL(r.Context(), appModel.GetFullURLRequest{
		Slug: model.Slug(slug),
	})
	if err != nil {
		if errors.Is(err, appModel.ErrURLNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.cfg.Logger.ErrorContext(r.Context(), "failed to get URL from slug", slog.Any(slogErrName, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, resp.URL, http.StatusTemporaryRedirect)
}
