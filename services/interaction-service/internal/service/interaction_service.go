package service

import (
	"context"
	"time"

	"instagram-clone/pkg/logger"
	"instagram-clone/services/interaction-service/internal/models"
	"instagram-clone/services/interaction-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "instagram-clone/api/grpc/auth"
	interactionpb "instagram-clone/api/grpc/interaction"
)

// InteractionService implements the interaction gRPC service
type InteractionService struct {
	interactionpb.UnimplementedInteractionServiceServer
	interactionRepo repository.InteractionRepository
	authClient      authpb.AuthServiceClient
}

// NewInteractionService creates a new InteractionService
func NewInteractionService(
	interactionRepo repository.InteractionRepository,
	authClient authpb.AuthServiceClient,
) interactionpb.InteractionServiceServer {
	return &InteractionService{
		interactionRepo: interactionRepo,
		authClient:      authClient,
	}
}

// LikePost handles liking a post
func (s *InteractionService) LikePost(ctx context.Context, req *interactionpb.LikePostRequest) (*interactionpb.LikePostResponse, error) {
	// Validate input
	if req.PostId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id and user_id are required")
	}

	// Validate user
	userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: req.UserId,
	})
	if err != nil {
		logger.Logger.Error("Failed to validate user", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Errorf(codes.Internal, "failed to validate user")
	}
	if userResp == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Add like
	like := &models.Like{
		PostID: req.PostId,
		UserID: req.UserId,
	}
	err = s.interactionRepo.AddLike(ctx, like)
	if err != nil {
		logger.Logger.Error("Failed to add like", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to add like")
	}

	// Get updated stats
	stats, err := s.interactionRepo.GetPostStats(ctx, req.PostId)
	if err != nil {
		logger.Logger.Error("Failed to get post stats", zap.Error(err))
		// Continue even if stats can't be retrieved
		stats = &models.PostStats{
			PostID:     req.PostId,
			LikesCount: 1, // At least 1 like (the one we just added)
		}
	}

	return &interactionpb.LikePostResponse{
		Success:    true,
		LikesCount: stats.LikesCount,
	}, nil
}

// UnlikePost handles unliking a post
func (s *InteractionService) UnlikePost(ctx context.Context, req *interactionpb.UnlikePostRequest) (*interactionpb.UnlikePostResponse, error) {
	// Validate input
	if req.PostId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id and user_id are required")
	}

	// Remove like
	err := s.interactionRepo.RemoveLike(ctx, req.PostId, req.UserId)
	if err != nil {
		logger.Logger.Error("Failed to remove like", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to remove like")
	}

	// Get updated stats
	stats, err := s.interactionRepo.GetPostStats(ctx, req.PostId)
	if err != nil {
		logger.Logger.Error("Failed to get post stats", zap.Error(err))
		// Continue even if stats can't be retrieved
		stats = &models.PostStats{
			PostID:     req.PostId,
			LikesCount: 0,
		}
	}

	return &interactionpb.UnlikePostResponse{
		Success:    true,
		LikesCount: stats.LikesCount,
	}, nil
}

// GetPostLikes handles retrieving likes for a post
func (s *InteractionService) GetPostLikes(ctx context.Context, req *interactionpb.GetPostLikesRequest) (*interactionpb.PostLikesResponse, error) {
	// Validate input
	if req.PostId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Get likes
	likes, total, err := s.interactionRepo.GetPostLikes(ctx, req.PostId, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get post likes", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get post likes")
	}

	// Extract user IDs
	userIDs := make(map[string]struct{})
	for _, like := range likes {
		userIDs[like.UserID] = struct{}{}
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

	// Build response
	result := make([]*interactionpb.Like, len(likes))
	for i, like := range likes {
		// Get username
		var username string
		if userInfo, ok := userInfoMap[like.UserID]; ok {
			username = userInfo.Username
		}

		result[i] = &interactionpb.Like{
			UserId:    like.UserID,
			Username:  username,
			CreatedAt: like.CreatedAt.Format(time.RFC3339),
		}
	}

	return &interactionpb.PostLikesResponse{
		Likes: result,
		Total: int32(total),
	}, nil
}

// CheckUserLiked handles checking if a user has liked a post
func (s *InteractionService) CheckUserLiked(ctx context.Context, req *interactionpb.CheckUserLikedRequest) (*interactionpb.CheckUserLikedResponse, error) {
	// Validate input
	if req.PostId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id and user_id are required")
	}

	// Check if user liked post
	liked, err := s.interactionRepo.CheckUserLiked(ctx, req.PostId, req.UserId)
	if err != nil {
		logger.Logger.Error("Failed to check if user liked post", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to check if user liked post")
	}

	return &interactionpb.CheckUserLikedResponse{
		Liked: liked,
	}, nil
}

// AddComment handles adding a comment to a post
func (s *InteractionService) AddComment(ctx context.Context, req *interactionpb.AddCommentRequest) (*interactionpb.AddCommentResponse, error) {
	// Validate input
	if req.PostId == "" || req.UserId == "" || req.Text == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id, user_id, and text are required")
	}

	// Validate user
	userResp, err := s.authClient.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		UserId: req.UserId,
	})
	if err != nil {
		logger.Logger.Error("Failed to validate user", zap.Error(err), zap.String("user_id", req.UserId))
		return nil, status.Errorf(codes.Internal, "failed to validate user")
	}
	if userResp == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Add comment
	comment := &models.Comment{
		PostID: req.PostId,
		UserID: req.UserId,
		Text:   req.Text,
	}
	err = s.interactionRepo.AddComment(ctx, comment)
	if err != nil {
		logger.Logger.Error("Failed to add comment", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to add comment")
	}

	return &interactionpb.AddCommentResponse{
		CommentId: comment.ID,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetPostComments handles retrieving comments for a post
func (s *InteractionService) GetPostComments(ctx context.Context, req *interactionpb.GetPostCommentsRequest) (*interactionpb.PostCommentsResponse, error) {
	// Validate input
	if req.PostId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Get comments
	comments, total, err := s.interactionRepo.GetPostComments(ctx, req.PostId, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get post comments", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get post comments")
	}

	// Extract user IDs
	userIDs := make(map[string]struct{})
	for _, comment := range comments {
		userIDs[comment.UserID] = struct{}{}
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

	// Build response
	result := make([]*interactionpb.Comment, len(comments))
	for i, comment := range comments {
		// Get username
		var username string
		if userInfo, ok := userInfoMap[comment.UserID]; ok {
			username = userInfo.Username
		}

		result[i] = &interactionpb.Comment{
			CommentId: comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Username:  username,
			Text:      comment.Text,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		}
	}

	return &interactionpb.PostCommentsResponse{
		Comments: result,
		Total:    int32(total),
	}, nil
}

// GetPostStats handles retrieving stats for a post
func (s *InteractionService) GetPostStats(ctx context.Context, req *interactionpb.GetPostStatsRequest) (*interactionpb.PostStatsResponse, error) {
	// Validate input
	if req.PostId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "post_id is required")
	}

	// Get stats
	stats, err := s.interactionRepo.GetPostStats(ctx, req.PostId)
	if err != nil {
		logger.Logger.Error("Failed to get post stats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get post stats")
	}

	if stats == nil {
		// Post has no interactions yet
		return &interactionpb.PostStatsResponse{
			PostId:        req.PostId,
			LikesCount:    0,
			CommentsCount: 0,
		}, nil
	}

	return &interactionpb.PostStatsResponse{
		PostId:        stats.PostID,
		LikesCount:    stats.LikesCount,
		CommentsCount: stats.CommentsCount,
	}, nil
}

