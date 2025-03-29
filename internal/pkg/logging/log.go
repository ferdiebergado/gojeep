package logging

import (
	"io"
	"log/slog"

	"github.com/ferdiebergado/gopherkit/env"
)

func SetLogger(out io.Writer, appEnv string) {
	logLevel := new(slog.LevelVar)
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler

	if appEnv == "production" {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		if env.GetBool("DEBUG", false) {
			logLevel.Set(slog.LevelDebug)
		}

		opts.AddSource = true
		handler = slog.NewTextHandler(out, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
