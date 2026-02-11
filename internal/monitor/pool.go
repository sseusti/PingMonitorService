package monitor

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type PoolConfig struct {
	Workers     int
	WithPreview bool
	RPS         int
	Retry       RetryConfig
}

func PingAllStable(
	ctx context.Context,
	client *http.Client,
	urls []string,
	cfg PoolConfig,
) []CheckResult {
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}

	interval := 200 * time.Millisecond
	if cfg.RPS > 0 {
		interval = time.Second / time.Duration(cfg.RPS)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	limit := ticker.C

	jobs := make(chan string)
	results := make(chan CheckResult)

	var wg sync.WaitGroup

	wg.Add(cfg.Workers)
	for w := 0; w < cfg.Workers; w++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case url, ok := <-jobs:
					if !ok {
						return
					}

					res := CheckURLStable(ctx, client, url, cfg.WithPreview, limit, cfg.Retry)

					select {
					case results <- res:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, url := range urls {
			select {
			case jobs <- url:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var checks []CheckResult
	for res := range results {
		checks = append(checks, res)
	}
	return checks
}
