package zh07

// calculateChecksum computes the checksum for sensor data validation.
func calculateChecksum(d *[]byte) int {
	var checksum byte
	for _, v := range (*d)[1 : len(*d)-1] {
		checksum += v
	}
	return int((^checksum) + 1)
}

// bytesToUint16BE converts 2 bytes to uint16 in big-endian format.
func bytesToUint16BE(data []byte) int {
	return int(data[1]) + (int(data[0]) << 8)
}
