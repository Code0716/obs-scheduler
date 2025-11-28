package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"obs-scheduler/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	scheduler, cleanup, err := InitializeScheduler(cfg)
	if err != nil {
		slog.Error("Failed to initialize scheduler", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	if err := scheduler.Run(ctx, cfg.StartTime, cfg.StopTime); err != nil {
		slog.Error("Error during execution", "error", err)
		os.Exit(1)
	}

	slog.Info("Recording session finished ✅")
}
