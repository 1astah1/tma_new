package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func Init(env string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "development" {
		opts.Level = slog.LevelDebug
		opts.AddSource = true
		Logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	} else {
		Logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	slog.SetDefault(Logger)
}
