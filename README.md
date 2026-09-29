# Orders API

[![Go](https://img.shields.io/badge/go-1.25-blue.svg)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

A small, production-minded Go service for managing customer orders, built around
clear boundaries and testability: clean architecture, offset pagination,
PostgreSQL with migrations, rate limiting, and graceful shutdown.

## Why this project exists

- An example of a real-world Go service that keeps domain logic framework-agnostic
  and easy to test.
- Focus: clear boundaries, mockable repositories, and predictable behaviour in
  production (rate limiting, graceful shutdown, connection pooling).

## Features

- Clean architecture: `application` (wiring/routing), `handler` (HTTP),
  `repository` (persistence), `infrastructure` (database), `model` (domain)
- CRUD for orders, with line items created and owned as part of the order
- Offset-based pagination for the list endpoint
- PostgreSQL with migrations and connection pooling
- Graceful shutdown on SIGINT/SIGTERM (HTTP server + database connection pool),
  request logging, panic recovery, per-request timeouts
- Config loader with validation and sensible defaults
- Configurable per-client rate limiting
- Health check endpoint returns JSON with database connectivity status
- Linting with golangci-lint
- Docker support (Dockerfile + docker-compose for local development)
- Test suite with 90%+ coverage on the `application`, `handler`, `repository`,
  and `middleware` packages (`go test -race ./...`)

## Project structure

```
.
├── cmd/api/            # Entrypoint + `migrate` subcommand
├── internal/
│   ├── application/    # Config, dependency wiring, routes, server lifecycle
│   ├── handler/        # HTTP handlers and JSON helpers
│   ├── infrastructure/ # Database connection and migrations
│   ├── middleware/     # Rate limiting
│   ├── model/          # Domain types (GORM models)
│   └── repository/     # Persistence contract + implementation
├── Dockerfile           # Multi-stage container build
├── docker-compose.yml   # Local development (API + PostgreSQL)
├── Makefile             # Common dev tasks
└── .golangci.yml        # Linting configuration
```

## Quick start

Requires Go 1.25+ and PostgreSQL.

### Option A: Local development

1. Copy the example environment file and edit it:

   ```bash
   cp .env.example .env
   ```

   ```env
   DATABASE_DSN="postgres://user:pass@localhost:5432/orders?sslmode=disable"
   SERVER_PORT=3000
   RATE_LIMIT_REQUESTS=10
   RATE_LIMIT_WINDOW_SECONDS=60
   ```

2. Apply the migrations:

   ```bash
   make migrate
   ```

3. Start the server:

   ```bash
   make run
   ```

### Option B: Docker Compose

```bash
docker compose up -d
```

This starts both the API and PostgreSQL. The API is available at `http://localhost:3000`.

## API

| Method   | Path        | Description                                  |
| -------- | ----------- | -------------------------------------------- |
| `GET`    | `/`         | Health check (returns JSON with DB status)   |
| `POST`   | `/orders`   | Create an order with its line items           |
| `GET`    | `/orders`   | List orders (`?cursor=<offset>`)              |
| `GET`    | `/orders/{id}`  | Fetch a single order                      |
| `PATCH`  | `/orders/{id}`  | Transition status to `shipped` or `completed` |
| `DELETE` | `/orders/{id}`  | Delete an order and its line items        |

### Examples

Create an order:

```bash
curl -X POST http://localhost:3000/orders \
  -H 'Content-Type: application/json' \
  -d '{
        "customer_id": "1",
        "line_items": [
          { "quantity": 2, "price": 1500 },
          { "quantity": 1, "price": 499 }
        ]
      }'
```

List orders (pass the returned `next` value back as `cursor`):

```bash
curl "http://localhost:3000/orders?cursor=0"
```

Transition an order to `shipped`, then `completed`:

```bash
curl -X PATCH http://localhost:3000/orders/1 -d '{"status":"shipped"}'
curl -X PATCH http://localhost:3000/orders/1 -d '{"status":"completed"}'
```

Errors are returned as `{"error": "message"}` with an appropriate status code.

## Configuration

The service reads configuration from environment variables. A `.env` file in the
working directory is loaded automatically for local development; real
environment variables take precedence.

| Variable                    | Required | Default | Description                     |
| --------------------------- | :------: | :-----: | ------------------------------- |
| `DATABASE_DSN`              |   Yes    |    —    | PostgreSQL DSN                  |
| `SERVER_PORT`               |    No    |  `3000` | HTTP port                       |
| `RATE_LIMIT_REQUESTS`       |    No    |   `10`  | Max requests per client per window |
| `RATE_LIMIT_WINDOW_SECONDS` |    No    |   `60`  | Window size in seconds          |

## Deployment notes

- The rate limiter keys on the client IP, which chi derives from
  `X-Forwarded-For`. Run the service behind a reverse proxy that **overwrites**
  `X-Forwarded-For` for every request. If the header is passed through from
  untrusted clients, clients can trivially bypass the limit by spoofing it.
- Rate limit state is held in memory per process. With multiple replicas the
  effective limit is per replica, not global.
- `PATCH /orders/{id}` only accepts the `shipped` and `completed` transitions,
  and enforces that an order is shipped before it is completed.

## Testing

```bash
make test
```

Or directly:

```bash
go test ./... -race -cover
```

Repository tests run against an in-memory SQLite database, so no PostgreSQL
instance is required to run the suite.

## Linting

```bash
make lint
```

## License

MIT — see [LICENSE](./LICENSE).
