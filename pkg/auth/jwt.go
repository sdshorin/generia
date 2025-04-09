package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sdshorin/generia/pkg/config"
)

// JWTClaims структура для хранения JWT-клеймов
type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// TokenService интерфейс для работы с JWT токенами
type TokenService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}

// JWTService реализует TokenService
type JWTService struct {
	config *config.Config
}

// NewJWTService создает новый JWTService
func NewJWTService(config *config.Config) TokenService {
	return &JWTService{
		config: config,
	}
}

// GenerateToken генерирует новый JWT токен
func (s *JWTService) GenerateToken(userID string) (string, error) {
	// Устанавливаем время истечения
	expirationTime := time.Now().Add(s.config.JWT.Expiration)

	// Создаем клеймы
	claims := &JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "github.com/sdshorin/generia",
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken проверяет JWT токен
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Получаем клеймы
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}