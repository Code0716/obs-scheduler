package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	StartTime  time.Time
	StopTime   time.Time
	Addr       string
	Password   string
	OBSAppName string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	startFlag := flag.String("start", "08:44", "Recording start time (HH:MM)")
	stopFlag := flag.String("stop", "10:00", "Recording stop time (HH:MM)")
	flag.Parse()

	addr := os.Getenv("OBS_ADDR")
	if addr == "" {
		return nil, fmt.Errorf("OBS_ADDR environment variable is required")
	}

	password := os.Getenv("OBS_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("OBS_PASSWORD environment variable is required")
	}

	obsAppName := os.Getenv("OBS_APP_NAME")
	if obsAppName == "" {
		obsAppName = "OBS"
	}

	now := time.Now()
	parseTargetTime := func(timeStr string) (time.Time, error) {
		t, err := time.Parse("15:04", timeStr)
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location()), nil
	}

	startTime, err := parseTargetTime(*startFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format '%s': %w", *startFlag, err)
	}

	stopTime, err := parseTargetTime(*stopFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid stop time format '%s': %w", *stopFlag, err)
	}

	return &Config{
		StartTime:  startTime,
		StopTime:   stopTime,
		Addr:       addr,
		Password:   password,
		OBSAppName: obsAppName,
	}, nil
}
