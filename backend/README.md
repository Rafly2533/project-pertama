# Intan Florist API

Production-oriented MVP REST API built with Go, Gin, GORM, PostgreSQL, JWT, and bcrypt.

## Requirements

- Go 1.26+
- PostgreSQL 14+

## PostgreSQL

With Docker:

```sh
docker compose up -d postgres
```

Without Docker:

```sql
CREATE DATABASE intan_florist;
CREATE USER intan_florist WITH ENCRYPTED PASSWORD 'strong-password';
GRANT ALL PRIVILEGES ON DATABASE intan_florist TO intan_florist;
```

## Configuration

```sh
cp .env.example .env
```

Set strong values for `DB_PASSWORD`, `JWT_SECRET`, and `ADMIN_PASSWORD`. `JWT_EXPIRES_IN` accepts Go durations such as `24h` or a number of seconds. `ALLOWED_ORIGINS` is a comma-separated list.

## Run

```sh
go mod download
go run .
```

Migrations and idempotent seed data run at startup. The health endpoint is `GET /health`; API routes use `/api/v1`.

## Verification

```sh
gofmt -w .
go test ./...
go vet ./...
```
