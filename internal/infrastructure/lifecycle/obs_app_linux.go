//go:build !windows && !darwin

package lifecycle

import (
	"context"
	"log/slog"

	"obs-scheduler/internal/domain"
)

type StubApp struct {
}

func NewOBSApp(appName, appPath string) domain.AppLifecycle {
	return &StubApp{}
}

func (a *StubApp) Start(ctx context.Context) error {
	slog.Warn("Start is not implemented on this platform")
	return nil
}

func (a *StubApp) Stop(ctx context.Context) error {
	slog.Warn("Stop is not implemented on this platform")
	return nil
}
