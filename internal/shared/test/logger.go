package test

import "go.uber.org/zap"

func NewTestZapLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger
}
