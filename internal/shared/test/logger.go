package test

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func NewTestZapLogger() *zap.Logger {
	logger, err := zap.NewDevelopment(zap.Fields(
		zap.String("trace_id", uuid.NewString()),
		zap.Uint32("user_id", 1),
	))
	if err != nil {
		panic(err)
	}
	return logger
}
