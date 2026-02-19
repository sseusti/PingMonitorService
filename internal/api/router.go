package api

import (
	"PingMonitorService/internal/api/handlers"
	"PingMonitorService/internal/api/middleware"
	"net/http"
	"strings"
)

func Router(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/api/v1/checks", h.CreateChecks)
	mux.HandleFunc("/api/v1/jobs/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/results") {
			h.GetJobResults(w, r)
			return
		}
		h.GetJob(w, r)
	})

	// Я добавил middleware слой: logger для наблюдаемости (status/latency на каждый запрос, в т.ч. 404/401) и recover,
	// чтобы паника в одном handler не уронила процесс. Это повышает надёжность и упрощает диагностику в проде.
	// Middleware подключаются к mux на уровне Router, чтобы гарантированно применяться ко всем маршрутам.”
	handler := middleware.Recover(middleware.Logger(mux))

	return handler
}
