package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original args and env vars
	origArgs := os.Args
	origEnvAddr := os.Getenv("OBS_ADDR")
	origEnvPass := os.Getenv("OBS_PASSWORD")
	defer func() {
		os.Args = origArgs
		if origEnvAddr != "" {
			_ = os.Setenv("OBS_ADDR", origEnvAddr)
		} else {
			_ = os.Unsetenv("OBS_ADDR")
		}
		if origEnvPass != "" {
			_ = os.Setenv("OBS_PASSWORD", origEnvPass)
		} else {
			_ = os.Unsetenv("OBS_PASSWORD")
		}
	}()

	tests := []struct {
		name      string
		args      []string
		env       map[string]string
		wantStart string // HH:MM
		wantStop  string // HH:MM
		wantErr   bool
	}{
		{
			name: "Valid config",
			args: []string{"cmd", "-start", "10:00", "-stop", "11:00"},
			env: map[string]string{
				"OBS_ADDR":     "localhost:4455",
				"OBS_PASSWORD": "password",
			},
			wantStart: "10:00",
			wantStop:  "11:00",
			wantErr:   false,
		},
		{
			name: "Missing OBS_ADDR",
			args: []string{"cmd"},
			env: map[string]string{
				"OBS_PASSWORD": "password",
			},
			wantErr: true,
		},
		{
			name: "Missing OBS_PASSWORD",
			args: []string{"cmd"},
			env: map[string]string{
				"OBS_ADDR": "localhost:4455",
			},
			wantErr: true,
		},
		{
			name: "Invalid start time",
			args: []string{"cmd", "-start", "invalid"},
			env: map[string]string{
				"OBS_ADDR":     "localhost:4455",
				"OBS_PASSWORD": "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = tt.args

			// Set env vars
			os.Clearenv()
			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				wantStart, _ := time.Parse("15:04", tt.wantStart)
				wantStop, _ := time.Parse("15:04", tt.wantStop)

				// Check time (ignoring date part for simplicity, or check full date)
				// The Load function sets date to today.
				if cfg.StartTime.Hour() != wantStart.Hour() || cfg.StartTime.Minute() != wantStart.Minute() {
					t.Errorf("StartTime = %v, want %v", cfg.StartTime.Format("15:04"), tt.wantStart)
				}
				if cfg.StopTime.Hour() != wantStop.Hour() || cfg.StopTime.Minute() != wantStop.Minute() {
					t.Errorf("StopTime = %v, want %v", cfg.StopTime.Format("15:04"), tt.wantStop)
				}
				if cfg.Addr != tt.env["OBS_ADDR"] {
					t.Errorf("Addr = %v, want %v", cfg.Addr, tt.env["OBS_ADDR"])
				}
			}
		})
	}
}
