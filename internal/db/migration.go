package db

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"
)

// RunMigrations выполняет миграции базы данных
func RunMigrations(logger *zap.SugaredLogger, db *gorm.DB) {
	logger.Info("Running migrations...")
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get database connection:", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		logger.Fatal("Failed to create migration driver:", err)
	}

	// Используйте абсолютный путь к папке migrations
	m, err := migrate.NewWithDatabaseInstance("file:///www/apps/backend/migrations", "postgres", driver)
	if err != nil {
		logger.Fatal("Failed to create migration instance:", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal("Migration failed:", err)
	}

	logger.Info("Migration completed...")
}
