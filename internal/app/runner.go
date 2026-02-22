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
	repo    *jobs.Repo
	pingAll PingAllFunc
	timeout time.Duration
	client  *http.Client
	cfg     monitor.PoolConfig
}

func NewRunner(repo *jobs.Repo, client *http.Client, cfg monitor.PoolConfig, timeout time.Duration, pingAll PingAllFunc) *Runner {
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
		repo:    repo,
		pingAll: pingAll,
		timeout: timeout,
		client:  client,
		cfg:     cfg,
	}
}

func (r *Runner) Run(jobID string, urls []string) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		log.Printf("runner panic for job %s: %v\n%s", jobID, rec, debug.Stack())
		if err := r.repo.MarkFailed(context.Background(), jobID, fmt.Sprintf("panic: %v", rec)); err != nil {
			log.Printf("repo.MarkFailed failed after panic, jobID: %s, err: %v", jobID, err)
		}
	}()

	_ = r.pingAll(ctx, r.client, urls, r.cfg)

	if err := ctx.Err(); err != nil {
		if markErr := r.repo.MarkFailed(ctx, jobID, err.Error()); markErr != nil {
			log.Printf("repo.MarkFailed failed, jobID: %s, err: %v", jobID, markErr)
		}
		return
	}

	if err := r.repo.MarkDone(ctx, jobID); err != nil {
		log.Printf("repo.MarkDone failed, jobID: %s, err: %v", jobID, err)
		return
	}
}
