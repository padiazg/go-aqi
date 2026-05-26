package zh07

import (
	"fmt"
	"time"

	"github.com/padiazg/go-aqi/domain"
)

// ModeType selects the ZH07 sensor communication mode.
type ModeType int

const (
	// ModeInitiative drives the sensor in initiative upload mode (sensor pushes data without polling).
	ModeInitiative ModeType = iota
	// ModeQA drives the sensor in query-response mode (host polls sensor for data).
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
