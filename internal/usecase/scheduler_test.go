package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"obs-scheduler/internal/domain"
)

type mockRecorder struct {
	connectFunc        func() error
	closeFunc          func() error
	startRecordingFunc func() error
	stopRecordingFunc  func() error
}

func (m *mockRecorder) Connect() error {
	if m.connectFunc != nil {
		return m.connectFunc()
	}
	return nil
}

func (m *mockRecorder) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockRecorder) StartRecording() error {
	if m.startRecordingFunc != nil {
		return m.startRecordingFunc()
	}
	return nil
}

func (m *mockRecorder) StopRecording() error {
	if m.stopRecordingFunc != nil {
		return m.stopRecordingFunc()
	}
	return nil
}

type mockAppLifecycle struct {
	startFunc func(ctx context.Context) error
	stopFunc  func(ctx context.Context) error
}

func (m *mockAppLifecycle) Start(ctx context.Context) error {
	if m.startFunc != nil {
		return m.startFunc(ctx)
	}
	return nil
}

func (m *mockAppLifecycle) Stop(ctx context.Context) error {
	if m.stopFunc != nil {
		return m.stopFunc(ctx)
	}
	return nil
}

func TestScheduler_Run(t *testing.T) {
	tests := []struct {
		name          string
		setupRecorder func() domain.Recorder
		setupApp      func() domain.AppLifecycle
		startDelay    time.Duration
		stopDelay     time.Duration
		cancelAfter   time.Duration // if > 0, cancel context after this duration
		wantErr       bool
		skipLaunch    bool
	}{
		{
			name: "Success",
			setupRecorder: func() domain.Recorder {
				return &mockRecorder{
					connectFunc:        func() error { return nil },
					startRecordingFunc: func() error { return nil },
					stopRecordingFunc:  func() error { return nil },
					closeFunc:          func() error { return nil },
				}
			},
			setupApp: func() domain.AppLifecycle {
				return &mockAppLifecycle{
					startFunc: func(ctx context.Context) error { return nil },
					stopFunc:  func(ctx context.Context) error { return nil },
				}
			},
			startDelay: 100 * time.Millisecond,
			stopDelay:  200 * time.Millisecond,
			wantErr:    false,
		},
		{
			name: "Connect Error",
			setupRecorder: func() domain.Recorder {
				return &mockRecorder{
					connectFunc: func() error { return errors.New("connect error") },
				}
			},
			setupApp: func() domain.AppLifecycle {
				return &mockAppLifecycle{
					startFunc: func(ctx context.Context) error { return nil },
				}
			},
			startDelay:  100 * time.Millisecond,
			stopDelay:   200 * time.Millisecond,
			cancelAfter: 150 * time.Millisecond, // Cancel to break retry loop
			wantErr:     true,
		},
		{
			name: "Start Recording Error",
			setupRecorder: func() domain.Recorder {
				return &mockRecorder{
					connectFunc:        func() error { return nil },
					startRecordingFunc: func() error { return errors.New("start error") },
				}
			},
			setupApp: func() domain.AppLifecycle {
				return &mockAppLifecycle{
					startFunc: func(ctx context.Context) error { return nil },
					stopFunc:  func(ctx context.Context) error { return nil },
				}
			},
			startDelay: 100 * time.Millisecond,
			stopDelay:  200 * time.Millisecond,
			wantErr:    true,
		},
		{
			name: "Context Cancelled Before Start",
			setupRecorder: func() domain.Recorder {
				return &mockRecorder{
					connectFunc: func() error { return nil },
				}
			},
			setupApp: func() domain.AppLifecycle {
				return &mockAppLifecycle{}
			},
			startDelay:  200 * time.Millisecond,
			stopDelay:   300 * time.Millisecond,
			cancelAfter: 50 * time.Millisecond,
			wantErr:     true, // ctx.Err()
		},
		{
			name: "Skip Launch",
			setupRecorder: func() domain.Recorder {
				return &mockRecorder{
					connectFunc:        func() error { return nil },
					startRecordingFunc: func() error { return nil },
					stopRecordingFunc:  func() error { return nil },
					closeFunc:          func() error { return nil },
				}
			},
			setupApp: func() domain.AppLifecycle {
				return &mockAppLifecycle{
					startFunc: func(ctx context.Context) error { return errors.New("should not be called") },
					stopFunc:  func(ctx context.Context) error { return errors.New("should not be called") },
				}
			},
			startDelay: 100 * time.Millisecond,
			stopDelay:  200 * time.Millisecond,
			wantErr:    false,
			skipLaunch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := tt.setupRecorder()
			app := tt.setupApp()
			scheduler := NewScheduler(recorder, app, tt.skipLaunch)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tt.cancelAfter > 0 {
				go func() {
					time.Sleep(tt.cancelAfter)
					cancel()
				}()
			}

			startTime := time.Now().Add(tt.startDelay)
			stopTime := time.Now().Add(tt.stopDelay)

			err := scheduler.Run(ctx, startTime, stopTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
