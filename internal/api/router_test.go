package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"PingMonitorService/internal/api"
	"PingMonitorService/internal/jobs"
)

func newTestServer(t *testing.T) (*httptest.Server, *jobs.Store) {
	t.Helper()

	store := jobs.NewStore()

	runJobStub := func(jobID string, urls []string) {}

	h := api.New(store, runJobStub)
	handler := api.Router(h)

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return srv, store
}

func TestHealth_OK(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d, body=%q", resp.StatusCode, string(body))
	}
}

func TestHealth_MethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/health", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Allow"); got != "GET" {
		t.Fatalf("expected Allow=GET, got %q", got)
	}
}

func TestCreateChecks_ThenGetJob(t *testing.T) {
	srv, _ := newTestServer(t)

	body := `{"urls":["https://example.com","https://google.com"]}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/checks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/v1/checks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d, body=%q", resp.StatusCode, string(b))
	}

	var created struct {
		JobID string `json:"job_id"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if created.JobID == "" {
		t.Fatalf("expected non-empty job_id")
	}

	jobResp, err := http.Get(srv.URL + "/api/v1/jobs/" + created.JobID)
	if err != nil {
		t.Fatalf("GET /api/v1/jobs/{id}: %v", err)
	}
	defer jobResp.Body.Close()

	if jobResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(jobResp.Body)
		t.Fatalf("expected 200, got %d, body=%q", jobResp.StatusCode, string(b))
	}

	var status struct {
		JobID  string `json:"job_id"`
		Status string `json:"status"`
		Total  int    `json:"total"`
		Done   int    `json:"done"`
	}
	if err = json.NewDecoder(jobResp.Body).Decode(&status); err != nil {
		t.Fatalf("decode job status: %v", err)
	}
	if status.JobID != created.JobID {
		t.Fatalf("expected job_id=%q, got %q", created.JobID, status.JobID)
	}
	if status.Status == "" {
		t.Fatalf("expected non-empty status")
	}
	if status.Total != 2 {
		t.Fatalf("expected total=2, got %d", status.Total)
	}
	if status.Done != 0 {
		t.Fatalf("expected done=0, got %d", status.Done)
	}
}
