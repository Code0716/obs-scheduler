package domain

import "context"

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock_domain

// AppLifecycle defines the interface for controlling the application lifecycle.
type AppLifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
