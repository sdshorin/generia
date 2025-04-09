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
}

// NewPostService creates a new PostService
func NewPostService(
	postRepo repository.PostRepository,
	authClient authpb.AuthServiceClient,
	mediaClient mediapb.MediaServiceClient,
	interactionClient interactionpb.InteractionServiceClient,
) postpb.PostServiceServer {
	return &PostService{
		postRepo:          postRepo,
		authClient:        authClient,
		mediaClient:       mediaClient,
		interactionClient: interactionClient,
	}
}

// CreatePost handles post creation
func (s *PostService) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*postpb.CreatePostResponse, error) {
	// Validate input
	if req.UserId == "" || req.MediaId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id and media_id are required")
	}

	// Validate user
	_, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: req.UserId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get user info", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Errorf(codes.Internal, "failed to validate user")
	}

	// Validate media
	mediaResp, err := s.mediaClient.GetMedia(ctx, &mediapb.GetMediaRequest{
		MediaId: req.MediaId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get media info", zap.Error(err), zap.String("media_id", req.MediaId))
		return nil, status.Errorf(codes.Internal, "failed to validate media")
	}

	// Check if media belongs to the user
	if mediaResp.UserId != req.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "media does not belong to the user")
	}

	// Create post
	post := &models.Post{
		UserID:   req.UserId,
		Caption:  req.Caption,
		MediaID:  req.MediaId,
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

	// Get user info
	userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: post.UserID,
	})
	if err != nil {
		logger.Logger.Error("Failed to get user info", zap.Error(err), zap.String("user_id", post.UserID))
		return nil, status.Errorf(codes.Internal, "failed to get user info")
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
		UserId:        post.UserID,
		Username:      userResp.Username,
		Caption:       post.Caption,
		MediaUrl:      mediaResp.Url,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		LikesCount:    statsResp.LikesCount,
		CommentsCount: statsResp.CommentsCount,
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

	// Get user info once for all posts
	userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: req.UserId,
	})
	if err != nil {
		logger.Logger.Error("Failed to get user info", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Errorf(codes.Internal, "failed to get user info")
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
			UserId:        post.UserID,
			Username:      userResp.Username,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
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

	// Extract unique user IDs
	userIDs := make(map[string]struct{})
	for _, post := range posts {
		userIDs[post.UserID] = struct{}{}
	}

	// Get user info for all users
	userInfoMap := make(map[string]*authpb.UserInfo)
	for userID := range userIDs {
		userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
			UserId: userID,
		})
		if err != nil {
			logger.Logger.Error("Failed to get user info", zap.Error(err), zap.String("user_id", userID))
			continue
		}
		userInfoMap[userID] = userResp
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
		// Get user info
		var username string
		if userInfo, ok := userInfoMap[post.UserID]; ok {
			username = userInfo.Username
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
			UserId:        post.UserID,
			Username:      username,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
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

	// Get posts
	posts, nextCursor, err := s.postRepo.GetGlobalFeed(ctx, limit, req.Cursor)
	if err != nil {
		logger.Logger.Error("Failed to get global feed", 
			zap.Error(err), 
			zap.Int("limit", limit), 
			zap.String("cursor", req.Cursor))
		return nil, status.Errorf(codes.Internal, "failed to get global feed")
	}

	if len(posts) == 0 {
		return &postpb.PostList{
			Posts: []*postpb.Post{},
			Total: 0,
			NextCursor: "",
		}, nil
	}

	// Extract unique user IDs
	userIDs := make(map[string]struct{})
	for _, post := range posts {
		userIDs[post.UserID] = struct{}{}
	}

	// Get user info for all users
	userInfoMap := make(map[string]*authpb.UserInfo)
	for userID := range userIDs {
		userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
			UserId: userID,
		})
		if err != nil {
			logger.Logger.Error("Failed to get user info", zap.Error(err), zap.String("user_id", userID))
			continue
		}
		userInfoMap[userID] = userResp
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
		// Get user info
		var username string
		if userInfo, ok := userInfoMap[post.UserID]; ok {
			username = userInfo.Username
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
			UserId:        post.UserID,
			Username:      username,
			Caption:       post.Caption,
			MediaUrl:      mediaURL,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    likesCount,
			CommentsCount: commentsCount,
		}
	}

	return &postpb.PostList{
		Posts: result,
		Total: int32(len(result)),
		NextCursor: nextCursor,
	}, nil
}

// HealthCheck implements health check
func (s *PostService) HealthCheck(ctx context.Context, req *postpb.HealthCheckRequest) (*postpb.HealthCheckResponse, error) {
	return &postpb.HealthCheckResponse{
		Status: postpb.HealthCheckResponse_SERVING,
	}, nil
}