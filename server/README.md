# Go Backend

Go backend starter using Gin, PostgreSQL, sqlc, Air, and Goose.

## Prerequisites

- Go 1.25 or newer
- PostgreSQL 16 or newer, running locally

## Setup

1. Copy the environment template:

   PowerShell:

   ```powershell
   Copy-Item .env.example .env
   ```

   macOS/Linux:

   ```sh
   cp .env.example .env
   ```

2. Start PostgreSQL and create the database and user expected by `.env.example`:

   ```sql
   CREATE USER app WITH PASSWORD 'app';
   CREATE DATABASE app OWNER app;
   ```

   Run these statements as a PostgreSQL administrator, or update `DATABASE_URL` in `.env` to use an existing database.

3. Download and verify Go dependencies:

   ```sh
   go mod tidy
   go test ./...
   ```

4. Run backend tasks through the PowerShell task dispatcher:

   PowerShell:

   ```powershell
   .\task.ps1 dev
   ```

   macOS/Linux:

   ```sh
   ./scripts/run.sh
   ```

The API listens on `http://localhost:8080` by default. The process handles `SIGINT` and `SIGTERM` and drains active HTTP requests before exiting. Startup and shutdown use a 10-second timeout by default.

## Server structure

The server uses a handler, service, and repository structure:

```text
cmd/server             application composition and process lifecycle
internal/handler       HTTP transport and response mapping
internal/service       application behavior and use cases
internal/repository    persistence interfaces and database adapters
internal/database      database connection setup and generated SQL code
internal/router         route registration and middleware setup
```

Dependencies flow from handler to service to repository. The composition point wires concrete adapters together, which keeps HTTP code independent of PostgreSQL and makes each layer replaceable in tests.

## Health check

```sh
curl http://localhost:8080/health
```

Expected response:

```json
{"success":true,"data":{"status":"ok"}}
```

`/health` checks that the process is running. `/ready` checks PostgreSQL and returns HTTP `503` if the application cannot reach it.

If a `ServerConfig` is constructed without timeout values, the server falls back to a 10-second shutdown timeout. The application entrypoint also restores the 10-second startup timeout before connecting to PostgreSQL.

## Tasks

On PowerShell, use `task.ps1` for common development commands:

```powershell
.\task.ps1 build
.\task.ps1 run
.\task.ps1 dev
.\task.ps1 tidy
.\task.ps1 generate
.\task.ps1 migrate-new -Name create_users
.\task.ps1 migrate-up
.\task.ps1 migrate-down
.\task.ps1 migrate-status
```

## Migrations for future schema changes

Add Goose migration files to `db/migrations` before running these commands. The health endpoint does not require migrations.

PowerShell:

```powershell
   .\task.ps1 migrate-up
   .\task.ps1 migrate-down
   .\task.ps1 migrate-status
```

macOS/Linux:

```sh
./scripts/migrate-up.sh
./scripts/migrate-down.sh
```

```sh
go run ./cmd/migrate up
go run ./cmd/migrate down
```

Migration files belong in `db/migrations` and are executed by Goose. The directory is currently empty because the health endpoint only checks whether PostgreSQL accepts a connection.

## sqlc

Application SQL queries belong in `db/queries`, and generated code is written to `internal/database/generated`.

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0 generate
```

The `Makefile` provides equivalent commands where GNU Make is available: `make run`, `make migrate-up`, `make migrate-down`, and `make sqlc`.
