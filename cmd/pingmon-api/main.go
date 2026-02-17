package main

import (
	"PingMonitorService/internal/api"
	"log"
	"net/http"
	"time"
)

func main() {
	// Я вынес роутинг и handlers из cmd в internal/api, чтобы точка входа была максимально тонкой: только сборка
	// зависимостей и запуск сервера. Это улучшает тестируемость (router/handlers тестируются через httptest),
	//снижает связность и упрощает расширение API — новые эндпоинты добавляются в одном месте без роста main.go.
	mux := api.Router()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
