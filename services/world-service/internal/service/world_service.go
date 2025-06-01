package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/pkg/temporal"
	"github.com/sdshorin/generia/services/world-service/internal/models"
	"github.com/sdshorin/generia/services/world-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
	mediapb "github.com/sdshorin/generia/api/grpc/media"
	postpb "github.com/sdshorin/generia/api/grpc/post"
	worldpb "github.com/sdshorin/generia/api/grpc/world"
)

// StageInfo represents information about generation stage status
type StageInfo struct {
	Name   string `bson:"name"`
	Status string `bson:"status"`
}

// WorldGenerationStatus represents the world generation progress information
type WorldGenerationStatus struct {
	ID                  string      `bson:"_id"`
	Status              string      `bson:"status"`
	CurrentStage        string      `bson:"current_stage"`
	Stages              []StageInfo `bson:"stages"`
	TasksTotal          int         `bson:"tasks_total"`
	TasksCompleted      int         `bson:"tasks_completed"`
	TasksFailed         int         `bson:"tasks_failed"`
	TaskPredicted       int         `bson:"task_predicted"`
	UsersCreated        int         `bson:"users_created"`
	PostsCreated        int         `bson:"posts_created"`
	UsersPredicted      int         `bson:"users_predicted"`
	PostsPredicted      int         `bson:"posts_predicted"`
	ApiCallLimitsLLM    int         `bson:"api_call_limits_LLM"`
	ApiCallLimitsImages int         `bson:"api_call_limits_images"`
	ApiCallsMadeLLM     int         `bson:"api_calls_made_LLM"`
	ApiCallsMadeImages  int         `bson:"api_calls_made_images"`
	LlmCostTotal        float64     `bson:"llm_cost_total"`
	ImageCostTotal      float64     `bson:"image_cost_total"`
	Parameters          bson.M      `bson:"parameters"`
	CreatedAt           time.Time   `bson:"created_at"`
	UpdatedAt           time.Time   `bson:"updated_at"`
}

// WorldService implements the world gRPC service
type WorldService struct {
	worldpb.UnimplementedWorldServiceServer
	worldRepo      repository.WorldRepository
	authClient     authpb.AuthServiceClient
	postClient     postpb.PostServiceClient
	mediaClient    mediapb.MediaServiceClient
	temporalClient *temporal.Client
}

// NewWorldService creates a new WorldService
func NewWorldService(
	worldRepo repository.WorldRepository,
	authClient authpb.AuthServiceClient,
	postClient postpb.PostServiceClient,
	mediaClient mediapb.MediaServiceClient,
	temporalHostPort string,
) worldpb.WorldServiceServer {
	temporalClient, err := temporal.NewClient(temporalHostPort)
	if err != nil {
		logger.Logger.Fatal("Failed to create Temporal client", zap.Error(err))
	}

	return &WorldService{
		worldRepo:      worldRepo,
		authClient:     authClient,
		postClient:     postClient,
		mediaClient:    mediaClient,
		temporalClient: temporalClient,
	}
}

