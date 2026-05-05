package sps30

import (
	"encoding/binary"
	"math"

	"github.com/sigurn/crc8"
)

var (
	CRC8_SPS30 = crc8.Params{
		Poly:   0x31,
		Init:   0xFF,
		RefIn:  false,
		RefOut: false,
		XorOut: 0x00,
		Check:  0x00,
		Name:   "CRC-8/SPS30",
	}

	table = crc8.MakeTable(CRC8_SPS30)
)

func crc(data []byte) byte {
	return crc8.Checksum(data, table)
}

func byteArrayToFloat32(data []byte) float32 {
	a0 := binary.BigEndian.Uint32(data)
	a1 := math.Float32frombits(a0)
	return a1
}
