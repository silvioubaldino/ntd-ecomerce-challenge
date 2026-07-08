package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ntd-ecomerce-api/internal/bootstrap"
	"ntd-ecomerce-api/migrations"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type config struct {
	port        string
	databaseURL string
}

func loadConfig() config {
	return config{
		port:        getEnv("API_PORT", "8080"),
		databaseURL: os.Getenv("DATABASE_URL"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if err := run(); err != nil {
		slog.Error("api exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := loadConfig()
	logger := configureLogger()
	slog.SetDefault(logger)

	db, err := openDB(cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	if err := runMigrations(cfg.databaseURL); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	r := setupGin()
	bootstrap.SetupComponents(r, db)

	return serve(r, cfg.port)
}

func configureLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func openDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(gormpostgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
}

func runMigrations(dsn string) error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("loading embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("initializing migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}

	return nil
}

func setupGin() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func serve(r *gin.Engine, port string) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("serving: %w", err)
	case <-stop:
		slog.Info("shutting down")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
