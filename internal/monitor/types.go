package monitor

import (
	"fmt"
	"time"
)

const PreviewBytes = 1024

type CheckResult struct {
	URL      string
	Status   int
	Duration time.Duration
	Preview  []byte
	Err      error
}

type HTTPStatusError struct {
	StatusCode int
	URL        string
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("http status %d for %s", e.StatusCode, e.URL)
}
