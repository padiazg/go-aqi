package domain

// TransportProvider provides bidirectional byte transport for sensor communication.
type TransportProvider interface {
	Write(out []byte) error
	Read(in []byte, full bool) (int, error)
}
