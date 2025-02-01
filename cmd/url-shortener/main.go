package main

import (
	"log/slog"
	"os"
	"url-shorter-REST-API/internal/config"
	"url-shorter-REST-API/internal/lib/logger/slpkg"
	"url-shorter-REST-API/internal/storage/posql"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("Debug messages are enabled now")

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

	//TODO: init router: chi (минималистичный очень, полностью совместим с net/http), "chi render"

	//TODO: run server:
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
		//Можно сделать только dev, если не хотим разделять серверы (по сути в данном случае так и нужно сделать), но я крутой и напишу на будущее сразу и dev и prod
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
