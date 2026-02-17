package main

import (
	"PingMonitorService/internal/api"
	"PingMonitorService/internal/jobs"
	"log"
	"net/http"
	"time"
)

func main() {
	store := jobs.NewStore()
	a := api.New(store)
	handler := api.Router(a)
	// Я вынес роутинг и handlers из cmd в internal/api, чтобы точка входа была максимально тонкой: только сборка
	// зависимостей и запуск сервера. Это улучшает тестируемость (router/handlers тестируются через httptest),
	//снижает связность и упрощает расширение API — новые эндпоинты добавляются в одном месте без роста main.go.
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
