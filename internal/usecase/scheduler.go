package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"obs-scheduler/internal/domain"
)

type Scheduler struct {
	recorder   domain.Recorder
	app        domain.AppLifecycle
	skipLaunch bool
}

func NewScheduler(recorder domain.Recorder, app domain.AppLifecycle, skipLaunch bool) *Scheduler {
	return &Scheduler{
		recorder:   recorder,
		app:        app,
		skipLaunch: skipLaunch,
	}
}

func (s *Scheduler) Run(ctx context.Context, startTime, stopTime time.Time) error {
	slog.Info("📅 Plan", "start", startTime.Format("2006/01/02 15:04:05"), "stop", stopTime.Format("2006/01/02 15:04:05"))

	if stopTime.Before(time.Now()) {
		return fmt.Errorf("stop time %s is already in the past", stopTime.Format("15:04:05"))
	}

	if err := s.prepareOBS(ctx, startTime); err != nil {
		return err
	}
	defer func() {
		if err := s.recorder.Close(); err != nil {
			slog.Error("Failed to close recorder", "error", err)
		}
	}()

	if err := s.executeRecording(ctx, startTime, stopTime); err != nil {
		return err
	}

	return s.cleanup(ctx)
}

func (s *Scheduler) prepareOBS(ctx context.Context, startTime time.Time) error {
	// 1. Wait for launch time (10 seconds before start time)
	launchTime := startTime.Add(-10 * time.Second)
	sleep := time.Until(launchTime)
	if sleep > 0 {
		slog.Info("Waiting for launch time...", "duration", sleep, "launchTime", launchTime)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// 2. Launch OBS
	if !s.skipLaunch {
		slog.Info("🚀 Launching OBS...")
		if err := s.app.Start(ctx); err != nil {
			return fmt.Errorf("failed to start OBS app: %w", err)
		}
	} else {
		slog.Info("🚀 Skipping OBS launch")
	}

	// 3. Connect to OBS (with retry)
	if err := s.connectWithRetry(ctx); err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	return nil
}

func (s *Scheduler) executeRecording(ctx context.Context, startTime, stopTime time.Time) error {
	// Wait for actual start time if we are early
	sleep := time.Until(startTime)
	if sleep > 0 {
		slog.Info("OBS connected, waiting for start time...", "duration", sleep)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Start Recording
	if err := s.startRecordingWithRetry(ctx); err != nil {
		return fmt.Errorf("failed to start recording: %w", err)
	}
	slog.Info("🎬 Recording started")

	// Wait for stop time
	sleep = time.Until(stopTime)
	if sleep > 0 {
		slog.Info("Recording... waiting for stop time...", "duration", sleep)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping recording...")
		}
	}

	// Stop Recording
	if err := s.recorder.StopRecording(); err != nil {
		return fmt.Errorf("failed to stop recording: %w", err)
	}
	slog.Info("🛑 Recording stopped")

	// Wait for file save (grace period)
	slog.Info("Waiting for file save confirmation...", "duration", "1s")
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
	}

	return nil
}

func (s *Scheduler) cleanup(ctx context.Context) error {
	if !s.skipLaunch {
		slog.Info("👋 Stopping OBS...")
		if err := s.app.Stop(ctx); err != nil {
			slog.Error("Failed to stop OBS app", "error", err)
			return err
		}
	}
	return nil
}

func (s *Scheduler) connectWithRetry(ctx context.Context) error {
	maxRetries := 15
	retryInterval := 2 * time.Second

	for i := range maxRetries {
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

func (s *Scheduler) startRecordingWithRetry(ctx context.Context) error {
	maxRetries := 10
	retryInterval := 2 * time.Second

	for i := range maxRetries {
		err := s.recorder.StartRecording()
		if err == nil {
			return nil
		}
		slog.Info("Failed to start recording (OBS might not be ready), retrying...", "attempt", i+1, "error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// continue
		}
	}
	return fmt.Errorf("failed to start recording after %d attempts", maxRetries)
}
