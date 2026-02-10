package monitor

import (
	"PingMonitorService/internal/netx"
	"context"
	"net/http"
	"time"
)

func CheckURLStable(
	ctx context.Context,
	client *http.Client,
	url string,
	withPreview bool,
	limit <-chan time.Time,
) CheckResult {
	start := time.Now()
	var last CheckResult
	attempts := 4
	baseDelay := 200 * time.Millisecond
	maxDelay := 2 * time.Second

	err := netx.DoWithRetryBackoffRateLimit(
		ctx,
		func(ctx context.Context) error {
			last = CheckURL(ctx, client, url, withPreview)
			return last.Err
		},
		attempts,
		baseDelay,
		maxDelay,
		limit,
		ShouldRetryHTTP,
	)

	last.Duration = time.Since(start)
	last.Err = err
	return last
}
