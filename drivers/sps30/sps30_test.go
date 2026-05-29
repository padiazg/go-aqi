package sps30

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/padiazg/go-aqi/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	sampleReadPayload = []byte{
		// Mass Concentration PM1.0 [µg/m³]
		0x40, 0xc2, // Upper two bytes (0,1)
		0x2d,       // Checksum for bytes 0, 1
		0x5f, 0xce, // Lower two bytes (3, 4)
		0xa7, // Checksum for bytes 3, 4
		// Mass Concentration PM2.5 [µg/m³]
		0x41, 0x86, // Upper two bytes (6,7)
		0x20,       // Checksum for bytes 6, 7
		0x4f, 0xaf, // Lower two bytes (9, 10)
		0x43, // Checksum for bytes 9, 10
		// Mass Concentration PM4.0 [µg/m³]
		0x41, 0xc9, // Upper two bytes (12, 13)
		0x33,       // Checksum for bytes 12, 13
		0x6c, 0xb0, // Lower two bytes (15, 16)
		0xdf, // Checksum for bytes 15, 16
		// Mass Concentration PM10 [µg/m³]
		0x41, 0xd6, // Upper two bytes (18, 19)
		0x5e,       // Checksum for bytes 18, 19
		0xd8, 0xe1, // Lower two bytes (21, 22)
		0x82, // Checksum for bytes 21, 22
		// Number Concentration PM0.5 [#/cm³]
		0x41, 0x8c, // Upper two bytes (24, 25)
		0xfb,       // Checksum for bytes 24, 25
		0x6a, 0x52, // Lower two bytes (27, 28)
		0x26, // Checksum for bytes 27, 28
		// Number Concentration PM1.0 [#/cm³]
		0x42, 0x11, // Upper two bytes (30, 31)
		0xa3,       // Checksum for bytes 30, 31
		0x47, 0x09, // Lower two bytes (33, 34)
		0x2e, // Checksum for bytes 33, 34
		// Number Concentration PM2.5 [#/cm³]
		0x42, 0x3f, // Upper two bytes (36, 27)
		0x3a,       // Checksum for bytes 36, 37
		0x37, 0xa1, // Lower two bytes (39, 40)
		0x50, // Checksum for bytes 39, 40
		// Number Concentration PM4.0 [#/cm³]
		0x42, 0x48, // Upper two bytes (42, 43)
		0x55,       // Checksum for bytes 42, 43
		0x8c, 0xb8, // Lower two bytes (45, 46)
		0x10, // Checksum for bytes 45, 46
		// Number Concentration PM10 [#/cm³]
		0x42, 0x49, // Upper two bytes (48, 49)
		0x64,       // Checksum for bytes 48, 49
		0xe4, 0x68, // Lower two bytes (51, 52)
		0x76, // Checksum for bytes 51, 52
		// Typical Particle Size [µm]
		0x3f, 0xa4, // Upper two bytes (54, 55)
		0x92,       // Checksum for bytes 54, 55
		0x05, 0xf2, // Lower two bytes (57, 58)
		0x16, // Checksum for bytes 57, 58
		0x00, // Padding byte to fill 60-byte buffer
	}
)

type NewFn func(*testing.T, *SPS30, error)

var checkNew = func(fns ...NewFn) []NewFn { return fns }

func checkNewError(want string) NewFn {
	return func(t *testing.T, _ *SPS30, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkNewError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkNewError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkNewError mismatch")
		}
	}
}

func checkNewID(want string) NewFn {
	return func(t *testing.T, s *SPS30, err error) {
		t.Helper()
		if s != nil {
			assert.Equal(t, want, s.id, "NewID: unexpected sensor ID")
		}
	}
}

func checkNewInterval(want time.Duration) NewFn {
	return func(t *testing.T, s *SPS30, err error) {
		t.Helper()
		if s != nil {
			assert.Equal(t, want, s.interval, "NewInterval: unexpected interval")
		}
	}
}

