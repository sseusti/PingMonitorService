package main

import "time"

func BackoffDelay(base time.Duration, attempt int, max time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if base < 0 {
		base = 0
	}

	delay := base * time.Duration(1<<attempt)
	if delay > max && max > 0 {
		delay = max
	}
	return delay
}
