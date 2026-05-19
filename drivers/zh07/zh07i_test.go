package zh07

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/padiazg/go-aqi/domain"
	"github.com/padiazg/go-aqi/internal/transport/serial"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	sampleInitiativePayload = []byte{
		0x42,       // Start byte 1
		0x4D,       // Start byte 2
		0x00, 0x1C, // Frame length
		0x00, 0x54, // Data  1 | Reserved
		0x00, 0x6E, // Data  2 | Reserved
		0x00, 0x7C, // Data  3 | Reserved
		0x00, 0x54, // Data  4 | PM1.0 concentration (ug/m3)
		0x00, 0x6E, // Data  5 | PM2.5 concentration (ug/m3)
		0x00, 0x7C, // Data  6 | PM10 concentration (ug/m3)
		0x00, 0x00, // Data  7 | Reserved
		0x00, 0x00, // Data  8 | Reserved
		0x00, 0x00, // Data  9 | Reserved
		0x00, 0x00, // Data 10 | Reserved
		0x00, 0x00, // Data 11 | Reserved
		0x00, 0x00, // Data 12 | Reserved
		0x00, 0x00, // Data 13 | Reserved
		0x03, 0x27, // Checksum
	}

	sampleInitiativeBadChecksum = []byte{
		0x42,       // Start byte 1
		0x4D,       // Start byte 2
		0x00, 0x1C, // Frame length
		0x00, 0x54, // Data  1 | Reserved
		0x00, 0x6E, // Data  2 | Reserved
		0x00, 0x7C, // Data  3 | Reserved
		0x00, 0x54, // Data  4 | PM1.0 concentration (ug/m3)
		0x00, 0x6E, // Data  5 | PM2.5 concentration (ug/m3)
		0x00, 0x7C, // Data  6 | PM10 concentration (ug/m3)
		0x00, 0x00, // Data  7 | Reserved
		0x00, 0x00, // Data  8 | Reserved
		0x00, 0x00, // Data  9 | Reserved
		0x00, 0x00, // Data 10 | Reserved
		0x00, 0x00, // Data 11 | Reserved
		0x00, 0x00, // Data 12 | Reserved
		0x00, 0x00, // Data 13 | Reserved
		0x03, 0x28, // Checksum
	} // badChecksum ...

)

type newZH07iFn func(*testing.T, *ZH07i)

var checknewZH07i = func(fns ...newZH07iFn) []newZH07iFn { return fns }

