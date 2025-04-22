package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/sdshorin/generia/api/grpc/character"
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/database"
	"github.com/sdshorin/generia/pkg/discovery"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/character-service/internal/repository"
	"github.com/sdshorin/generia/services/character-service/internal/service"
)

func main() {
	// Initialize logger
	if err := logger.InitDevelopment(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Logger.Sync()

	// Load configuration from env variables
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to database
	dbConfig := database.PostgresConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Username: cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repository
	characterRepo := repository.NewCharacterRepository(db.DB)

	// Initialize service
	characterService := service.NewCharacterService(characterRepo)

	// Get port from config or environment
	// Using port 8089 as specified in docker-compose.yml
	port := 8089
	
	// Initialize gRPC server
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	
	// Register services
	pb.RegisterCharacterServiceServer(grpcServer, characterService)
	grpc_health_v1.RegisterHealthServer(grpcServer, characterService)
	reflection.Register(grpcServer)

	// Register with service discovery
	consul, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Consul", zap.Error(err))
	}

	serviceID := "character-service-1"
	err = consul.Register(serviceID, "character-service", "character-service", port, []string{"character", "service"})
	if err != nil {
		logger.Logger.Fatal("Failed to register service", zap.Error(err))
	}

	// Start server
	logger.Logger.Info("Starting character service", zap.Int("port", port))
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			logger.Logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	logger.Logger.Info("Shutting down...")

	// Deregister from service discovery
	if err := consul.Deregister(serviceID); err != nil {
		logger.Logger.Error("Failed to deregister from Consul", zap.Error(err))
	}

	// Graceful shutdown
	grpcServer.GracefulStop()
	logger.Logger.Info("Server stopped")
}