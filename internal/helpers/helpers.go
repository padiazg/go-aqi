package helpers

import (
	"context"
	"fmt"
	"time"
)

type RetryConfig struct {
	Fn       func() error
	CountAs  func(err error) bool // nil = everything counts
	Interval time.Duration        // 0 = no retry limits
	Timeout  time.Duration        // 0 = no timeout (respects father ctx)
	Times    int
}

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
