package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sdshorin/generia/pkg/kafka"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/world-service/internal/models"
	"github.com/sdshorin/generia/services/world-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
	postpb "github.com/sdshorin/generia/api/grpc/post"
	worldpb "github.com/sdshorin/generia/api/grpc/world"
)

// WorldService implements the world gRPC service
type WorldService struct {
	worldpb.UnimplementedWorldServiceServer
	worldRepo      repository.WorldRepository
	authClient     authpb.AuthServiceClient
	postClient     postpb.PostServiceClient
	kafkaProducer  *kafka.Producer
}

// NewWorldService creates a new WorldService
func NewWorldService(
	worldRepo repository.WorldRepository,
	authClient authpb.AuthServiceClient,
	postClient postpb.PostServiceClient,
	kafkaBrokers []string,
) worldpb.WorldServiceServer {
	return &WorldService{
		worldRepo:     worldRepo,
		authClient:    authClient,
		postClient:    postClient,
		kafkaProducer: kafka.NewProducer(kafkaBrokers),
	}
}

// CreateWorld handles creating a new world
func (s *WorldService) CreateWorld(ctx context.Context, req *worldpb.CreateWorldRequest) (*worldpb.WorldResponse, error) {
	// Validate input
	if req.UserId == "" || req.Name == "" || req.Prompt == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id, name, and prompt are required")
	}

	// Validate user exists
	_, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: req.UserId,
	})
	if err != nil {
		logger.Logger.Error("Failed to validate user", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Errorf(codes.Internal, "failed to validate user")
	}

	// Create world
	world := &models.World{
		Name:             req.Name,
		Description:      req.Description,
		Prompt:           req.Prompt,
		CreatorID:        req.UserId,
		GenerationStatus: models.GenerationStatusPending,
		Status:           models.WorldStatusActive,
	}

	err = s.worldRepo.Create(ctx, world)
	if err != nil {
		logger.Logger.Error("Failed to create world", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create world")
	}

	// Add creator to world members
	err = s.worldRepo.AddUserToWorld(ctx, req.UserId, world.ID)
	if err != nil {
		logger.Logger.Error("Failed to add creator to world",
			zap.Error(err),
			zap.String("user_id", req.UserId),
			zap.String("world_id", world.ID))
		return nil, status.Errorf(codes.Internal, "failed to add creator to world")
	}

	// Get world stats
	usersCount, postsCount, err := s.worldRepo.GetWorldStats(ctx, world.ID)
	if err != nil {
		logger.Logger.Error("Failed to get world stats", zap.Error(err), zap.String("world_id", world.ID))
		// Not a critical error, continue with zeros
		usersCount = 0
		postsCount = 0
	}

	// Create and enqueue content generation tasks
	s.createInitialGenerationTasks(ctx, world.ID)

	// Build response
	return &worldpb.WorldResponse{
		Id:               world.ID,
		Name:             world.Name,
		Description:      world.Description,
		Prompt:           world.Prompt,
		CreatorId:        world.CreatorID,
		GenerationStatus: world.GenerationStatus,
		Status:           world.Status,
		UsersCount:       int32(usersCount),
		PostsCount:       int32(postsCount),
		CreatedAt:        world.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        world.UpdatedAt.Format(time.RFC3339),
		IsJoined:         true,
	}, nil
}

// GetWorld handles getting a world by ID
func (s *WorldService) GetWorld(ctx context.Context, req *worldpb.GetWorldRequest) (*worldpb.WorldResponse, error) {
	// Validate input
	if req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "world_id is required")
	}

	// Get world
	world, err := s.worldRepo.GetByID(ctx, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get world")
	}

	if world == nil {
		return nil, status.Errorf(codes.NotFound, "world not found")
	}

	// Get world stats
	usersCount, postsCount, err := s.worldRepo.GetWorldStats(ctx, world.ID)
	if err != nil {
		logger.Logger.Error("Failed to get world stats", zap.Error(err), zap.String("world_id", world.ID))
		// Not a critical error, continue with zeros
		usersCount = 0
		postsCount = 0
	}

	// Check if user has joined this world
	isJoined := false
	if req.UserId != "" {
		hasAccess, err := s.worldRepo.CheckUserWorld(ctx, req.UserId, world.ID)
		if err != nil {
			logger.Logger.Error("Failed to check user access to world",
				zap.Error(err),
				zap.String("user_id", req.UserId),
				zap.String("world_id", world.ID))
			// Not a critical error, continue
		} else {
			isJoined = hasAccess
		}
	}

	// Build response
	return &worldpb.WorldResponse{
		Id:               world.ID,
		Name:             world.Name,
		Description:      world.Description,
		Prompt:           world.Prompt,
		CreatorId:        world.CreatorID,
		GenerationStatus: world.GenerationStatus,
		Status:           world.Status,
		UsersCount:       int32(usersCount),
		PostsCount:       int32(postsCount),
		CreatedAt:        world.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        world.UpdatedAt.Format(time.RFC3339),
		IsJoined:         isJoined,
	}, nil
}

