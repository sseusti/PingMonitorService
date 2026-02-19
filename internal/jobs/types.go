package jobs

import "time"

type Status string

const (
	Running Status = "running"
	Done    Status = "done"
	Failed  Status = "failed"
)

type Job struct {
	ID         string
	Status     Status
	CreatedAt  time.Time
	Total      int
	Done       int
	Results    []Result
	Error      string
	FinishedAt *time.Time
}

type Result struct {
	URL        string
	StatusCode int
	OK         bool
	Duration   time.Duration
	Error      string
}
