package domain

import (
	"context"
)

// SensorProvider drives a particulate matter sensor through its lifecycle:
// Init (one-time setup), Run (periodic reading loop), Read (single shot), and Stop.
type SensorProvider interface {
	Init(ctx context.Context) error
	Run(ctx context.Context) <-chan *ReadingEvent
	Read(ctx context.Context) *ReadingEvent
	Stop()
}
