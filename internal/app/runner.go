package app

import (
	"PingMonitorService/internal/jobs"
	"PingMonitorService/internal/monitor"
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type PingAllFunc func(ctx context.Context, client *http.Client, urls []string, cfg monitor.PoolConfig) []monitor.CheckResult

type Runner struct {
	store   *jobs.Store
	pingAll PingAllFunc
	timeout time.Duration
	client  *http.Client
	cfg     monitor.PoolConfig
}

func NewRunner(store *jobs.Store, client *http.Client, cfg monitor.PoolConfig, timeout time.Duration, pingAll PingAllFunc) *Runner {
	if pingAll == nil {
		pingAll = monitor.PingAllStable
	}
	if client == nil {
		client = http.DefaultClient
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Runner{
		store:   store,
		pingAll: pingAll,
		timeout: timeout,
		client:  client,
		cfg:     cfg,
	}
}

func (r *Runner) Run(jobID string, urls []string) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		log.Printf("runner panic for job %s: %v\n%s", jobID, rec, debug.Stack())
		r.store.SetFailed(jobID, fmt.Errorf("panic: %v", rec))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	raws := r.pingAll(ctx, r.client, urls, r.cfg)
	if err := ctx.Err(); err != nil {
		if !r.store.SetFailed(jobID, err) {
			log.Printf("store.SetFailed failed, jobID: %s", jobID)
		}
		return
	}

	results := make([]jobs.Result, 0, len(raws))

	for _, raw := range raws {
		errText := ""
		if raw.Err != nil {
			errText = raw.Err.Error()
		}

		res := jobs.Result{
			URL:        raw.URL,
			StatusCode: raw.Status,
			OK:         raw.Err == nil,
			Duration:   raw.Duration,
			Error:      errText,
		}

		results = append(results, res)
	}

	ok := r.store.SetResults(jobID, results)
	if !ok {
		log.Printf("store.SetResults failed, jobID: %s", jobID)
		return
	}
}
