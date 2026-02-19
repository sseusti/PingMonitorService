package httpx

import (
	"context"
	"net/http"
)

func DoRequestOnce(ctx context.Context, client *http.Client, url string) (*http.Response, error) {

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
