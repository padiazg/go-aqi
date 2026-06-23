package serial

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

// errorReader simulates an error at the reader
type errorReader struct{ err error }

func (e *errorReader) Read([]byte) (int, error) { return 0, e.err }

// errorWriter simulates an error at the writer
type errorWriter struct{ err error }

func (e *errorWriter) Write([]byte) (int, error) { return 0, e.err }

// check helpers
type checkFn func(t *testing.T, trn *SerialTransport, writeBuf io.Writer, got []byte, err error)

var (
	check = func(fns ...checkFn) []checkFn { return fns }

	checkError = func(want bool) checkFn {
		return func(t *testing.T, _ *SerialTransport, _ io.Writer, _ []byte, err error) {
			t.Helper()
			if want {
				assert.NotNil(t, err, "hasError: error expected, none produced")
			} else {
				assert.Nil(t, err, "hasError = [+%v], no error expected")
			}
		}
	}

	checkWrite = func(want []byte) checkFn {
		return func(t *testing.T, trn *SerialTransport, writeBuf io.Writer, _ []byte, err error) {
			t.Helper()
			got := writeBuf.(*bytes.Buffer).Bytes()
			assert.True(t, bytes.Equal(got, want))
		}
	}

	checkReader = func(want []byte) checkFn {
		return func(t *testing.T, trn *SerialTransport, _ io.Writer, got []byte, err error) {
			t.Helper()
			assert.True(t, bytes.Equal(got, want))
		}
	}
)

func newTransport(t *testing.T, readBuf io.Reader, writeBuf io.Writer, prefillSize int) (*SerialTransport, io.Writer) {
	t.Helper()

	if readBuf == nil {
		readBuf = bytes.NewBuffer([]byte{})
	}

	if writeBuf == nil {
		writeBuf = new(bytes.Buffer)
	}

	var bw *bufio.Writer

	if prefillSize > 0 {
		bw = bufio.NewWriterSize(writeBuf, prefillSize)
		_, err := bw.Write(make([]byte, prefillSize))
		if err != nil {
			t.Fatal("Writting prefill")
		}
	} else {
		bw = bufio.NewWriter(writeBuf)
	}

	rw := bufio.NewReadWriter(
		bufio.NewReader(readBuf),
		bw,
	)

	return New(rw), writeBuf
}

func TestSerialTransport_Write(t *testing.T) {
	tests := []struct {
		name        string
		checks      []checkFn
		out         []byte
		prefillSize int
		writeBuf    io.Writer
	}{
		{
			name: "success",
			out:  []byte{0x7E, 0x00, 0x01},
			checks: check(
				checkError(false),
				checkWrite([]byte{0x7E, 0x00, 0x01}),
			),
		},
		{
			name: "fail- writes empty slice",
			out:  []byte{},
			checks: check(
				checkError(false),
				checkWrite([]byte{}),
			),
		},
		{
			name:        "fail - write error",
			out:         []byte{0x01, 0x2},
			prefillSize: 1,
			writeBuf:    &errorWriter{err: io.ErrClosedPipe},
			checks: check(
				checkError(true),
			),
		},
		{
			name:     "fail - write/flush error",
			out:      []byte{0x01, 0x2},
			writeBuf: &errorWriter{err: io.ErrClosedPipe},
			checks: check(
				checkError(true),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, writeBuf := newTransport(t, nil, tt.writeBuf, tt.prefillSize)
			err := transport.Write(tt.out)

			for _, c := range tt.checks {
				c(t, transport, writeBuf, nil, err)
			}
		})
	}
}

func TestSerialTransport_Read(t *testing.T) {

	sample := []byte{0x01, 0x02, 0x30}

	tests := []struct {
		name    string
		checks  []checkFn
		readBuf io.Reader
		full    bool
		before  func(t *SerialTransport, readBuf io.Reader)
	}{
		{
			name:    "fail - reader error",
			readBuf: &errorReader{err: io.ErrUnexpectedEOF},
			checks: check(
				checkError(true),
			),
		},
		{
			name:    "fail - reader error - full",
			readBuf: &errorReader{err: io.ErrUnexpectedEOF},
			full:    true,
			checks: check(
				checkError(true),
			),
		},
		{
			name:    "fail - reader error - full",
			readBuf: bytes.NewBuffer([]byte{0x01, 0x02}),
			full:    true,
			checks: check(
				checkError(true),
			),
		},
		{
			name:    "success",
			readBuf: bytes.NewBuffer(sample),
			checks: check(
				checkError(false),
				checkReader(sample),
			),
		},
		{
			name:    "success - full",
			readBuf: bytes.NewBuffer(sample),
			full:    true,
			checks: check(
				checkError(false),
				checkReader(sample),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, _ := newTransport(t, tt.readBuf, nil, 0)

			got := make([]byte, 3)
			count, err := transport.Read(got, tt.full)

			for _, c := range tt.checks {
				c(t, transport, nil, got[:count], err)
			}
		})
	}
}
