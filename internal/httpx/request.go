package httpx

import (
	"context"
	"net/http"
)

// Я всегда протягиваю context.Context до http.NewRequestWithContext, чтобы верхний уровень (handler/worker/shutdown)
// мог отменять запросы и задавать дедлайны. При этом client.Timeout оставляю как общий защитный предел, чтобы запросы не зависали бесконечно.
func DoRequestOnce(ctx context.Context, client *http.Client, url string) (*http.Response, error) {

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
