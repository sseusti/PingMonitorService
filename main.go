// Я разложил код по смысловым файлам: транспорт/клиент, IO-утилиты, доменные типы и логика проверки. Это снижает
//связность, упрощает тестирование и подготавливает проект к росту (worker pool / server / storage), не меняя поведение программы.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

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
		Timeout: 10 * time.Second,
		Transport: &loggingRoundTripper{
			logger: os.Stdout,
			next:   http.DefaultTransport,
		},
	}

	result := checkURL(ctx, client, "http://www.google.com", true)
	if result.Err != nil {
		log.Fatal(result.Err)
	}
	fmt.Println(result)
}
