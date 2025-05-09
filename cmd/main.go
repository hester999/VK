package main

import (
	"VK/internal/config"
	"log/slog"
	"os"
)

func main() {

	cfg := config.LoadConfig()
	log := NewLogger()

	log.Info("app start", slog.Any("cfg", cfg))

}

func NewLogger() *slog.Logger {

	var logger *slog.Logger

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return logger
}
