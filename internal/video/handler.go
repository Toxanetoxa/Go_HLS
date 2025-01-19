package video

import (
	"github.com/pkg/errors"
	"github.com/toxanetoxa/gohls/internal/user"
	"net/http"
	"os"
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

	// Получаем имя пользователя из контекста (установленного в middleware)
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Ищем пользователя по имени
	var u user.User
	if err := h.DB.Where("username = ?", username).First(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user", "details": err.Error()})
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
		AuthorID: u.ID,
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

// StreamVideo обрабатывает запрос на стриминг видео.
func (h *Handler) StreamVideo(c *gin.Context) {
	// Получаем ID видео из параметров запроса
	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
		return
	}

	// Ищем видео в базе данных
	var v Video
	if err := h.DB.First(&v, videoID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch video", "details": err.Error()})
		return
	}

	// Получаем IP-адрес пользователя
	ip := c.ClientIP()

	// Проверяем, был ли IP-адрес уже учтён
	var existingView View
	if err := h.DB.Where("video_id = ? AND ip_address = ?", v.ID, ip).First(&existingView).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// IP-адрес не найден, добавляем новый просмотр
			view := View{
				VideoID:   v.ID,
				IPAddress: ip,
			}

			if err := h.DB.Create(&view).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save view", "details": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check view", "details": err.Error()})
			return
		}
	}

	// Открываем файл
	file, err := os.Open(v.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open video file", "details": err.Error()})
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close file", "details": err.Error()})
		}
	}(file)

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info", "details": err.Error()})
		return
	}

	// Устанавливаем заголовки для стриминга
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", "video/mp4") // Укажите правильный MIME-тип для вашего видео
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// Используем http.ServeContent для обработки Range-запросов
	http.ServeContent(c.Writer, c.Request, fileInfo.Name(), fileInfo.ModTime(), file)
}

// GetVideoViews получение общего количества просмотров
func (h *Handler) GetVideoViews(c *gin.Context) {
	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
		return
	}

	var count int64
	if err := h.DB.Model(&View{}).Where("video_id = ?", videoID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch view count", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"views": count})
}
