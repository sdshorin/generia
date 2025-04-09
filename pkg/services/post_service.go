package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"instagram-clone/internal/repositories"
	"instagram-clone/pkg/config"
	"instagram-clone/pkg/logger"
	"instagram-clone/pkg/models"
	"go.uber.org/zap"
)

// PostService интерфейс для работы с постами
type PostService interface {
	CreatePost(ctx context.Context, userID string, input *models.CreatePostRequest) (*models.Post, error)
	GetPostByID(ctx context.Context, id string, userID string) (*models.Post, error)
	GetGlobalFeed(ctx context.Context, userID string, limit, offset int) ([]*models.Post, error)
	GetUserPosts(ctx context.Context, userID string, requestUserID string, limit, offset int) ([]*models.Post, error)
	
	// Комментарии
	CreateComment(ctx context.Context, postID, userID string, input *models.CreateCommentRequest) (*models.Comment, error)
	GetPostComments(ctx context.Context, postID string, limit, offset int) ([]*models.Comment, error)
	
	// Лайки
	LikePost(ctx context.Context, postID, userID string) error
	UnlikePost(ctx context.Context, postID, userID string) error
	CheckUserLiked(ctx context.Context, postID, userID string) (bool, error)
}

// PostServiceImpl реализует PostService
type PostServiceImpl struct {
	postRepo      repositories.PostRepository
	userRepo      repositories.UserRepository
	config        *config.Config
	minioClient   *minio.Client
}

// NewPostService создает новый сервис постов
func NewPostService(postRepo repositories.PostRepository, userRepo repositories.UserRepository, config *config.Config) (PostService, error) {
	// Инициализируем клиент MinIO
	minioClient, err := minio.New(config.Storage.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Storage.AccessKey, config.Storage.SecretKey, ""),
		Secure: config.Storage.UseSSL,
	})
	if err != nil {
		logger.Logger.Error("Failed to initialize MinIO client", zap.Error(err))
		return nil, err
	}

	// Проверяем существует ли бакет
	exists, err := minioClient.BucketExists(context.Background(), config.Storage.Bucket)
	if err != nil {
		logger.Logger.Error("Failed to check if bucket exists", zap.Error(err))
		return nil, err
	}

	// Если бакет не существует, создаем его
	if !exists {
		err = minioClient.MakeBucket(context.Background(), config.Storage.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			logger.Logger.Error("Failed to create bucket", zap.Error(err))
			return nil, err
		}
		logger.Logger.Info("Created new bucket", zap.String("bucket", config.Storage.Bucket))

		// Устанавливаем политику доступа на чтение для всех
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, config.Storage.Bucket)

		err = minioClient.SetBucketPolicy(context.Background(), config.Storage.Bucket, policy)
		if err != nil {
			logger.Logger.Error("Failed to set bucket policy", zap.Error(err))
		}
	}

	return &PostServiceImpl{
		postRepo:    postRepo,
		userRepo:    userRepo,
		config:      config,
		minioClient: minioClient,
	}, nil
}

// CreatePost создает новый пост
func (s *PostServiceImpl) CreatePost(ctx context.Context, userID string, input *models.CreatePostRequest) (*models.Post, error) {
	// Проверяем наличие пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err), zap.String("userID", userID))
		return nil, errors.New("user not found")
	}

	// Проверяем наличие изображения
	if input.Image == "" {
		return nil, errors.New("image is required")
	}

	// Декодируем Base64 изображение
	imageData, err := s.decodeBase64Image(input.Image)
	if err != nil {
		logger.Logger.Error("Failed to decode image", zap.Error(err))
		return nil, errors.New("invalid image format")
	}

	// Генерируем имя файла
	fileName := fmt.Sprintf("posts/%s/%s.jpg", userID, uuid.New().String())

	// Загружаем изображение в MinIO
	_, err = s.minioClient.PutObject(
		ctx,
		s.config.Storage.Bucket,
		fileName,
		imageData.data,
		imageData.size,
		minio.PutObjectOptions{ContentType: imageData.contentType},
	)
	if err != nil {
		logger.Logger.Error("Failed to upload image to MinIO", zap.Error(err))
		return nil, errors.New("failed to upload image")
	}

	// Формируем URL изображения
	imageURL := fmt.Sprintf("http://%s/%s/%s", s.config.Storage.Endpoint, s.config.Storage.Bucket, fileName)

	// Создаем новый пост
	post := &models.Post{
		UserID:   userID,
		Username: user.Username,
		Caption:  input.Caption,
		ImageURL: imageURL,
	}

	// Сохраняем пост в базе данных
	err = s.postRepo.Create(ctx, post)
	if err != nil {
		logger.Logger.Error("Failed to create post", zap.Error(err))
		return nil, errors.New("failed to create post")
	}

	return post, nil
}

// GetPostByID получает пост по ID
func (s *PostServiceImpl) GetPostByID(ctx context.Context, id string, userID string) (*models.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("id", id))
		return nil, errors.New("post not found")
	}

	// Проверяем, лайкнул ли пользователь этот пост
	if userID != "" {
		liked, err := s.postRepo.CheckUserLiked(ctx, id, userID)
		if err != nil {
			logger.Logger.Warn("Failed to check if user liked post", zap.Error(err))
		} else {
			post.UserLiked = liked
		}
	}

	return post, nil
}

