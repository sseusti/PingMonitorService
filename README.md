# PingMonitorService

Concurrent HTTP ping monitor with retries, backoff, and optional response previews.

Russian version: `README.ru.md`

## Build

```sh
go build -o pingmon ./cmd/pingmon
```

## Run

```sh
./pingmon [flags] <url...>
```

If no URLs are provided, the program uses a small built-in list of sample URLs.

### Example

```sh
./pingmon -workers=8 -timeout=5s -rps=10 https://example.com https://httpstat.us/200
```

### Flags

- `-workers` (default: 4) number of workers
- `-timeout` (default: 10s) per-request HTTP client timeout
- `-preview` (default: false) read and log preview bytes from the response body
- `-rps` (default: 5) global requests per second (`0` = no limit)
- `-attempts` (default: 4) retry attempts
- `-base-delay` (default: 200ms) base backoff delay
- `-max-delay` (default: 2s) max backoff delay
