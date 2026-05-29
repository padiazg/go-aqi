package sps30

import (
	"github.com/stretchr/testify/mock"
)

// mockTransportProvider implements domain.TransportProvider for testing.
type mockTransportProvider struct {
	mock.Mock
}

func (m *mockTransportProvider) Read(in []byte, full bool) (int, error) {
	args := m.Called(in, full)
	r0 := args.Get(0).(int)
	r1 := args.Error(1)
	return r0, r1
}

func (m *mockTransportProvider) Write(out []byte) error {
	args := m.Called(out)
	return args.Error(0)
}
