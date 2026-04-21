//go:build windows

package lifecycle

import (
	"obs-scheduler/internal/domain"
	"obs-scheduler/internal/infrastructure/windows"
)

// NewOBSApp returns the Windows implementation of the AppLifecycle.
func NewOBSApp(appName, appPath string) domain.AppLifecycle {
	return windows.NewOBSApp(appName, appPath)
}
