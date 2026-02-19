package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"PingMonitorService/internal/api/middleware"
)

func TestRecover_ConvertsPanicTo500_AndServerStaysAlive(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := middleware.Recover(mux)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/panic")
	if err != nil {
		t.Fatalf("GET /panic: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 500, got %d, body=%q", resp.StatusCode, string(b))
	}

	resp2, err := http.Get(srv.URL + "/ok")
	if err != nil {
		t.Fatalf("GET /ok: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 200, got %d, body=%q", resp2.StatusCode, string(b))
	}
}
