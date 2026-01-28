package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
)

// Init initializes the global logger
func Init(service string, env string, level slog.Level) *slog.Logger {
	var handler slog.Handler

	if env == "production" {
		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts).
			WithAttrs([]slog.Attr{
				slog.String("service", service),
				slog.String("env", env),
			})
	} else {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: "15:04:05",
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// ParseLevelFromEnv reads LOG_LEVEL from environment or defaults to INFO
func ParseLevelFromEnv() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// L returns the default global logger
func L() *slog.Logger {
	return slog.Default()
}
