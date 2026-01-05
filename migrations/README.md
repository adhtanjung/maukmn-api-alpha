# Database Migrations with Goose

This project uses [goose](https://github.com/pressly/goose) for database migrations.

## Setup

1. Configure your database connection in `.env`:

```bash
DATABASE_URL=postgresql://username:password@host:port/database?sslmode=require
```

## Usage

### Run Migrations

Apply all pending migrations:

```bash
go run cmd/migrate/main.go up
```

Or using goose CLI directly:

```bash
goose -dir migrations postgres "$DATABASE_URL" up
```

### Rollback Last Migration

```bash
go run cmd/migrate/main.go down
```

### Check Migration Status

```bash
go run cmd/migrate/main.go status
```

### Reset Database

**⚠️ Warning: This will delete all data!**

```bash
go run cmd/migrate/main.go reset
```

### Create New Migration

```bash
goose -dir migrations create add_new_table sql
```

This creates a new file like `migrations/00002_add_new_table.sql`.

## Migration Files

- `00001_init_schema.sql` - Initial schema with PostGIS, all tables, indexes, and seed data

## Available Commands

| Command           | Description                               |
| ----------------- | ----------------------------------------- |
| `up`              | Apply all pending migrations              |
| `up-by-one`       | Apply only the next pending migration     |
| `up-to VERSION`   | Apply migrations up to VERSION            |
| `down`            | Roll back the last migration              |
| `down-to VERSION` | Roll back migrations to VERSION           |
| `redo`            | Roll back and re-apply the last migration |
| `reset`           | Roll back all migrations                  |
| `status`          | Show migration status                     |
| `version`         | Show current database version             |

## Notes

- Migrations are stored in the `migrations/` directory
- Migration files use goose annotations (`-- +goose Up`, `-- +goose Down`)
- The schema follows `SPEC.md` architecture with PostGIS support
