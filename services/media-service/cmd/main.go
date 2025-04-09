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
	"instagram-clone/pkg/config"
	"instagram-clone/pkg/database"
	"instagram-clone/pkg/discovery"
	"instagram-clone/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	authpb "instagram-clone/api/grpc/auth"
	mediapb "instagram-clone/api/grpc/media"
)

// MediaService implements the media service
type MediaService struct {
	mediapb.UnimplementedMediaServiceServer
	logger      *zap.Logger
	authClient  authpb.AuthServiceClient
	minioClient *minio.Client
	db          *sqlx.DB
}

// UploadMedia implements the UploadMedia method (streaming)
func (s *MediaService) UploadMedia(stream mediapb.MediaService_UploadMediaServer) error {
	// Placeholder implementation
	s.logger.Info("UploadMedia called (stream)")
	
	// In a real implementation, we would:
	// 1. Receive first message with metadata
	// 2. Receive subsequent messages with data chunks
	// 3. Save to MinIO
	// 4. Return response with media ID and variants
	
	// Just return a placeholder response
	return stream.SendAndClose(&mediapb.UploadMediaResponse{
		MediaId: "media-id-placeholder",
		Variants: []*mediapb.MediaVariant{
			{
				Name:   "original",
				Url:    "http://placeholder/media/original.jpg",
				Width:  1080,
				Height: 1080,
			},
			{
				Name:   "thumbnail",
				Url:    "http://placeholder/media/thumbnail.jpg",
				Width:  320,
				Height: 320,
			},
		},
	})
}

// GetMedia implements the GetMedia method
func (s *MediaService) GetMedia(ctx context.Context, req *mediapb.GetMediaRequest) (*mediapb.Media, error) {
	// Placeholder implementation
	s.logger.Info("GetMedia called")
	return &mediapb.Media{
		MediaId:      req.MediaId,
		UserId:      "user-id-placeholder",
		Filename:    "placeholder.jpg",
		ContentType: "image/jpeg",
		Size:        1024,
		Variants: []*mediapb.MediaVariant{
			{
				Name:   "original",
				Url:    "http://placeholder/media/original.jpg",
				Width:  1080,
				Height: 1080,
			},
			{
				Name:   "thumbnail",
				Url:    "http://placeholder/media/thumbnail.jpg",
				Width:  320,
				Height: 320,
			},
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// GetMediaURL implements the GetMediaURL method
func (s *MediaService) GetMediaURL(ctx context.Context, req *mediapb.GetMediaURLRequest) (*mediapb.GetMediaURLResponse, error) {
	// Placeholder implementation
	s.logger.Info("GetMediaURL called", 
		zap.String("media_id", req.MediaId), 
		zap.String("variant", req.Variant),
		zap.Int64("expires_in", req.ExpiresIn))
	
	// Generate a presigned URL for the media
	// In a real implementation, we would fetch the media details from the database
	// and generate a presigned URL for the appropriate object in MinIO
	
	var url string
	switch req.Variant {
	case "thumbnail":
		url = "http://placeholder/media/thumbnail.jpg"
	case "medium":
		url = "http://placeholder/media/medium.jpg"
	case "original":
		url = "http://placeholder/media/original.jpg"
	default:
		url = "http://placeholder/media/medium.jpg"
	}
	
	return &mediapb.GetMediaURLResponse{
		Url:       url,
		ExpiresAt: time.Now().Add(time.Duration(req.ExpiresIn) * time.Second).Unix(),
	}, nil
}

// OptimizeImage implements the OptimizeImage method
func (s *MediaService) OptimizeImage(ctx context.Context, req *mediapb.OptimizeImageRequest) (*mediapb.OptimizeImageResponse, error) {
	// Placeholder implementation
	s.logger.Info("OptimizeImage called", zap.String("media_id", req.MediaId))
	
	// In a real implementation, we would fetch the original media, generate the requested variants,
	// save them to MinIO, and return the details
	
	variants := make([]*mediapb.MediaVariant, 0)
	
	for _, variantName := range req.VariantsToCreate {
		var width, height int32
		
		switch variantName {
		case "thumbnail":
			width, height = 320, 320
		case "medium":
			width, height = 768, 768
		case "large":
			width, height = 1024, 1024
		default:
			width, height = 500, 500
		}
		
		variants = append(variants, &mediapb.MediaVariant{
			Name:   variantName,
			Url:    fmt.Sprintf("http://placeholder/media/%s.jpg", variantName),
			Width:  width,
			Height: height,
		})
	}
	
	return &mediapb.OptimizeImageResponse{
		Variants: variants,
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
	tp, err := initTracer(cfg.Jaeger.Host)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
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

func initTracer(jaegerHost string) (*tracesdk.TracerProvider, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(jaegerHost)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("media-service"),
		)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
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