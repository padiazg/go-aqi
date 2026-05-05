package zh07

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/padiazg/go-aqi/domain"
)

const (
	checksum = 0x0327
)

type checkFn func(t *testing.T, r *domain.ReadingEvent, err error)

var (
	check = func(fns ...checkFn) []checkFn { return fns }

	checkError = func(want bool) checkFn {
		return func(t *testing.T, _ *domain.ReadingEvent, err error) {
			t.Helper()
			if want {
				assert.NotNil(t, err, "hasError: error expected, none produced")
			} else {
				assert.Nil(t, err, "hasError = [+%v], no error expected")
			}
		}
	}
)
