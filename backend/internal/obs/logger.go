package obs

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	defaultLogger *slog.Logger
	defaultMu     sync.RWMutex
)

// NewLogger returns a JSON slog.Logger writing to stderr at the given
// level. Levels accepted: "debug" | "info" | "warn" | "error". Anything
// else falls back to "info".
//
// JSON is chosen because Fly captures stderr verbatim and forwards it
// to any configured log drain (Honeycomb, Loki, Grafana Cloud, etc.) —
// structured fields parse cleanly downstream. Locally `fly logs` still
// displays each line readable as a one-liner.
func NewLogger(level string) *slog.Logger {
	lv := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lv})
	return slog.New(h)
}

// SetDefault installs lg as the package-level logger used by Default()
// and as Go's slog default. Idempotent + goroutine-safe so a test
// suite can swap loggers without races.
func SetDefault(lg *slog.Logger) {
	defaultMu.Lock()
	defaultLogger = lg
	defaultMu.Unlock()
	slog.SetDefault(lg)
}

// Default returns the package logger. Falls back to slog.Default() if
// SetDefault has never been called — that fallback isn't strictly
// useful in production (we always set one in main) but keeps tests
// importable in any order.
func Default() *slog.Logger {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	if defaultLogger == nil {
		return slog.Default()
	}
	return defaultLogger
}
