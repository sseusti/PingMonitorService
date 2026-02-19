# PingMonitorService

Concurrent HTTP URL monitor with retries, backoff, worker pool, and an async job API.

Russian version: `README.ru.md`

## Build Binaries

```sh
go build -o pingmon ./cmd/pingmon
go build -o pingmon-api ./cmd/pingmon-api
```

## CLI Usage (`cmd/pingmon`)

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

## API Usage (`cmd/pingmon-api`)

Run server:

```sh
./pingmon-api
```

Server listens on `:8080`.

### Endpoints

- `GET /health`
- `POST /api/v1/checks`
- `GET /api/v1/jobs/{job_id}`
- `GET /api/v1/jobs/{job_id}/results`

### `POST /api/v1/checks`

Request:

```json
{ "urls": ["https://example.com", "https://google.com"] }
```

Rules:

- Method must be `POST`
- Content type: `application/json` or `application/json; charset=utf-8`
- `urls` length: `1..1000`
- Max body size: 1 MiB

Response:

- `201 Created`
- Body:

```json
{ "job_id": "..." }
```

### `GET /api/v1/jobs/{job_id}`

Response:

- `200 OK` with job status (`running`, `done`, `failed`)
- `404 Not Found` if job does not exist

Example:

```json
{
  "job_id": "...",
  "status": "running",
  "created_at": "2026-02-19T17:36:40Z",
  "total": 2,
  "done": 0
}
```

### `GET /api/v1/jobs/{job_id}/results`

Response behavior by job status:

- `202 Accepted` when job is still `running`
- `200 OK` when job is `done`
- `200 OK` when job is `failed` (includes `error`)
- `404 Not Found` when job does not exist

Response shape:

```json
{
  "job_id": "...",
  "status": "done",
  "results": [
    {
      "url": "https://example.com",
      "status": 200,
      "duration_ms": 123
    }
  ]
}
```

Failed example:

```json
{
  "job_id": "...",
  "status": "failed",
  "error": "partial: context deadline exceeded",
  "results": []
}
```

### Quick API Flow

Create job:

```sh
curl -i -X POST localhost:8080/api/v1/checks \
  -H 'Content-Type: application/json' \
  -d '{"urls":["https://example.com","https://httpstat.us/500"]}'
```

Poll status:

```sh
curl -i localhost:8080/api/v1/jobs/<job_id>
```

Fetch results:

```sh
curl -i localhost:8080/api/v1/jobs/<job_id>/results
```
