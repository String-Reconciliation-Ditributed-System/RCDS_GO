package log

import (
	"log/slog"
	"os"
	"strings"
)

// SetLogLevel sets the process-wide structured logger level. Valid levels include
// "debug", "info", "warn", and "error". Legacy zap levels above error are
// accepted and mapped to error.
func SetLogLevel(logLevel []byte) error {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(string(logLevel))) {
	case "dpanic", "panic", "fatal":
		level = slog.LevelError
	default:
		if err := level.UnmarshalText(logLevel); err != nil {
			return err
		}
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
	return nil
}
