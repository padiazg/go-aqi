package i2c

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type checkI2CTransportReadFn func(*testing.T, []byte, int, error)

var checkI2CTransportRead = func(fns ...checkI2CTransportReadFn) []checkI2CTransportReadFn { return fns }

func checkReadError(want string) checkI2CTransportReadFn {
	return func(t *testing.T, _ []byte, _ int, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkReadError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkReadError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkReadError mismatch")
		}
	}
}

func checkBytes(want []byte) checkI2CTransportReadFn {
	return func(t *testing.T, in []byte, i int, err error) {
		t.Helper()
		assert.Equal(t, want, in)
	}
}
func TestI2CTransport_Read(t *testing.T) {
	tests := []struct {
		name   string
		in     []byte
		checks []checkI2CTransportReadFn
		before func(*I2CTransport)
	}{
		{
			name: "success",
			in:   make([]byte, 3),
			before: func(ic *I2CTransport) {
				ic.i2c.(*mockI2CBus).
					On("ReadBytes", mock.Anything).
					Run(func(args mock.Arguments) {
						in := args.Get(0).([]byte)
						in[0] = 0x01
						in[1] = 0x02
					}).
					Return(1, nil)
			},
			checks: checkI2CTransportRead(
				checkReadError(""),
				checkBytes([]byte{0x01, 0x02, 0x00}),
			),
		},
		{
			name: "fail",
			in:   make([]byte, 3),
			before: func(ic *I2CTransport) {
				ic.i2c.(*mockI2CBus).
					On("ReadBytes", mock.Anything).
					Return(0, fmt.Errorf("test-error"))
			},
			checks: checkI2CTransportRead(
				checkReadError("test-error"),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockI2CBus)
			s := New(mock)
			if tt.before != nil {
				tt.before(s)
			}
			r, err := s.Read(tt.in, false)
			for _, c := range tt.checks {
				c(t, tt.in, r, err)
			}
		})
	}
}

type checkI2CTransportWriteFn func(*testing.T, error)

var checkI2CTransportWrite = func(fns ...checkI2CTransportWriteFn) []checkI2CTransportWriteFn { return fns }

func checkWriteError(want string) checkI2CTransportWriteFn {
	return func(t *testing.T, err error) {
		t.Helper()
		if want == "" {
			assert.NoErrorf(t, err, "checkWriteError: expected no error, got %v", err)
			return
		}
		if assert.Errorf(t, err, "checkWriteError: expected error %q", want) {
			assert.Containsf(t, err.Error(), want, "checkWriteError mismatch")
		}
	}
}
func TestI2CTransport_Write(t *testing.T) {
	tests := []struct {
		name   string
		out    []byte
		checks []checkI2CTransportWriteFn
		before func(*I2CTransport)
	}{
		{
			name: "success",
			out:  []byte{0x00, 0x01},
			before: func(ic *I2CTransport) {
				ic.i2c.(*mockI2CBus).
					On("WriteBytes", mock.Anything).
					Return(2, nil)
			},
			checks: checkI2CTransportWrite(
				checkWriteError(""),
			),
		},
		{
			name: "fail",
			out:  []byte{0x00, 0x01},
			before: func(ic *I2CTransport) {
				ic.i2c.(*mockI2CBus).
					On("WriteBytes", mock.Anything).
					Return(0, fmt.Errorf("test-error"))
			},
			checks: checkI2CTransportWrite(
				checkWriteError("test-error"),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mock := new(mockI2CBus)
			s := New(mock)
			if tt.before != nil {
				tt.before(s)
			}
			err := s.Write(tt.out)
			for _, c := range tt.checks {
				c(t, err)
			}
		})
	}
}
