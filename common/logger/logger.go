package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/DeRuina/timberjack"
)

const (
	maxFileSize      = 100
	maxBackups       = 7
	maxAge           = 30
	rotationInterval = 24
)

func Setup() *slog.Logger {
	logRotator := timberjack.Logger{
		Filename:         "./logs/api.log",
		MaxSize:          maxFileSize,
		MaxBackups:       maxBackups,
		MaxAge:           maxAge,
		RotationInterval: rotationInterval * time.Hour,
	}

	return slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, &logRotator), nil))
}
