package main

import (
	"fmt"
	"github.com/toxanetoxa/gohls/internal/db"
	"github.com/toxanetoxa/gohls/pkg/logger"
	"go.uber.org/zap"
	"os"
)

func main() {
	l := logger.InitLogger()
	defer func(myLogger *zap.SugaredLogger) {
		err := myLogger.Sync()
		if err != nil {
			_ = fmt.Errorf(err.Error())
		}
	}(l)

	// LoadEnv(l)

	// Получаем конфигурацию для подключения к бд
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSL"),
	)

	// Подключаемся к базе
	dbConn := db.ConnectDB(l, dsn)

	// Запускаем миграции
	db.RunMigrations(l, dbConn)
}
