# Orders API 🚀


[![Go](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]()
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)]()


One-line pitch: I built a small, production-minded Go service to manage customer orders — clean architecture, cursor pagination, PostgreSQL with migrations, rate limiting, and tested.

Why this project exists
- I wanted an example of a real-world service that keeps domain logic framework-agnostic and easy to test.
- Focus: clear boundaries, testability (mockable repositories), and predictable behavior in production (rate limiting, graceful shutdown).

Highlights ✨
- Clean architecture: separation of application, infrastructure, and handler layers
- CRUD for Orders and Line Items
- Cursor-based pagination for list endpoints (safe for large datasets)
- PostgreSQL with migrations and connection pooling
- Graceful shutdown and structured logging
- Config loader with validation and sensible defaults
- Comprehensive test suite and mockable repository layer
- Configurable rate limiting

Project structure 🧱
```
├── cmd/
│   └── api/             # Main entrypoint
├── internal/
│   ├── application/     
│   ├── handler/         
│   ├── repository/      
│   ├── infrastructure/ 
│   └── model/           
```

Quick start ▶️
Go 1.25+, PostgreSQL.

1) Copy a local .env
```env
DATABASE_DSN="postgres://user:pass@localhost:5432/orders?sslmode=disable"
SERVER_PORT=8080
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_WINDOW_SECONDS=60
```

2) Run migrations
```bash
go run ./cmd/api migrate
```

3) Start the server
```bash
go run ./cmd/api
# or build and run:
go build -o orders ./cmd/api && ./orders
```

Example requests 🔌

Get orders (cursor pagination):
```bash
curl "http://localhost:3000/orders?cursor=3"

```

Configuration ⚙️
The service reads environment variables (supports `.env` for local development).

| Variable                     | Required | Default | Description |
|-----------------------------:|:--------:|:-------:|------------|
| `DATABASE_DSN`               | Yes      | —       | PostgreSQL DSN (e.g. `postgres://user:pass@localhost:5432/orders?sslmode=disable`) |
| `SERVER_PORT`                | No       | `3000`  | HTTP port |
| `RATE_LIMIT_REQUESTS`        | No       | `10`    | Max requests per window |
| `RATE_LIMIT_WINDOW_SECONDS`  | No       | `60`    | Window size in seconds |

Testing 🧪
Run the full test suite:
```bash
go test ./... -v
```

Database & migrations 🗄️
Run migrations from the repo root:
```bash
go run ./cmd/api migrate
```

License 📜
MIT — see LICENSE.
