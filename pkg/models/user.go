package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User представляет модель пользователя
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Не возвращаем пароль в JSON
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
	User  User   `json:"user"`
	Token string `json:"token"`
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