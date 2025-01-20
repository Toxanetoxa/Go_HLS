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

	// Регистрация
	r.POST("/register", auth.RegisterHandler(connectDB))
	// Авторизация
	r.POST("/login", auth.LoginHandler(connectDB))

	// Защищенные эндпоинты
	authGroup := r.Group("/")
	authGroup.Use(auth.AuthMiddleware())
	{
		// Маршрут для загрузки видео
		authGroup.POST("/videos/upload", videoHandler.UploadVideo)
	}

	// TODO Remove
	//r.POST("/videos/upload", videoHandler.UploadVideo)

	// Маршрут для стриминга видео
	r.GET("/videos/:id/stream", videoHandler.StreamVideo)
	r.GET("/videos/:id/views", videoHandler.GetVideoViews)
	r.GET("/video/:id/info", videoHandler.GetVideoInfo)
	r.GET("/video/:id/chunk", videoHandler.GetVideoChunk)

	err := r.Run(":8080")
	l.Info("Starting server on :8080")
	if err != nil {
		l.Fatal(err.Error())
		return
	}
}
