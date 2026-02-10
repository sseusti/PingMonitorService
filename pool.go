// Я реализовал worker pool по схеме fan-out/fan-in: задания через jobs, результаты через results, завершение через
//WaitGroup и закрытие каналов. Каждый воркер использует устойчивую проверку с retry/backoff и общий rate limit
//через один time.Ticker, а все операции уважают context для корректной отмены

package main

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type PoolConfig struct {
	Workers     int
	WithPreview bool
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

	ticker := time.NewTicker(200 * time.Millisecond)
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

					res := checkURLStable(ctx, client, url, cfg.WithPreview, limit)

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
