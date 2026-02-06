package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type loggingRoundTripper struct {
	logger io.Writer
	next   http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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

func doRequestPreview(client *http.Client, url string) (int, time.Duration, []byte, error) {
	start := time.Now()

	//Если интервьюер спросит «зачем NewRequest», ты теперь можешь сказать:
	//Потому что http.NewRequest позволяет:
	//управлять заголовками
	//работать с context.Context
	//переиспользовать один и тот же клиент
	//писать middleware поверх запроса
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, time.Since(start), nil, err
	}

	req.Header.Set("User-Agent", "MyGoClient/1.0")

	res, err := client.Do(req)
	if err != nil {
		return 0, time.Since(start), nil, err
	}

	// В HTTP-клиенте тело ответа — это поток, связанный с TCP-соединением.
	// Если читать только часть тела и не дочитать остаток, соединение не возвращается в пул.
	// Поэтому при частичном чтении нужно сначала использовать io.LimitReader, а затем дочитать остаток в io.Discard.
	defer res.Body.Close()

	limited := io.LimitReader(res.Body, 1024)
	preview, err := io.ReadAll(limited)
	if err != nil {
		return 0, time.Since(start), nil, err
	}
	_, _ = io.Copy(io.Discard, res.Body)

	return res.StatusCode, time.Since(start), preview, nil
}

func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("stopped after 3 redirects: last=%s next=%s", via[len(via)-1].URL, req.URL)
			}
			return nil
		},
		Timeout: time.Second * 10,
		Transport: &loggingRoundTripper{
			logger: os.Stdout,
			next:   http.DefaultTransport,
		},
	}

	status, duration, preview, err := doRequestPreview(client, "http://www.google.com")
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Printf("status: %d, duration: %s\n", status, duration)
	fmt.Printf("preview (first 1024 bytes):\n%s\n", string(preview))
}
