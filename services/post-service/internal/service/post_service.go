package service

import (
	"context"
	"time"

	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/post-service/internal/models"
	"github.com/sdshorin/generia/services/post-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
	characterpb "github.com/sdshorin/generia/api/grpc/character"
	interactionpb "github.com/sdshorin/generia/api/grpc/interaction"
	mediapb "github.com/sdshorin/generia/api/grpc/media"
	postpb "github.com/sdshorin/generia/api/grpc/post"
)

// PostService implements the post gRPC service
type PostService struct {
	postpb.UnimplementedPostServiceServer
	postRepo          repository.PostRepository
	authClient        authpb.AuthServiceClient
	mediaClient       mediapb.MediaServiceClient
	interactionClient interactionpb.InteractionServiceClient
	characterClient   characterpb.CharacterServiceClient
}

// NewPostService creates a new PostService
func NewPostService(
	postRepo repository.PostRepository,
	authClient authpb.AuthServiceClient,
	mediaClient mediapb.MediaServiceClient,
	interactionClient interactionpb.InteractionServiceClient,
	characterClient characterpb.CharacterServiceClient,
) postpb.PostServiceServer {
	return &PostService{
		postRepo:          postRepo,
		authClient:        authClient,
		mediaClient:       mediaClient,
		interactionClient: interactionClient,
		characterClient:   characterClient,
	}
}

// CreatePost handles post creation
func (s *PostService) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*postpb.CreatePostResponse, error) {
	// Validate input
	if req.UserId == "" || req.MediaId == "" || req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id, media_id, and world_id are required")
	}

	var characterID string

	// If character ID is provided, validate that it belongs to the user
	if req.CharacterId != "" {
		// Get character
		characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
			CharacterId: req.CharacterId,
		})
		if err != nil {
			logger.Logger.Error("Failed to get character", zap.Error(err), zap.String("character_id", req.CharacterId))
			return nil, status.Errorf(codes.Internal, "failed to validate character")
		}

		// Ensure character belongs to the user
		if characterResp.RealUserId == nil || *characterResp.RealUserId != req.UserId {
			return nil, status.Errorf(codes.PermissionDenied, "character does not belong to the user")
		}

		characterID = req.CharacterId
	} else {
		// Get or create a character for the user in this world
		characters, err := s.characterClient.GetUserCharactersInWorld(ctx, &characterpb.GetUserCharactersInWorldRequest{
			UserId:  req.UserId,
			WorldId: req.WorldId,
		})
		if err != nil || len(characters.Characters) == 0 {
			// Need to create a character for this user in this world
			// First get user info to use as display name
			userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
				UserId: req.UserId,
			})
			if err != nil {
				logger.Logger.Error("Failed to get user info", zap.Error(err), zap.String("user_id", req.UserId))
				return nil, status.Errorf(codes.Internal, "failed to validate user")
			}

			// Create a character
			realUserID := req.UserId
			newCharacter, err := s.characterClient.CreateCharacter(ctx, &characterpb.CreateCharacterRequest{
				WorldId:     req.WorldId,
				RealUserId:  &realUserID,
				DisplayName: userResp.Username,
			})
			if err != nil {
				logger.Logger.Error("Failed to create character", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to create character")
			}
			characterID = newCharacter.Id
		} else {
			// Use the first character
			characterID = characters.Characters[0].Id
		}
	}

	// Validate media
	mediaResp, err := s.mediaClient.GetMedia(ctx, &mediapb.GetMediaRequest{
		MediaId: req.MediaId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get media info", zap.Error(err), zap.String("media_id", req.MediaId))
		return nil, status.Errorf(codes.Internal, "failed to validate media")
	}

	// Check if media belongs to the character
	if mediaResp.CharacterId != characterID {
		return nil, status.Errorf(codes.PermissionDenied, "media does not belong to the character")
	}

	// Create post
	post := &models.Post{
		CharacterID: characterID,
		IsAI:        req.IsAi,
		WorldID:     req.WorldId,
		Caption:     req.Caption,
		MediaID:     req.MediaId,
	}

	err = s.postRepo.Create(ctx, post)
	if err != nil {
		logger.Logger.Error("Failed to create post", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create post")
	}

	return &postpb.CreatePostResponse{
		PostId:    post.ID,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetPost handles post retrieval by ID
func (s *PostService) GetPost(ctx context.Context, req *postpb.GetPostRequest) (*postpb.Post, error) {
	// Validate input
	if req.PostId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id is required")
	}

	// Get post
	post, err := s.postRepo.GetByID(ctx, req.PostId)
	if err != nil {
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("post_id", req.PostId))
		return nil, status.Errorf(codes.Internal, "failed to get post")
	}

	if post == nil {
		return nil, status.Errorf(codes.NotFound, "post not found")
	}

	// Get character info
	characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
		CharacterId: post.CharacterID,
	})
	if err != nil {
		logger.Logger.Error("Failed to get character info", zap.Error(err), zap.String("character_id", post.CharacterID))
		return nil, status.Errorf(codes.Internal, "failed to get character info")
	}

	// Get media URL
	mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
		MediaId:   post.MediaID,
		Variant:   "medium", // Default to medium size
		ExpiresIn: 3600,     // 1 hour
	})
	if err != nil {
		logger.Logger.Error("Failed to get media URL", zap.Error(err), zap.String("media_id", post.MediaID))
		return nil, status.Errorf(codes.Internal, "failed to get media URL")
	}

	// Get interaction stats
	statsResp, err := s.interactionClient.GetPostStats(ctx, &interactionpb.GetPostStatsRequest{
		PostId: post.ID,
	})
	if err != nil {
		logger.Logger.Error("Failed to get post stats", zap.Error(err), zap.String("post_id", post.ID))
		// Continue even if stats can't be retrieved
		statsResp = &interactionpb.PostStatsResponse{
			PostId:        post.ID,
			LikesCount:    0,
			CommentsCount: 0,
		}
	}

	return &postpb.Post{
		PostId:        post.ID,
		CharacterId:   post.CharacterID,
		DisplayName:   characterResp.DisplayName,
		Caption:       post.Caption,
		MediaUrl:      mediaResp.Url,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		LikesCount:    statsResp.LikesCount,
		CommentsCount: statsResp.CommentsCount,
		WorldId:       post.WorldID,
		IsAi:          post.IsAI,
	}, nil
}