// GetWorlds handles getting all worlds available to a user
func (s *WorldService) GetWorlds(ctx context.Context, req *worldpb.GetWorldsRequest) (*worldpb.WorldsResponse, error) {
	// Validate input
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	// Set default limit if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Get worlds
	worlds, total, err := s.worldRepo.GetByUser(ctx, req.UserId, limit, offset, req.Status)
	if err != nil {
		logger.Logger.Error("Failed to get worlds",
			zap.Error(err),
			zap.String("user_id", req.UserId),
			zap.Int("limit", limit),
			zap.Int("offset", offset))
		return nil, status.Errorf(codes.Internal, "failed to get worlds")
	}

	// Build response
	worldResponses := make([]*worldpb.WorldResponse, len(worlds))
	for i, world := range worlds {
		// Get world stats
		usersCount, postsCount, err := s.worldRepo.GetWorldStats(ctx, world.ID)
		if err != nil {
			logger.Logger.Error("Failed to get world stats", zap.Error(err), zap.String("world_id", world.ID))
			// Not a critical error, continue with zeros
			usersCount = 0
			postsCount = 0
		}

		worldResponses[i] = &worldpb.WorldResponse{
			Id:               world.ID,
			Name:             world.Name,
			Description:      world.Description,
			Prompt:           world.Prompt,
			CreatorId:        world.CreatorID,
			GenerationStatus: world.GenerationStatus,
			Status:           world.Status,
			UsersCount:       int32(usersCount),
			PostsCount:       int32(postsCount),
			CreatedAt:        world.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        world.UpdatedAt.Format(time.RFC3339),
			IsJoined:         true, // User has access since this is from user-specific query
		}
	}

	return &worldpb.WorldsResponse{
		Worlds: worldResponses,
		Total:  int32(total),
	}, nil
}

// JoinWorld handles a user joining a world
func (s *WorldService) JoinWorld(ctx context.Context, req *worldpb.JoinWorldRequest) (*worldpb.JoinWorldResponse, error) {
	// Validate input
	if req.UserId == "" || req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id and world_id are required")
	}

	// Validate user exists
	_, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: req.UserId,
	})
	if err != nil {
		logger.Logger.Error("Failed to validate user", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Errorf(codes.Internal, "failed to validate user")
	}

	// Validate world exists
	world, err := s.worldRepo.GetByID(ctx, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get world")
	}

	if world == nil {
		return nil, status.Errorf(codes.NotFound, "world not found")
	}

	// Check if user already has access
	hasAccess, err := s.worldRepo.CheckUserWorld(ctx, req.UserId, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to check user access to world",
			zap.Error(err),
			zap.String("user_id", req.UserId),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to check user access to world")
	}

	if hasAccess {
		return &worldpb.JoinWorldResponse{
			Success: true,
			Message: "User already has access to this world",
		}, nil
	}

	// Add user to world
	err = s.worldRepo.AddUserToWorld(ctx, req.UserId, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to add user to world",
			zap.Error(err),
			zap.String("user_id", req.UserId),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to add user to world")
	}

	return &worldpb.JoinWorldResponse{
		Success: true,
		Message: "User joined world successfully",
	}, nil
}

