package zh07

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/padiazg/go-aqi/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	sampleQAPayload = []byte{
		0xFF,       // starting
		0x86,       // command
		0x00, 0x85, // pm2.5
		0x00, 0x96, // pm10
		0x00, 0x65, // pm1.0
		0xFA, // checksum
	}
	sampleQABadChecksum = []byte{
		0xFF,       // starting
		0x86,       // command
		0x00, 0x85, // pm2.5
		0x00, 0x96, // pm10
		0x00, 0x65, // pm1.0
		0xFB, // checksum
	}
)

type newZH07qFn func(*testing.T, *ZH07q)

var checknewZH07q = func(fns ...newZH07qFn) []newZH07qFn { return fns }

func checkTypeZH07q(t *testing.T, sp *ZH07q) {
	if assert.NotNil(t, sp) {
		assert.IsType(t, &ZH07q{}, sp)
	}
}

func checkConfigZH07q(t *testing.T, sp *ZH07q) {
	assert.NotNil(t, sp.transport)
	assert.NotNil(t, sp.interval)
	if assert.NotNil(t, sp.data) {
		assert.Equal(t, 9, len(sp.data))
	}
}

func Test_newZH07q(t *testing.T) {
	checkID := func(want string) newZH07qFn {
		return func(t *testing.T, z *ZH07q) {
			t.Helper()
			assert.Equalf(t, want, z.id, "checkID id = %s, expected: %s", z.id, want)
		}
	}

	tests := []struct {
		name   string
		config *Config
		checks []newZH07qFn
	}{
		{
			name: "success",
			checks: checknewZH07q(
				checkTypeZH07q,
				checkConfigZH07q,
				checkID("zh07-q"),
			),
		},
		{
			name:   "custom-id",
			config: &Config{ID: "custom-id"},
			checks: checknewZH07q(
				checkTypeZH07q,
				checkConfigZH07q,
				checkID("custom-id"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := newZH07q(tt.config)
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

type checkZH07qInitFn func(*testing.T, error)

var checkZH07qInit = func(fns ...checkZH07qInitFn) []checkZH07qInitFn { return fns }

func checkZH07qInitError(want string) checkZH07qInitFn {
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

func TestZH07q_Init(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkZH07qInitFn
		before func(*ZH07q)
	}{
		{
			name: "success",
			before: func(z *ZH07q) {
				z.transport.(*mockTransportProvider).
					On("Write", mock.Anything).
					Return(nil)
			},
			checks: checkZH07qInit(
				checkZH07qInitError(""),
			),
		},
		{
			name: "fail",
			before: func(z *ZH07q) {
				z.transport.(*mockTransportProvider).
					On("Write", mock.Anything).
					Return(fmt.Errorf("transport-write-error"))
			},
			checks: checkZH07qInit(
				checkZH07qInitError("transport-write-error"),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := newZH07q(&Config{Transport: new(mockTransportProvider)})

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

var checkZH07qRead = func(fns ...checkReadFn) []checkReadFn { return fns }

func TestZH07q_Read(t *testing.T) {

	checkZH07qReadError := func(want string) checkReadFn {
		return func(t *testing.T, re *domain.ReadingEvent) {
			t.Helper()
			if want == "" {
				assert.NoErrorf(t, re.Err, "checkZH07qReadError: expected no error, got %v", re.Err)
				return
			}
			if assert.Errorf(t, re.Err, "checkZH07qReadError: expected error %q", want) {
				assert.Containsf(t, re.Err.Error(), want, "checkZH07qReadError mismatch")
			}
		}
	}

	tests := []struct {
		name   string
		checks []checkReadFn
		before func(*ZH07q)
	}{
		{
			name: "fail - write error",
			before: func(z *ZH07q) {
				mk := z.transport.(*mockTransportProvider)

				mk.On("Write", mock.Anything).
					Return(fmt.Errorf("write-error"))

				mk.On("Read", mock.Anything, mock.Anything).
					Return(nil)
			},

			checks: checkZH07qRead(
				checkZH07qReadError("write-error"),
			),
		},
		{
			name: "fail - read error",
			before: func(z *ZH07q) {
				mk := z.transport.(*mockTransportProvider)

				mk.On("Write", mock.Anything).
					Return(nil)

				mk.On("Read", mock.Anything, mock.Anything).
					Return(0, fmt.Errorf("read-error"))
			},

			checks: checkZH07qRead(
				checkZH07qReadError("read-error"),
			),
		},
		{
			name: "fail - invalid read",
			before: func(z *ZH07q) {
				mk := z.transport.(*mockTransportProvider)

				mk.On("Write", mock.Anything).
					Return(nil)

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleQABadChecksum)
					}).
					Return(9, nil)
			},

			checks: checkZH07qRead(
				checkZH07qReadError("checksum mismatch"),
			),
		},
		{
			name: "success",
			before: func(z *ZH07q) {
				mk := z.transport.(*mockTransportProvider)

				mk.On("Write", mock.Anything).
					Return(nil)

				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleQAPayload)
					}).
					Return(9, nil)
			},
			checks: checkZH07qRead(
				checkZH07qReadError(""),
				checkReadValues(101.0, 133.0, 150.0),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := newZH07q(&Config{Transport: new(mockTransportProvider)})
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

// --------------- Run tests ---------------

var checkZH07qRun = func(fns ...checkReadFn) []checkReadFn { return fns }

func checkZH07qRunError(want string) checkReadFn {
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

func TestZH07q_Run(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkReadFn
		before func(*ZH07q)
		after  func(*ZH07q, context.CancelFunc)
	}{
		{
			name: "emits valid readings on tick",
			before: func(z *ZH07q) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Write", mock.Anything).Return(nil).Maybe()
				mk.On("Read", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						copy(in, sampleQAPayload)
					}).
					Return(9, nil).Maybe()
			},
			after: func(z *ZH07q, _ context.CancelFunc) {
				z.Stop()
			},
			checks: checkZH07qRun(
				checkZH07qRunError(""),
				checkReadValues(101.0, 133.0, 150.0),
			),
		},
		{
			name: "emits error readings on transport failure",
			before: func(z *ZH07q) {
				mk := z.transport.(*mockTransportProvider)
				mk.On("Write", mock.Anything).
					Return(fmt.Errorf("write-error")).Maybe()
				mk.On("Read", mock.Anything, mock.Anything).
					Return(0, nil).Maybe()
			},
			after: func(z *ZH07q, _ context.CancelFunc) {
				z.Stop()
			},
			checks: checkZH07qRun(
				checkZH07qRunError("write-error"),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := newZH07q(&Config{
				Transport: new(mockTransportProvider),
				Interval:  10 * time.Millisecond,
			})

			if tt.before != nil {
				tt.before(s)
			}

			ch := s.Run(context.Background())

			// Wait for exactly one reading from the Run loop, then stop
			// immediately. This avoids the race where a second tick fires
			// before s.Stop() cancels the context (which would cause an
			// unexpected mock call when using .Once() expectations).
			var got *domain.ReadingEvent

			select {
			case reading, ok := <-ch:
				if !ok {
					t.Fatal("channel closed unexpectedly")
				}
				got = reading
			case <-time.After(30 * time.Second):
				t.Fatal("timed out waiting for first reading")
			}

			s.Stop()

			// Confirm the Run goroutine exited by waiting for the channel
			// to close. If the channel is still open, drain leftover
			// events to unblock any concurrent sends before the test
			// finishes.
			select {
			case _, ok := <-ch:
				if ok {
					for range ch {
					}
				}
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for channel to close")
			}

			for _, c := range tt.checks {
				c(t, got)
			}

		})
	}
}

func TestZH07q_getChecksum(t *testing.T) {
	s := newZH07q(&Config{Transport: new(mockTransportProvider)})
	s.data = sampleQAPayload

	r := s.getChecksum()

	assert.Equal(t, 0xFA, r)
}

func TestZH07q_calculateChecksum(t *testing.T) {
	s := newZH07q(&Config{Transport: new(mockTransportProvider)})
	s.data = sampleQAPayload

	// Act
	r := s.calculateChecksum()

	assert.Equal(t, int(sampleQAPayload[8]), r)
}

func TestZH07q_IsReadingValid(t *testing.T) {
	tests := []struct {
		name string
		want bool
		data []byte
	}{
		{
			name: "success",
			data: sampleQAPayload,
			want: true,
		},
		{
			name: "bad checksum",
			data: sampleQABadChecksum,
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := newZH07q(&Config{Transport: new(mockTransportProvider)})
			s.data = tt.data
			r := s.IsReadingValid()
			assert.Equal(t, tt.want, r)
		})
	}
}
