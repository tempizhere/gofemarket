package logger

import (
	"go.uber.org/zap"
)

// NewLogger создает новый логгер с уровнем Info.
func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	return cfg.Build()
}
