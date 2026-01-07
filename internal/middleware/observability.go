package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"maukemana-backend/internal/utils"
)

// Observability returns a middleware that handles:
// 1. Request ID generation/propagation
// 2. Access logging (JSON)
// 3. Centralized error logging
// 4. Panic recovery
func Observability() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 1. Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// 2. Trace/Span extraction from context (set by otelgin in router)
		span := trace.SpanFromContext(c.Request.Context())

		// Set Request ID on span
		span.SetAttributes(attribute.String("request_id", requestID))

		// Panic Recovery
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				slog.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(stack)),
					slog.String("request_id", requestID),
					slog.String("method", c.Request.Method),
					slog.String("path", path),
				)

				// Record error in span
				span := trace.SpanFromContext(c.Request.Context())
				span.RecordError(fmt.Errorf("%v", err))
				span.SetAttributes(
					semconv.ExceptionStacktrace(string(stack)),
				)

				// Use project's existing error utility if possible
				utils.SendError(c, http.StatusInternalServerError, "Internal Server Error", nil)
				c.Abort()
			}
		}()

		// Process request
		c.Next()

		// Skip logging for health and docs
		if path == "/health" || path == "/api" {
			return
		}

		// 2. Request Completion Log
		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		fields := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		// Add Trace ID and Span ID for log correlation
		if span.SpanContext().IsValid() {

			fields = append(fields,
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)
		}

		// Add user_id if present (from auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, slog.Any("user_id", userID))
		}

		// 3. Centralized Error Log
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				slog.Error("request error",
					append(fields, slog.String("error", e.Error()))...,
				)
			}
		} else {
			slog.Info("request completed", fields...)
		}
	}
}
