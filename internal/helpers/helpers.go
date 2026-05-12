package helpers

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig controls the behavior of WithRetry.
type RetryConfig struct {
	Fn       func() error           // function to retry
	CountAs  func(err error) bool   // nil means all errors count toward retry limit
	Interval time.Duration          // 0 means no retry interval
	Timeout  time.Duration          // 0 means no timeout (respects parent context)
	Times    int                    // max retry attempts; 0 means unlimited
}

// WithRetry retries fn up to config.Times times with config.Interval between attempts.
// If config.Timeout > 0, a timeout context wraps the parent ctx.
func WithRetry(ctx context.Context, config *RetryConfig) error {
	timeoutCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if config.Timeout > 0 {
		timeoutCtx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	var (
		count   int
		lastErr error
	)

	for {
		if lastErr = config.Fn(); lastErr == nil {
			return nil
		}

		if config.CountAs == nil || config.CountAs(lastErr) {
			count++
			if config.Times > 0 && count >= config.Times {
				return fmt.Errorf("failed after %d attempts: %w", count, lastErr)
			}
		}

		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout after %d attempts: %w", count, lastErr)
		case <-ticker.C:
		}
	}
}
