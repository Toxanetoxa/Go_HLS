package video

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler обрабатывает запросы, связанные с видео.
type Handler struct {
	DB *gorm.DB
}

// NewVideoHandler создаёт новый экземпляр Handler.
func NewVideoHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

// UploadVideo обрабатывает загрузку видео.
func (h *Handler) UploadVideo(c *gin.Context) {
	// Убедимся, что папка uploads существует
	if err := EnsureUploadsDir(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory", "details": err.Error()})
		return
	}

	// Получаем файл из запроса
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file", "details": err.Error()})
		return
	}

	// Получаем заголовок видео
	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	// Получаем ID автора
	authorID, err := strconv.Atoi(c.PostForm("author_id"))
	if err != nil || authorID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	// Сохраняем файл на сервере
	filePath := filepath.Join("uploads", file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file", "details": err.Error()})
		return
	}

	// Создаём запись о видео в базе данных
	video := Video{
		Title:    title,
		FilePath: filePath,
		AuthorID: uint(authorID),
	}

	if err := h.DB.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video metadata", "details": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"message":  "Video uploaded successfully",
		"video_id": video.ID,
	})
}
