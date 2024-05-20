package fxlogger

import (
	"time"

	"github.com/ecumenos-social/schemas/formats"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(serviceName string, zapConfig zap.Config) (*zap.Logger, error) {
	zapConfig.EncoderConfig.TimeKey = "time"
	zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(formats.FormatDateTime(t))
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.WithOptions(zap.Fields(zap.String("service", serviceName)))
	return logger, nil
}

func NewProductionLogger(serviceName string) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	return newLogger(serviceName, zapConfig)
}

func NewDevelopmentLogger(serviceName string) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapConfig.Sampling = nil
	return newLogger(serviceName, zapConfig)
}