// GenerateContent handles generating AI content for a world
func (s *WorldService) GenerateContent(ctx context.Context, req *worldpb.GenerateContentRequest) (*worldpb.GenerateContentResponse, error) {
	// Validate input
	if req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "world_id is required")
	}

	// Validate world exists
	world, err := s.worldRepo.GetByID(ctx, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get world")
	}

	if world == nil {
		return nil, status.Errorf(codes.NotFound, "world not found")
	}

	// Check if world is already being generated and not forced
	if world.GenerationStatus == models.GenerationStatusInProgress {
		return &worldpb.GenerateContentResponse{
			Success: false,
			Message: "World generation is already in progress",
		}, nil
	}

	// Update world generation status
	err = s.worldRepo.UpdateGenerationStatus(ctx, req.WorldId, models.GenerationStatusInProgress)
	if err != nil {
		logger.Logger.Error("Failed to update world generation status",
			zap.Error(err),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to update world generation status")
	}

	// Create generation tasks
	usersCount := int(req.UsersCount)
	if usersCount <= 0 || usersCount > 100 {
		usersCount = 10 // Default to 10 AI users
	}

	postsCount := int(req.PostsCount)
	if postsCount <= 0 || postsCount > 700 {
		postsCount = 50 // Default to 50 posts
	}

	// Create users generation task
	usersTask := &models.WorldGenerationTask{
		WorldID:    req.WorldId,
		TaskType:   models.TaskTypeUsers,
		Status:     models.TaskStatusPending,
		Parameters: fmt.Sprintf(`{"count": %d, "world_prompt": "%s"}`, usersCount, world.Prompt),
	}

	err = s.worldRepo.CreateGenerationTask(ctx, usersTask)
	if err != nil {
		logger.Logger.Error("Failed to create users generation task",
			zap.Error(err),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to create users generation task")
	}

	// Create posts generation task
	postsTask := &models.WorldGenerationTask{
		WorldID:    req.WorldId,
		TaskType:   models.TaskTypePosts,
		Status:     models.TaskStatusPending,
		Parameters: fmt.Sprintf(`{"count": %d, "world_prompt": "%s"}`, postsCount, world.Prompt),
	}

	err = s.worldRepo.CreateGenerationTask(ctx, postsTask)
	if err != nil {
		logger.Logger.Error("Failed to create posts generation task",
			zap.Error(err),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to create posts generation task")
	}

	// In a real implementation, these tasks would be processed by a background worker
	// For this implementation, we'll do a minimal simulation of the process
	go s.simulateContentGeneration(context.Background(), world.ID, usersCount, postsCount)

	return &worldpb.GenerateContentResponse{
		Success: true,
		Message: "Content generation started",
		TaskId:  usersTask.ID, // Return the first task's ID for reference
	}, nil
}

// GetGenerationStatus handles getting the generation status of a world
func (s *WorldService) GetGenerationStatus(ctx context.Context, req *worldpb.GetGenerationStatusRequest) (*worldpb.GetGenerationStatusResponse, error) {
	// Validate input
	if req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "world_id is required")
	}

	// Get world
	world, err := s.worldRepo.GetByID(ctx, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get world")
	}

	if world == nil {
		return nil, status.Errorf(codes.NotFound, "world not found")
	}

	// Get tasks
	tasks, err := s.worldRepo.GetGenerationTasks(ctx, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get generation tasks",
			zap.Error(err),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get generation tasks")
	}

	// Calculate progress
	var usersGenerated, postsGenerated int
	var message string

	for _, task := range tasks {
		if task.TaskType == models.TaskTypeUsers && task.Status == models.TaskStatusCompleted {
			// Parse JSON result to get count
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
				if count, ok := result["generated_count"].(float64); ok {
					usersGenerated = int(count)
				}
			}
		} else if task.TaskType == models.TaskTypePosts && task.Status == models.TaskStatusCompleted {
			// Parse JSON result to get count
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
				if count, ok := result["generated_count"].(float64); ok {
					postsGenerated = int(count)
				}
			}
		}

		// Get the most recent task message
		if task.Result != "" {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(task.Result), &result); err == nil {
				if msg, ok := result["message"].(string); ok && msg != "" {
					message = msg
				}
			}
		}
	}

	// Calculate progress percentage
	var progressPercentage float32
	if world.GenerationStatus == models.GenerationStatusCompleted {
		progressPercentage = 100.0
	} else if world.GenerationStatus == models.GenerationStatusFailed {
		progressPercentage = 0.0
	} else {
		totalTasks := len(tasks)
		completedTasksCount := 0
		inProgressTasksCount := 0

		for _, task := range tasks {
			if task.Status == models.TaskStatusCompleted {
				completedTasksCount++
			} else if task.Status == models.TaskStatusInProgress {
				inProgressTasksCount++
			}
		}

		if totalTasks > 0 {
			// Count completed tasks fully and in-progress tasks as half complete
			progressPercentage = float32(completedTasksCount) / float32(totalTasks) * 100.0
			// Add 50% credit for in-progress tasks
			progressPercentage += float32(inProgressTasksCount) / float32(totalTasks) * 50.0
		}
	}

	return &worldpb.GetGenerationStatusResponse{
		Status:             world.GenerationStatus,
		Message:            message,
		UsersGenerated:     int32(usersGenerated),
		PostsGenerated:     int32(postsGenerated),
		ProgressPercentage: progressPercentage,
	}, nil
}