// GetUserPosts handles retrieval of posts by user ID
func (s *PostService) GetUserPosts(ctx context.Context, req *postpb.GetUserPostsRequest) (*postpb.PostList, error) {
	// Validate input
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Get posts
	posts, total, err := s.postRepo.GetByUserID(ctx, req.UserId, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get user posts",
			zap.Error(err),
			zap.String("user_id", req.UserId),
			zap.Int("limit", limit),
			zap.Int("offset", offset))
		return nil, status.Errorf(codes.Internal, "failed to get user posts")
	}

	if len(posts) == 0 {
		return &postpb.PostList{
			Posts: []*postpb.Post{},
			Total: 0,
		}, nil
	}

	// Get character IDs
	characterIDs := make(map[string]struct{})
	for _, post := range posts {
		characterIDs[post.CharacterID] = struct{}{}
	}

	// Get character info for all characters
	characterInfoMap := make(map[string]*characterpb.Character)
	for characterID := range characterIDs {
		characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
			CharacterId: characterID,
		})
		if err != nil {
			logger.Logger.Error("Failed to get character info", zap.Error(err), zap.String("character_id", characterID))
			continue
		}
		characterInfoMap[characterID] = characterResp
	}

	// Prepare post IDs for batch operations
	postIDs := make([]string, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Get stats for all posts in batch
	var statsResp *interactionpb.PostsStatsResponse
	if len(postIDs) > 0 {
		statsResp, err = s.interactionClient.GetPostsStats(ctx, &interactionpb.GetPostsStatsRequest{
			PostIds: postIDs,
		})
		if err != nil {
			logger.Logger.Error("Failed to get posts stats", zap.Error(err))
			// Continue even if stats can't be retrieved
			statsResp = &interactionpb.PostsStatsResponse{
				Stats: make(map[string]*interactionpb.PostStatsResponse),
			}
		}
	} else {
		statsResp = &interactionpb.PostsStatsResponse{
			Stats: make(map[string]*interactionpb.PostStatsResponse),
		}
	}

	// Build response
	result := make([]*postpb.Post, len(posts))
	for i, post := range posts {
		// Get character info
		var displayName string
		if characterInfo, ok := characterInfoMap[post.CharacterID]; ok {
			displayName = characterInfo.DisplayName
		}

		// Get media URL
		var mediaURL string
		mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   post.MediaID,
			Variant:   "medium", // Default to medium size
			ExpiresIn: 3600,     // 1 hour
		})
		if err != nil {
			logger.Logger.Error("Failed to get media URL", zap.Error(err), zap.String("media_id", post.MediaID))
			mediaURL = "" // Default empty URL
		} else {
			mediaURL = mediaResp.Url
		}

		// Get stats
		var likesCount, commentsCount int32
		if stats, ok := statsResp.Stats[post.ID]; ok {
			likesCount = stats.LikesCount
			commentsCount = stats.CommentsCount
		}

		result[i] = &postpb.Post{
			PostId:        post.ID,
			CharacterId:   post.CharacterID,
			DisplayName:   displayName,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			WorldId:       post.WorldID,
			IsAi:          post.IsAI,
		}
	}

	return &postpb.PostList{
		Posts: result,
		Total: int32(total),
	}, nil
}