// GetGlobalFeed получает глобальную ленту постов
func (s *PostServiceImpl) GetGlobalFeed(ctx context.Context, userID string, limit, offset int) ([]*models.Post, error) {
	posts, err := s.postRepo.GetGlobalFeed(ctx, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get global feed", zap.Error(err))
		return nil, errors.New("failed to get posts")
	}

	// Проверяем, лайкнул ли пользователь каждый пост
	if userID != "" {
		for _, post := range posts {
			liked, err := s.postRepo.CheckUserLiked(ctx, post.ID, userID)
			if err != nil {
				logger.Logger.Warn("Failed to check if user liked post", zap.Error(err))
			} else {
				post.UserLiked = liked
			}
		}
	}

	return posts, nil
}

// GetUserPosts получает посты пользователя
func (s *PostServiceImpl) GetUserPosts(ctx context.Context, userID string, requestUserID string, limit, offset int) ([]*models.Post, error) {
	posts, err := s.postRepo.GetUserPosts(ctx, userID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get user posts", zap.Error(err), zap.String("userID", userID))
		return nil, errors.New("failed to get posts")
	}

	// Проверяем, лайкнул ли запрашивающий пользователь каждый пост
	if requestUserID != "" {
		for _, post := range posts {
			liked, err := s.postRepo.CheckUserLiked(ctx, post.ID, requestUserID)
			if err != nil {
				logger.Logger.Warn("Failed to check if user liked post", zap.Error(err))
			} else {
				post.UserLiked = liked
			}
		}
	}

	return posts, nil
}

// CreateComment создает новый комментарий
func (s *PostServiceImpl) CreateComment(ctx context.Context, postID, userID string, input *models.CreateCommentRequest) (*models.Comment, error) {
	// Проверяем наличие поста
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("postID", postID))
		return nil, errors.New("post not found")
	}

	// Проверяем наличие пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err), zap.String("userID", userID))
		return nil, errors.New("user not found")
	}

	// Создаем новый комментарий
	comment := &models.Comment{
		PostID:   postID,
		UserID:   userID,
		Username: user.Username,
		Text:     input.Text,
	}

	// Сохраняем комментарий в базе данных
	err = s.postRepo.CreateComment(ctx, comment)
	if err != nil {
		logger.Logger.Error("Failed to create comment", zap.Error(err))
		return nil, errors.New("failed to create comment")
	}

	return comment, nil
}

// GetPostComments получает комментарии к посту
func (s *PostServiceImpl) GetPostComments(ctx context.Context, postID string, limit, offset int) ([]*models.Comment, error) {
	comments, err := s.postRepo.GetPostComments(ctx, postID, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to get post comments", zap.Error(err), zap.String("postID", postID))
		return nil, errors.New("failed to get comments")
	}

	return comments, nil
}

// LikePost ставит лайк посту
func (s *PostServiceImpl) LikePost(ctx context.Context, postID, userID string) error {
	// Проверяем наличие поста
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("postID", postID))
		return errors.New("post not found")
	}

	// Проверяем наличие пользователя
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err), zap.String("userID", userID))
		return errors.New("user not found")
	}

	// Проверяем, не поставил ли пользователь уже лайк
	liked, err := s.postRepo.CheckUserLiked(ctx, postID, userID)
	if err != nil {
		logger.Logger.Error("Failed to check if user liked post", zap.Error(err))
		return errors.New("failed to like post")
	}

	if liked {
		return errors.New("post already liked")
	}

	// Создаем новый лайк
	like := &models.Like{
		PostID: postID,
		UserID: userID,
	}

	// Сохраняем лайк в базе данных
	err = s.postRepo.CreateLike(ctx, like)
	if err != nil {
		logger.Logger.Error("Failed to create like", zap.Error(err))
		return errors.New("failed to like post")
	}

	return nil
}

// UnlikePost удаляет лайк с поста
func (s *PostServiceImpl) UnlikePost(ctx context.Context, postID, userID string) error {
	// Проверяем наличие поста
	_, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		logger.Logger.Error("Failed to get post", zap.Error(err), zap.String("postID", postID))
		return errors.New("post not found")
	}

	// Проверяем наличие пользователя
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Logger.Error("Failed to get user", zap.Error(err), zap.String("userID", userID))
		return errors.New("user not found")
	}

	// Проверяем, поставил ли пользователь лайк
	liked, err := s.postRepo.CheckUserLiked(ctx, postID, userID)
	if err != nil {
		logger.Logger.Error("Failed to check if user liked post", zap.Error(err))
		return errors.New("failed to unlike post")
	}

	if !liked {
		return errors.New("post not liked")
	}

	// Удаляем лайк из базы данных
	err = s.postRepo.DeleteLike(ctx, postID, userID)
	if err != nil {
		logger.Logger.Error("Failed to delete like", zap.Error(err))
		return errors.New("failed to unlike post")
	}

	return nil
}

// CheckUserLiked проверяет, поставил ли пользователь лайк посту
func (s *PostServiceImpl) CheckUserLiked(ctx context.Context, postID, userID string) (bool, error) {
	return s.postRepo.CheckUserLiked(ctx, postID, userID)
}

type imageData struct {
	data        io.Reader
	size        int64
	contentType string
}

// decodeBase64Image декодирует Base64 изображение
func (s *PostServiceImpl) decodeBase64Image(base64Image string) (*imageData, error) {
	// Удаляем префикс "data:image/jpeg;base64," если он есть
	var b64data string
	if strings.Contains(base64Image, ",") {
		b64data = strings.Split(base64Image, ",")[1]
	} else {
		b64data = base64Image
	}

	// Декодируем Base64
	decoded, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return nil, err
	}

	return &imageData{
		data:        strings.NewReader(string(decoded)),
		size:        int64(len(decoded)),
		contentType: "image/jpeg",
	}, nil
}