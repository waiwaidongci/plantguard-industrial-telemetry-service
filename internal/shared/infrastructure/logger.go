package infrastructure

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func NewLogger(w io.Writer, level, format string) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	var parsed slog.Level
	switch strings.ToLower(level) {
	case "debug":
		parsed = slog.LevelDebug
	case "warn":
		parsed = slog.LevelWarn
	case "error":
		parsed = slog.LevelError
	default:
		parsed = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: parsed}
	if strings.EqualFold(format, "json") {
		return slog.New(slog.NewJSONHandler(w, opts))
	}
	return slog.New(slog.NewTextHandler(w, opts))
}