// GetCharacterPosts handles retrieval of posts by character ID
func (s *PostService) GetCharacterPosts(ctx context.Context, req *postpb.GetCharacterPostsRequest) (*postpb.PostList, error) {
	// Validate input
	if req.CharacterId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "character_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Get posts
	posts, total, err := s.postRepo.GetByCharacterID(ctx, req.CharacterId, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get character posts",
			zap.Error(err),
			zap.String("character_id", req.CharacterId),
			zap.Int("limit", limit),
			zap.Int("offset", offset))
		return nil, status.Errorf(codes.Internal, "failed to get character posts")
	}

	if len(posts) == 0 {
		return &postpb.PostList{
			Posts: []*postpb.Post{},
			Total: 0,
		}, nil
	}

	// Get character info
	characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
		CharacterId: req.CharacterId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get character info", zap.Error(err), zap.String("character_id", req.CharacterId))
		return nil, status.Errorf(codes.Internal, "failed to get character info")
	}

	// Prepare post IDs for batch operations
	postIDs := make([]string, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Get stats for all posts in batch
	var statsResp *interactionpb.PostsStatsResponse
	if len(postIDs) > 0 {
		statsResp, err = s.interactionClient.GetPostsStats(ctx, &interactionpb.GetPostsStatsRequest{
			PostIds: postIDs,
		})
		if err != nil {
			logger.Logger.Error("Failed to get posts stats", zap.Error(err))
			// Continue even if stats can't be retrieved
			statsResp = &interactionpb.PostsStatsResponse{
				Stats: make(map[string]*interactionpb.PostStatsResponse),
			}
		}
	} else {
		statsResp = &interactionpb.PostsStatsResponse{
			Stats: make(map[string]*interactionpb.PostStatsResponse),
		}
	}

	// Build response
	result := make([]*postpb.Post, len(posts))
	for i, post := range posts {
		// Get media URL
		var mediaURL string
		mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   post.MediaID,
			Variant:   "medium", // Default to medium size
			ExpiresIn: 3600,     // 1 hour
		})
		if err != nil {
			logger.Logger.Error("Failed to get media URL", zap.Error(err), zap.String("media_id", post.MediaID))
			mediaURL = "" // Default empty URL
		} else {
			mediaURL = mediaResp.Url
		}

		// Get stats
		var likesCount, commentsCount int32
		if stats, ok := statsResp.Stats[post.ID]; ok {
			likesCount = stats.LikesCount
			commentsCount = stats.CommentsCount
		}

		result[i] = &postpb.Post{
			PostId:        post.ID,
			CharacterId:   post.CharacterID,
			DisplayName:   characterResp.DisplayName,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			WorldId:       post.WorldID,
			IsAi:          post.IsAI,
		}
	}

	return &postpb.PostList{
		Posts: result,
		Total: int32(total),
	}, nil
}

