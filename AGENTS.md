# VerbalForge Backend

Go/Gin API for the GRE verbal-learning application.

## Architecture

- `cmd/server` is the composition root: it loads configuration, initializes MongoDB and Redis, wires the application, and starts Gin.
- `internal/routes` registers endpoints. HTTP work lives in `internal/handlers` and `internal/controllers`, business logic in `internal/services`, and MongoDB access in `internal/repository`.
- Shared data shapes live in `internal/models`; cross-cutting request behavior lives in `internal/middleware`.

## Commands

Run commands from the repository root:

- `go mod download` downloads module dependencies.
- `go run ./cmd/server` starts the API.
- `go build -o /tmp/verbalforge-backend ./cmd/server` verifies the build without creating a
  repository artifact. `go build ./cmd/server` instead writes an unignored root `server` binary;
  remove it after verification and never commit it.
- `go test ./...` currently validates compilation but no behavior because the repository has no
  `_test.go` files.
- `go vet ./...` runs the standard Go analyzer.
- `gofmt -w path/to/changed.go` formats a changed Go file; `gofmt -l .` is repository-wide
  inspection only. Do not rewrite unrelated files.
- `air` is optional when installed; `.air.toml` rebuilds `./cmd/server/main.go` into `tmp/`.

## Runtime And Configuration

- Startup requires reachable MongoDB and Redis services.
- Uploaded files are stored locally under the configured upload directory and served from `/uploads`.
- `GET /health` is the health endpoint.
- Configuration is loaded by `internal/config/config.go` from the process environment after checking `.env`, `../.env`, and `../../.env` in order.
- Effective environment variable names are `MONGODB_URI`, `REDIS_URL`, `JWT_SECRET`,
  `FRONTEND_URLS`, `UPLOAD_DIR`, `PORT`, `GOOGLE_CLIENT_ID`, and `GOOGLE_CLIENT_SECRET`.
  `GOOGLE_REDIRECT_URL` is loaded but unused; OAuth uses the first `FRONTEND_URLS` entry.

## Conventions

- Keep Go code formatted with standard `gofmt`.
- Follow existing `NewType` constructors when wiring repositories, services, handlers, and controllers.
- Bind request bodies with Gin's binding APIs and return responses through the shared `internal/utils.APIResponse` helpers.
- Give persisted and API-facing model fields explicit `json` and `bson` tags.
- MongoDB schema is model-driven; this repository has no migration tooling.

## Safety

- Never read, print, or expose values from `.env` files.
- Do not run `setup.sh` as routine local setup; use the documented Go commands instead.
- Treat `server`, `tmp/`, `uploads/`, logs, and coverage output (`*.out`, `coverage.out`, and
  `coverage.html`) as generated or runtime output; do not edit or commit them. The root `server`
  binary is not ignored.
