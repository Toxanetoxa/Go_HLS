package db

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB устанавливает соединение с базой данных и возвращает *gorm.DB
func ConnectDB(logger *zap.SugaredLogger, dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database:", err)
	}
	logger.Info("Connected to the database successfully")

	return db
}
