package monitor

import (
	"fmt"
	"time"
)

const PreviewBytes = 1024

// Я ввёл тип CheckResult как единый контракт результата проверки URL (статус, latency, опциональное превью, ошибка).
// Это упрощает параллельную обработку в worker pool и даёт готовую модель для хранения в БД и выдачи через HTTP API.
type CheckResult struct {
	URL      string
	Status   int
	Duration time.Duration
	Preview  []byte
	Err      error
}

// Я типизирую HTTP-ошибки через HTTPStatusError, чтобы отделить “сервер ответил 5xx” от сетевых ошибок. При этом поле
// Err оставляю типа error, чтобы не терять другие классы ошибок, и использую errors.As для принятия решений в retry-политике.
type HTTPStatusError struct {
	StatusCode int
	URL        string
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("http status %d for %s", e.StatusCode, e.URL)
}
