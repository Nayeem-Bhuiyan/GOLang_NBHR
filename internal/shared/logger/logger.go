package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

// New initializes and returns a structured zap logger.
func New(env string) *zap.Logger {
	once.Do(func() {
		var cfg zap.Config
		if env == "production" {
			cfg = zap.NewProductionConfig()
		} else {
			cfg = zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}

		var err error
		instance, err = cfg.Build(zap.AddCallerSkip(0))
		if err != nil {
			// Fallback to no-op logger on build failure
			instance = zap.NewNop()
			os.Stderr.WriteString("failed to build logger: " + err.Error() + "\n")
		}
	})
	return instance
}

// Get returns the initialized logger instance.
func Get() *zap.Logger {
	if instance == nil {
		return New("development")
	}
	return instance
}