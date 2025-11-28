package domain

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock_domain

type Recorder interface {
	Connect() error
	Close() error
	StartRecording() error
	StopRecording() error
}
