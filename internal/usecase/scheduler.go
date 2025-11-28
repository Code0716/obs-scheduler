package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"obs-scheduler/internal/domain"
)

type Scheduler struct {
	recorder domain.Recorder
	app      domain.AppLifecycle
}

func NewScheduler(recorder domain.Recorder, app domain.AppLifecycle) *Scheduler {
	return &Scheduler{
		recorder: recorder,
		app:      app,
	}
}

func (s *Scheduler) Run(ctx context.Context, startTime, stopTime time.Time) error {
	slog.Info("📅 Plan", "start", startTime.Format("15:04:05"), "stop", stopTime.Format("15:04:05"))

	// 1. Wait for start time
	sleep := time.Until(startTime)
	if sleep > 0 {
		slog.Info("Waiting for start time...", "duration", sleep)
		select {
		case <-time.After(sleep):
			// continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// 2. Launch OBS
	slog.Info("🚀 Launching OBS...")
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start OBS app: %w", err)
	}

	// 3. Connect to OBS (with retry)
	if err := s.connectWithRetry(ctx); err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}
	defer s.recorder.Close()

	// 4. Start Recording
	if err := s.recorder.StartRecording(); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}
	slog.Info("🎬 Recording started")

	// 5. Wait for stop time
	sleep = time.Until(stopTime)
	if sleep > 0 {
		slog.Info("Recording... waiting for stop time...", "duration", sleep)
		select {
		case <-time.After(sleep):
			// continue
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping recording...")
		}
	}

	// 6. Stop Recording
	if err := s.recorder.StopRecording(); err != nil {
		return fmt.Errorf("failed to stop recording: %w", err)
	}
	slog.Info("🛑 Recording stopped")

	// 7. Wait for file save (grace period)
	slog.Info("Waiting for file save...", "duration", "5s")
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
	}

	// 8. Stop OBS
	slog.Info("👋 Stopping OBS...")
	if err := s.app.Stop(ctx); err != nil {
		slog.Error("Failed to stop OBS app", "error", err)
	}

	return nil
}

func (s *Scheduler) connectWithRetry(ctx context.Context) error {
	maxRetries := 15
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if err := s.recorder.Connect(); err == nil {
			return nil
		}
		slog.Info("Waiting for OBS to be ready...", "attempt", i+1)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// continue
		}
	}
	return fmt.Errorf("failed to connect to OBS after %d attempts", maxRetries)
}
