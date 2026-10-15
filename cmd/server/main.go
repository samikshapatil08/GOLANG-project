package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"switchboard/internal/config"
	"switchboard/internal/handlers"
	"switchboard/internal/repository"
	"switchboard/internal/routes"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	var listingsRepo repository.ListingRepository
	var usersRepo repository.UserRepository
	var rawDB *sql.DB

	if cfg.DatabaseURL != "" {
		dbConn, dbErr := sql.Open("postgres", cfg.DatabaseURL)
		if dbErr == nil {
			pingErr := dbConn.Ping()
			if pingErr == nil {
				slog.Info("connected to postgres database successfully")
				postgresStore := repository.NewPostgresStore(dbConn)
				listingsRepo = postgresStore
				usersRepo = postgresStore
				rawDB = dbConn
			} else {
				slog.Warn("failed to ping database, falling back to in-memory store", "error", pingErr)
			}
		} else {
			slog.Warn("failed to open database, falling back to in-memory store", "error", dbErr)
		}
	}

	if listingsRepo == nil || usersRepo == nil {
		inMemoryStore := repository.NewInMemoryStore()
		listingsRepo = inMemoryStore
		usersRepo = inMemoryStore
	}

	defer func() {
		if rawDB != nil {
			_ = rawDB.Close()
		}
	}()

	handler := handlers.NewHandler(listingsRepo, usersRepo, cfg.JWTSecret)
	router := routes.Setup(handler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	slog.Info("shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
