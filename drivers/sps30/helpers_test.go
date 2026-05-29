package sps30

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_crc8Checksum(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want byte
	}{
		{
			name: "empty data returns initial value",
			data: nil,
			want: 0xFF,
		},
		{
			name: "single zero byte",
			data: []byte{0x00},
			want: 0xAC,
		},
		{
			name: "single 0x42 byte",
			data: []byte{0x42},
			want: 0xF3,
		},
		{
			name: "Mass Concentration PM1.0 - upper bytes",
			data: sampleReadPayload[0:2],
			want: sampleReadPayload[2],
		},
		{
			name: "Mass Concentration PM1.0 - lower bytes",
			data: sampleReadPayload[3:5],
			want: sampleReadPayload[5],
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := crc8Checksum(tt.data)
			assert.Equal(t, tt.want, r)
		})
	}
}

func Test_bytesToFloat32(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float32
	}{
		{
			name: "zero",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: 0,
		},
		{
			name: "one",
			data: []byte{0x3F, 0x80, 0x00, 0x00},
			want: 1,
		},
		{
			name: "two",
			data: []byte{0x40, 0x00, 0x00, 0x00},
			want: 2,
		},
		{
			name: "three",
			data: []byte{0x40, 0x40, 0x00, 0x00},
			want: 3,
		},
		{
			name: "point three two",
			data: []byte{0x3E, 0xA3, 0xD7, 0x0A},
			want: 0.32,
		},
		{
			name: "one with bit pattern 0x00000001",
			data: []byte{0x00, 0x00, 0x00, 0x01},
			want: 1e-45,
		},
		{
			name: "SPS30 PM1 raw 0x00005400",
			data: []byte{0x00, 0x00, 0x54, 0x00},
			want: 3.0134e-41,
		},
		{
			name: "max finite float32",
			data: []byte{0x7F, 0x7F, 0xFF, 0xFF},
			want: math.MaxFloat32,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := bytesToFloat32(tt.data)
			assert.Equal(t, tt.want, r)
		})
	}
}
