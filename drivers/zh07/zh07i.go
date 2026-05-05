package zh07

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/padiazg/go-aqi/domain"
	"github.com/padiazg/go-aqi/transport/serial"
)

var _ domain.SensorProvider = (*ZH07i)(nil)

type ZH07i struct {
	transport domain.TransportProvider
	cancel    context.CancelFunc
	interval  time.Duration
	data      []byte
}

// New creates a new sensor object
func newZH07i(config *Config) *ZH07i {
	if config == nil {
		config = &Config{}
	}

	if config.Transport == nil {
		config.Transport = serial.New(bufio.NewReadWriter(bufio.NewReader(bytes.NewReader([]byte{})), nil))
	}

	return &ZH07i{
		data:      make([]byte, 32),
		transport: config.Transport,
		interval:  config.Interval,
	}
}

// ---------------------------- Interface
// Init initializes the sensor for initiative upload mode.
func (z *ZH07i) Init(ctx context.Context) error {
	if err := z.transport.Write(commandSetInitiativeUploadMode); err != nil {
		return err
	}

	time.Sleep(sleepAfterWrite) // wait command to be executed

	return nil
}

// Read reads particulate matter data from the sensor in initiative upload mode.
func (z *ZH07i) Read(ctx context.Context) *domain.ReadingEvent {
	var (
		b0   = make([]byte, 1)
		b1   = make([]byte, 3)
		data = make([]byte, 28)
		err  error
	)

	// read byte by byte until we find the 1st start character (0x42)
	if _, err = z.transport.Read(b0, false); err != nil {
		return &domain.ReadingEvent{Err: fmt.Errorf("%w: %v", ErrSensorCommunication, err)}
	}

	if b0[0] != 0x42 {
		return nil
	}

	// then we read the next 3 bytes, which should be:
	//   2nd character start (0x4d)
	//   frame length high bits
	//   frame length low bits
	if _, err = z.transport.Read(b1, false); err != nil {
		return &domain.ReadingEvent{Err: fmt.Errorf("%w: %v", ErrSensorCommunication, err)}
	}

	if b1[0] != 0x4d {
		return nil
	}

	// frame length must be 0x00 0x1C => 28
	if byteToInt(b1[1:3]) != 28 {
		return nil
	}

	// if everything matches so far, we read the remaining data (28 bytes)
	// if _, err = io.ReadFull(z.rw, data); err != nil {
	if _, err = z.transport.Read(data, true); err != nil {
		return &domain.ReadingEvent{Err: fmt.Errorf("%w: %v", ErrSensorCommunication, err)}
	}

	// let's concatenate all the bytes read into a single slice
	z.data = append(append(b0, b1...), data...)

	return &domain.ReadingEvent{
		Reading: &domain.AirQualityReading{
			NumberPM1:  float32(byteToInt(z.data[10:12])),
			NumberPM25: float32(byteToInt(z.data[12:14])),
			NumberPM10: float32(byteToInt(z.data[14:16])),
		},
	}
}

// Run starts the sensor reading loop and returns a channel for reading events.
func (s *ZH07i) Run(ctx context.Context) <-chan *domain.ReadingEvent {
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
func (s *ZH07i) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