// CreateWorld handles creating a new world
func (s *WorldService) CreateWorld(ctx context.Context, req *worldpb.CreateWorldRequest) (*worldpb.WorldResponse, error) {
	// Validate input
	if req.UserId == "" || req.Name == "" || req.Prompt == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id, name, and prompt are required")
	}

	// Set default values if not provided
	charactersCount := req.CharactersCount
	if charactersCount <= 0 {
		charactersCount = 25
	}
	postsCount := req.PostsCount
	if postsCount <= 0 {
		postsCount = 150
	}

	// Validate ranges
	if charactersCount < 1 || charactersCount > 40 {
		return nil, status.Errorf(codes.InvalidArgument, "characters_count must be between 1 and 40")
	}
	if postsCount < 1 || postsCount > 250 {
		return nil, status.Errorf(codes.InvalidArgument, "posts_count must be between 1 and 250")
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
	usersCount, postsCountStats, err := s.worldRepo.GetWorldStats(ctx, world.ID)
	if err != nil {
		logger.Logger.Error("Failed to get world stats", zap.Error(err), zap.String("world_id", world.ID))
		// Not a critical error, continue with zeros
		usersCount = 0
		postsCountStats = 0
	}

	// Create and enqueue content generation tasks with specified parameters
	s.createInitialGenerationTasks(ctx, world.ID, int(charactersCount), int(postsCount))

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
		PostsCount:       int32(postsCountStats),
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

	// Get image URLs if UUIDs exist
	var imageUrl, iconUrl string

	// Get background image URL
	if world.ImageUUID.Valid && world.ImageUUID.String != "" {
		// Get media URL for world background image
		mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   world.ImageUUID.String,
			Variant:   "original", // Use original variant for world background image
			ExpiresIn: 3600,       // 1 hour
		})
		if err == nil && mediaResp != nil {
			imageUrl = mediaResp.Url
			logger.Logger.Debug("Got image URL for world",
				zap.String("world_id", world.ID),
				zap.String("image_url", imageUrl))
		} else {
			logger.Logger.Warn("Failed to get image URL for world",
				zap.String("world_id", world.ID),
				zap.Error(err))
		}
	}

	// Get icon image URL
	if world.IconUUID.Valid && world.IconUUID.String != "" {
		// Get media URL for world icon image
		mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   world.IconUUID.String,
			Variant:   "original", // Use original variant for world icon image
			ExpiresIn: 3600,       // 1 hour
		})
		if err == nil && mediaResp != nil {
			iconUrl = mediaResp.Url
			logger.Logger.Debug("Got icon URL for world",
				zap.String("world_id", world.ID),
				zap.String("icon_url", iconUrl))
		} else {
			logger.Logger.Warn("Failed to get icon URL for world",
				zap.String("world_id", world.ID),
				zap.Error(err))
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
		ImageUrl:         imageUrl,
		IconUrl:          iconUrl,
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

		// Get image URL if image UUID exists
		var imageUrl string
		if world.ImageUUID.Valid && world.ImageUUID.String != "" {
			// Get media URL for world image
			mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
				MediaId:   world.ImageUUID.String,
				Variant:   "original", // Use original variant for world background image
				ExpiresIn: 3600,       // 1 hour
			})
			if err == nil && mediaResp != nil {
				imageUrl = mediaResp.Url
				logger.Logger.Debug("Got image URL for world in list",
					zap.String("world_id", world.ID),
					zap.String("image_url", imageUrl))
			} else {
				logger.Logger.Warn("Failed to get image URL for world in list",
					zap.String("world_id", world.ID),
					zap.Error(err))
			}
		}

		// Get icon URL if icon UUID exists
		var iconUrl string
		if world.IconUUID.Valid && world.IconUUID.String != "" {
			// Get media URL for world icon
			mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
				MediaId:   world.IconUUID.String,
				Variant:   "original", // Use original variant for world icon image
				ExpiresIn: 3600,       // 1 hour
			})
			if err == nil && mediaResp != nil {
				iconUrl = mediaResp.Url
				logger.Logger.Debug("Got icon URL for world in list",
					zap.String("world_id", world.ID),
					zap.String("icon_url", iconUrl))
			} else {
				logger.Logger.Warn("Failed to get icon URL for world in list",
					zap.String("world_id", world.ID),
					zap.Error(err))
			}
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
			ImageUrl:         imageUrl,
			IconUrl:          iconUrl,
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

