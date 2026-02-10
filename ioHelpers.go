package main

import "io"

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
