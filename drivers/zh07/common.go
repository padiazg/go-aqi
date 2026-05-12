package zh07

import (
	"errors"
	"time"
)

var (

	// ErrChecksumMismatch is returned when the calculated checksum doesn't match the received checksum
	ErrChecksumMismatch = errors.New("checksum mismatch")
	// ErrInvalidFrame is returned when the received data frame is invalid
	ErrInvalidFrame = errors.New("invalid data frame")
	// ErrSensorCommunication is returned when communication with the sensor fails
	ErrSensorCommunication = errors.New("sensor communication failed")
	//ErrUnknownMode is returned when an uknown mode is requested
	ErrUnknownMode = errors.New("unknown mode")

	commandSetInitiativeUploadMode = []byte{
		0xFF,
		0x01,
		0x78,
		0x40, // first byte for initiative mode
		0x00,
		0x00,
		0x00,
		0x00,
		0x47, // second byte for initiative mode
	}

	commandSetQAMode = []byte{
		0xFF,
		0x01,
		0x78,
		0x41, // first byte for q&a mode
		0x00,
		0x00,
		0x00,
		0x00,
		0x46, // second byte for q&a mode
	}

	commandQuery = []byte{ // q&a mode - query the sensor
		0xFF,
		0x01,
		0x86,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x79,
	}

	sleepAfterWrite = 250 * time.Millisecond
)
