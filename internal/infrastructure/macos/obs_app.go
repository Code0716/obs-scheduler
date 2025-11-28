package macos

import (
	"context"
	"fmt"
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
	script := fmt.Sprintf("quit app \"%s\"", a.appName)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop OBS app (%s): %w", a.appName, err)
	}
	return nil
}
