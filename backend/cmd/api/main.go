package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alongtu/backend/internal/app"
	"alongtu/backend/internal/platform/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("application initialization failed", "error", err)
		os.Exit(1)
	}
	defer a.Close()

	if err := a.Run(ctx, 15*time.Second); err != nil {
		log.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
