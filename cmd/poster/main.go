package main

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/kostromin59/poster/internal/apps/poster"
	"github.com/kostromin59/poster/internal/configs"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn(".env not found", slog.String("err", err.Error()))
	}

	var cfg configs.Poster
	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}

	if err := poster.Run(&cfg); err != nil {
		panic(err)
	}
}
