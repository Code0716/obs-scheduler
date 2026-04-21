//go:build darwin

package lifecycle

import (
	"obs-scheduler/internal/domain"
	"obs-scheduler/internal/infrastructure/macos"
)

// NewOBSApp returns the macOS implementation of the AppLifecycle.
func NewOBSApp(appName, appPath string) domain.AppLifecycle {
	return macos.NewOBSApp(appName, appPath)
}
