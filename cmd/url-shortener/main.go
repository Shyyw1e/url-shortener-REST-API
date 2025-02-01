package main

import (
	"log/slog"
	"os"
	"url-shorter-REST-API/internal/config"
	mwLogger "url-shorter-REST-API/internal/http-server/middleware/logger"
	"url-shorter-REST-API/internal/lib/logger/handlers/slogpretty"
	"url-shorter-REST-API/internal/lib/logger/slpkg"
	"url-shorter-REST-API/internal/storage/posql"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting url-shortener",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	storage, err := posql.New(cfg.DatabaseDSN)
	if err != nil {
		log.Error("failed to init storage", slpkg.Err(err))
		os.Exit(1)
	}

	log.Info("Database connected successfully")

	urlToSave := "https://example.com"
	alias := "ex"

	err = storage.SaveURL(urlToSave, alias)
	if err != nil {
		log.Error("failed to save URL", slpkg.Err(err))
	} else {
		log.Info("URL saved successfully", slog.String("url", urlToSave), slog.String("alias", alias))
	}

	_ = storage

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger) // записывает в свой chi логгер. Надо проверить, можно ли переопределить (скорее всего нельзя)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//TODO: run server:

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()

	case envDev: //Можно сделать только dev, если не хотим разделять серверы (по сути в данном случае так и нужно сделать), но я крутой и напишу на будущее сразу и dev и prod
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
