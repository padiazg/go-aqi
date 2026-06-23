package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAirQualityReading_Defaults(t *testing.T) {
	r := &AirQualityReading{
		Timestamp: time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC),
		SensorID:  "test",
		MassPM25:  12.5,
	}

	assert.Equal(t, "test", r.SensorID)
	assert.Equal(t, float32(12.5), r.MassPM25)
	assert.WithinDuration(t, time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC), r.Timestamp, 0)
}

func TestReadingEvent_Err(t *testing.T) {
	errSentinel := errors.New("sensor error")
	evt := &ReadingEvent{Err: errSentinel}

	assert.True(t, errors.Is(evt.Err, errSentinel))
	assert.Nil(t, evt.Reading)
}

func TestReadingEvent_Success(t *testing.T) {
	r := &AirQualityReading{SensorID: "test", MassPM10: 25.0}
	evt := &ReadingEvent{Reading: r}

	assert.NoError(t, evt.Err)
	assert.Equal(t, float32(25.0), evt.Reading.MassPM10)
}
