package repository

import (
	"context"
	"time"

	"instagram-clone/pkg/logger"
	"instagram-clone/services/interaction-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// InteractionRepository handles database operations for interactions
type InteractionRepository interface {
	// Likes
	AddLike(ctx context.Context, like *models.Like) error
	RemoveLike(ctx context.Context, postID, userID string) error
	GetPostLikes(ctx context.Context, postID string, limit, offset int) ([]*models.Like, int, error)
	CheckUserLiked(ctx context.Context, postID, userID string) (bool, error)
	
	// Comments
	AddComment(ctx context.Context, comment *models.Comment) error
	GetPostComments(ctx context.Context, postID string, limit, offset int) ([]*models.Comment, int, error)
	
	// Stats
	GetPostStats(ctx context.Context, postID string) (*models.PostStats, error)
	GetPostsStats(ctx context.Context, postIDs []string) (map[string]*models.PostStats, error)
	UpdatePostStats(ctx context.Context, postID string) error
}

type interactionRepository struct {
	db             *mongo.Database
	likesCol       *mongo.Collection
	commentsCol    *mongo.Collection
	statsCol       *mongo.Collection
}

// NewInteractionRepository creates a new InteractionRepository
func NewInteractionRepository(db *mongo.Database) InteractionRepository {
	return &interactionRepository{
		db:             db,
		likesCol:       db.Collection("likes"),
		commentsCol:    db.Collection("comments"),
		statsCol:       db.Collection("stats"),
	}
}

// AddLike adds a like to a post
func (r *interactionRepository) AddLike(ctx context.Context, like *models.Like) error {
	like.CreatedAt = time.Now()
	
	// Check if like already exists
	filter := bson.M{"post_id": like.PostID, "user_id": like.UserID}
	count, err := r.likesCol.CountDocuments(ctx, filter)
	if err != nil {
		logger.Logger.Error("Failed to check if like exists", zap.Error(err))
		return err
	}
	
	if count > 0 {
		// Like already exists, nothing to do
		return nil
	}
	
	// Insert like
	_, err = r.likesCol.InsertOne(ctx, like)
	if err != nil {
		logger.Logger.Error("Failed to add like", zap.Error(err))
		return err
	}
	
	// Update post stats
	err = r.UpdatePostStats(ctx, like.PostID)
	if err != nil {
		logger.Logger.Error("Failed to update post stats after adding like", zap.Error(err))
		// Continue even if stats update fails
	}
	
	return nil
}

// RemoveLike removes a like from a post
func (r *interactionRepository) RemoveLike(ctx context.Context, postID, userID string) error {
	filter := bson.M{"post_id": postID, "user_id": userID}
	
	// Delete like
	result, err := r.likesCol.DeleteOne(ctx, filter)
	if err != nil {
		logger.Logger.Error("Failed to remove like", zap.Error(err))
		return err
	}
	
	if result.DeletedCount > 0 {
		// Update post stats
		err = r.UpdatePostStats(ctx, postID)
		if err != nil {
			logger.Logger.Error("Failed to update post stats after removing like", zap.Error(err))
			// Continue even if stats update fails
		}
	}
	
	return nil
}

// GetPostLikes gets likes for a post with pagination
func (r *interactionRepository) GetPostLikes(ctx context.Context, postID string, limit, offset int) ([]*models.Like, int, error) {
	filter := bson.M{"post_id": postID}
	
	// Get total count
	total, err := r.likesCol.CountDocuments(ctx, filter)
	if err != nil {
		logger.Logger.Error("Failed to count likes", zap.Error(err))
		return nil, 0, err
	}
	
	// Define sort and pagination options
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))
	
	// Execute query
	cursor, err := r.likesCol.Find(ctx, filter, opts)
	if err != nil {
		logger.Logger.Error("Failed to get likes", zap.Error(err))
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	// Decode results
	likes := []*models.Like{}
	if err := cursor.All(ctx, &likes); err != nil {
		logger.Logger.Error("Failed to decode likes", zap.Error(err))
		return nil, 0, err
	}
	
	return likes, int(total), nil
}

// CheckUserLiked checks if a user has liked a post
func (r *interactionRepository) CheckUserLiked(ctx context.Context, postID, userID string) (bool, error) {
	filter := bson.M{"post_id": postID, "user_id": userID}
	
	count, err := r.likesCol.CountDocuments(ctx, filter)
	if err != nil {
		logger.Logger.Error("Failed to check if user liked post", zap.Error(err))
		return false, err
	}
	
	return count > 0, nil
}

// AddComment adds a comment to a post
func (r *interactionRepository) AddComment(ctx context.Context, comment *models.Comment) error {
	comment.CreatedAt = time.Now()
	
	// Generate ObjectID if not provided
	if comment.ID == "" {
		objID := primitive.NewObjectID()
		comment.ID = objID.Hex()
	}
	
	// Insert comment
	_, err := r.commentsCol.InsertOne(ctx, comment)
	if err != nil {
		logger.Logger.Error("Failed to add comment", zap.Error(err))
		return err
	}
	
	// Update post stats
	err = r.UpdatePostStats(ctx, comment.PostID)
	if err != nil {
		logger.Logger.Error("Failed to update post stats after adding comment", zap.Error(err))
		// Continue even if stats update fails
	}
	
	return nil
}

