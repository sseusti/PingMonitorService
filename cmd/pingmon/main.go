package main

import (
	"PingMonitorService/internal/httpx"
	"PingMonitorService/internal/monitor"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	workers := flag.Int("workers", 4, "number of workers")
	timeout := flag.Duration("timeout", 10*time.Second, "overall timeout for the run")
	preview := flag.Bool("preview", false, "read preview bytes")
	rps := flag.Int("rps", 5, "global requests per second (0 = no limit)")

	attempts := flag.Int("attempts", 4, "retry attempts")
	baseDelay := flag.Duration("base-delay", 200*time.Millisecond, "base backoff delay")
	maxDelay := flag.Duration("max-delay", 2*time.Second, "max backoff delay")

	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		urls = []string{
			"https://www.google.com",
			"https://www.facebook.com",
			"https://httpstat.us/500",
			"https://httpstat.us/200?sleep=1000",
			"https://example.com",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("stopped after 3 redirects: last=%s next=%s", via[len(via)-1].URL, req.URL)
			}
			return nil
		},
		Timeout: *timeout,
		Transport: &httpx.LoggingRoundTripper{
			Logger: os.Stdout,
			Next:   http.DefaultTransport,
		},
	}

	cfg := monitor.PoolConfig{
		Workers:     *workers,
		WithPreview: *preview,
		RPS:         *rps,
		Retry: monitor.RetryConfig{
			Attempts:  *attempts,
			BaseDelay: *baseDelay,
			MaxDelay:  *maxDelay,
		},
	}

	out := monitor.PingAllStable(ctx, client, urls, cfg)
	for _, o := range out {
		if o.Err != nil {
			fmt.Printf("[%s] error: %s | %v | %d\n", o.URL, o.Err, o.Duration, o.Status)
		} else {
			fmt.Printf("[%s] %v | %d\n", o.URL, o.Duration, o.Status)
		}
	}
}