func TestSPS30_New(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []NewFn
	}{
		{
			name:   "fail - nil config defaults to empty then missing transport",
			config: nil,
			checks: checkNew(
				checkNewError("transport not provided"),
			),
		},
		{
			name:   "fail - nil transport",
			config: &Config{},
			checks: checkNew(
				checkNewError("transport not provided"),
			),
		},
		{
			name: "success - transport provided, default ID",
			config: &Config{
				Transport: new(mockTransportProvider),
			},
			checks: checkNew(
				checkNewError(""),
				checkNewID("sps30"),
			),
		},
		{
			name: "success - transport with custom ID",
			config: &Config{
				Transport: new(mockTransportProvider),
				ID:        "sensor-01",
			},
			checks: checkNew(
				checkNewError(""),
				checkNewID("sensor-01"),
			),
		},
		{
			name: "success - transport with interval",
			config: &Config{
				Transport: new(mockTransportProvider),
				ID:        "sensor-02",
				Interval:  5 * time.Second,
			},
			checks: checkNew(
				checkNewError(""),
				checkNewID("sensor-02"),
				checkNewInterval(5*time.Second),
			),
		},
		{
			name: "success - transport with zero interval",
			config: &Config{
				Transport: new(mockTransportProvider),
			},
			checks: checkNew(
				checkNewError(""),
				checkNewID("sps30"),
				checkNewInterval(0),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(tt.config)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}

// ---- Read tests ----

type checkSPS30ReadFn func(*testing.T, *domain.ReadingEvent)

var checkSPS30Read = func(fns ...checkSPS30ReadFn) []checkSPS30ReadFn { return fns }

type checkReadFn = checkSPS30ReadFn

func checkReadError(want string) checkSPS30ReadFn {
	return func(t *testing.T, re *domain.ReadingEvent) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, re.Err, "checkReadError: expected no error, got %v", re.Err)
			return
		}
		if assert.Error(t, re.Err, "checkReadError: expected error") {
			assert.Containsf(t, re.Err.Error(), want, "checkReadError mismatch")
		}
	}
}

func checkReadNil(t *testing.T, re *domain.ReadingEvent) {
	t.Helper()
	assert.Nil(t, re.Reading, "checkReadNil: expected reading to be nil")
}

func checkReadValues(pm1, pm25, pm4, pm10, n05, n1, n25, n4, n10, tps float32) checkSPS30ReadFn {
	return func(t *testing.T, re *domain.ReadingEvent) {
		t.Helper()
		if assert.NotNilf(t, re.Reading, "checkReadValues: reading is nil") {
			assert.InDeltaf(t, pm1, re.Reading.MassPM1, 0.01, "MassPM1 mismatch")
			assert.InDeltaf(t, pm25, re.Reading.MassPM25, 0.01, "MassPM25 mismatch")
			assert.InDeltaf(t, pm4, re.Reading.MassPM4, 0.01, "MassPM4 mismatch")
			assert.InDeltaf(t, pm10, re.Reading.MassPM10, 0.01, "MassPM10 mismatch")
			assert.InDeltaf(t, n05, re.Reading.NumberPM05, 0.01, "NumberPM05 mismatch")
			assert.InDeltaf(t, n1, re.Reading.NumberPM1, 0.01, "NumberPM1 mismatch")
			assert.InDeltaf(t, n25, re.Reading.NumberPM25, 0.01, "NumberPM25 mismatch")
			assert.InDeltaf(t, n4, re.Reading.NumberPM4, 0.01, "NumberPM4 mismatch")
			assert.InDeltaf(t, n10, re.Reading.NumberPM10, 0.01, "NumberPM10 mismatch")
			assert.InDeltaf(t, tps, re.Reading.TypicalParticleSize, 0.01, "TypicalParticleSize mismatch")
		}
	}
}

func TestSPS30_Read(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkSPS30ReadFn
		before func(*SPS30) *SPS30
	}{
		{
			name: "success - measurement read",
			before: func(_ *SPS30) *SPS30 {
				m := new(mockTransportProvider)
				s, _ := New(&Config{
					Transport: m,
				})

				m.On("Write", cmdStartMeasurement).Return(nil)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[0] = 0x00
						buf[1] = 0x01
						buf[2] = 0x00
					}).
					Return(3, nil).Once()

				m.On("Write", cmdReadMeasurement).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 60 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						copy(buf, sampleReadPayload)
					}).
					Return(60, nil).Once()

				m.On("Write", cmdStopMeasurement).Return(nil)
				return s
			},
			checks: checkSPS30Read(
				checkReadError(""),
				checkReadValues(6.074195, 16.788908, 25.17807, 26.855898, 17.551914, 36.31937, 47.804325, 50.13742, 50.473053, 1.2814314),
			),
		},
		{
			name: "fail - start measurement write error",
			before: func(_ *SPS30) *SPS30 {
				m := new(mockTransportProvider)
				s, _ := New(&Config{
					Transport: m,
				})

				m.On("Write", cmdStartMeasurement).Return(errors.New("write failed"))
				m.On("Write", cmdStopMeasurement).Return(nil)
				return s
			},
			checks: checkSPS30Read(
				checkReadError("StartMeasurement"),
				checkReadNil,
			),
		},
		{
			name: "fail - is data ready read error",
			before: func(_ *SPS30) *SPS30 {
				m := new(mockTransportProvider)
				s, _ := New(&Config{
					Transport: m,
				})

				m.On("Write", cmdStartMeasurement).Return(nil)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.Anything, false).Return(0, errors.New("transport read error"))

				m.On("Write", cmdStopMeasurement).Return(nil)
				return s
			},
			checks: checkSPS30Read(
				checkReadError("is data ready"),
				checkReadNil,
			),
		},
		{
			name: "fail - not ready after retries",
			before: func(_ *SPS30) *SPS30 {
				m := new(mockTransportProvider)
				s, _ := New(&Config{
					Transport: m,
				})

				m.On("Write", cmdStartMeasurement).Return(nil)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.Anything, false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[0] = 0x00
						buf[1] = 0x00
						buf[2] = 0x00
					}).
					Return(3, nil)

				m.On("Write", cmdStopMeasurement).Return(nil)
				return s
			},
			checks: checkSPS30Read(
				checkReadError("not ready"),
				checkReadNil,
			),
		},
		{
			name: "fail - read measurement write error",
			before: func(_ *SPS30) *SPS30 {
				m := new(mockTransportProvider)
				s, _ := New(&Config{
					Transport: m,
				})

				m.On("Write", cmdStartMeasurement).Return(nil)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.Anything, false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[1] = 0x01
					}).
					Return(3, nil)

				m.On("Write", cmdReadMeasurement).Return(errors.New("transport write error"))

				m.On("Write", cmdStopMeasurement).Return(nil)
				return s
			},
			checks: checkSPS30Read(
				checkReadError("read"),
				checkReadNil,
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := new(SPS30)
			if tt.before != nil {
				s = tt.before(s)
			}
			r := s.Read(context.Background())
			for _, c := range tt.checks {
				c(t, r)
			}
		})
	}
}

