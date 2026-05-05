package i2c

type I2CBus interface {
	WriteBytes(buf []byte) (int, error)
	ReadBytes(buf []byte) (int, error)
	ReadRegBytes(reg byte, n int) ([]byte, int, error)
	ReadRegU8(reg byte) (byte, error)
	WriteRegU8(reg byte, value byte) error
	ReadRegU16BE(reg byte) (uint16, error)
	ReadRegU16LE(reg byte) (uint16, error)
	ReadRegS16BE(reg byte) (int16, error)
	ReadRegS16LE(reg byte) (int16, error)
	WriteRegU16BE(reg byte, value uint16) error
	WriteRegU16LE(reg byte, value uint16) error
	WriteRegS16BE(reg byte, value int16) error
	WriteRegS16LE(reg byte, value int16) error
	Close() error
}