// GetPostsByIds handles retrieval of posts by IDs
func (s *PostService) GetPostsByIds(ctx context.Context, req *postpb.GetPostsByIdsRequest) (*postpb.PostList, error) {
	// Validate input
	if len(req.PostIds) == 0 {
		return &postpb.PostList{
			Posts: []*postpb.Post{},
			Total: 0,
		}, nil
	}

	// Get posts
	posts, err := s.postRepo.GetByIDs(ctx, req.PostIds)
	if err != nil {
		logger.Logger.Error("Failed to get posts by IDs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get posts by IDs")
	}

	if len(posts) == 0 {
		return &postpb.PostList{
			Posts: []*postpb.Post{},
			Total: 0,
		}, nil
	}

	// Extract unique character IDs
	characterIDs := make(map[string]struct{})
	for _, post := range posts {
		characterIDs[post.CharacterID] = struct{}{}
	}

	// Get character info for all characters
	characterInfoMap := make(map[string]*characterpb.Character)
	for characterID := range characterIDs {
		characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
			CharacterId: characterID,
		})
		if err != nil {
			logger.Logger.Error("Failed to get character info", zap.Error(err), zap.String("character_id", characterID))
			continue
		}
		characterInfoMap[characterID] = characterResp
	}

	// Prepare post IDs for batch operations
	postIDs := make([]string, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Get stats for all posts in batch
	statsResp, err := s.interactionClient.GetPostsStats(ctx, &interactionpb.GetPostsStatsRequest{
		PostIds: postIDs,
	})
	if err != nil {
		logger.Logger.Error("Failed to get posts stats", zap.Error(err))
		// Continue even if stats can't be retrieved
		statsResp = &interactionpb.PostsStatsResponse{
			Stats: make(map[string]*interactionpb.PostStatsResponse),
		}
	}

	// Build response
	result := make([]*postpb.Post, len(posts))
	for i, post := range posts {
		// Get character info
		var displayName string
		if characterInfo, ok := characterInfoMap[post.CharacterID]; ok {
			displayName = characterInfo.DisplayName
		}

		// Get media URL
		var mediaURL string
		mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   post.MediaID,
			Variant:   "medium", // Default to medium size
			ExpiresIn: 3600,     // 1 hour
		})
		if err != nil {
			logger.Logger.Error("Failed to get media URL", zap.Error(err), zap.String("media_id", post.MediaID))
			mediaURL = "" // Default empty URL
		} else {
			mediaURL = mediaResp.Url
		}

		// Get stats
		var likesCount, commentsCount int32
		if stats, ok := statsResp.Stats[post.ID]; ok {
			likesCount = stats.LikesCount
			commentsCount = stats.CommentsCount
		}

		result[i] = &postpb.Post{
			PostId:        post.ID,
			CharacterId:   post.CharacterID,
			DisplayName:   displayName,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			WorldId:       post.WorldID,
			IsAi:          post.IsAI,
		}
	}

	return &postpb.PostList{
		Posts: result,
		Total: int32(len(result)),
	}, nil
}