func checkRunError(want string) checkReadFn {
	return func(t *testing.T, re *domain.ReadingEvent) {
		t.Helper()
		if want == "" && re.Err != nil {
			assert.NoErrorf(t, re.Err, "checkRunError: expected no error, got %v", re.Err)
			return
		}
		if want != "" && assert.NotNil(t, re.Err) {
			if assert.Errorf(t, re.Err, "checkRunError: expected error %q", want) {
				assert.Containsf(t, re.Err.Error(), want, "checkRunError mismatch")
			}
		}
	}
}

var checkReadRun = func(fns ...checkReadFn) []checkReadFn { return fns }

func TestSPS30_Run(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkReadFn
		before func(*SPS30)
		after  func(*SPS30, context.CancelFunc)
	}{
		{
			name: "emits valid readings on tick",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdStartMeasurement).Return(nil).Maybe()
				m.On("Write", cmdIsDataReady).Return(nil).Maybe()
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[1] = 0x01
					}).
					Return(3, nil).Maybe()
				m.On("Write", cmdReadMeasurement).Return(nil).Maybe()
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 60 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						copy(buf, sampleReadPayload)
					}).
					Return(60, nil).Maybe()
				m.On("Write", cmdStopMeasurement).Return(nil).Maybe()
			},
			after: func(s *SPS30, _ context.CancelFunc) {
				s.Stop()
			},
			checks: checkReadRun(
				checkRunError(""),
				checkReadValues(6.074195, 16.788908, 25.17807, 26.855898, 17.551914, 36.31937, 47.804325, 50.13742, 50.473053, 1.2814314),
			),
		},
		{
			name: "closes channel on context cancellation",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdStopMeasurement).Return(nil).Maybe()
				s.interval = 500 * time.Millisecond
			},
			after: func(s *SPS30, cancel context.CancelFunc) {
				cancel()
			},
			checks: checkReadRun(),
		},
		{
			name: "emits error readings on transport failure",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdStartMeasurement).Return(errors.New("write-error")).Maybe()
				m.On("Write", cmdStopMeasurement).Return(nil).Maybe()
			},
			after: func(s *SPS30, _ context.CancelFunc) {
				s.Stop()
			},
			checks: checkReadRun(
				checkRunError("StartMeasurement"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			m := new(mockTransportProvider)
			s, err := New(&Config{
				Transport: m,
				ID:        "sps30",
				Interval:  10 * time.Millisecond,
			})
			assert.NoError(t, err)

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

type checkSPS30IsDataReadyFn func(*testing.T, bool, error)

var checkSPS30IsDataReady = func(fns ...checkSPS30IsDataReadyFn) []checkSPS30IsDataReadyFn { return fns }

func checkIsDataReadyError(want string) checkSPS30IsDataReadyFn {
	return func(t *testing.T, _ bool, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkIsDataReadyError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkIsDataReadyError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkIsDataReadyError mismatch")
		}
	}
}

func checkIsDataReadyReady(want bool) checkSPS30IsDataReadyFn {
	return func(t *testing.T, got bool, err error) {
		t.Helper()
		assert.Equal(t, want, got, "checkIsDataReadyReady: ready mismatch")
	}
}

func TestSPS30_IsDataReady(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkSPS30IsDataReadyFn
		before func(*SPS30)
	}{
		{
			name: "fail - sendCommand error",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdIsDataReady).Return(fmt.Errorf("write-error"))
			},
			checks: checkSPS30IsDataReady(
				checkIsDataReadyError("sendCommand sending code"),
			),
		},
		{
			name: "no new measurements available",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[0] = 0x00
						buf[1] = 0x00
						buf[2] = 0x81
					}).
					Return(3, nil)
			},
			checks: checkSPS30IsDataReady(
				checkIsDataReadyError(""),
				checkIsDataReadyReady(false),
			),
		},
		{
			name: "new measurements ready to read",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[0] = 0x00
						buf[1] = 0x01
						buf[2] = 0xb0
					}).
					Return(3, nil)
			},
			checks: checkSPS30IsDataReady(
				checkIsDataReadyError(""),
				checkIsDataReadyReady(true),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTransportProvider)
			s, err := New(&Config{
				Transport: m,
				ID:        "sps30",
			})
			assert.NoError(t, err)

			if tt.before != nil {
				tt.before(s)
			}

			r, err := s.IsDataReady()
			for _, c := range tt.checks {
				c(t, r, err)
			}

			m.AssertExpectations(t)
		})
	}
}

