package sps30

import (
	"context"
	"errors"
	"fmt"
	"time"

	i2c "github.com/d2r2/go-i2c"
	"github.com/padiazg/go-aqi/domain"
	"github.com/padiazg/go-aqi/internal/helpers"
	i2ctransport "github.com/padiazg/go-aqi/internal/transport/i2c"
)

var _ domain.SensorProvider = (*SPS30)(nil)

var ErrNotReady = errors.New("not ready")

type SPS30 struct {
	cancel    context.CancelFunc
	transport domain.TransportProvider
	interval  time.Duration
}

// New creates a new sensor object
func New(i2c *i2c.I2C, interval time.Duration) *SPS30 {
	return &SPS30{
		transport: i2ctransport.New(i2c),
		interval:  interval,
	}
}

// ---------------------------- Interface

// Init initializes the sensor. No-op — SPS30 starts on first measurement request.
func (s *SPS30) Init(ctx context.Context) error {
	return nil
}

func (s *SPS30) Read(ctx context.Context) *domain.ReadingEvent {
	var reading *domain.AirQualityReading

	if err := s.StartMeasurement(); err != nil {
		return &domain.ReadingEvent{Err: err}
	}

	defer s.StopMeasurement()

	err := helpers.WithRetry(ctx, &helpers.RetryConfig{
		Interval: 500 * time.Millisecond,
		Times:    3,
		Timeout:  2 * time.Second,
		CountAs:  func(err error) bool { return !errors.Is(err, ErrNotReady) }, // ErrNotReady doesn't count for retry
		Fn: func() error {
			if s.IsDataReady() != 1 {
				return ErrNotReady
			}

			var err error
			reading, err = s.ReadMeasurement()
			if err != nil {
				return fmt.Errorf("read: %w", err)
			}

			return nil
		},
	})

	if err != nil {
		return &domain.ReadingEvent{Err: err}
	}

	return &domain.ReadingEvent{Reading: reading}
}

// Run starts the sensor reading loop and returns a channel for reading events.
func (s *SPS30) Run(ctx context.Context) <-chan *domain.ReadingEvent {
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	ch := make(chan *domain.ReadingEvent, 10)

	go func() {
		ticker := time.NewTicker(s.interval)

		defer close(ch)
		defer cancel()
		defer ticker.Stop()

		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				ch <- s.Read(runCtx)
			}
		}
	}()

	return ch
}

// Stop halts the sensor reading loop.
func (s *SPS30) Stop() {
	if s.cancel != nil {
		s.StopMeasurement()
		s.cancel()
	}
}

// ---------------------------- SPS30 specifics

// ReadArticleCode reads the sensor's article code via I2C address 0xD025.
// The response is packed in a 47-byte buffer with every 3rd byte as padding (skipped).
func (s *SPS30) ReadArticleCode() (string, error) {
	var code []byte
	in := make([]byte, 47)

	if err := s.readFromAddress([]byte{0xD0, 0x25}, in); err != nil {
		return "", fmt.Errorf("ReadArticleCode: %w", err)
	}

	for i := range in {
		if in[i] == 0 {
			break
		}

		if (i % 3) != 2 {
			code = append(code, in[i])
		}
	}

	return string(code), nil
}

// ReadSerial reads sensor serial
func (s *SPS30) ReadSerial() (string, error) {
	var serial []byte
	in := make([]byte, 47)
	err := s.readFromAddress([]byte{0xD0, 0x33}, in)
	if err != nil {
		return "", fmt.Errorf("ReadSerial: %w", err)
	}

	for i := range in {
		if (i % 3) != 2 {
			serial = append(serial, in[i])
		}
	}

	return string(serial), nil
}

// ReadCleaningInterval reads cleaning interval from sensor
func (s *SPS30) ReadCleaningInterval() (int64, error) {
	in := make([]byte, 6)
	if err := s.readFromAddress([]byte{0x80, 0x04}, in); err != nil {
		return -1, fmt.Errorf("ReadCleaningInterval: %w", err)
	}

	return int64(in[4]) + (int64(in[3]) << 8) + (int64(in[1]) << 16) + (int64(in[0]) << 24), nil
}

// StartMeasurement starts measurement
func (s *SPS30) StartMeasurement() error {
	err := s.transport.Write([]byte{0x00, 0x10, 0x03, 0x00, crc8Checksum([]byte{0x03, 0x00})})
	if err != nil {
		return fmt.Errorf("StartMeasurement: %w", err)
	}
	return nil
}

// StopMeasurement stops measurements
func (s *SPS30) StopMeasurement() error {
	err := s.transport.Write([]byte{0x01, 0x04})
	if err != nil {
		return fmt.Errorf("StopMeasurement: %w", err)
	}
	return nil
}

// Read reads measurements
func (s *SPS30) ReadMeasurement() (*domain.AirQualityReading, error) {
	in := make([]byte, 60)
	if err := s.readFromAddress([]byte{0x03, 0x00}, in); err != nil {
		return nil, fmt.Errorf("Read: %w", err)
	}

	return &domain.AirQualityReading{
		MassPM1:             bytesToFloat32([]byte{in[0], in[1], in[3], in[4]}),
		MassPM25:            bytesToFloat32([]byte{in[6], in[7], in[9], in[10]}),
		MassPM4:             bytesToFloat32([]byte{in[12], in[13], in[15], in[16]}),
		MassPM10:            bytesToFloat32([]byte{in[18], in[19], in[21], in[22]}),
		NumberPM05:          bytesToFloat32([]byte{in[24], in[25], in[27], in[28]}),
		NumberPM1:           bytesToFloat32([]byte{in[30], in[31], in[33], in[34]}),
		NumberPM25:          bytesToFloat32([]byte{in[36], in[37], in[39], in[40]}),
		NumberPM4:           bytesToFloat32([]byte{in[42], in[43], in[45], in[46]}),
		NumberPM10:          bytesToFloat32([]byte{in[48], in[49], in[51], in[52]}),
		TypicalParticleSize: bytesToFloat32([]byte{in[54], in[55], in[57], in[58]}),
	}, nil
}

// IsDataReady reads data-ready flag
// 0x00: no new measurements available
// 0x01: new measurements ready to read
func (s *SPS30) IsDataReady() int {
	in := make([]byte, 3)
	if err := s.readFromAddress([]byte{0x02, 0x02}, in); err != nil {
		return -1
	}

	return int(in[1])
}

// Reset sends reset command
func (s *SPS30) Reset() error {
	err := s.transport.Write([]byte{0xD3, 0x04})
	if err != nil {
		return fmt.Errorf("Reset: %w", err)
	}

	return nil
}

func (s *SPS30) readFromAddress(addr []byte, in []byte) error {
	if err := s.transport.Write(addr); err != nil {
		return fmt.Errorf("readFromAddress sending code %X: %w", addr, err)
	}

	_, err := s.transport.Read(in, false)
	if err != nil {
		return fmt.Errorf("readFromAddress reading response for %X: %w", addr, err)
	}

	return nil
}
