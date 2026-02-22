package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type createChecksRequest struct {
	URLs []string `json:"urls"`
}

type createChecksResponse struct {
	JobID string `json:"job_id"`
}

func (h *Handler) CreateChecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != "application/json" && ct != "application/json; charset=utf-8" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createChecksRequest
	if err := dec.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 || len(req.URLs) > 1000 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := h.repo.Create(r.Context(), len(req.URLs))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if h.runJob != nil {
		urlsCopy := append([]string(nil), req.URLs...)
		go h.runJob(job.ID, urlsCopy)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(createChecksResponse{JobID: job.ID}); err != nil {
		log.Printf("encode create checks response: %v", err)
	}
}
