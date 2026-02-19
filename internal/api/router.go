package api

import (
	"PingMonitorService/internal/api/handlers"
	"PingMonitorService/internal/api/middleware"
	"net/http"
)

func Router(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/api/v1/checks", h.CreateChecks)
	mux.HandleFunc("/api/v1/jobs/", h.JobDispatcher)

	handler := middleware.Recover(middleware.Logger(mux))

	return handler
}
