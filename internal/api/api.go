package api

import (
	"PingMonitorService/internal/jobs"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	store *jobs.Store
}

type JobStatusResponse struct {
	JobID     string      `json:"job_id"`
	Status    jobs.Status `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	Total     int         `json:"total"`
	Done      int         `json:"done"`
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	if id == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	job, ok := h.store.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := JobStatusResponse{JobID: job.ID, Status: job.Status, CreatedAt: job.CreatedAt, Total: job.Total, Done: job.Done}
	json.NewEncoder(w).Encode(resp)
}

func New(store *jobs.Store) *Handler {
	return &Handler{
		store: store,
	}
}
