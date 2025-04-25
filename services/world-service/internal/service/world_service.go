package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sdshorin/generia/pkg/kafka"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/world-service/internal/models"
	"github.com/sdshorin/generia/services/world-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	worldRepo     repository.WorldRepository
	authClient    authpb.AuthServiceClient
	postClient    postpb.PostServiceClient
	kafkaProducer *kafka.Producer
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
		Name:        req.Name,
		Description: req.Description,
		Prompt:      req.Prompt,
		CreatorID:   req.UserId,
		Status:      models.WorldStatusActive,
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
		GenerationStatus: "", // world.GenerationStatus,
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
		GenerationStatus: "", // world.GenerationStatus,
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

// // GetGenerationStatus handles getting the generation status of a world
// func (s *WorldService) GetGenerationStatus(ctx context.Context, req *worldpb.GetGenerationStatusRequest) (*worldpb.GetGenerationStatusResponse, error) {
// 	// Validate input
// 	if req.WorldId == "" {
// 		return nil, status.Errorf(codes.InvalidArgument, "world_id is required")
// 	}

// 	// Get world
// 	world, err := s.worldRepo.GetByID(ctx, req.WorldId)
// 	if err != nil {
// 		logger.Logger.Error("Failed to get world", zap.Error(err), zap.String("world_id", req.WorldId))
// 		return nil, status.Errorf(codes.Internal, "failed to get world")
// 	}

// 	if world == nil {
// 		return nil, status.Errorf(codes.NotFound, "world not found")
// 	}

// 	// todo: load status from mongo - services/ai-worker/src/utils/progress.py

// 	return &worldpb.GetGenerationStatusResponse{
// 		Status:             world.GenerationStatus,
// 		Message:            message,
// 		UsersGenerated:     int32(usersGenerated),
// 		PostsGenerated:     int32(postsGenerated),
// 		ProgressPercentage: progressPercentage,
// 	}, nil
// }

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

	// Parameters for initialization task
	parameters := map[string]interface{}{
		"user_prompt": world.Prompt,
		"users_count": 20,  // Default value
		"posts_count": 100, // Default value
		"created_at":  time.Now().Format(time.RFC3339),
	}

	// Save task to MongoDB
	err = s.createMongoDBTask(ctx, taskID, worldID, parameters)
	if err != nil {
		logger.Logger.Error("Failed to create task in MongoDB",
			zap.Error(err),
			zap.String("world_id", worldID),
			zap.String("task_id", taskID))
		return
	}

	// Send message to Kafka
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
			zap.String("world_id", worldID),
			zap.String("task_id", taskID))
	} else {
		logger.Logger.Info("Successfully sent AI generation task to Kafka",
			zap.String("task_id", taskID),
			zap.String("world_id", worldID),
			zap.String("prompt", world.Prompt),
			zap.Int("users_count", parameters["users_count"].(int)),
			zap.Int("posts_count", parameters["posts_count"].(int)))
	}
}

// createMongoDBTask creates a task record in MongoDB
func (s *WorldService) createMongoDBTask(ctx context.Context, taskID, worldID string, parameters map[string]interface{}) error {
	// Get MongoDB configuration from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27017" // Default value if not set
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Select database
	db := client.Database("generia_ai_worker")

	// Create task document
	now := time.Now()
	task := bson.M{
		"_id":           taskID,
		"world_id":      worldID,
		"type":          "init_world_creation",
		"status":        "pending",
		"parameters":    parameters,
		"created_at":    now,
		"updated_at":    now,
		"attempt_count": 0,
		"worker_id":     nil,
		"result":        nil,
		"error":         nil,
	}

	// Insert document into tasks collection
	_, err = db.Collection("tasks").InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to insert task into MongoDB: %w", err)
	}

	logger.Logger.Info("Task created in MongoDB",
		zap.String("task_id", taskID),
		zap.String("world_id", worldID))

	return nil
}
