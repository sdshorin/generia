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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"instagram-clone/pkg/config"
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
	feedpb "instagram-clone/api/grpc/feed"
	postpb "instagram-clone/api/grpc/post"
)

// FeedService implements the feed service
type FeedService struct {
	feedpb.UnimplementedFeedServiceServer
	logger     *zap.Logger
	authClient authpb.AuthServiceClient
	postClient postpb.PostServiceClient
}

// GetGlobalFeed implements the GetGlobalFeed method
func (s *FeedService) GetGlobalFeed(ctx context.Context, req *feedpb.GetGlobalFeedRequest) (*feedpb.GetGlobalFeedResponse, error) {
	// Placeholder implementation
	s.logger.Info("GetGlobalFeed called")
	return &feedpb.GetGlobalFeedResponse{
		Posts: []*feedpb.PostInfo{
			{
				Id:        "post-id-1",
				UserId:    "user-id-1",
				Caption:   "Example post 1",
				MediaId:   "media-id-1",
				CreatedAt: time.Now().Add(-24 * time.Hour).Unix(),
				User: &feedpb.UserInfo{
					Id:       "user-id-1",
					Username: "user1",
				},
				Stats: &feedpb.PostStats{
					LikesCount:    10,
					CommentsCount: 5,
				},
			},
			{
				Id:        "post-id-2",
				UserId:    "user-id-2",
				Caption:   "Example post 2",
				MediaId:   "media-id-2",
				CreatedAt: time.Now().Add(-12 * time.Hour).Unix(),
				User: &feedpb.UserInfo{
					Id:       "user-id-2",
					Username: "user2",
				},
				Stats: &feedpb.PostStats{
					LikesCount:    5,
					CommentsCount: 2,
				},
			},
		},
		NextCursor: "",
		HasMore:    false,
	}, nil
}

// GetUserFeed implements the GetUserFeed method
func (s *FeedService) GetUserFeed(ctx context.Context, req *feedpb.GetUserFeedRequest) (*feedpb.GetUserFeedResponse, error) {
	// Placeholder implementation
	s.logger.Info("GetUserFeed called")
	return &feedpb.GetUserFeedResponse{
		Posts: []*feedpb.PostInfo{
			{
				Id:        "post-id-3",
				UserId:    "user-id-3",
				Caption:   "Example post 3",
				MediaId:   "media-id-3",
				CreatedAt: time.Now().Add(-2 * time.Hour).Unix(),
				User: &feedpb.UserInfo{
					Id:       "user-id-3",
					Username: "user3",
				},
				Stats: &feedpb.PostStats{
					LikesCount:    20,
					CommentsCount: 8,
				},
			},
		},
		NextCursor: "",
		HasMore:    false,
	}, nil
}

// InvalidateFeedCache implements the InvalidateFeedCache method
func (s *FeedService) InvalidateFeedCache(ctx context.Context, req *feedpb.InvalidateFeedCacheRequest) (*feedpb.InvalidateFeedCacheResponse, error) {
	// Placeholder implementation
	s.logger.Info("InvalidateFeedCache called", 
		zap.String("type", req.Type.String()),
		zap.String("id", req.Id))
	return &feedpb.InvalidateFeedCacheResponse{
		Success: true,
	}, nil
}

// HealthCheck implements the HealthCheck method
func (s *FeedService) HealthCheck(ctx context.Context, req *feedpb.HealthCheckRequest) (*feedpb.HealthCheckResponse, error) {
	// Placeholder implementation
	return &feedpb.HealthCheckResponse{
		Status: feedpb.HealthCheckResponse_SERVING,
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

	// Initialize service discovery client
	discoveryClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to create service discovery client", zap.Error(err))
	}

	// Register service with Consul
	serviceID := fmt.Sprintf("%s-%s", cfg.Service.Name, cfg.Service.Host)
	err = discoveryClient.Register(serviceID, cfg.Service.Name, cfg.Service.Host, cfg.Service.Port, []string{"feed", "api"})
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

	postConn, postClient, err := createPostClient(discoveryClient)
	if err != nil {
		logger.Logger.Fatal("Failed to create post client", zap.Error(err))
	}
	defer postConn.Close()

	// Initialize feed service
	feedService := &FeedService{
		logger:     logger.Logger,
		authClient: authClient,
		postClient: postClient,
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
	feedpb.RegisterFeedServiceServer(grpcServer, feedService)
	
	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("feed.FeedService", grpc_health_v1.HealthCheckResponse_SERVING)

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

	logger.Logger.Info("Starting feed service", 
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

	logger.Logger.Info("Shutting down feed service...")
	grpcServer.GracefulStop()
	logger.Logger.Info("Feed service stopped")
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
			semconv.ServiceName("feed-service"),
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

func createPostClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, postpb.PostServiceClient, error) {
	// Get service address from Consul
	serviceAddress, err := discoveryClient.ResolveService("post-service")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve post service: %w", err)
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
		return nil, nil, fmt.Errorf("failed to connect to post service: %w", err)
	}

	// Create client
	client := postpb.NewPostServiceClient(conn)

	return conn, client, nil
}