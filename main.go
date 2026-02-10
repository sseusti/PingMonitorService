// Я разложил код по смысловым файлам: транспорт/клиент, IO-утилиты, доменные типы и логика проверки. Это снижает
//связность, упрощает тестирование и подготавливает проект к росту (worker pool / server / storage), не меняя поведение программы.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	urls := []string{
		"https://www.google.com",
		"https://www.facebook.com",
		"https://httpstat.us/500",
		"https://httpstat.us/200?sleep=1000",
		"https://example.com",
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("stopped after 3 redirects: last=%s next=%s", via[len(via)-1].URL, req.URL)
			}
			return nil
		},
		Timeout: 10 * time.Second,
		Transport: &LoggingRoundTripper{
			logger: os.Stdout,
			next:   http.DefaultTransport,
		},
	}

	// Я вынес retry/backoff/rate limit в независимую обёртку, а решение “ретраить или нет” — в отдельную политику
	// ShouldRetryHTTP. Это позволяет переиспользовать один и тот же механизм для HTTP/SQL, и безопасно контролировать
	// нагрузку при параллельной работе воркеров.
	// Я построил устойчивую проверку URL как композицию: checkURL отвечает только за один HTTP-запрос, а
	//retry/backoff/rate-limit вынесены в универсальную обёртку. Политика повторов инкапсулирована в ShouldRetryHTTP,
	//что позволяет менять правила без изменения механики. Все ожидания и ретраи уважают context, поэтому код безопасен для worker pool и graceful shutdown.
	//result := checkURLStable(ctx, client, "http://www.google.com", true, limit)
	//if result.Err != nil {
	//	log.Fatal(result.Err)
	//}
	//fmt.Println(result)

	out := PingAllStable(ctx, client, urls, PoolConfig{Workers: 4, WithPreview: false})
	for _, o := range out {
		if o.Err != nil {
			fmt.Printf("[%s] error: %s | %v | %d\n", o.URL, o.Err, o.Duration, o.Status)
		} else {
			fmt.Printf("[%s] %v | %d\n", o.URL, o.Duration, o.Status)
		}
	}
}
