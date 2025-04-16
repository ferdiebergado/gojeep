package logging

import (
	"io"
	"log/slog"
	"strings"

	"github.com/ferdiebergado/gopherkit/env"
)

func SetLogger(out io.Writer, appEnv string, level string) {
	if level == "" {
		level = env.Get("SERVER_LOG_LEVEL", "info")
	}

	opts := &slog.HandlerOptions{
		Level: stringToLogLevel(level),
	}

	var handler slog.Handler

	if appEnv == "production" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		opts.AddSource = true
		handler = slog.NewTextHandler(out, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func stringToLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARNING", "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
