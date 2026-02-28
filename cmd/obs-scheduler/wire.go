//go:build wireinject

package main

import (
	"obs-scheduler/internal/config"
	"obs-scheduler/internal/domain"
	"obs-scheduler/internal/infrastructure/lifecycle"
	"obs-scheduler/internal/infrastructure/obs"
	"obs-scheduler/internal/usecase"

	"github.com/google/wire"
)

func InitializeScheduler(cfg *config.Config) (*usecase.Scheduler, func(), error) {
	wire.Build(
		provideOBSClient,
		provideOBSApp,
		provideSkipLaunch,
		usecase.NewScheduler,
		wire.Bind(new(domain.Recorder), new(*obs.Client)),
	)
	return nil, nil, nil
}

func provideOBSClient(cfg *config.Config) *obs.Client {
	return obs.NewClient(cfg.Addr, cfg.Password)
}

func provideOBSApp(cfg *config.Config) domain.AppLifecycle {
	return lifecycle.NewOBSApp(cfg.OBSAppName)
}

func provideSkipLaunch(cfg *config.Config) bool {
	return cfg.SkipLaunch
}
