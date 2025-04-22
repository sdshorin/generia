package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/database"
	"github.com/sdshorin/generia/pkg/discovery"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/pkg/telemetry"
	"github.com/sdshorin/generia/services/media-service/internal/repository"
	"github.com/sdshorin/generia/services/media-service/internal/service"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
	mediapb "github.com/sdshorin/generia/api/grpc/media"
)

// MediaService implements the media service
type MediaService struct {
	mediapb.UnimplementedMediaServiceServer
	logger      *zap.Logger
	authClient  authpb.AuthServiceClient
	minioClient *minio.Client
	db          *sqlx.DB
	bucket      string
}

// GetPresignedUploadURL generates a presigned URL for direct upload to storage
func (s *MediaService) GetPresignedUploadURL(ctx context.Context, req *mediapb.GetPresignedUploadURLRequest) (*mediapb.GetPresignedUploadURLResponse, error) {
	s.logger.Info("GetPresignedUploadURL called",
		zap.String("character_id", req.CharacterId),
		zap.String("filename", req.Filename),
		zap.String("content_type", req.ContentType),
		zap.Int64("size", req.Size))

	// Create media service instance
	mediaRepo := repository.NewPostgresMediaRepository(s.db, s.minioClient)
	mediaService := service.NewMediaService(mediaRepo, s.minioClient, s.bucket, s.logger)

	// Generate presigned URL
	media, presignedURL, expiresAt, err := mediaService.GeneratePresignedPutURL(
		ctx,
		req.CharacterId,
		req.Filename,
		req.ContentType,
		req.Size,
	)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", zap.Error(err))
		return nil, err
	}

	return &mediapb.GetPresignedUploadURLResponse{
		MediaId:   media.ID,
		UploadUrl: presignedURL,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// ConfirmUpload confirms that a media file has been uploaded via presigned URL
func (s *MediaService) ConfirmUpload(ctx context.Context, req *mediapb.ConfirmUploadRequest) (*mediapb.ConfirmUploadResponse, error) {
	s.logger.Info("ConfirmUpload called",
		zap.String("media_id", req.MediaId),
		zap.String("character_id", req.CharacterId))

	// Create media service instance
	mediaRepo := repository.NewPostgresMediaRepository(s.db, s.minioClient)
	mediaService := service.NewMediaService(mediaRepo, s.minioClient, s.bucket, s.logger)

	// Check if media exists and belongs to the character
	media, err := mediaRepo.GetMediaByID(ctx, req.MediaId)
	if err != nil {
		s.logger.Error("Failed to get media from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get media from database: %w", err)
	}

	// Confirm upload
	err = mediaService.ConfirmMediaUpload(ctx, req.MediaId, req.CharacterId)
	if err != nil {
		s.logger.Error("Failed to confirm upload", zap.Error(err))
		return nil, err
	}

	// Generate variants asynchronously (in a real application, this could be done via Kafka)
	// For now, we'll just generate them synchronously
	variants, err := mediaService.GenerateVariants(ctx, req.MediaId, []string{"thumbnail", "medium"})
	if err != nil {
		s.logger.Error("Failed to generate variants", zap.Error(err))
		// Continue even if variant generation fails
	}

	// Return response with variants (if any)
	variantsProto := make([]*mediapb.MediaVariant, 0, len(variants))
	for _, v := range variants {
		variantsProto = append(variantsProto, &mediapb.MediaVariant{
			Name:   v.Name,
			Url:    v.URL,
			Width:  v.Width,
			Height: v.Height,
		})
	}

	// Add the original as a variant
	urlStr, _, err := mediaService.GetPresignedURL(ctx, media, "original", time.Hour)
	if err == nil {
		variantsProto = append(variantsProto, &mediapb.MediaVariant{
			Name: "original",
			Url:  urlStr,
		})
	}

	return &mediapb.ConfirmUploadResponse{
		Success:  true,
		Variants: variantsProto,
	}, nil
}

// GetMedia implements the GetMedia method
func (s *MediaService) GetMedia(ctx context.Context, req *mediapb.GetMediaRequest) (*mediapb.Media, error) {
	s.logger.Info("GetMedia called", zap.String("media_id", req.MediaId))

	// Create media service instance
	mediaRepo := repository.NewPostgresMediaRepository(s.db, s.minioClient)
	mediaService := service.NewMediaService(mediaRepo, s.minioClient, s.bucket, s.logger)

	// Get media from database
	media, variants, err := mediaService.GetMedia(ctx, req.MediaId)
	if err != nil {
		s.logger.Error("Failed to get media", zap.Error(err))
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	// Convert variants to proto format
	variantsProto := make([]*mediapb.MediaVariant, 0, len(variants))
	for _, v := range variants {
		variantsProto = append(variantsProto, &mediapb.MediaVariant{
			Name:   v.Name,
			Url:    v.URL,
			Width:  v.Width,
			Height: v.Height,
		})
	}

	// Add the original as a variant if not already included
	originalExists := false
	for _, v := range variantsProto {
		if v.Name == "original" {
			originalExists = true
			break
		}
	}

	if !originalExists {
		// Generate a URL for the original
		urlStr, _, err := mediaService.GetPresignedURL(ctx, media, "original", time.Hour)
		if err == nil {
			variantsProto = append(variantsProto, &mediapb.MediaVariant{
				Name: "original",
				Url:  urlStr,
			})
		}
	}

	return &mediapb.Media{
		MediaId:     media.ID,
		CharacterId: media.CharacterId,
		Filename:    media.Filename,
		ContentType: media.ContentType,
		Size:        media.Size,
		Variants:    variantsProto,
		CreatedAt:   media.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetMediaURL implements the GetMediaURL method
func (s *MediaService) GetMediaURL(ctx context.Context, req *mediapb.GetMediaURLRequest) (*mediapb.GetMediaURLResponse, error) {
	s.logger.Info("GetMediaURL called",
		zap.String("media_id", req.MediaId),
		zap.String("variant", req.Variant),
		zap.Int64("expires_in", req.ExpiresIn))

	// Create media service instance
	mediaRepo := repository.NewPostgresMediaRepository(s.db, s.minioClient)
	mediaService := service.NewMediaService(mediaRepo, s.minioClient, s.bucket, s.logger)

	// Get media from database
	media, err := mediaRepo.GetMediaByID(ctx, req.MediaId)
	if err != nil {
		s.logger.Error("Failed to get media from database", zap.Error(err))
		return nil, fmt.Errorf("failed to get media from database: %w", err)
	}

	// Generate presigned URL
	expiresIn := time.Duration(req.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = time.Hour // Default expiry
	}

	urlStr, expiresAt, err := mediaService.GetPresignedURL(ctx, media, req.Variant, expiresIn)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL", zap.Error(err))
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &mediapb.GetMediaURLResponse{
		Url:       urlStr,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

// OptimizeImage implements the OptimizeImage method
func (s *MediaService) OptimizeImage(ctx context.Context, req *mediapb.OptimizeImageRequest) (*mediapb.OptimizeImageResponse, error) {
	s.logger.Info("OptimizeImage called", zap.String("media_id", req.MediaId))

	// Create media service instance
	mediaRepo := repository.NewPostgresMediaRepository(s.db, s.minioClient)
	mediaService := service.NewMediaService(mediaRepo, s.minioClient, s.bucket, s.logger)

	// Generate variants
	variants, err := mediaService.GenerateVariants(ctx, req.MediaId, req.VariantsToCreate)
	if err != nil {
		s.logger.Error("Failed to generate variants", zap.Error(err))
		return nil, fmt.Errorf("failed to generate variants: %w", err)
	}

	// Convert variants to proto format
	variantsProto := make([]*mediapb.MediaVariant, 0, len(variants))
	for _, v := range variants {
		variantsProto = append(variantsProto, &mediapb.MediaVariant{
			Name:   v.Name,
			Url:    v.URL,
			Width:  v.Width,
			Height: v.Height,
		})
	}

	return &mediapb.OptimizeImageResponse{
		Variants: variantsProto,
	}, nil
}

// HealthCheck implements the HealthCheck method
func (s *MediaService) HealthCheck(ctx context.Context, req *mediapb.HealthCheckRequest) (*mediapb.HealthCheckResponse, error) {
	// Placeholder implementation
	s.logger.Info("HealthCheck called")

	return &mediapb.HealthCheckResponse{
		Status: mediapb.HealthCheckResponse_SERVING,
	}, nil
}

func main() {
	// Initialize logger
	if err := logger.InitProduction(); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize OpenTelemetry
	tp, err := telemetry.InitTracer(&cfg.Telemetry)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(ctx, tp); err != nil {
			logger.Logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	// Initialize database
	db, err := database.NewPostgresDB(database.PostgresConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Username: cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize MinIO client
	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		logger.Logger.Fatal("Failed to create MinIO client", zap.Error(err))
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.Minio.Bucket)
	if err != nil {
		logger.Logger.Fatal("Failed to check if bucket exists", zap.Error(err))
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.Minio.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			logger.Logger.Fatal("Failed to create bucket", zap.Error(err))
		}
		logger.Logger.Info("Created bucket", zap.String("bucket", cfg.Minio.Bucket))
	}

	// Initialize service discovery client
	discoveryClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to create service discovery client", zap.Error(err))
	}

	// Register service with Consul
	serviceID := fmt.Sprintf("%s-%s", cfg.Service.Name, cfg.Service.Host)
	err = discoveryClient.Register(serviceID, cfg.Service.Name, cfg.Service.Host, cfg.Service.Port, []string{"media", "api"})
	if err != nil {
		logger.Logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer discoveryClient.Deregister(serviceID)

	// Initialize gRPC clients for other services
	authConn, authClient, err := createAuthClient(discoveryClient)
	if err != nil {
		logger.Logger.Fatal("Failed to create auth client", zap.Error(err))
	}
	defer authConn.Close()

	// Initialize media service
	mediaService := &MediaService{
		logger:      logger.Logger,
		authClient:  authClient,
		minioClient: minioClient,
		db:          db,
		bucket:      cfg.Minio.Bucket,
	}

	// Create gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger),
			otelgrpc.UnaryServerInterceptor(),
		)),
	)

	// Register services
	mediapb.RegisterMediaServiceServer(grpcServer, mediaService)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("media.MediaService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Initialize metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := cfg.Service.Port + 10000 // Metrics on port+10000
		logger.Logger.Info("Starting metrics server", zap.Int("port", metricsPort))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			logger.Logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port))
	if err != nil {
		logger.Logger.Fatal("Failed to listen", zap.Error(err))
	}

	logger.Logger.Info("Starting media service",
		zap.String("host", cfg.Service.Host),
		zap.Int("port", cfg.Service.Port))

	// Handle graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down media service...")
	grpcServer.GracefulStop()
	logger.Logger.Info("Media service stopped")
}

func createAuthClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, authpb.AuthServiceClient, error) {
	// Get service address from Consul
	serviceAddress, err := discoveryClient.ResolveService("auth-service")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve auth service: %w", err)
	}

	// Create gRPC connection
	conn, err := grpc.Dial(
		serviceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	// Create client
	client := authpb.NewAuthServiceClient(conn)

	return conn, client, nil
}
