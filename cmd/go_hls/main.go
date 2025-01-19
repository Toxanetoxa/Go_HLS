package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/toxanetoxa/gohls/internal/auth"
	"github.com/toxanetoxa/gohls/internal/db"
	"github.com/toxanetoxa/gohls/internal/video"
	"github.com/toxanetoxa/gohls/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func loadEnv(logger *zap.SugaredLogger) {
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using system environment variables")
	} else {
		logger.Info(".env file loaded successfully")
	}
}

func main() {
	l := logger.InitLogger()
	defer func(myLogger *zap.SugaredLogger) {
		err := myLogger.Sync()
		if err != nil {
			_ = fmt.Errorf(err.Error())
		}
	}(l)

	loadEnv(l)

	// Получаем конфиг для подключения к бд
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
	connectDB := db.ConnectDB(l, dsn)

	r := gin.Default()

	r.MaxMultipartMemory = 100 << 20

	videoHandler := video.NewVideoHandler(connectDB)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Success"})
	})

	// Маршрут для видео
	r.POST("/videos/upload", videoHandler.UploadVideo)
	r.GET("/videos/:id/stream", videoHandler.StreamVideo)

	// Регистрация
	r.POST("/register", auth.RegisterHandler(connectDB))
	// Авторизация
	r.POST("/login", auth.LoginHandler(connectDB))

	// Защищенные эндпоинты
	authGroup := r.Group("/")
	authGroup.Use(auth.AuthMiddleware())
	{
		authGroup.GET("/protected", func(c *gin.Context) {
			username := c.GetString("username")
			c.JSON(http.StatusOK, gin.H{"message": "Hello, " + username})
		})
	}

	err := r.Run(":8080")
	l.Info("Starting server on :8080")
	if err != nil {
		l.Fatal(err.Error())
		return
	}
}
