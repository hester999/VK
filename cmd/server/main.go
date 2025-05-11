package main

import (
	"VK/internal/app"
	"VK/internal/config"
	"VK/internal/pkg/logger"
)

func main() {
	log := logger.New()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "error", err)
		return
	}

	app := app.NewApp(cfg, log)
	if err := app.Run(cfg.Server.Addr, cfg.Server.Port); err != nil {
		log.Error("failed to run app", "error", err)
	}
}
