package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/generia/api/grpc/character"
	"github.com/generia/pkg/config"
	"github.com/generia/pkg/database"
	"github.com/generia/pkg/discovery"
	"github.com/generia/pkg/logger"
	"github.com/generia/services/character-service/internal/repository"
	"github.com/generia/services/character-service/internal/service"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("configs")
	if err != nil {
		logger.Fatal("Failed to load config", err)
	}

	// Connect to database
	db, err := database.NewPostgresDB(cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Initialize repository
	characterRepo := repository.NewCharacterRepository(db, logger)

	// Initialize service
	characterService := service.NewCharacterService(characterRepo, logger)

	// Initialize gRPC server
	list, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	
	// Register services
	pb.RegisterCharacterServiceServer(grpcServer, characterService)
	grpc_health_v1.RegisterHealthServer(grpcServer, characterService)
	reflection.Register(grpcServer)

	// Register with service discovery
	consul, err := discovery.NewConsulClient(cfg.Consul)
	if err != nil {
		logger.Fatal("Failed to connect to Consul", err)
	}

	serviceID := "character-service-1"
	err = consul.Register(serviceID, "character-service", 50051)
	if err != nil {
		logger.Fatal("Failed to register service", err)
	}

	// Start server
	logger.Info("Starting character service on :50051")
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			logger.Fatal("Failed to serve", err)
		}
	}()

	// Wait for termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	logger.Info("Shutting down...")

	// Deregister from service discovery
	if err := consul.Deregister(serviceID); err != nil {
		logger.Error("Failed to deregister from Consul", err)
	}

	// Graceful shutdown
	grpcServer.GracefulStop()
	logger.Info("Server stopped")
}