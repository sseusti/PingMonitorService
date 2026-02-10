package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoggingRoundTripper struct {
	logger io.Writer
	next   http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := l.next.RoundTrip(req)

	dur := time.Since(start)

	if err != nil {
		fmt.Fprintf(l.logger, "[%s] %s %s -> error (%v) (%v)\n", time.Now().Format(time.ANSIC), req.Method, req.URL, err, dur)
		return nil, err
	}
	fmt.Fprintf(l.logger, "[%s] %s %s -> %d (%v)\n", time.Now().Format(time.ANSIC), req.Method, req.URL, resp.StatusCode, dur)
	return resp, nil
}

// Я всегда протягиваю context.Context до http.NewRequestWithContext, чтобы верхний уровень (handler/worker/shutdown)
// мог отменять запросы и задавать дедлайны. При этом client.Timeout оставляю как общий защитный предел, чтобы запросы не зависали бесконечно.
func doRequestOnce(ctx context.Context, client *http.Client, url string) (*http.Response, error) {

	//Если интервьюер спросит «зачем NewRequest», ты теперь можешь сказать:
	//Потому что http.NewRequest позволяет:
	//управлять заголовками
	//работать с context.Context
	//переиспользовать один и тот же клиент
	//писать middleware поверх запроса
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "MyGoClient/1.0")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
