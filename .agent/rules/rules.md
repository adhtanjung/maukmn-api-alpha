# Maukemana API Development Rules

This document outlines the patterns and rules for developing the Maukemana API. All contributors must follow these guidelines to ensure consistency and maintainability.

## 1. Project Structure

The project follows a standard Go project layout with logic residing in the `internal/` directory:

- `internal/auth`: Authentication logic and Clerk integration.
- `internal/database`: Database connection and health check logic.
- `internal/handlers`: HTTP request handlers (controllers).
- `internal/models`: (Optional) Shared data structures.
- `internal/repositories`: Database abstraction layer (SQL logic).
- `internal/router`: Gin-gonic route definitions and middleware setup.
- `internal/utils`: Shared utilities for responses and requests.

## 2. Routing Pattern

Routes are managed in `internal/router/router.go` using Gin.

- Use route groups for versioning (e.g., `/api/v1`) and feature areas.
- Apply middleware (like `AuthMiddleware`) at the group level when possible.

## 3. Handler Implementation

Handlers should focus on request parsing and response formatting.

- **Dependency Injection**: Pass repositories into handlers via constructors.
- **Request Binding**: Use local `struct`s with `json` tags for `c.ShouldBindJSON`.
- **Consistent Request Handling**: Use `internal/utils/request.go` for common tasks:
  - `page, limit := utils.GetPagination(c)` to handle pagination predictably.
- **Consistent Response Wrapping**: Always use `internal/utils/response.go` to ensure a uniform API contract:
  - `utils.SendSuccess(c, "Message", data)` for 200 OK.
  - `utils.SendCreated(c, "Message", data)` for 201 Created.
  - `utils.SendPaginated(c, "Message", data, page, limit, total)` for paginated lists.
  - `utils.SendError(c, code, "Message", err)` for custom errors.
  - `utils.SendValidationError(c, err)` for binding or validation failures.
  - `utils.SendInternalError(c, err)` for server-side failures.

> [!IMPORTANT] > **Never use `c.JSON` or `c.AbortWithStatusJSON` directly.**
> Always use the wrappers in `internal/utils` to maintain the standard response structure:
> `{ "success": bool, "message": string, "data": optional, "error": optional, "meta": optional }`

## 4. Validation

The API uses `github.com/go-playground/validator/v10` (via Gin).

- **Struct Tags**: Use `binding` tags for request validation (e.g., `binding:"required,url"`).
- **Handling**: Always check the error from `c.ShouldBindJSON` and use `utils.SendValidationError(c, err)`.

## 5. Configuration & Environment

- **Environment Variables**: Use `.env` for local development.
- **Loading**: Variables are loaded via `godotenv` in `cmd/server/main.go`.
- **Access**: Use `os.Getenv("KEY")` or follow the `getEnv` pattern for defaults.

## 6. Repository Pattern

All database interactions MUST go through repositories in `internal/repositories`.

- **sqlx**: Use `sqlx` for named parameters and struct scanning.
- **Raw SQL**: Write clean, readable raw SQL. Use PostGIS functions (e.g., `ST_SetSRID`, `ST_MakePoint`) for location data.
- **Transactions**: Use `db.BeginTxx` for operations involving multiple queries to ensure atomicity.

## 7. Data Modeling & Types

- **IDs**: Always use `uuid.UUID` for primary and foreign keys.
- **Arrays**: Use `pq.StringArray` for simple Postgres string arrays.
- **JSON**: Use `json.RawMessage` or `map[string]interface{}` for JSONB columns.

## 8. Migration Management

> [!IMPORTANT] > **Always use `goose cli` to generate and run migrations.**
> Manual schema changes are strictly prohibited. Use `make migrate-create name=<name>` to generate new migrations.

## 9. Authentication

- Use the `AuthMiddleware` to protect routes.
- Access user information from the Gin context after authentication (e.g., `c.MustGet("userID")`).

## 10. Common Commands

Use the `Makefile` for standard tasks:

- `make dev`: Run with hot reload (Air).
- `make migrate`: Apply migrations.
- `make migrate-create name=<name>`: Create a new migration file.
- `make build`: Build the server binary.
- `make test`: Run the test suite.

## 11. Conventions

- **File Naming**: Use `snake_case.go` for all source files.
- **Function/Type Naming**: Use `PascalCase` for exported and `camelCase` for private symbols.
- **Logging**: Use the standard `log` package. Prefer `log.Printf` or `log.Println` with descriptive prefixes (e.g., `✓`, `🚀`, `⚠️`).

---

_Note: This document is intended to be a living guide and should be updated as the project evolves._
