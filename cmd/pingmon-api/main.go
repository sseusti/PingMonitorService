package main

import (
	"PingMonitorService/internal/api"
	"PingMonitorService/internal/app"
	"PingMonitorService/internal/db"
	"PingMonitorService/internal/httpx"
	"PingMonitorService/internal/jobs"
	"PingMonitorService/internal/monitor"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	database, err := db.OpenPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.EnsureSchema(database)
	if err != nil {
		log.Fatal(err)
	}

	store := jobs.NewStore()
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &httpx.LoggingRoundTripper{
			Logger: os.Stdout,
			Next:   http.DefaultTransport,
		},
	}

	cfg := monitor.PoolConfig{
		Workers:     4,
		WithPreview: false,
		RPS:         5,
		Retry: monitor.RetryConfig{
			Attempts:  3,
			BaseDelay: 200 * time.Millisecond,
			MaxDelay:  2 * time.Second,
		},
	}

	runner := app.NewRunner(store, client, cfg, 30*time.Second, nil)
	a := api.New(store, runner.Run)
	handler := api.Router(a)
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