type checkSPS30StopMeasurementFn func(*testing.T, error)

var checkSPS30StopMeasurement = func(fns ...checkSPS30StopMeasurementFn) []checkSPS30StopMeasurementFn { return fns }

func checkStopMeasurementError(want string) checkSPS30StopMeasurementFn {
	return func(t *testing.T, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkStopMeasurementError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkStopMeasurementError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkStopMeasurementError mismatch")
		}
	}
}
func TestSPS30_StopMeasurement(t *testing.T) {
	tests := []struct {
		name   string
		checks []checkSPS30StopMeasurementFn
		before func(*SPS30)
	}{
		{
			name: "fail - transport write error",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdStopMeasurement).Return(fmt.Errorf("write-error"))
			},
			checks: checkSPS30StopMeasurement(
				checkStopMeasurementError("StopMeasurement"),
			),
		},
		{
			name: "success",
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdStopMeasurement).Return(nil)
			},
			checks: checkSPS30StopMeasurement(
				checkStopMeasurementError(""),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTransportProvider)
			s, err := New(&Config{
				Transport: m,
				ID:        "sps30",
			})
			assert.NoError(t, err)

			if tt.before != nil {
				tt.before(s)
			}

			err = s.StopMeasurement()
			for _, c := range tt.checks {
				c(t, err)
			}

			m.AssertExpectations(t)
		})
	}
}

type checkSPS30sendCommandFn func(*testing.T, error)

var checkSPS30sendCommand = func(fns ...checkSPS30sendCommandFn) []checkSPS30sendCommandFn { return fns }

func checksendCommandError(want string) checkSPS30sendCommandFn {
	return func(t *testing.T, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checksendCommandError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checksendCommandError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checksendCommandError mismatch")
		}
	}
}
func TestSPS30_sendCommand(t *testing.T) {
	tests := []struct {
		name string
		addr []byte
		// in     []byte
		checks []checkSPS30sendCommandFn
		before func(*SPS30)
	}{
		{
			name: "fail - transport write error",
			addr: cmdIsDataReady,
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdIsDataReady).Return(fmt.Errorf("write-error"))
			},
			checks: checkSPS30sendCommand(
				checksendCommandError("sendCommand sending code"),
			),
		},
		{
			name: "fail - transport read error",
			addr: cmdIsDataReady,
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.Anything, false).Return(0, fmt.Errorf("read-error"))
			},
			checks: checkSPS30sendCommand(
				checksendCommandError("sendCommand reading response"),
			),
		},
		{
			name: "success",
			addr: cmdIsDataReady,
			before: func(s *SPS30) {
				m := s.transport.(*mockTransportProvider)
				m.On("Write", cmdIsDataReady).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[0] = 0x00
						buf[1] = 0x01
						buf[2] = 0xb0
					}).
					Return(3, nil)
			},
			checks: checkSPS30sendCommand(
				checksendCommandError(""),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTransportProvider)
			s, err := New(&Config{
				Transport: m,
				ID:        "sps30",
			})
			assert.NoError(t, err)

			if tt.before != nil {
				tt.before(s)
			}

			in := make([]byte, 3)
			err = s.sendCommand(tt.addr, in)
			for _, c := range tt.checks {
				c(t, err)
			}

			m.AssertExpectations(t)
		})
	}
}
