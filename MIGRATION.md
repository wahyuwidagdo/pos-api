# Database Migrations

This project uses **golang-migrate** for applying migrations and **Atlas** for generating them from GORM models.

## Prerequisites
1.  **Atlas CLI**: [Install Atlas](https://atlasgo.io/getting-started/) to generate migrations.
    ```bash
    curl -sSf https://atlasgo.sh | sh
    ```
2.  **Golang Migrate**: The application runs migrations on startup, but you can also install the CLI for manual control.
    ```bash
3.  **Docker**: Required for Atlas to spin up a temporary dev database for safe schema diffing.
    - If you don't use Docker, you must configure a local Postgres database as `dev-url` in `atlas.hcl`.

## Workflow

### 1. Modify Schema
Make changes to your GORM models in `internal/models/*.go`.

### 2. Generate Migration (Using Atlas)
Run Atlas to inspect your GORM schema and generate the SQL migration files using the configured `atlas.hcl`.

```bash
# Uses the 'gorm' environment defined in atlas.hcl
atlas migrate diff name_of_change --env gorm
```
*Note: This runs `go run ./cmd/atlas` under the hood to load your GORM schema.*

### 3. Apply Migrations
The application automatically applies pending migrations when it starts:
```bash
go run cmd/api/main.go
```

### 4. Rollback
To revert the last migration, use the `migrate` CLI:
```bash
migrate -path database/migrations -database "postgres://postgres:postgres@localhost:5432/pos?sslmode=disable" down 1
```

## Migration History
- **000001_init_schema**: Initial schema dump from GORM.
