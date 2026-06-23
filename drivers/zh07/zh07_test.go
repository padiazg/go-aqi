package zh07

import (
	"testing"

	"github.com/padiazg/go-aqi/domain"
	"github.com/stretchr/testify/assert"
)

type NewFn func(*testing.T, domain.SensorProvider, error)

var checkNew = func(fns ...NewFn) []NewFn { return fns }

func checkType(want domain.SensorProvider) NewFn {
	return func(t *testing.T, sp domain.SensorProvider, err error) {
		t.Helper()
		if want == nil {
			assert.NotNil(t, err)
			return
		}
		assert.Nil(t, err)
		assert.IsType(t, want, sp)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		checks []NewFn
	}{
		{
			name:   "fail - unknown mode",
			config: &Config{Mode: -1},
			checks: checkNew(checkType(nil)),
		},
		{
			name:   "success - ModeInitiative",
			config: &Config{Mode: ModeInitiative},
			checks: checkNew(checkType(&ZH07i{})),
		},
		{
			name:   "success - ModeQA",
			config: &Config{Mode: ModeQA},
			checks: checkNew(checkType(&ZH07q{})),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(tt.config)
			for _, c := range tt.checks {
				c(t, r, err)
			}
		})
	}
}
