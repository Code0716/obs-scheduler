package macos

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"obs-scheduler/internal/domain"
)

type OBSApp struct {
	appName string
}

// NewOBSApp creates a new instance of OBSApp.
func NewOBSApp(appName string) domain.AppLifecycle {
	return &OBSApp{appName: appName}
}

func (a *OBSApp) Start(ctx context.Context) error {
	// "open -a OBS" launches the application on macOS
	// Using "open" command returns immediately after sending the launch request,
	// so we don't need to worry about blocking.
	cmd := exec.CommandContext(ctx, "open", "-a", a.appName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start OBS app (%s): %w", a.appName, err)
	}
	return nil
}

func (a *OBSApp) Stop(ctx context.Context) error {
	// Use AppleScript to gracefully quit the application
	script := fmt.Sprintf("tell application \"%s\" to quit", a.appName)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// Fallback to pkill if AppleScript fails (e.g. user cancelled or dialog shown)
	slog.Warn("Failed to quit OBS via AppleScript, trying pkill...", "error", err, "output", string(output))

	// Try to kill by process name "OBS"
	// Note: appName might be "OBS Studio" but process name is usually "OBS"
	pkillCmd := exec.CommandContext(ctx, "pkill", "-x", "OBS")
	if err := pkillCmd.Run(); err != nil {
		return fmt.Errorf("failed to stop OBS app via pkill: %w (AppleScript error: %s)", err, string(output))
	}

	return nil
}
