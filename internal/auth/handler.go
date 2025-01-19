package auth

import (
	"github.com/toxanetoxa/gohls/internal/user"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var jwtKey = []byte("your_secret_key") // Замените на ваш секретный ключ

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем, существует ли пользователь с таким же именем или email
		var existingUser user.User
		if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
			return
		}

		// Создаем нового пользователя
		u := user.User{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		}

		// Хешируем пароль
		if err := u.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Сохраняем пользователя в базу данных
		if err := db.Create(&u).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}

func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ищем пользователя по имени
		var u user.User
		if err := db.Where("username = ?", req.Username).First(&u).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Проверяем пароль
		if !u.CheckPassword(req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Создаем JWT
		expirationTime := time.Now().Add(24 * time.Hour) // Токен действителен 24 часа
		claims := &jwt.RegisteredClaims{
			Subject:   u.Username,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Возвращаем токен
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}
