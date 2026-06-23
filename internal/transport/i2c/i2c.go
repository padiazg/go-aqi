package i2c

import (
	"fmt"

	"github.com/d2r2/go-i2c"
	"github.com/padiazg/go-aqi/domain"
)

var _ domain.TransportProvider = (*I2CTransport)(nil)
var _ I2CBus = (*i2c.I2C)(nil)

type I2CTransport struct {
	i2c I2CBus
}

// New creates a new I2C transport instance.
func New(i2c I2CBus) *I2CTransport {
	return &I2CTransport{
		i2c: i2c,
	}
}

// Read reads bytes from the I2C bus into the provided buffer.
func (i *I2CTransport) Read(in []byte, _ bool) (int, error) {
	count, err := i.i2c.ReadBytes(in)
	if err != nil {
		return 0, fmt.Errorf("i2c read: %w", err)
	}

	return count, err
}

// Write writes data to the I2C bus.
func (i *I2CTransport) Write(out []byte) error {
	_, err := i.i2c.WriteBytes(out)
	if err != nil {
		return fmt.Errorf("i2c write %X: %w", out, err)
	}

	return nil
}
