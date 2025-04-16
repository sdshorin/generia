package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	// "strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/discovery"
	"github.com/sdshorin/generia/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"github.com/sdshorin/generia/pkg/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	cachepb "github.com/sdshorin/generia/api/grpc/cache"
)

// CacheService implements the cache service
type CacheService struct {
	cachepb.UnimplementedCacheServiceServer
	logger      *zap.Logger
	redisClient *redis.Client
}

// Get implements the Get method
func (s *CacheService) Get(ctx context.Context, req *cachepb.GetRequest) (*cachepb.GetResponse, error) {
	// Basic implementation
	val, err := s.redisClient.Get(ctx, req.Key).Result()
	if err == redis.Nil {
		return &cachepb.GetResponse{
			Exists: false,
			Value:  []byte(""),
		}, nil
	} else if err != nil {
		s.logger.Error("Failed to get key from Redis", zap.String("key", req.Key), zap.Error(err))
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	return &cachepb.GetResponse{
		Exists: true,
		Value:  []byte(val),
	}, nil
}

// Set implements the Set method
func (s *CacheService) Set(ctx context.Context, req *cachepb.SetRequest) (*cachepb.SetResponse, error) {
	// Basic implementation
	var err error
	if req.Ttl > 0 {
		err = s.redisClient.Set(ctx, req.Key, req.Value, time.Duration(req.Ttl)*time.Second).Err()
	} else {
		err = s.redisClient.Set(ctx, req.Key, req.Value, 0).Err()
	}

	if err != nil {
		s.logger.Error("Failed to set key in Redis", zap.String("key", req.Key), zap.Error(err))
		return nil, fmt.Errorf("failed to set key: %w", err)
	}

	return &cachepb.SetResponse{
		Success: true,
	}, nil
}

// Delete implements the Delete method
func (s *CacheService) Delete(ctx context.Context, req *cachepb.DeleteRequest) (*cachepb.DeleteResponse, error) {
	// Basic implementation
	result, err := s.redisClient.Del(ctx, req.Key).Result()
	if err != nil {
		s.logger.Error("Failed to delete key from Redis", zap.String("key", req.Key), zap.Error(err))
		return nil, fmt.Errorf("failed to delete key: %w", err)
	}

	return &cachepb.DeleteResponse{
		Success: result > 0,
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

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Ping Redis to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Logger.Info("Connected to Redis", zap.String("address", cfg.Redis.Address))

	// Initialize service discovery client
	discoveryClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to create service discovery client", zap.Error(err))
	}

	// Register service with Consul
	serviceID := fmt.Sprintf("%s-%s", cfg.Service.Name, cfg.Service.Host)
	err = discoveryClient.Register(serviceID, cfg.Service.Name, cfg.Service.Host, cfg.Service.Port, []string{"cache", "api"})
	if err != nil {
		logger.Logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer discoveryClient.Deregister(serviceID)

	// Initialize cache service
	cacheService := &CacheService{
		logger:      logger.Logger,
		redisClient: redisClient,
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
	cachepb.RegisterCacheServiceServer(grpcServer, cacheService)
	
	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("cache.CacheService", grpc_health_v1.HealthCheckResponse_SERVING)

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

	logger.Logger.Info("Starting cache service", 
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

	logger.Logger.Info("Shutting down cache service...")
	grpcServer.GracefulStop()
	logger.Logger.Info("Cache service stopped")
}

