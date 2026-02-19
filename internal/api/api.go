package api

import (
	"PingMonitorService/internal/jobs"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	store  *jobs.Store
	runJob func(jobID string, urls []string)
}

type JobStatusResponse struct {
	JobID     string      `json:"job_id"`
	Status    jobs.Status `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	Total     int         `json:"total"`
	Done      int         `json:"done"`
}

type JobResultDTO struct {
	URL      string `json:"url"`
	Status   int    `json:"status"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
}

type JobResultsResponse struct {
	JobID   string         `json:"job_id"`
	Results []JobResultDTO `json:"results"`
}

func (h *Handler) JobDispatcher(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/results") {
		h.GetJobResults(w, r)
		return
	}
	h.GetJob(w, r)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	job, ok := h.store.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp := JobStatusResponse{JobID: job.ID, Status: job.Status, CreatedAt: job.CreatedAt, Total: job.Total, Done: job.Done}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("encode get job response: %v", err)
	}
}

func (h *Handler) GetJobResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	if !strings.HasSuffix(path, "/results") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	id := strings.TrimSuffix(path, "/results")
	if id == "" || strings.Contains(id, "/") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	job, ok := h.store.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	results := make([]JobResultDTO, 0, len(job.Results))
	for _, result := range job.Results {
		results = append(results, JobResultDTO{
			URL:      result.URL,
			Status:   result.StatusCode,
			Error:    result.Error,
			Duration: result.Duration.Milliseconds(),
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(JobResultsResponse{
		JobID:   job.ID,
		Results: results,
	}); err != nil {
		log.Printf("encode get job results response: %v", err)
	}
}

func New(store *jobs.Store, runJob ...func(jobID string, urls []string)) *Handler {
	var runner func(jobID string, urls []string)
	if len(runJob) > 0 {
		runner = runJob[0]
	}

	return &Handler{
		store:  store,
		runJob: runner,
	}
}