// HealthCheck implements health check
func (s *WorldService) HealthCheck(ctx context.Context, req *worldpb.HealthCheckRequest) (*worldpb.HealthCheckResponse, error) {
	return &worldpb.HealthCheckResponse{
		Status: worldpb.HealthCheckResponse_SERVING,
	}, nil
}

// Helper methods

// createInitialGenerationTasks creates the initial task for world generation in Kafka and MongoDB
func (s *WorldService) createInitialGenerationTasks(ctx context.Context, worldID string) {
	// Get world details
	world, err := s.worldRepo.GetByID(ctx, worldID)
	if err != nil {
		logger.Logger.Error("Failed to get world details",
			zap.Error(err),
			zap.String("world_id", worldID))
		return
	}

	if world == nil {
		logger.Logger.Error("World not found when creating generation tasks",
			zap.String("world_id", worldID))
		return
	}

	// Create task ID
	taskID := uuid.New().String()

	// Параметры для задачи инициализации
	parameters := map[string]interface{}{
		"user_prompt": world.Prompt,
		"users_count": 10,  // Значение по умолчанию
		"posts_count": 50,  // Значение по умолчанию
	}

	// Преобразуем параметры в JSON
	paramsJSON, err := json.Marshal(parameters)
	if err != nil {
		logger.Logger.Error("Failed to marshal task parameters",
			zap.Error(err),
			zap.String("world_id", worldID))
		return
	}

	// Создаем задачу инициализации в БД
	task := &models.WorldGenerationTask{
		ID:         taskID,
		WorldID:    worldID,
		TaskType:   "init_world_creation",
		Status:     models.TaskStatusPending,
		Parameters: string(paramsJSON),
	}

	err = s.worldRepo.CreateGenerationTask(ctx, task)
	if err != nil {
		logger.Logger.Error("Failed to create initialization task",
			zap.Error(err),
			zap.String("world_id", worldID))
		return
	}

	logger.Logger.Info("Created initial generation task for world",
		zap.String("world_id", worldID),
		zap.String("task_id", taskID))

	// Отправляем сообщение в Kafka
	kafkaMessage := map[string]interface{}{
		"event_type": "task_created",
		"task_id":    taskID,
		"task_type":  "init_world_creation",
		"world_id":   worldID,
		"parameters": parameters,
	}

	err = s.kafkaProducer.SendJSON("generia-tasks", kafkaMessage)
	if err != nil {
		logger.Logger.Error("Failed to send task to Kafka",
			zap.Error(err),
			zap.String("world_id", worldID))
	} else {
		logger.Logger.Info("Successfully sent AI generation task to Kafka",
			zap.String("task_id", taskID),
			zap.String("world_id", worldID))
	}
}

