package api

import "PingMonitorService/internal/jobs"

type Handler struct {
	store *jobs.Store
}

func New(store *jobs.Store) *Handler {
	return &Handler{
		store: store,
	}
}
