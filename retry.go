package main

import (
	"context"
	"errors"
	"time"
)

func DoWithRetryBackoffRateLimit(
	ctx context.Context,
	do func(context.Context) error,
	attempts int,
	baseDelay time.Duration,
	maxDelay time.Duration,
	limit <-chan time.Time,
	shouldRetry func(error) bool,
) error {
	if attempts <= 0 {
		return errors.New("attempts must be greater than zero")
	}
	if do == nil {
		return errors.New("do must not be nil")
	}
	if shouldRetry == nil {
		shouldRetry = func(err error) bool {
			return true
		}
	}

	if baseDelay < 0 {
		baseDelay = 0
	}
	if maxDelay < 0 {
		maxDelay = baseDelay
	}
	if maxDelay < baseDelay {
		maxDelay = baseDelay
	}

	var lastErr error

	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		select {
		case <-limit:
		case <-ctx.Done():
			return ctx.Err()
		}

		lastErr = do(ctx)
		if lastErr == nil {
			return nil
		}

		if !shouldRetry(lastErr) {
			return lastErr
		}
		if i == attempts-1 {
			break
		}

		sleep := BackoffDelay(baseDelay, i, maxDelay)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return lastErr
}
