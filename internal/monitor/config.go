package monitor

import "time"

type RetryConfig struct {
	Attempts  int
	MaxDelay  time.Duration
	BaseDelay time.Duration
}
