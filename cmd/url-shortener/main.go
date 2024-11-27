package main

import (
	"log/slog"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stepan41k/FullRestAPI/internal/config"
	nwLogger "github.com/stepan41k/FullRestAPI/internal/http-server/middleware/logger"
	"github.com/stepan41k/FullRestAPI/internal/lib/logger/handlers/slogpretty"
	"github.com/stepan41k/FullRestAPI/internal/lib/logger/sl"
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

	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(nwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)


	//middleware

	//TODO: init router: chi, chi-render

	//TODO: run server:

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