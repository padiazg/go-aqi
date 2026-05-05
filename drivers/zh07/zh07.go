package zh07

import (
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
}

func New(config *Config) domain.SensorProvider {
	switch config.Mode {
	case ModeInitiative:
		return newZH07i(config)
	case ModeQA:
		return newZH07q(config)
	default:
		return nil
	}
}
