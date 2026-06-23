package helpers

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestWithRetry(t *testing.T) {
	t.Parallel()

	errFail := errors.New("fail")

	tests := []struct {
		name     string
		makeFn   func(*int) func() error
		config   func(*int) *RetryConfig
		wantErr  string
		wantCalls int
	}{
		{
			name: "success - first attempt",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					return nil
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Millisecond,
				}
			},
			wantCalls: 1,
		},
		{
			name: "success - after retries",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					if *calls < 3 {
						return errFail
					}
					return nil
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Millisecond,
					Times:    5,
				}
			},
			wantCalls: 3,
		},
		{
			name: "fail - exhausted retries",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					return errFail
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Millisecond,
					Times:    3,
				}
			},
			wantErr:   "failed after 3 attempts",
			wantCalls: 3,
		},
		{
			name: "fail - timeout",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					return errFail
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Millisecond,
					Timeout:  50 * time.Millisecond,
					Times:    0,
				}
			},
			wantErr:   "timeout after",
			wantCalls: 0,
		},
		{
			name: "fail - context canceled",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					return errFail
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Hour,
					Times:    10,
				}
			},
			wantErr:   "timeout after",
			wantCalls: 1,
		},
		{
			name: "CountAs skips non-countable errors - timeout",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					return errFail
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Millisecond,
					Timeout:  50 * time.Millisecond,
					Times:    10,
					CountAs:  func(err error) bool { return false },
				}
			},
			wantErr:   "timeout after",
		},
		{
			name: "CountAs counts matching errors",
			makeFn: func(calls *int) func() error {
				return func() error {
					*calls++
					return errFail
				}
			},
			config: func(calls *int) *RetryConfig {
				return &RetryConfig{
					Interval: time.Millisecond,
					Times:    3,
					CountAs:  func(err error) bool { return err.Error() == "fail" },
				}
			},
			wantErr:   "failed after 3 attempts",
			wantCalls: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls int
			cfg := tt.config(&calls)
			if tt.makeFn != nil {
				cfg.Fn = tt.makeFn(&calls)
			}

			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.name == "fail - context canceled" {
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := WithRetry(ctx, cfg)

			if tt.wantErr != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
			if tt.wantCalls > 0 {
				assert.Equal(t, tt.wantCalls, calls, fmt.Sprintf("expected %d calls, got %d", tt.wantCalls, calls))
			}
		})
	}
}

func BenchmarkWithRetry_Success(b *testing.B) {
	errFail := errors.New("fail")
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		var calls int
		_ = WithRetry(ctx, &RetryConfig{
			Fn: func() error {
				calls++
				if calls < 3 {
					return errFail
				}
				return nil
			},
			Interval: time.Nanosecond,
			Times:    5,
		})
	}
}

func BenchmarkWithRetry_Exhausted(b *testing.B) {
	errFail := errors.New("fail")
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		_ = WithRetry(ctx, &RetryConfig{
			Fn:       func() error { return errFail },
			Interval: time.Nanosecond,
			Times:    3,
		})
	}
}
