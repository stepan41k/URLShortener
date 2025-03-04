package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stepan41k/FullRestAPI/internal/config"
	"github.com/stepan41k/FullRestAPI/internal/http-server/handlers/redirect"
	"github.com/stepan41k/FullRestAPI/internal/http-server/handlers/url/delete"
	"github.com/stepan41k/FullRestAPI/internal/http-server/handlers/url/save"
	nwLogger "github.com/stepan41k/FullRestAPI/internal/http-server/middleware/logger"
	"github.com/stepan41k/FullRestAPI/internal/lib/logger/handlers/slogpretty"
	"github.com/stepan41k/FullRestAPI/internal/lib/logger/sl"
	eventsender "github.com/stepan41k/FullRestAPI/internal/services/event-sender"
	"github.com/stepan41k/FullRestAPI/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")
	log.Error("error messages are enabled")

	storage, err := postgres.New(fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.PSQL.Host, cfg.PSQL.Port, cfg.PSQL.Username, cfg.PSQL.DBName, os.Getenv("DB_PASSWORD"), cfg.PSQL.SSLMode))
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	//router.Use(middleware.RealIP)
	// router.Use(middleware.Logger)
	router.Use(nwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.Server.User: cfg.Server.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))
	
	log.Info("starting server", slog.String("address", cfg.Server.Address))

	srv := &http.Server{
		Addr: cfg.Server.Address,
		Handler: router,
		ReadTimeout: cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout: cfg.Server.IdleTimeout,
	}

	sender := eventsender.New(storage, log)
	sender.StartProcessEvents(context.Background(), 5*time.Second)

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log

}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}