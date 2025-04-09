package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"instagram-clone/pkg/logger"
	"instagram-clone/pkg/middleware"
	"instagram-clone/pkg/models"
	"instagram-clone/pkg/services"
	"go.uber.org/zap"
)

// UserHandler обработчик для работы с пользователями
type UserHandler struct {
	userService services.UserService
	validate    *validator.Validate
}

// NewUserHandler создает новый обработчик пользователей
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// Register обрабатывает запрос на регистрацию пользователя
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.UserRegister

	// Декодируем JSON запрос
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		logger.Logger.Warn("Failed to decode registration request", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Валидируем данные
	err = h.validate.Struct(input)
	if err != nil {
		logger.Logger.Warn("Validation error", zap.Error(err))
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя
	response, err := h.userService.Register(r.Context(), &input)
	if err != nil {
		logger.Logger.Error("Registration error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login обрабатывает запрос на аутентификацию пользователя
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.UserLogin

	// Декодируем JSON запрос
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		logger.Logger.Warn("Failed to decode login request", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Валидируем данные
	err = h.validate.Struct(input)
	if err != nil {
		logger.Logger.Warn("Validation error", zap.Error(err))
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Аутентифицируем пользователя
	response, err := h.userService.Login(r.Context(), &input)
	if err != nil {
		logger.Logger.Error("Login error", zap.Error(err))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Me обрабатывает запрос на получение информации о текущем пользователе
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста запроса
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err), zap.String("userID", userID))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}