// GetGlobalFeed handles retrieval of global feed
func (s *PostService) GetGlobalFeed(ctx context.Context, req *postpb.GetGlobalFeedRequest) (*postpb.PostList, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Validate world_id
	if req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "world_id is required")
	}

	// Get posts
	posts, nextCursor, err := s.postRepo.GetGlobalFeed(ctx, limit, req.Cursor, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get global feed",
			zap.Error(err),
			zap.Int("limit", limit),
			zap.String("cursor", req.Cursor),
			zap.String("world_id", req.WorldId))
		return nil, status.Errorf(codes.Internal, "failed to get global feed")
	}

	if len(posts) == 0 {
		return &postpb.PostList{
			Posts:      []*postpb.Post{},
			Total:      0,
			NextCursor: "",
		}, nil
	}

	// Extract unique character IDs
	characterIDs := make(map[string]struct{})
	for _, post := range posts {
		characterIDs[post.CharacterID] = struct{}{}
	}

	// Get character info for all characters
	characterInfoMap := make(map[string]*characterpb.Character)
	for characterID := range characterIDs {
		characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
			CharacterId: characterID,
		})
		if err != nil {
			logger.Logger.Error("Failed to get character info", zap.Error(err), zap.String("character_id", characterID))
			continue
		}
		characterInfoMap[characterID] = characterResp
	}

	// Prepare post IDs for batch operations
	postIDs := make([]string, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Get stats for all posts in batch
	statsResp, err := s.interactionClient.GetPostsStats(ctx, &interactionpb.GetPostsStatsRequest{
		PostIds: postIDs,
	})
	if err != nil {
		logger.Logger.Error("Failed to get posts stats", zap.Error(err))
		// Continue even if stats can't be retrieved
		statsResp = &interactionpb.PostsStatsResponse{
			Stats: make(map[string]*interactionpb.PostStatsResponse),
		}
	}

	// Build response
	result := make([]*postpb.Post, len(posts))
	for i, post := range posts {
		// Get character info
		var displayName string
		if characterInfo, ok := characterInfoMap[post.CharacterID]; ok {
			displayName = characterInfo.DisplayName
		}

		// Get media URL
		var mediaURL string
		mediaResp, err := s.mediaClient.GetMediaURL(ctx, &mediapb.GetMediaURLRequest{
			MediaId:   post.MediaID,
			Variant:   "medium", // Default to medium size
			ExpiresIn: 3600,     // 1 hour
		})
		if err != nil {
			logger.Logger.Error("Failed to get media URL", zap.Error(err), zap.String("media_id", post.MediaID))
			mediaURL = "" // Default empty URL
		} else {
			mediaURL = mediaResp.Url
		}

		// Get stats
		var likesCount, commentsCount int32
		if stats, ok := statsResp.Stats[post.ID]; ok {
			likesCount = stats.LikesCount
			commentsCount = stats.CommentsCount
		}

		result[i] = &postpb.Post{
			PostId:        post.ID,
			CharacterId:   post.CharacterID,
			DisplayName:   displayName,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
			WorldId:       post.WorldID,
			IsAi:          post.IsAI,
		}
	}

	return &postpb.PostList{
		Posts:      result,
		Total:      int32(len(result)),
		NextCursor: nextCursor,
	}, nil
}

// HealthCheck implements health check
func (s *PostService) HealthCheck(ctx context.Context, req *postpb.HealthCheckRequest) (*postpb.HealthCheckResponse, error) {
	return &postpb.HealthCheckResponse{
		Status: postpb.HealthCheckResponse_SERVING,
	}, nil
}

// CreateAIPost handles AI post creation (internal method, not requiring user auth)
func (s *PostService) CreateAIPost(ctx context.Context, req *postpb.CreateAIPostRequest) (*postpb.CreatePostResponse, error) {
	// Validate input
	if req.CharacterId == "" || req.MediaId == "" || req.WorldId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "character_id, media_id, and world_id are required")
	}

	// Verify character exists
	characterResp, err := s.characterClient.GetCharacter(ctx, &characterpb.GetCharacterRequest{
		CharacterId: req.CharacterId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get character", zap.Error(err), zap.String("character_id", req.CharacterId))
		return nil, status.Errorf(codes.NotFound, "Character not found")
	}

	// Verify character is an AI character
	if !characterResp.IsAi {
		return nil, status.Errorf(codes.PermissionDenied, "Cannot create AI post for non-AI character")
	}

	// Validate media
	mediaResp, err := s.mediaClient.GetMedia(ctx, &mediapb.GetMediaRequest{
		MediaId: req.MediaId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get media info", zap.Error(err), zap.String("media_id", req.MediaId))
		return nil, status.Errorf(codes.Internal, "failed to validate media")
	}

	// Check if media belongs to the character
	if mediaResp.CharacterId != req.CharacterId {
		return nil, status.Errorf(codes.PermissionDenied, "media does not belong to the character")
	}

	// Create post
	post := &models.Post{
		CharacterID: req.CharacterId,
		IsAI:        true, // Always true for AI posts
		WorldID:     req.WorldId,
		Caption:     req.Caption,
		MediaID:     req.MediaId,
	}

	err = s.postRepo.Create(ctx, post)
	if err != nil {
		logger.Logger.Error("Failed to create AI post", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create post")
	}

	return &postpb.CreatePostResponse{
		PostId:    post.ID,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
	}, nil
}
