/*
 * Contract ID: CTR-006
 * Service Name: SupermarketService
 * Description: Zap structured logger bootstrapper. Configured to emit JSON formatted
 *              structure logs to stdout for ELK/Graylog ingestion.
 */

package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger initializes and returns a configured *zap.Logger.
// It maps the log level from the config and formats logs in production-ready structured JSON.
// Implements requirements outlined in HLD and LLD Section 10.1.
func NewLogger(cfg *Config) (*zap.Logger, error) {
	var level zapcore.Level

	// Parse configured log level (defaulting to info if invalid or empty)
	if err := level.UnmarshalText([]byte(cfg.LOG_LEVEL)); err != nil {
		level = zapcore.InfoLevel
	}

	// Build production-ready encoder config matching Graylog/ELK indexing patterns
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder, // E.g., "2025-02-14T15:30:10.456Z"
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Direct logs to stdout
	writeSyncer := zapcore.AddSync(os.Stdout)

	// Create JSON core with specified encoder config and level
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		level,
	)

	// Build the logger, adding Caller info and automatic stacktraces for Errors and above
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}