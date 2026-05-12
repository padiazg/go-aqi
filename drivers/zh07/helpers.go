package zh07

import (
	"fmt"
)

// calculateChecksum computes the checksum for sensor data validation.
func calculateChecksum(d *[]byte) int {
	var checksum byte
	for _, v := range (*d)[1 : len(*d)-1] {
		checksum += v
	}
	return int((^checksum) + 1)
}

// byteToInt converts 2 bytes to int in big-endian format.
func bytesToUint16BE(data []byte) int {
	return int(data[1]) + (int(data[0]) << 8)
}

// toHex formats byte slice as hexadecimal string for debugging.
func toHex(data []byte) string {
	var result = ""

	for _, c := range data {
		result = fmt.Sprintf("%s%#02x ", result, c)
	}

	return result
}
