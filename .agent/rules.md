# Maukemana API Development Rules

This document outlines the patterns and rules for developing the Maukemana API. All contributors must follow these guidelines to ensure consistency, security, and maintainability.

## 1) API Contract Rules (Stability & Structure)

- **Versioning**: Always version your API (e.g., `/api/v1`) to allow for backward compatibility. Use route groups in `internal/router/router.go`.
- **Consistent Response Envelope**:
  - Utilize `internal/utils/response.go` for all responses to maintain the contract:
    ```json
    { "success": true, "message": "...", "data": {...}, "error": null }
    ```
  - **Success**: Use `utils.SendSuccess(c, "Message", data)` or `utils.SendCreated`.
  - **Error**: Use `utils.SendError(c, code, "Message", err)`.
- **Pagination**: Always paginate large datasets.
  - Use helpers: `page, limit := utils.GetPagination(c)`.
  - Return metadata using `utils.SendPaginated(c, "Message", data, page, limit, total)`.
- **Input Validation**: Use **Gin/Validator** struct tags.
  - Define structs with proper binding tags (e.g., `binding:"required,url"`).
  - Always handle `c.ShouldBindJSON` errors with `utils.SendValidationError`.
- **Idempotency**: For critical resource creation, consider implementing `Idempotency-Key` headers to prevent duplicate submissions.

## 2) Database Discipline (SQLX + PostGIS)

- **Use SQLX**: We use `sqlx` for database interactions.
  - **Named Parameters**: Always use named parameters (`:name`) over positional placeholders (`$1`) for readability.
  - **Type Handling**: Use `pq.StringArray` for arrays and `json.RawMessage` for JSONB.
- **Migrations**: Use **goose** for schema migrations.
  - All schema changes must be done via migrations: `make migrate-create name=<task>`.
  - **Never modify the database schema manually in production.**
- **Indexes**: Create indexes for frequently filtered and sorted columns (e.g., `WHERE created_at > ?` or PostGIS `GIST` indexes for locations).
- **Transactions**: Always wrap multiple write operations in a transaction using `db.BeginTxx`.
- **Soft Deletes**: Implement soft deletes using a `deleted_at` field where appropriate. Ensure queries filter out deleted records.

## 3) Security (Core Principles)

- **Authentication**: Use **Clerk** via `internal/auth` middleware.
  - Protect routes using the `AuthMiddleware`.
  - Access permissions/IDs via Gin context (e.g., `c.MustGet("userID")`).
- **Authorization**: Enforce role-based access control (RBAC) in middleware or handlers before performing sensitive actions.
- **Input Validation & Sanitization**:
  - Never trust user input. Validate strict types and formats.
  - Sanitize inputs to prevent XSS and SQL injection (use SQLX placeholders).
- **Rate Limiting**: Protect critical endpoints using `golang.org/x/time` based rate limiters (or Gin middleware equivalents).

## 4) Error Handling and Observability

- **Structured Error Responses**: Standardize error returns using `internal/utils/response.go`.
  - Map internal errors to user-friendly messages.
  - Maintain specific HTTP status codes (2xx, 4xx, 5xx).
- **Logging**: Use consistent logging (Standard `log` or structured logger).
  - Log request details, errors, and context for debugging.
- **Observability (OpenTelemetry)**:
  - The project is instrumented with **OpenTelemetry**.
  - Use tracing to monitor request flow and database performance (`otelsqlx`).

## 5) Performance & Scalability

- **Connection Pooling**: Configure `sqlx` connection pool settings (`SetMaxOpenConns`, `SetMaxIdleConns`) appropriately for the environment.
- **Asynchronous Processing**: Use Go routines for non-blocking tasks (logging, metrics, background jobs).
- **Caching**: Implement caching for expensive read operations (e.g., complex geospatial queries).
- **Optimize Queries**:
  - Use `EXPLAIN ANALYZE` to debug slow queries.
  - Avoid N+1 query problems.

## 6) API Design & Best Practices

- **RESTful Endpoints**:
  - `GET`: Fetch resources.
  - `POST`: Create new resources.
  - `PUT/PATCH`: Update resources.
  - `DELETE`: Remove protocols (Soft or Hard).
- **Folder Structure**:
  - `/internal/handlers`: Controllers/Logic.
  - `/internal/repositories`: Database interactions.
  - `/internal/models`: Data structures.
  - `/internal/router`: Route definitions.

## 7) Code Quality & Maintainability

- **Conventions**:
  - Files: `snake_case.go`.
  - Functions: `PascalCase` (Public), `camelCase` (Private).
- **Dependency Injection**: Pass repositories to handlers to facilitate testing.
- **Testing**:
  - Unit tests for core logic.
  - Integration tests for handlers and repositories using the `Makefile` (`make test`).

## 8) CI/CD & Deployment

- **Pipeline**: Ensure `make build` and `make test` pass before merging.
- **Environment**: Manage configuration via `.env` files (local) and environment variables (production).
- **Zero-Downtime**: Ensure migrations are backward compatible.

## 9) Security and Compliance

- **HTTPS**: Enforce TLS.
- **Data Encryption**: Encrypt sensitive data at rest and in transit.
