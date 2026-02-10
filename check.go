package main

import (
	"context"
	"net/http"
	"time"
)

func checkURL(ctx context.Context, client *http.Client, url string, withPreview bool) CheckResult {
	start := time.Now()

	res, err := doRequestOnce(ctx, client, url)
	if err != nil {
		return CheckResult{
			URL:      url,
			Status:   0,
			Duration: time.Since(start),
			Err:      err,
		}
	}

	// Важно: при любом return после получения res — закрыть/дочитать body
	if res.StatusCode >= 500 {
		// можно оставить preview nil сейчас; улучшим на шаге 5
		drainAndClose(res.Body)
		return CheckResult{
			URL:      url,
			Status:   res.StatusCode,
			Duration: time.Since(start),
			Err:      HTTPStatusError{StatusCode: res.StatusCode, URL: url},
		}
	}

	if withPreview {
		preview, pErr := readPreviewAndDrain(res.Body, previewBytes)
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

	drainAndClose(res.Body)
	return CheckResult{
		URL:      url,
		Status:   res.StatusCode,
		Duration: time.Since(start),
	}
}
