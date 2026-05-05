package zh07

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
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

func TestZH07q_Read(t *testing.T) {
	var (
		tests = []struct {
			name     string
			response []byte
			checks   []checkFn
			before   func(z *ZH07q)
		}{
			{
				name:     "success",
				response: sampleQAPayload,
				checks: check(
					checkError(false),
					pm(0x85, 0x96, 0x65),
				),
			},
			{
				name:     "fail-checksum-mismatch",
				response: sampleQABadChecksum,
				checks: check(
					checkError(true),
				),
			},
			{
				name: "fail-sendcommand",
				before: func(z *ZH07q) {
					z.writeAndRead = func(_ *bufio.ReadWriter, _ []byte) ([]byte, error) {
						return nil, fmt.Errorf("test error from sendCommand")
					}
				},
				checks: check(
					checkError(true),
				),
			},
		}
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				b = &bytes.Buffer{}
				z = NewZH07q(&Config{
					RW: bufio.NewReadWriter(bufio.NewReader(b), bufio.NewWriter(b)),
				})
				got *Reading
				err error
			)

			if tt.before != nil {
				tt.before(z)
			}

			go dummyCommandResponder(t, z.rw, commandQuery, tt.response)

			got, err = z.Read()
			for _, c := range tt.checks {
				c(t, got, err)
			}
		})
	}
}

func TestZH07q_getChecksum(t *testing.T) {
	var z *ZH07q = &ZH07q{data: sampleQAPayload}
	if cs := z.getChecksum(); cs != 0xFA {
		t.Errorf("TestGetChecksumQA, got %d, expected %d", cs, checksum)
	}
}
