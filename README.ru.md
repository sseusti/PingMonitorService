# PingMonitorService

Конкурентный HTTP монитор URL с ретраями, backoff, worker pool и асинхронным job API.

## Сборка бинарников

```sh
go build -o pingmon ./cmd/pingmon
go build -o pingmon-api ./cmd/pingmon-api
```

## CLI (`cmd/pingmon`)

```sh
./pingmon [flags] <url...>
```

Если URL не указаны, используются несколько тестовых URL по умолчанию.

### Пример

```sh
./pingmon -workers=8 -timeout=5s -rps=10 https://example.com https://httpstat.us/200
```

### Флаги

- `-workers` (по умолчанию: 4) число воркеров
- `-timeout` (по умолчанию: 10s) таймаут HTTP-клиента на запрос
- `-preview` (по умолчанию: false) читать и логировать превью байты из тела ответа
- `-rps` (по умолчанию: 5) глобальный лимит запросов в секунду (`0` = без лимита)
- `-attempts` (по умолчанию: 4) количество ретраев
- `-base-delay` (по умолчанию: 200ms) базовая задержка backoff
- `-max-delay` (по умолчанию: 2s) максимальная задержка backoff

## API (`cmd/pingmon-api`)

Запуск сервера:

```sh
export DATABASE_URL='postgres://postgres:password@127.0.0.1:5432/postgres?sslmode=disable'
./pingmon-api
```

Сервер слушает `:8080`.
Переменная `DATABASE_URL` обязательна.

Текущий режим хранения: `jobs-only` в PostgreSQL:
- Метаданные job (`id`, `status`, время, счётчики, ошибка) сохраняются в Postgres.
- `GET /api/v1/jobs/{job_id}` переживает перезапуск сервера.
- Детальные результаты по URL пока не сохраняются в БД.

### Эндпоинты

- `GET /health`
- `POST /api/v1/checks`
- `GET /api/v1/jobs/{job_id}`
- `GET /api/v1/jobs/{job_id}/results`

### `POST /api/v1/checks`

Тело запроса:

```json
{ "urls": ["https://example.com", "https://google.com"] }
```

Ограничения:

- Метод только `POST`
- `Content-Type`: `application/json` или `application/json; charset=utf-8`
- Количество `urls`: `1..1000`
- Максимальный размер body: 1 MiB

Ответ:

- `201 Created`
- Тело:

```json
{ "job_id": "..." }
```

### `GET /api/v1/jobs/{job_id}`

Ответ:

- `200 OK` со статусом job (`running`, `done`, `failed`)
- `404 Not Found`, если job не существует

Пример:

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

Временное поведение в режиме `jobs-only`:
- эндпоинт возвращает статус job и ошибку (для `failed`),
- поле `results` сейчас пустое, потому что per-URL результаты пока не пишутся в БД.

Поведение по статусу job:

- `202 Accepted`, если job ещё `running`
- `200 OK`, если job `done`
- `200 OK`, если job `failed` (с полем `error`)
- `404 Not Found`, если job не существует

Формат ответа:

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

Пример ошибки:

```json
{
  "job_id": "...",
  "status": "failed",
  "error": "partial: context deadline exceeded",
  "results": []
}
```

### Быстрый сценарий API

Создать job:

```sh
curl -i -X POST localhost:8080/api/v1/checks \
  -H 'Content-Type: application/json' \
  -d '{"urls":["https://example.com","https://httpstat.us/500"]}'
```

Проверить статус:

```sh
curl -i localhost:8080/api/v1/jobs/<job_id>
```

Получить результаты:

```sh
curl -i localhost:8080/api/v1/jobs/<job_id>/results
```
