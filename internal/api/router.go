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
	mux.HandleFunc("/api/v1/jobs/", h.GetJob)

	// Я добавил middleware слой: logger для наблюдаемости (status/latency на каждый запрос, в т.ч. 404/401) и recover,
	// чтобы паника в одном handler не уронила процесс. Это повышает надёжность и упрощает диагностику в проде.
	// Middleware подключаются к mux на уровне Router, чтобы гарантированно применяться ко всем маршрутам.”
	handler := middleware.Recover(middleware.Logger(mux))

	return handler
}
