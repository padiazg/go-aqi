package zh07

// bytesToUint16BE converts 2 bytes to uint16 in big-endian format.
func bytesToUint16BE(data []byte) int {
	return int(data[1]) + (int(data[0]) << 8)
}
