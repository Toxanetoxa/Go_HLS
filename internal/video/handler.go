package video

import (
	"errors"
	"fmt"
	"github.com/toxanetoxa/gohls/internal/user"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// TODO перенести сигнатуры в video.go

type ActiveViewers struct {
	mu      sync.Mutex
	viewers map[string]map[*websocket.Conn]bool // videoID -> connections
}

func NewActiveViewers() *ActiveViewers {
	return &ActiveViewers{
		viewers: make(map[string]map[*websocket.Conn]bool),
	}
}

// AddViewer добавляет зрителя для указанного видео.
func (av *ActiveViewers) AddViewer(videoID string, conn *websocket.Conn) {
	av.mu.Lock()
	defer av.mu.Unlock()

	if av.viewers[videoID] == nil {
		av.viewers[videoID] = make(map[*websocket.Conn]bool)
	}
	av.viewers[videoID][conn] = true
}

// RemoveViewer удаляет зрителя для указанного видео.
func (av *ActiveViewers) RemoveViewer(videoID string, conn *websocket.Conn) {
	av.mu.Lock()
	defer av.mu.Unlock()

	delete(av.viewers[videoID], conn)
	if len(av.viewers[videoID]) == 0 {
		delete(av.viewers, videoID)
	}
}

// GetViewers возвращает количество активных зрителей для указанного видео.
func (av *ActiveViewers) GetViewers(videoID string) int {
	av.mu.Lock()
	defer av.mu.Unlock()

	return len(av.viewers[videoID])
}

// Handler обрабатывает запросы, связанные с видео.
type Handler struct {
	DB            *gorm.DB
	ActiveViewers *ActiveViewers
}

// NewVideoHandler создаёт новый экземпляр Handler.
func NewVideoHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:            db,
		ActiveViewers: NewActiveViewers(),
	}
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

	// Отправляем количество активных зрителей через WebSocket
	go func() {
		for {
			time.Sleep(5 * time.Second) // Обновляем каждые 5 секунд
			activeViewers := h.ActiveViewers.GetViewers(videoID)
			for conn := range h.ActiveViewers.viewers[videoID] {
				if err := conn.WriteJSON(gin.H{"active_viewers": activeViewers}); err != nil {
					h.ActiveViewers.RemoveViewer(videoID, conn)
				}
			}
		}
	}()
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешить все источники (в реальном проекте настройте правильно)
	},
}

// ActiveViewersWS обрабатывает WebSocket-соединения для отслеживания активных зрителей.
func (h *Handler) ActiveViewersWS(c *gin.Context) {
	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID is required"})
		return
	}

	// Обновляем HTTP-соединение до WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket", "details": err.Error()})
		return
	}
	defer conn.Close()

	// Добавляем зрителя
	h.ActiveViewers.AddViewer(videoID, conn)
	defer h.ActiveViewers.RemoveViewer(videoID, conn)

	// Бесконечный цикл для поддержания соединения
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// GetVideoInfo возвращает информацию о видео.
func (h *Handler) GetVideoInfo(c *gin.Context) {
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

	// Получаем информацию о файле
	fileInfo, err := os.Stat(v.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info", "details": err.Error()})
		return
	}

	// Возвращаем информацию о видео
	c.JSON(http.StatusOK, gin.H{
		"video_id":  v.ID,
		"title":     v.Title,
		"file_size": fileInfo.Size(),
		"duration":  "TODO: Добавьте логику для получения длительности видео", // TODO: Добавьте логику для получения длительности видео
	})
}

// GetVideoChunk возвращает конкретную часть видеофайла.
func (h *Handler) GetVideoChunk(c *gin.Context) {
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

	// Открываем файл
	file, err := os.Open(v.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open video file", "details": err.Error()})
		return
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info", "details": err.Error()})
		return
	}

	// Обрабатываем Range-заголовок
	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Range header is required"})
		return
	}

	// Парсим Range-заголовок
	ranges, err := parseRange(rangeHeader, fileInfo.Size())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid range header", "details": err.Error()})
		return
	}

	// Если запрошен один диапазон
	if len(ranges) == 1 {
		start := ranges[0].start
		end := ranges[0].end

		// Устанавливаем заголовки для частичного контента
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
		c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
		c.Status(http.StatusPartialContent)

		// Читаем и отправляем запрошенную часть файла
		_, err = file.Seek(start, io.SeekStart)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seek file", "details": err.Error()})
			return
		}

		_, err = io.CopyN(c.Writer, file, end-start+1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send file chunk", "details": err.Error()})
			return
		}
	} else {
		// Если запрошено несколько диапазонов, возвращаем ошибку (можно реализовать поддержку множественных диапазонов)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Multiple ranges are not supported"})
		return
	}
}

// parseRange парсит Range-заголовок и возвращает список диапазонов.
func parseRange(rangeHeader string, fileSize int64) ([]struct{ start, end int64 }, error) {
	// Пример Range-заголовка: "bytes=0-499"
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil, errors.New("invalid range header")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	ranges := strings.Split(rangeStr, ",")

	var result []struct{ start, end int64 }

	for _, r := range ranges {
		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, errors.New("invalid range format")
		}

		start, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, errors.New("invalid start value")
		}

		end, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, errors.New("invalid end value")
		}

		if start < 0 || end >= fileSize || start > end {
			return nil, errors.New("invalid range values")
		}

		result = append(result, struct{ start, end int64 }{start, end})
	}

	return result, nil
}
