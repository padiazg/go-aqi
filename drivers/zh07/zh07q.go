package zh07

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/padiazg/go-aqi/domain"
	"github.com/padiazg/go-aqi/internal/transport/serial"
)

var _ domain.SensorProvider = (*ZH07q)(nil)

// ZH07q drives the ZH07 sensor in query-response mode (host polls sensor for data).
type ZH07q struct {
	id        string
	transport domain.TransportProvider
	cancel    context.CancelFunc
	interval  time.Duration
	data      []byte
}

// newZH07q creates a new sensor object.
func newZH07q(config *Config) *ZH07q {
	if config == nil {
		config = &Config{}
	}

	if config.Transport == nil {
		// default is a no-op transport for testing (reads empty byte slice)
		config.Transport = serial.New(bufio.NewReadWriter(bufio.NewReader(bytes.NewReader([]byte{})), nil))
	}

	id := config.ID
	if id == "" {
		id = "zh07-q"
	}

	return &ZH07q{
		id:        id,
		transport: config.Transport,
		data:      make([]byte, 9),
		interval:  config.Interval,
	}
}

// Init initializes the sensor for question and answer mode.
func (z *ZH07q) Init(ctx context.Context) error {
	if err := z.transport.Write(commandSetQAMode); err != nil {
		return err
	}

	time.Sleep(sleepAfterWrite) // wait command to be executed

	return nil
}

// Read sends a query command and reads particulate matter data from the sensor.
func (z *ZH07q) Read(ctx context.Context) *domain.ReadingEvent {
	if err := z.writeAndRead(commandQuery, z.data); err != nil {
		return &domain.ReadingEvent{Err: err}
	}

	if !z.IsReadingValid() {
		return &domain.ReadingEvent{Err: fmt.Errorf("%w: received=%X, calculated=%X", ErrChecksumMismatch, z.getChecksum(), z.CalculateChecksum())}
	}

	return &domain.ReadingEvent{
		Reading: &domain.AirQualityReading{
			Timestamp:  time.Now(),
			SensorID:   z.id,
			NumberPM1:  float32(bytesToUint16BE(z.data[6:8])),
			NumberPM25: float32(bytesToUint16BE(z.data[2:4])),
			NumberPM10: float32(bytesToUint16BE(z.data[4:6])),
		},
	}
}

// Run starts the sensor reading loop and returns a channel for reading events.
func (s *ZH07q) Run(ctx context.Context) <-chan *domain.ReadingEvent {
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
				event := s.Read(runCtx)
				if event != nil {
					ch <- event
				}
			}
		}
	}()

	return ch
}

// Stop halts the sensor reading loop.
func (s *ZH07q) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// ---------------------------- ZH07q specifics
func (z *ZH07q) writeAndRead(command, in []byte) error {
	if err := z.transport.Write(command); err != nil {
		return fmt.Errorf("writeAndRead sending code: %X: %w", command, err)
	}

	time.Sleep(sleepAfterWrite) // wait for the response

	if _, err := z.transport.Read(in, false); err != nil { // read response from tty
		return fmt.Errorf("writeAndRead reading response for %X: %w", command, err)
	}

	return nil
}

// IsReadingValid checks if the calculated checksum matches the payload checksum.
func (z *ZH07q) IsReadingValid() bool {
	return z.CalculateChecksum() == z.getChecksum()
}

// CalculateChecksum calculates the checksum from the payload.
func (z *ZH07q) CalculateChecksum() int {
	return calculateChecksum(z.data)
}

// getChecksum recovers the checksum from the payload.
func (z *ZH07q) getChecksum() int {
	return int(z.data[8])
}