// GetPostComments gets comments for a post with pagination
func (r *interactionRepository) GetPostComments(ctx context.Context, postID string, limit, offset int) ([]*models.Comment, int, error) {
	filter := bson.M{"post_id": postID}
	
	// Get total count
	total, err := r.commentsCol.CountDocuments(ctx, filter)
	if err != nil {
		logger.Logger.Error("Failed to count comments", zap.Error(err))
		return nil, 0, err
	}
	
	// Define sort and pagination options
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))
	
	// Execute query
	cursor, err := r.commentsCol.Find(ctx, filter, opts)
	if err != nil {
		logger.Logger.Error("Failed to get comments", zap.Error(err))
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	// Decode results
	comments := []*models.Comment{}
	if err := cursor.All(ctx, &comments); err != nil {
		logger.Logger.Error("Failed to decode comments", zap.Error(err))
		return nil, 0, err
	}
	
	return comments, int(total), nil
}

// GetPostStats gets the stats for a post
func (r *interactionRepository) GetPostStats(ctx context.Context, postID string) (*models.PostStats, error) {
	filter := bson.M{"_id": postID}
	
	var stats models.PostStats
	err := r.statsCol.FindOne(ctx, filter).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No stats found, let's calculate them
			return r.calculatePostStats(ctx, postID)
		}
		
		logger.Logger.Error("Failed to get post stats", zap.Error(err))
		return nil, err
	}
	
	return &stats, nil
}

// GetPostsStats gets the stats for multiple posts
func (r *interactionRepository) GetPostsStats(ctx context.Context, postIDs []string) (map[string]*models.PostStats, error) {
	if len(postIDs) == 0 {
		return make(map[string]*models.PostStats), nil
	}
	
	filter := bson.M{"_id": bson.M{"$in": postIDs}}
	
	// Execute query
	cursor, err := r.statsCol.Find(ctx, filter)
	if err != nil {
		logger.Logger.Error("Failed to get posts stats", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)
	
	// Decode results
	stats := []*models.PostStats{}
	if err := cursor.All(ctx, &stats); err != nil {
		logger.Logger.Error("Failed to decode posts stats", zap.Error(err))
		return nil, err
	}
	
	// Convert to map
	result := make(map[string]*models.PostStats)
	for _, stat := range stats {
		result[stat.PostID] = stat
	}
	
	// Check if we need to calculate stats for any missing posts
	for _, postID := range postIDs {
		if _, ok := result[postID]; !ok {
			// Stats not found, calculate them
			stat, err := r.calculatePostStats(ctx, postID)
			if err != nil {
				logger.Logger.Error("Failed to calculate post stats", zap.Error(err), zap.String("post_id", postID))
				// Continue with next post
				continue
			}
			
			result[postID] = stat
		}
	}
	
	return result, nil
}

// UpdatePostStats updates the stats for a post
func (r *interactionRepository) UpdatePostStats(ctx context.Context, postID string) error {
	// Calculate stats
	likesCount, err := r.likesCol.CountDocuments(ctx, bson.M{"post_id": postID})
	if err != nil {
		logger.Logger.Error("Failed to count likes for post stats", zap.Error(err))
		return err
	}
	
	commentsCount, err := r.commentsCol.CountDocuments(ctx, bson.M{"post_id": postID})
	if err != nil {
		logger.Logger.Error("Failed to count comments for post stats", zap.Error(err))
		return err
	}
	
	// Update stats
	now := time.Now()
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$set": bson.M{
			"likes_count":    likesCount,
			"comments_count": commentsCount,
			"updated_at":     now,
		},
	}
	opts := options.Update().SetUpsert(true)
	
	_, err = r.statsCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Logger.Error("Failed to update post stats", zap.Error(err))
		return err
	}
	
	return nil
}

// calculatePostStats calculates the stats for a post
func (r *interactionRepository) calculatePostStats(ctx context.Context, postID string) (*models.PostStats, error) {
	// Calculate stats
	likesCount, err := r.likesCol.CountDocuments(ctx, bson.M{"post_id": postID})
	if err != nil {
		logger.Logger.Error("Failed to count likes for post stats", zap.Error(err))
		return nil, err
	}
	
	commentsCount, err := r.commentsCol.CountDocuments(ctx, bson.M{"post_id": postID})
	if err != nil {
		logger.Logger.Error("Failed to count comments for post stats", zap.Error(err))
		return nil, err
	}
	
	// Create stats object
	now := time.Now()
	stats := &models.PostStats{
		PostID:        postID,
		LikesCount:    int32(likesCount),
		CommentsCount: int32(commentsCount),
		UpdatedAt:     now,
	}
	
	// Save to database
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$set": stats,
	}
	opts := options.Update().SetUpsert(true)
	
	_, err = r.statsCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Logger.Error("Failed to update post stats", zap.Error(err))
		return nil, err
	}
	
	return stats, nil
}