package zh07

import (
	"fmt"
	"time"

	"github.com/padiazg/go-aqi/domain"
)

type ModeType int

const (
	ModeInitiative ModeType = iota
	ModeQA
)

// Config holds configuration options for sensor instances.
type Config struct {
	Transport domain.TransportProvider
	Interval  time.Duration
	Mode      ModeType
	ID        string
}

func New(config *Config) (domain.SensorProvider, error) {
	switch config.Mode {
	case ModeInitiative:
		return newZH07i(config), nil
	case ModeQA:
		return newZH07q(config), nil
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnknownMode, config.Mode)
	}
}
