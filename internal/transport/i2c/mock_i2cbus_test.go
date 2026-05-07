package i2c

import (
	"github.com/stretchr/testify/mock"
)

// mockI2CBus implements .I2CBus for testing.
type mockI2CBus struct {
	mock.Mock
}

func (m *mockI2CBus) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockI2CBus) ReadBytes(buf []byte) (int, error) {
	args := m.Called(buf)
	r0 := args.Get(0).(int)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) ReadRegBytes(reg byte, n int) ([]byte, int, error) {
	args := m.Called(reg, n)
	r0, _ := args.Get(0).([]byte)
	r1 := args.Get(1).(int)
	r2 := args.Error(2)
	return r0, r1, r2
}

func (m *mockI2CBus) ReadRegS16BE(reg byte) (int16, error) {
	args := m.Called(reg)
	r0 := args.Get(0).(int16)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) ReadRegS16LE(reg byte) (int16, error) {
	args := m.Called(reg)
	r0 := args.Get(0).(int16)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) ReadRegU16BE(reg byte) (uint16, error) {
	args := m.Called(reg)
	r0 := args.Get(0).(uint16)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) ReadRegU16LE(reg byte) (uint16, error) {
	args := m.Called(reg)
	r0 := args.Get(0).(uint16)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) ReadRegU8(reg byte) (byte, error) {
	args := m.Called(reg)
	r0 := args.Get(0).(byte)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) WriteBytes(buf []byte) (int, error) {
	args := m.Called(buf)
	r0 := args.Get(0).(int)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockI2CBus) WriteRegS16BE(reg byte, value int16) error {
	args := m.Called(reg, value)
	return args.Error(0)
}

func (m *mockI2CBus) WriteRegS16LE(reg byte, value int16) error {
	args := m.Called(reg, value)
	return args.Error(0)
}

func (m *mockI2CBus) WriteRegU16BE(reg byte, value uint16) error {
	args := m.Called(reg, value)
	return args.Error(0)
}

func (m *mockI2CBus) WriteRegU16LE(reg byte, value uint16) error {
	args := m.Called(reg, value)
	return args.Error(0)
}

func (m *mockI2CBus) WriteRegU8(reg byte, value byte) error {
	args := m.Called(reg, value)
	return args.Error(0)
}