func Test_newZH07i(t *testing.T) {

	checkNotNil := func(t *testing.T, z *ZH07i) {
		t.Helper()
		assert.NotNil(t, z)
	}

	checkTransportType := func(want domain.TransportProvider) func(t *testing.T, z *ZH07i) {
		return func(t *testing.T, z *ZH07i) {
			t.Helper()
			assert.NotNil(t, want)
			assert.IsType(t, want, z.transport)
		}
	}

	checkId := func(want string) func(t *testing.T, z *ZH07i) {
		return func(t *testing.T, z *ZH07i) {
			t.Helper()
			assert.Equal(t, want, z.id)
		}
	}

	tests := []struct {
		name   string
		config *Config
		checks []newZH07iFn
	}{
		{
			name: "success - defaults",
			checks: checknewZH07i(
				checkNotNil,
				checkTransportType(&serial.SerialTransport{}),
				checkId("zh07-i"),
			),
		},
		{
			name: "success - custom",
			config: &Config{
				Transport: new(mockTransportProvider),
				ID:        "zh07i-01",
			},
			checks: checknewZH07i(
				checkNotNil,
				checkTransportType(&mockTransportProvider{}),
				checkId("zh07i-01"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := newZH07i(tt.config)
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

type checkZH07iInitFn func(*testing.T, error)

var checkZH07iInit = func(fns ...checkZH07iInitFn) []checkZH07iInitFn { return fns }

func checkInitError(want string) checkZH07iInitFn {
	return func(t *testing.T, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkInitError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkInitError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkInitError mismatch")
		}
	}
}
func TestZH07i_Init(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkZH07iInitFn
		before func(*ZH07i)
	}{
		{
			name: "fail - transport.write",
			before: func(z *ZH07i) {
				z.transport.(*mockTransportProvider).
					On("Write", mock.Anything).
					Return(fmt.Errorf("transport-write-error"))
			},
			checks: checkZH07iInit(
				checkInitError("transport-write-error"),
			),
		},
		{
			name: "success",
			before: func(z *ZH07i) {
				z.transport.(*mockTransportProvider).
					On("Write", mock.Anything).
					Return(nil)
			},
			checks: checkZH07iInit(
				checkInitError(""),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := newZH07i(&Config{Transport: new(mockTransportProvider)})

			if tt.before != nil {
				tt.before(s)
			}

			err := s.Init(context.Background())
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}

type checkReadFn func(*testing.T, *domain.ReadingEvent)

var checkZH07iRead = func(fns ...checkReadFn) []checkReadFn { return fns }

func checkReadValues(pm1, pm25, pm10 float32) checkReadFn {
	return func(t *testing.T, re *domain.ReadingEvent) {
		t.Helper()
		assert.Equalf(t, pm1, re.Reading.NumberPM1, "PM1 expected %.2f, got %.2f", pm1, re.Reading.NumberPM1)
		assert.Equalf(t, pm25, re.Reading.NumberPM25, "PM25 expected %.2f, got %.2f", pm25, re.Reading.NumberPM25)
		assert.Equalf(t, pm10, re.Reading.NumberPM10, "PM10 expected %.2f, got %.2f", pm10, re.Reading.NumberPM10)
	}
}

func TestZH07i_Read(t *testing.T) {

	checkZH07iReadError := func(want string) checkReadFn {
		return func(t *testing.T, re *domain.ReadingEvent) {
			t.Helper()
			if want == "" {
				assert.NoErrorf(t, re.Err, "checkZH07iReadError: expected no error, got %v", re.Err)
				return
			}
			if assert.Errorf(t, re.Err, "checkZH07iReadError: expected error %q", want) {
				assert.Containsf(t, re.Err.Error(), want, "checkZH07iReadError mismatch")
			}
		}
	}

	tests := []struct {
		name   string
		checks []checkReadFn
		before func(*ZH07i)
	}{
		{
			name: "fail - start character read error",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Return(0, fmt.Errorf("read-error"))
			},
			checks: checkZH07iRead(
				checkZH07iReadError("read-error"),
			),
		},
		{
			name: "fail - incorrect start character",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = 0x00
					}).
					Return(1, nil)
			},
			checks: checkZH07iRead(
				checkZH07iReadError("expected 0x42"),
			),
		},
		{
			name: "fail - read next 3 bytes",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = sampleInitiativePayload[0]
					}).
					Return(1, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Return(0, fmt.Errorf("next-3-bytes-error")).
					Once()
			},
			checks: checkZH07iRead(
				checkZH07iReadError("next-3-bytes-error"),
			),
		},
		{
			name: "fail - invalid frame signature",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = sampleInitiativePayload[0]
					}).
					Return(1, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, []byte{0x00, 0x00, 0x00})
					}).
					Return(3, nil).
					Once()
			},
			checks: checkZH07iRead(
				checkZH07iReadError("expected 0x4d"),
			),
		},
		{
			name: "fail - invalid frame length",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = sampleInitiativePayload[0]
					}).
					Return(1, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, []byte{
							sampleInitiativePayload[1],
							sampleInitiativePayload[2],
							0x00,
						})
					}).
					Return(3, nil).
					Once()
			},
			checks: checkZH07iRead(
				checkZH07iReadError("expected length 28"),
			),
		},
		{
			name: "fail - read remaining",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = sampleInitiativePayload[0]
					}).
					Return(1, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleInitiativePayload[1:4])
					}).
					Return(3, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Return(0, fmt.Errorf("read-remaining-error")).
					Once()
			},
			checks: checkZH07iRead(
				checkZH07iReadError("read-remaining-error"),
			),
		},
		{
			name: "success",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = sampleInitiativePayload[0]
					}).
					Return(1, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleInitiativePayload[1:4])
					}).
					Return(3, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleInitiativePayload[4:])
					}).
					Return(28, nil).
					Once()
			},
			checks: checkZH07iRead(
				checkZH07iReadError(""),
				checkReadValues(84, 110, 124),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := newZH07i(&Config{Transport: new(mockTransportProvider)})
			if tt.before != nil {
				tt.before(s)
			}
			r := s.Read(context.Background())
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

var checkZH07iRun = func(fns ...checkReadFn) []checkReadFn { return fns }

func TestZH07i_Run(t *testing.T) {
	checkZH07iRunError := func(want string) checkReadFn {
		return func(t *testing.T, re *domain.ReadingEvent) {
			t.Helper()
			if want == "" && re.Err != nil {
				assert.NoErrorf(t, re.Err, "checkInitError: expected no error, got %v", re.Err)
				return
			}
			if want != "" && assert.NotNil(t, re.Err) {
				if assert.Errorf(t, re.Err, "checkInitError: expected error %q", want) {
					assert.Containsf(t, re.Err.Error(), want, "checkInitError mismatch")
				}
			}
		}
	}

	tests := []struct {
		name   string
		checks []checkReadFn
		before func(*ZH07i)
	}{
		{
			name: "emits valid readings on tick",
			before: func(z *ZH07i) {
				mk := z.transport.(*mockTransportProvider)

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = sampleInitiativePayload[0]
					}).
					Return(1, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleInitiativePayload[1:4])
					}).
					Return(3, nil).
					Once()

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleInitiativePayload[4:])
					}).
					Return(28, nil).
					Once()
			},
			checks: checkZH07iRun(
				checkZH07iRunError(""),
				checkReadValues(84, 110, 124),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			s := newZH07i(&Config{
				Transport: new(mockTransportProvider),
				Interval:  10 * time.Millisecond,
			})

			if tt.before != nil {
				tt.before(s)
			}

			ch := s.Run(ctx)

			var (
				wg  sync.WaitGroup
				got *domain.ReadingEvent
			)

			wg.Go(func() {
				timeout := time.After(1 * time.Second)
				for {
					select {
					case reading, ok := <-ch:
						if !ok {
							return // channel closed as expected
						}
						got = reading
					case <-timeout:
						cancel()
						t.Fatal("timed out waiting for channel to close")
					}
				}
			})

			go func() {
				time.Sleep(20 * time.Millisecond)
				s.Stop()
			}()

			wg.Wait()

			for _, c := range tt.checks {
				c(t, got)
			}
		})
	}
}

func TestZH07i_getChecksum(t *testing.T) {
	s := newZH07i(&Config{Transport: new(mockTransportProvider)})
	s.data = sampleInitiativePayload
	r := s.getChecksum()

	assert.Equal(t, bytesToUint16BE([]byte{0x03, 0x27}), r)
}

// TODO: check checksum algorithm for interactive mode
// func TestZH07i_CalculateChecksum(t *testing.T) {
// 	s := newZH07i(&Config{Transport: new(mockTransportProvider)})
// 	s.data = sampleInitiativePayload

// 	// Act
// 	r := s.CalculateChecksum()

// 	assert.Equal(t, bytesToUint16BE([]byte{0x03, 0x27}), r)
// }
