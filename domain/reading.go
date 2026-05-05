package domain

import "time"

// AirQualityReading holds an AQI reading from a sensor
type AirQualityReading struct {
	Timestamp           time.Time
	SensorID            string
	MassPM1             float32
	MassPM10            float32
	MassPM25            float32
	MassPM4             float32
	NumberPM05          float32
	NumberPM1           float32
	NumberPM10          float32
	NumberPM25          float32
	NumberPM4           float32
	TypicalParticleSize float32
}

// ReadingEvent is the result from a sensor reading
type ReadingEvent struct {
	Reading *AirQualityReading
	Err     error
}
