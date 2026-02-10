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

func checkURL(ctx context.Context, client *http.Client, url string, withPreview bool) CheckResult {
	start := time.Now()

	res, doErr := doRequestOnce(ctx, client, url)
	if doErr != nil {
		return CheckResult{
			URL:      url,
			Status:   0,
			Duration: time.Since(start),
			Preview:  nil,
			Err:      doErr,
		}
	}

	if res.StatusCode >= 500 {
		drainAndClose(res.Body)
		return CheckResult{
			URL:      url,
			Status:   res.StatusCode,
			Duration: time.Since(start),
			Preview:  nil,
			Err: HTTPStatusError{
				StatusCode: res.StatusCode,
				URL:        url,
			},
		}
	}

	if withPreview == true {
		preview, previewErr := readPreviewAndDrain(res.Body, previewBytes)
		if previewErr != nil {
			return CheckResult{
				URL:      url,
				Status:   res.StatusCode,
				Duration: time.Since(start),
				Preview:  nil,
				Err:      previewErr,
			}
		}
		return CheckResult{
			URL:      url,
			Status:   res.StatusCode,
			Duration: time.Since(start),
			Preview:  preview,
			Err:      nil,
		}
	}

	drainAndClose(res.Body)

	return CheckResult{
		URL:      url,
		Status:   res.StatusCode,
		Duration: time.Since(start),
		Preview:  nil,
		Err:      nil,
	}
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

func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}

// Я разделил ответственность: doRequestOnce делает один HTTP-запрос и возвращает *http.Response, а readPreviewAndDrain
// полностью отвечает за чтение ограниченного превью и обязательный drain/close для возврата соединения в пул. Так код проще тестировать и безопаснее использовать в параллельном worker pool.
//func doRequestPreview(ctx context.Context, client *http.Client, url string) (int, time.Duration, []byte, error) {
//	res, duration, err := doRequestOnce(ctx, client, url)
//	if err != nil {
//		return 0, duration, nil, err
//	}
//
//	preview, err := readPreviewAndDrain(res.Body, previewBytes)
//	if err != nil {
//		return res.StatusCode, duration, preview, err
//	}
//
//	return res.StatusCode, duration, preview, nil
//}

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

	var result CheckResult
	result = checkURL(ctx, client, "http://www.google.com", true)
	if result.Err != nil {
		log.Fatal("error:", result.Err)
	}

	fmt.Println(result)
	//fmt.Printf("status: %d, duration: %s\n", status, duration)
	//fmt.Printf("preview (first 1024 bytes):\n%s\n", string(preview))
}
