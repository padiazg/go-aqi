package domain

import (
	"context"
)

type SensorProvider interface {
	Init(ctx context.Context) error
	Run(ctx context.Context) <-chan *ReadingEvent
	Read(ctx context.Context) *ReadingEvent
	Stop()
}
