package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User представляет модель пользователя
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"-"` // Не возвращаем пароль в JSON
	IsAI      bool      `json:"is_ai"`
	WorldID   string    `json:"world_id,omitempty"` // ID мира для AI-пользователей
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRegister представляет тело запроса на регистрацию пользователя
type UserRegister struct {
	Username string `json:"username" validate:"required,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserLogin представляет тело запроса на аутентификацию пользователя
type UserLogin struct {
	EmailOrUsername string `json:"email_or_username" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

// AuthResponse представляет ответ на запрос аутентификации
type AuthResponse struct {
	User       User   `json:"user"`
	Token      string `json:"token"`
	ActiveWorld string `json:"active_world,omitempty"` // ID активного мира для пользователя
}

// World представляет модель мира
type World struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	Prompt           string    `json:"prompt"`
	CreatorID        string    `json:"creator_id,omitempty"`
	GenerationStatus string    `json:"generation_status"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// WorldCreateRequest представляет тело запроса на создание мира
type WorldCreateRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description"`
	Prompt      string `json:"prompt" validate:"required,min=10"`
}

// WorldResponse представляет ответ с информацией о мире
type WorldResponse struct {
	World       World  `json:"world"`
	UsersCount  int    `json:"users_count"`
	PostsCount  int    `json:"posts_count"`
	IsJoined    bool   `json:"is_joined"`
	HasAccess   bool   `json:"has_access"`
	IsActive    bool   `json:"is_active"`
}

// SetActiveWorldRequest представляет тело запроса на установку активного мира
type SetActiveWorldRequest struct {
	WorldID string `json:"world_id" validate:"required"`
}

// HashPassword хеширует пароль пользователя
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// ComparePasswords сравнивает предоставленный пароль с хешированным паролем пользователя
func (u *User) ComparePasswords(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}