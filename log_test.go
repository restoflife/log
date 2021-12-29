package log

import (
	"errors"
	"go.uber.org/zap"
	"testing"
)

func TestNew(t *testing.T) {
	New(&Config{
		Level:    "error",
		Filename: "error.log",
	})
	defer Sync()
	Info("info", zap.String("level", "info"))
	Debug("debug", zap.String("level", "debug"))
	Error("error", zap.Error(errors.New("error")))
}
