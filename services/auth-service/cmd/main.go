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
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/database"
	"github.com/sdshorin/generia/pkg/discovery"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
	"github.com/sdshorin/generia/services/auth-service/internal/repository"
	"github.com/sdshorin/generia/services/auth-service/internal/service"
)

func main() {
	// Initialize logger
	if err := logger.InitDevelopment(); err != nil {
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

	// Initialize service discovery client
	discoveryClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to create service discovery client", zap.Error(err))
	}

	// Register service with Consul
	serviceID := fmt.Sprintf("%s-%s-%d", cfg.Service.Name, cfg.Service.Name, cfg.Service.Port)
	err = discoveryClient.Register(serviceID, cfg.Service.Name, cfg.Service.Host, cfg.Service.Port, []string{"auth", "api"})
	if err != nil {
		logger.Logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer discoveryClient.Deregister(serviceID)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration)

	// Create gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger.Logger),
			otelgrpc.UnaryServerInterceptor(),
		)),
	)

	// Register services
	authpb.RegisterAuthServiceServer(grpcServer, authService)
	
	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("auth.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)

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

	logger.Logger.Info("Starting auth service", 
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

	logger.Logger.Info("Shutting down auth service...")
	grpcServer.GracefulStop()
	logger.Logger.Info("Auth service stopped")
}

