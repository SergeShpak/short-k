package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"

	"shortik/internal/core/app"
	"shortik/internal/core/service/randgen"
	"shortik/internal/infra/api/rest"
	"shortik/internal/infra/store/db"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	cfg, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	logger := initLogger()

	g, ctx := errgroup.WithContext(rootCtx)
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.Run.ShutdownTimeout)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	d, err := db.NewDB(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to open a DB connection: %w", err)
	}

	gen := randgen.NewGenerator()

	a := app.NewApp(&app.Config{
		DB:           d,
		RandGen:      gen,
		ConfigParams: cfg.App,
	})
	srv := rest.NewServer(&rest.ServerConfig{
		ServerConfigParams: cfg.HTTP,
		Handler: rest.HandlerConfig{
			App:                 a,
			Logger:              logger.With(slog.String("component", "handler")),
			HandlerConfigParams: cfg.Handler,
		},
	})

	// server
	g.Go(func() error {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("HTTP server has failed: %w", err)
			}
		}
		return nil
	})

	// server watcher
	g.Go(func() error {
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(
			context.Background(),
			cfg.Run.HTTPServerShutdownTimeout,
		)
		defer cancelShutdownTimeoutCtx()

		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			return fmt.Errorf("an error occurred during server shutdown: %w", err)
		}
		return nil
	})

	// DB closer
	g.Go(func() error {
		<-ctx.Done()

		dbCloseCtx, cancelDBCloseCtx := context.WithTimeout(context.Background(), cfg.Run.DBCloseTimeoout)
		defer cancelDBCloseCtx()

		if err := d.Close(dbCloseCtx); err != nil {
			return fmt.Errorf("failed to properly close the DB conection: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("a component has thrown an errorL %w", err)
	}

	return nil
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
