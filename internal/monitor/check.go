package monitor

import (
	"PingMonitorService/internal/httpx"
	"context"
	"net/http"
	"time"
)

func CheckURL(ctx context.Context, client *http.Client, url string, withPreview bool) CheckResult {
	start := time.Now()

	res, err := httpx.DoRequestOnce(ctx, client, url)
	if err != nil {
		return CheckResult{
			URL:      url,
			Status:   0,
			Duration: time.Since(start),
			Err:      err,
		}
	}

	if res.StatusCode >= 500 {
		httpx.DrainAndClose(res.Body)
		return CheckResult{
			URL:      url,
			Status:   res.StatusCode,
			Duration: time.Since(start),
			Err:      HTTPStatusError{StatusCode: res.StatusCode, URL: url},
		}
	}

	if withPreview {
		preview, pErr := httpx.ReadPreviewAndDrain(res.Body, PreviewBytes)
		if pErr != nil {
			return CheckResult{
				URL:      url,
				Status:   res.StatusCode,
				Duration: time.Since(start),
				Err:      pErr,
			}
		}
		return CheckResult{
			URL:      url,
			Status:   res.StatusCode,
			Duration: time.Since(start),
			Preview:  preview,
		}
	}

	httpx.DrainAndClose(res.Body)
	return CheckResult{
		URL:      url,
		Status:   res.StatusCode,
		Duration: time.Since(start),
	}
}
