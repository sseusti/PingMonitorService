package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const previewBytes = 1024

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

// Я всегда протягиваю context.Context до http.NewRequestWithContext, чтобы верхний уровень (handler/worker/shutdown)
// мог отменять запросы и задавать дедлайны. При этом client.Timeout оставляю как общий защитный предел, чтобы запросы не зависали бесконечно.
func doRequestOnce(ctx context.Context, client *http.Client, url string) (*http.Response, time.Duration, error) {
	start := time.Now()

	//Если интервьюер спросит «зачем NewRequest», ты теперь можешь сказать:
	//Потому что http.NewRequest позволяет:
	//управлять заголовками
	//работать с context.Context
	//переиспользовать один и тот же клиент
	//писать middleware поверх запроса
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, time.Since(start), err
	}

	req.Header.Set("User-Agent", "MyGoClient/1.0")

	res, err := client.Do(req)
	if err != nil {
		return nil, time.Since(start), err
	}

	return res, time.Since(start), nil
}

func readPreviewAndDrain(body io.ReadCloser, n int64) ([]byte, error) {
	defer body.Close()

	// В HTTP-клиенте тело ответа — это поток, связанный с TCP-соединением.
	// Если читать только часть тела и не дочитать остаток, соединение не возвращается в пул.
	// Поэтому при частичном чтении нужно сначала использовать io.LimitReader, а затем дочитать остаток в io.Discard.
	limited := io.LimitReader(body, n)
	preview, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	_, _ = io.Copy(io.Discard, body)

	return preview, nil
}

// Я разделил ответственность: doRequestOnce делает один HTTP-запрос и возвращает *http.Response, а readPreviewAndDrain
// полностью отвечает за чтение ограниченного превью и обязательный drain/close для возврата соединения в пул. Так код проще тестировать и безопаснее использовать в параллельном worker pool.
func doRequestPreview(ctx context.Context, client *http.Client, url string) (int, time.Duration, []byte, error) {
	res, duration, err := doRequestOnce(ctx, client, url)
	if err != nil {
		return 0, duration, nil, err
	}

	preview, err := readPreviewAndDrain(res.Body, previewBytes)
	if err != nil {
		return res.StatusCode, duration, preview, err
	}

	return res.StatusCode, duration, preview, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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

	status, duration, preview, err := doRequestPreview(ctx, client, "http://www.google.com")
	if err != nil {
		log.Fatal("error:", err)
	}

	fmt.Printf("status: %d, duration: %s\n", status, duration)
	fmt.Printf("preview (first 1024 bytes):\n%s\n", string(preview))
}
