package serial

import (
	"bufio"
	"io"

	"github.com/padiazg/go-aqi/domain"
)

var _ domain.TransportProvider = (*SerialTransport)(nil)

type SerialTransport struct {
	rw *bufio.ReadWriter
}

// New creates a new serial transport instance.
func New(rw *bufio.ReadWriter) *SerialTransport {
	return &SerialTransport{
		rw: rw,
	}
}

// Read reads data from the serial port.
func (s *SerialTransport) Read(in []byte, full bool) (int, error) {
	if full {
		return io.ReadFull(s.rw, in)
	}
	return s.rw.Read(in)
}

// Write writes data to the serial port.
func (s *SerialTransport) Write(out []byte) error {
	if _, err := s.rw.Write(out); err != nil { // send command
		return err
	}

	return s.rw.Flush() // flush write buffer
}