// UpdateWorldImage handles updating the world's background image and icon
func (s *WorldService) UpdateWorldImage(ctx context.Context, req *worldpb.UpdateWorldImageRequest) (*worldpb.UpdateWorldImageResponse, error) {
	// Validate input
	if req.WorldId == "" || req.ImageUuid == "" || req.IconUuid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "world_id, image_uuid, and icon_uuid are required")
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
	logger.Logger.Info("Updating world with images",
		zap.String("world_id", req.WorldId),
		zap.String("image_uuid", req.ImageUuid),
		zap.String("icon_uuid", req.IconUuid))

	// Update world with image and icon UUIDs
	world.ImageUUID = sql.NullString{String: req.ImageUuid, Valid: req.ImageUuid != ""}
	world.IconUUID = sql.NullString{String: req.IconUuid, Valid: req.IconUuid != ""}
	world.UpdatedAt = time.Now()

	// Save updated world
	err = s.worldRepo.Update(ctx, world)
	if err != nil {
		logger.Logger.Error("Failed to update world with images",
			zap.Error(err),
			zap.String("world_id", req.WorldId),
			zap.String("image_uuid", req.ImageUuid),
			zap.String("icon_uuid", req.IconUuid))
		return nil, status.Errorf(codes.Internal, "failed to update world with images")
	}

	logger.Logger.Info("Updated world with images",
		zap.String("world_id", req.WorldId),
		zap.String("image_uuid", req.ImageUuid),
		zap.String("icon_uuid", req.IconUuid))

	return &worldpb.UpdateWorldImageResponse{
		Success: true,
		Message: "Successfully updated world images",
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

	// Get generation status from MongoDB
	generationStatus, err := s.getGenerationStatusFromMongo(ctx, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get generation status from MongoDB", zap.Error(err), zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get generation status")
	}

	if generationStatus == nil {
		// No generation status found, return default response
		return &worldpb.GetGenerationStatusResponse{
			Status:         "not_started",
			CurrentStage:   "",
			Stages:         []*worldpb.StageInfo{},
			TasksTotal:     0,
			TasksCompleted: 0,
			TasksFailed:    0,
			TaskPredicted:  0,
			UsersCreated:   0,
			PostsCreated:   0,
			UsersPredicted: 0,
			PostsPredicted: 0,
		}, nil
	}

	// Convert stages to protobuf format
	stages := make([]*worldpb.StageInfo, len(generationStatus.Stages))
	for i, stage := range generationStatus.Stages {
		stages[i] = &worldpb.StageInfo{
			Name:   stage.Name,
			Status: stage.Status,
		}
	}

	return &worldpb.GetGenerationStatusResponse{
		Status:              generationStatus.Status,
		CurrentStage:        generationStatus.CurrentStage,
		Stages:              stages,
		TasksTotal:          int32(generationStatus.TasksTotal),
		TasksCompleted:      int32(generationStatus.TasksCompleted),
		TasksFailed:         int32(generationStatus.TasksFailed),
		TaskPredicted:       int32(generationStatus.TaskPredicted),
		UsersCreated:        int32(generationStatus.UsersCreated),
		PostsCreated:        int32(generationStatus.PostsCreated),
		UsersPredicted:      int32(generationStatus.UsersPredicted),
		PostsPredicted:      int32(generationStatus.PostsPredicted),
		ApiCallLimitsLlm:    int32(generationStatus.ApiCallLimitsLLM),
		ApiCallLimitsImages: int32(generationStatus.ApiCallLimitsImages),
		ApiCallsMadeLlm:     int32(generationStatus.ApiCallsMadeLLM),
		ApiCallsMadeImages:  int32(generationStatus.ApiCallsMadeImages),
		LlmCostTotal:        generationStatus.LlmCostTotal,
		ImageCostTotal:      generationStatus.ImageCostTotal,
		CreatedAt:           generationStatus.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           generationStatus.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// HealthCheck implements health check
func (s *WorldService) HealthCheck(ctx context.Context, req *worldpb.HealthCheckRequest) (*worldpb.HealthCheckResponse, error) {
	return &worldpb.HealthCheckResponse{
		Status: worldpb.HealthCheckResponse_SERVING,
	}, nil
}

// Helper methods

// createInitialGenerationTasks starts the world creation workflow using Temporal
func (s *WorldService) createInitialGenerationTasks(ctx context.Context, worldID string, charactersCount, postsCount int) {
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

	// Create input for Temporal workflow
	input := temporal.InitWorldCreationInput{
		WorldID:         worldID,
		WorldName:       world.Name,
		WorldPrompt:     world.Prompt,
		CharactersCount: charactersCount,
		PostsCount:      postsCount,
	}

	// Execute Temporal workflow
	workflowRun, err := s.temporalClient.ExecuteInitWorldCreationWorkflow(ctx, input)
	if err != nil {
		logger.Logger.Error("Failed to start Temporal workflow",
			zap.Error(err),
			zap.String("world_id", worldID))
		return
	}

	logger.Logger.Info("Successfully started world creation workflow",
		zap.String("workflow_id", workflowRun.GetID()),
		zap.String("run_id", workflowRun.GetRunID()),
		zap.String("world_id", worldID),
		zap.String("prompt", world.Prompt),
		zap.Int("users_count", charactersCount),
		zap.Int("posts_count", postsCount))
}

// getGenerationStatusFromMongo retrieves the world generation status from MongoDB
func (s *WorldService) getGenerationStatusFromMongo(ctx context.Context, worldID string) (*WorldGenerationStatus, error) {
	// Get MongoDB configuration from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongodb:27017" // Default value if not set
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	// Select database and collection
	db := client.Database("generia_ai_worker")
	collection := db.Collection("world_generation_status")

	// Find generation status by world ID
	var generationStatus WorldGenerationStatus
	err = collection.FindOne(ctx, bson.M{"_id": worldID}).Decode(&generationStatus)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No generation status found
		}
		return nil, fmt.Errorf("failed to find generation status: %w", err)
	}

	return &generationStatus, nil
}
