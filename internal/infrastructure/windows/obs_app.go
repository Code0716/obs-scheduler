//go:build windows

package windows

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"obs-scheduler/internal/domain"
)

type OBSApp struct {
	appName string
	appPath string
}

// NewOBSApp creates a new instance of OBSApp for Windows.
func NewOBSApp(appName, appPath string) domain.AppLifecycle {
	return &OBSApp{appName: appName, appPath: appPath}
}

// Start launches the OBS application on Windows.
func (a *OBSApp) Start(ctx context.Context) error {
	obsPath := a.appPath
	if obsPath == "" {
		obsPath = "C:\\Program Files\\obs-studio\\bin\\64bit\\obs64.exe"
	}

	// Use "cmd /C start" to launch the application without blocking.
	// The first empty argument "" is for the window title.
	cmd := exec.CommandContext(ctx, "cmd", "/C", "start", "", obsPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start OBS app (%s): %w", a.appName, err)
	}
	slog.Info("Successfully sent start command to OBS", "appName", a.appName)
	return nil
}

// Stop terminates the OBS application on Windows.
func (a *OBSApp) Stop(ctx context.Context) error {
	// Use taskkill to gracefully terminate the application.
	// /IM allows specifying the process by its image name.
	processName := filepath.Base(a.appName)
	if processName == "." || processName == "/" {
		// As a fallback, use a common process name for OBS
		processName = "obs64.exe"
	}

	cmd := exec.CommandContext(ctx, "taskkill", "/IM", processName, "/T")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If graceful termination fails, try to force-quit.
		slog.Warn("Failed to gracefully stop OBS, trying to force-kill...", "error", err, "output", string(output))
		forceCmd := exec.CommandContext(ctx, "taskkill", "/IM", processName, "/F", "/T")
		if forceErr := forceCmd.Run(); forceErr != nil {
			return fmt.Errorf("failed to force-stop OBS app: %w (original error: %s)", forceErr, err)
		}
	}
	slog.Info("Successfully sent stop command to OBS", "processName", processName)
	return nil
}