// GetPostsStats handles retrieving stats for multiple posts
func (s *InteractionService) GetPostsStats(ctx context.Context, req *interactionpb.GetPostsStatsRequest) (*interactionpb.PostsStatsResponse, error) {
	// Validate input
	if len(req.PostIds) == 0 {
		return &interactionpb.PostsStatsResponse{
			Stats: make(map[string]*interactionpb.PostStatsResponse),
		}, nil
	}

	// Get stats
	stats, err := s.interactionRepo.GetPostsStats(ctx, req.PostIds)
	if err != nil {
		logger.Logger.Error("Failed to get posts stats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get posts stats")
	}

	// Build response
	result := make(map[string]*interactionpb.PostStatsResponse)
	for _, postID := range req.PostIds {
		var stat *models.PostStats
		var ok bool
		if stat, ok = stats[postID]; !ok {
			// Post has no interactions yet
			result[postID] = &interactionpb.PostStatsResponse{
				PostId:        postID,
				LikesCount:    0,
				CommentsCount: 0,
			}
			continue
		}

		result[postID] = &interactionpb.PostStatsResponse{
			PostId:        stat.PostID,
			LikesCount:    stat.LikesCount,
			CommentsCount: stat.CommentsCount,
		}
	}

	return &interactionpb.PostsStatsResponse{
		Stats: result,
	}, nil
}

// HealthCheck handles health check requests
func (s *InteractionService) HealthCheck(ctx context.Context, req *interactionpb.HealthCheckRequest) (*interactionpb.HealthCheckResponse, error) {
	return &interactionpb.HealthCheckResponse{
		Status: interactionpb.HealthCheckResponse_SERVING,
	}, nil
}