package zh07

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_bytesToUint16BE(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{
			name: "case 0",
			data: []byte{0x00, 0x00},
			want: 0,
		},
		{
			name: "case 1",
			data: []byte{0x00, 0x01},
			want: 1,
		},
		{
			name: "case 255",
			data: []byte{0x00, 0xff},
			want: 255,
		},
		{
			name: "case 256",
			data: []byte{0x01, 0x00},
			want: 256,
		},
		{
			name: "case 512",
			data: []byte{0x02, 0x00},
			want: 512,
		},
		{
			name: "case 65535",
			data: []byte{0xff, 0xff},
			want: 65535,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := bytesToUint16BE(tt.data)
			assert.Equal(t, tt.want, r)
		})
	}
}
