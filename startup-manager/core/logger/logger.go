package logger

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(message string, args ...zap.Field)
	Info(message string, args ...zap.Field)
	Warn(message string, args ...zap.Field)
	Error(message string, args ...zap.Field)
	Fatal(message string, args ...zap.Field)
}

var defaultZapConfig = zap.Config{
	Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
	Development: false,
	Sampling: &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	},
	Encoding: "json",
	EncoderConfig: zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	},
	OutputPaths:      []string{"stdout"},
	ErrorOutputPaths: []string{"stderr"},
}

// NewDefaultLogger returns you a new otelzap logger with core specific default settings
func NewDefaultLogger() (*otelzap.Logger, error) {
	zapLogger, err := defaultZapConfig.Build()
	if err != nil {
		return nil, err
	}

	zapLogger.WithOptions(zap.AddStacktrace(zap.PanicLevel))

	logger := otelzap.New(zapLogger, otelzap.WithTraceIDField(true), otelzap.WithMinLevel(zap.DebugLevel))

	return logger, nil
}
