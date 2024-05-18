package db

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	coreModel "shortik/internal/core/model"
	"shortik/internal/infra/store/db/internal/queries"
	"shortik/internal/infra/store/db/model"
)

type handler interface {
	GetURL(ctx context.Context, slug string) (string, error)
	InsertURL(ctx context.Context, arg queries.InsertURLParams) (queries.InsertURLRow, error)
}

// DB is the handler to a SQL database.
type DB struct {
	pool    *pgxpool.Pool
	handler handler
}

// ConfigParams is the set of parameters for DB passed during the service initialization.
type ConfigParams struct {
	DSN string `yaml:"-" validate:"required"`
}

// GetDefaultConfigParams return default configuraiton parameters.
func GetDefaultConfigParams() ConfigParams {
	return ConfigParams{}
}

// NewDB initializes a new handler to a SQL database.
// It runs the migrations and pings the DB using a provided connection string.
func NewDB(ctx context.Context, cfg ConfigParams) (*DB, error) {
	if err := runMigrations(cfg.DSN); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create the pool connection config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	h := queries.New(pool)
	return &DB{
		pool:    pool,
		handler: h,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func getProblemWithSlugMsg(slug string) string {
	return "problem with slug " + slug
}

func newErrSlugAlreadyExists(slug string) error {
	return fmt.Errorf("%s: %w", getProblemWithSlugMsg(slug), model.ErrSlugAlreadyExists)
}

// StoreURL stores a full URL and a slug associated with it in the DB.
// If a slug already exists it returns model.ErrSlugAlreadyExists.
// If a URL already exists it returns the slug associated with it.
// Otherwise, it returns the passed full URL and slug.
func (db *DB) StoreURL(ctx context.Context, req model.StoreURLRequest) (model.StoreURLResponse, error) {
	var resp model.StoreURLResponse
	res, err := db.handler.InsertURL(ctx, queries.InsertURLParams{
		Url:  string(req.URL),
		Slug: string(req.Slug),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "unique_slug" {
				return resp, newErrSlugAlreadyExists(string(req.Slug))
			}
		}
		return resp, fmt.Errorf("failed to store the URL: %w", err)
	}
	resp.URL = coreModel.URL(res.Url)
	resp.Slug = coreModel.Slug(res.Slug)
	resp.IsNewSlugInserted = resp.Slug == req.Slug
	return resp, nil
}

func newErrSlugNotFound(slug string) error {
	return fmt.Errorf("%s: %w", getProblemWithSlugMsg(slug), model.ErrSlugNotFound)
}

// GetURL gets a full URL associated with the given slug.
// If a slug does not exist it returns model.ErrSlugNotFound.
func (db *DB) GetURL(ctx context.Context, req model.GetURLRequest) (model.GetURLResponse, error) {
	resp := model.GetURLResponse{}
	url, err := db.handler.GetURL(ctx, string(req.Slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, newErrSlugNotFound(string(req.Slug))
		}
		return resp, fmt.Errorf("failed to get a URL by slug %s: %w", string(req.Slug), err)
	}
	resp.FullURL = coreModel.URL(url)
	return resp, nil
}

func (db *DB) Close(_ context.Context) error {
	db.pool.Close()
	return nil
}