// simulateContentGeneration simulates generating content for a world
// This is a placeholder for the actual AI content generation
func (s *WorldService) simulateContentGeneration(ctx context.Context, worldID string, usersCount, postsCount int) {
	logger.Logger.Info("Starting content generation simulation",
		zap.String("world_id", worldID),
		zap.Int("users_count", usersCount),
		zap.Int("posts_count", postsCount))

	// Get tasks
	tasks, err := s.worldRepo.GetGenerationTasks(ctx, worldID)
	if err != nil {
		logger.Logger.Error("Failed to get generation tasks",
			zap.Error(err),
			zap.String("world_id", worldID))
		return
	}

	var usersTask, postsTask *models.WorldGenerationTask
	for _, task := range tasks {
		if task.TaskType == models.TaskTypeUsers && task.Status == models.TaskStatusPending {
			usersTask = task
		} else if task.TaskType == models.TaskTypePosts && task.Status == models.TaskStatusPending {
			postsTask = task
		}
	}

	// Update users task to in-progress
	if usersTask != nil {
		err = s.worldRepo.UpdateGenerationTask(
			ctx,
			usersTask.ID,
			models.TaskStatusInProgress,
			`{"message": "Generating AI users..."}`)
		if err != nil {
			logger.Logger.Error("Failed to update users task status",
				zap.Error(err),
				zap.String("task_id", usersTask.ID))
		}

		// Simulate generating users (in a real implementation, this would create actual users in the database)
		time.Sleep(2 * time.Second)

		// For demo purposes, let's generate a few test users
		generatedUsers := make([]string, 0, usersCount)
		for i := 0; i < usersCount; i++ {
			userID := uuid.New().String()
			generatedUsers = append(generatedUsers, userID)
			// In a real implementation, you would create these users in the database
			logger.Logger.Info("Generated AI user",
				zap.String("user_id", userID),
				zap.String("world_id", worldID))
		}

		// Mark users task as completed
		resultData := map[string]interface{}{
			"message":         "Generated AI users successfully",
			"generated_count": len(generatedUsers),
			"user_ids":        generatedUsers,
		}
		resultJSON, _ := json.Marshal(resultData)

		err = s.worldRepo.UpdateGenerationTask(
			ctx,
			usersTask.ID,
			models.TaskStatusCompleted,
			string(resultJSON))
		if err != nil {
			logger.Logger.Error("Failed to update users task status",
				zap.Error(err),
				zap.String("task_id", usersTask.ID))
		}
	}

	// Update posts task to in-progress
	if postsTask != nil {
		err = s.worldRepo.UpdateGenerationTask(
			ctx,
			postsTask.ID,
			models.TaskStatusInProgress,
			`{"message": "Generating AI posts..."}`)
		if err != nil {
			logger.Logger.Error("Failed to update posts task status",
				zap.Error(err),
				zap.String("task_id", postsTask.ID))
		}

		// Simulate generating posts
		time.Sleep(3 * time.Second)

		// For demo purposes, let's just log some test post generation
		generatedPosts := make([]string, 0, postsCount)
		for i := 0; i < postsCount; i++ {
			postID := uuid.New().String()
			generatedPosts = append(generatedPosts, postID)
			// In a real implementation, you would create these posts in the database
			logger.Logger.Info("Generated AI post",
				zap.String("post_id", postID),
				zap.String("world_id", worldID))
		}

		// Mark posts task as completed
		resultData := map[string]interface{}{
			"message":         "Generated AI posts successfully",
			"generated_count": len(generatedPosts),
			"post_ids":        generatedPosts,
		}
		resultJSON, _ := json.Marshal(resultData)

		err = s.worldRepo.UpdateGenerationTask(
			ctx,
			postsTask.ID,
			models.TaskStatusCompleted,
			string(resultJSON))
		if err != nil {
			logger.Logger.Error("Failed to update posts task status",
				zap.Error(err),
				zap.String("task_id", postsTask.ID))
		}
	}

	// Update world generation status to completed
	err = s.worldRepo.UpdateGenerationStatus(ctx, worldID, models.GenerationStatusCompleted)
	if err != nil {
		logger.Logger.Error("Failed to update world generation status",
			zap.Error(err),
			zap.String("world_id", worldID))
	}

	logger.Logger.Info("Completed content generation simulation", zap.String("world_id", worldID))
}
