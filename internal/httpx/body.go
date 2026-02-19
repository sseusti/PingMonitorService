package httpx

import "io"

func ReadPreviewAndDrain(body io.ReadCloser, n int64) ([]byte, error) {
	defer body.Close()

	limited := io.LimitReader(body, n)
	preview, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	_, _ = io.Copy(io.Discard, body)

	return preview, nil
}

func DrainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}
