package i2c

// ---------------- mock I2C
var _ I2CBus = (*mockI2C)(nil)

type mockI2C struct {
	writeBytesFn func(buf []byte) (int, error)
	readBytesFn  func(buf []byte) (int, error)
	closeFn      func() error
}

func (m *mockI2C) WriteBytes(buf []byte) (int, error) {
	if m.writeBytesFn != nil {
		return m.writeBytesFn(buf)
	}
	return len(buf), nil
}

func (m *mockI2C) ReadBytes(buf []byte) (int, error) {
	if m.readBytesFn != nil {
		return m.readBytesFn(buf)
	}
	return 0, nil
}

func (m *mockI2C) ReadRegBytes(reg byte, n int) ([]byte, int, error) {
	return make([]byte, n), n, nil
}

func (m *mockI2C) ReadRegU8(reg byte) (byte, error)       { return 0, nil }
func (m *mockI2C) WriteRegU8(reg byte, value byte) error  { return nil }
func (m *mockI2C) ReadRegU16BE(reg byte) (uint16, error)  { return 0, nil }
func (m *mockI2C) ReadRegU16LE(reg byte) (uint16, error)  { return 0, nil }
func (m *mockI2C) ReadRegS16BE(reg byte) (int16, error)   { return 0, nil }
func (m *mockI2C) ReadRegS16LE(reg byte) (int16, error)   { return 0, nil }
func (m *mockI2C) WriteRegU16BE(reg byte, v uint16) error { return nil }
func (m *mockI2C) WriteRegU16LE(reg byte, v uint16) error { return nil }
func (m *mockI2C) WriteRegS16BE(reg byte, v int16) error  { return nil }
func (m *mockI2C) WriteRegS16LE(reg byte, v int16) error  { return nil }
func (m *mockI2C) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}
