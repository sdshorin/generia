package services

import (
	"context"
	"errors"

	"github.com/sdshorin/generia/internal/repositories"
	"github.com/sdshorin/generia/pkg/auth"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/pkg/models"
	"go.uber.org/zap"
)

// UserService интерфейс для работы с пользователями
type UserService interface {
	Register(ctx context.Context, input *models.UserRegister) (*models.AuthResponse, error)
	Login(ctx context.Context, input *models.UserLogin) (*models.AuthResponse, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

// UserServiceImpl реализует UserService
type UserServiceImpl struct {
	userRepo    repositories.UserRepository
	tokenService auth.TokenService
}

// NewUserService создает новый сервис пользователей
func NewUserService(userRepo repositories.UserRepository, tokenService auth.TokenService) UserService {
	return &UserServiceImpl{
		userRepo:    userRepo,
		tokenService: tokenService,
	}
}

// Register регистрирует нового пользователя
func (s *UserServiceImpl) Register(ctx context.Context, input *models.UserRegister) (*models.AuthResponse, error) {
	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already in use")
	}

	// Проверяем, существует ли пользователь с таким username
	existingUser, err = s.userRepo.GetByUsername(ctx, input.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already in use")
	}

	// Создаем нового пользователя
	user := &models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}

	// Хешируем пароль
	if err := user.HashPassword(); err != nil {
		logger.Logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.New("failed to create user")
	}

	// Сохраняем пользователя в базе данных
	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Logger.Error("Failed to create user", zap.Error(err))
		return nil, errors.New("failed to create user")
	}

	// Генерируем JWT токен
	token, err := s.tokenService.GenerateToken(user.ID)
	if err != nil {
		logger.Logger.Error("Failed to generate token", zap.Error(err))
		return nil, errors.New("failed to generate authentication token")
	}

	// Очищаем пароль для ответа
	user.Password = ""

	return &models.AuthResponse{
		User:  *user,
		Token: token,
	}, nil
}

// Login аутентифицирует пользователя
func (s *UserServiceImpl) Login(ctx context.Context, input *models.UserLogin) (*models.AuthResponse, error) {
	// Получаем пользователя по email или имени пользователя
	user, err := s.userRepo.GetByEmailOrUsername(ctx, input.EmailOrUsername)
	if err != nil {
		logger.Logger.Error("Failed to get user by email or username", 
			zap.Error(err), 
			zap.String("emailOrUsername", input.EmailOrUsername),
		)
		return nil, errors.New("invalid credentials")
	}

	// Проверяем пароль
	if !user.ComparePasswords(input.Password) {
		logger.Logger.Warn("Invalid password attempt", zap.String("username", user.Username))
		return nil, errors.New("invalid credentials")
	}

	// Генерируем JWT токен
	token, err := s.tokenService.GenerateToken(user.ID)
	if err != nil {
		logger.Logger.Error("Failed to generate token", zap.Error(err))
		return nil, errors.New("failed to generate authentication token")
	}

	// Очищаем пароль для ответа
	user.Password = ""

	return &models.AuthResponse{
		User:  *user,
		Token: token,
	}, nil
}

// GetUserByID получает пользователя по ID
func (s *UserServiceImpl) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		logger.Logger.Error("Failed to get user by ID", zap.Error(err), zap.String("id", id))
		return nil, errors.New("user not found")
	}

	// Очищаем пароль для ответа
	user.Password = ""

	return user, nil
}