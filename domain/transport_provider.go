package domain

type TransportProvider interface {
	Write(out []byte) error
	Read(in []byte, full bool) (int, error)
}
