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
	Views     []View    `gorm:"foreignKey:VideoID"` // Связь с таблицей video_views
}

type View struct {
	ID        uint      `gorm:"primaryKey"`
	VideoID   uint      `gorm:"not null"`
	IPAddress string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
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
