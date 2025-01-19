package video

import (
	"os"
	"path/filepath"
	"time"
)

type Video struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"not null"`
	FilePath  string    `gorm:"not null"`
	AuthorID  uint      `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// EnsureUploadsDir проверяет наличие папки uploads и создаёт её, если она отсутствует.
func EnsureUploadsDir() error {
	uploadsDir := filepath.Join(".", "uploads")
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
