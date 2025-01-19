package logger

import (
	"go.uber.org/zap"
)

// Logger глобальный экземпляр логгера
var Logger *zap.SugaredLogger

// InitLogger инициализирует логгер и возвращает его
func InitLogger() *zap.SugaredLogger {
	log, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	Logger = log.Sugar()
	return Logger
}
