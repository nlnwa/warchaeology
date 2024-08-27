package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/nlnwa/warchaeology/cmd/internal/flag"
)

func getWriter(w io.WriteCloser) (io.WriteCloser, error) {
	logFileName := flag.LogFileName()
	if logFileName == "-" {
		return w, nil
	}
	logFile, err := os.Create(logFileName)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func InitLogger(w io.WriteCloser) (io.Closer, error) {
	w, err := getWriter(w)
	if err != nil {
		return nil, err
	}

	levelVar := new(slog.LevelVar)
	levelVar.Set(toLogLevel(flag.LogLevel()))

	opts := &slog.HandlerOptions{Level: levelVar}

	var handler slog.Handler
	if flag.LogFormat() == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	slog.SetDefault(slog.New(handler))

	return w, nil
}

func toLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
