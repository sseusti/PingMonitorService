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
	rc RetryConfig,
) CheckResult {
	start := time.Now()
	var last CheckResult

	err := netx.DoWithRetryBackoffRateLimit(
		ctx,
		func(ctx context.Context) error {
			last = CheckURL(ctx, client, url, withPreview)
			return last.Err
		},
		rc.Attempts,
		rc.BaseDelay,
		rc.MaxDelay,
		limit,
		ShouldRetryHTTP,
	)

	last.Duration = time.Since(start)
	last.Err = err
	return last
}
