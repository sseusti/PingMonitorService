package httpx

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoggingRoundTripper struct {
	Logger io.Writer
	Next   http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := l.Next.RoundTrip(req)

	dur := time.Since(start)

	if err != nil {
		fmt.Fprintf(l.Logger, "[%s] %s %s -> error (%v) (%v)\n", time.Now().Format(time.ANSIC), req.Method, req.URL, err, dur)
		return nil, err
	}
	fmt.Fprintf(l.Logger, "[%s] %s %s -> %d (%v)\n", time.Now().Format(time.ANSIC), req.Method, req.URL, resp.StatusCode, dur)
	return resp, nil
}
