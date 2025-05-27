package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sdshorin/generia/pkg/config"
	"github.com/sdshorin/generia/pkg/discovery"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/pkg/telemetry"
	"github.com/sdshorin/generia/services/api-gateway/handlers"
	"github.com/sdshorin/generia/services/api-gateway/middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	authpb "github.com/sdshorin/generia/api/grpc/auth"
	cachepb "github.com/sdshorin/generia/api/grpc/cache"
	cdnpb "github.com/sdshorin/generia/api/grpc/cdn"
	characterpb "github.com/sdshorin/generia/api/grpc/character"
	interactionpb "github.com/sdshorin/generia/api/grpc/interaction"
	mediapb "github.com/sdshorin/generia/api/grpc/media"
	postpb "github.com/sdshorin/generia/api/grpc/post"
	worldpb "github.com/sdshorin/generia/api/grpc/world"
)

// grpcClients contains all gRPC clients for interacting with microservices
type grpcClients struct {
	authClient        authpb.AuthServiceClient
	postClient        postpb.PostServiceClient
	mediaClient       mediapb.MediaServiceClient
	interactionClient interactionpb.InteractionServiceClient
	cacheClient       cachepb.CacheServiceClient
	cdnClient         cdnpb.CDNServiceClient
	worldClient       worldpb.WorldServiceClient
	characterClient   characterpb.CharacterServiceClient
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

	// Initialize OpenTelemetry tracer
	tp, err := telemetry.InitTracer(&cfg.Telemetry)
	if err != nil {
		logger.Logger.Warn("Failed to initialize tracer, continuing without tracing", zap.Error(err))
		// Create a no-op tracer provider instead of failing
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(ctx, tp); err != nil {
			logger.Logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()
	tracer := otel.Tracer("api-gateway")

	// Initialize service discovery client
	discoveryClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		logger.Logger.Fatal("Failed to create service discovery client", zap.Error(err))
	}

	// Initialize gRPC clients
	var clients *grpcClients
	// Retry a few times if gRPC initialization fails
	for attempt := 1; attempt <= 5; attempt++ {
		clients, err = initGrpcClients(discoveryClient)
		if err == nil {
			break
		}

		logger.Logger.Warn("Failed to initialize gRPC clients, retrying...",
			zap.Error(err),
			zap.Int("attempt", attempt),
			zap.Int("maxAttempts", 5))

		// Wait before retrying
		time.Sleep(5 * time.Second)
	}

	if clients == nil {
		logger.Logger.Fatal("Failed to initialize gRPC clients after multiple attempts", zap.Error(err))
	}

	// Initialize JWT middleware
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWT.Secret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(clients.authClient, tracer)
	postHandler := handlers.NewPostHandler(clients.postClient, clients.mediaClient, clients.interactionClient, tracer)
	mediaHandler := handlers.NewMediaHandler(clients.mediaClient, clients.cdnClient, tracer)
	interactionHandler := handlers.NewInteractionHandler(clients.interactionClient, tracer)

	worldHandler := handlers.NewWorldHandler(clients.worldClient, 30*time.Second, cfg.JWT.Secret)
	characterHandler := handlers.NewCharacterHandler(clients.characterClient, 30*time.Second)

	// Initialize router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.TracingMiddleware(tracer))
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.CORSMiddleware)

	// Health and metrics endpoints
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/health", handlers.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/ready", handlers.ReadinessCheckHandler).Methods("GET")

	// Auth routes
	router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
	router.Handle("/api/v1/auth/me", jwtMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me))).Methods("GET")
	router.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken).Methods("POST")

	// Post routes
	router.Handle("/api/v1/worlds/{world_id}/post", jwtMiddleware.RequireAuth(http.HandlerFunc(postHandler.CreatePost))).Methods("POST")
	router.Handle("/api/v1/worlds/{world_id}/posts", jwtMiddleware.Optional(http.HandlerFunc(postHandler.GetGlobalPosts))).Methods("GET")
	router.Handle("/api/v1/worlds/{world_id}/posts/{id}", jwtMiddleware.Optional(http.HandlerFunc(postHandler.GetPost))).Methods("GET")
	router.Handle("/api/v1/worlds/{world_id}/users/{user_id}/posts", jwtMiddleware.Optional(http.HandlerFunc(postHandler.GetUserPosts))).Methods("GET")
	router.Handle("/api/v1/worlds/{world_id}/character/{character_id}/posts", jwtMiddleware.Optional(http.HandlerFunc(postHandler.GetCharacterPosts))).Methods("GET")

	// Media routes - Legacy and Direct Upload
	router.Handle("/api/v1/media/upload-url", jwtMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.GetUploadURL))).Methods("POST")
	router.Handle("/api/v1/media/confirm", jwtMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.ConfirmUpload))).Methods("POST")
	router.HandleFunc("/api/v1/media/{id}", mediaHandler.GetMediaURLs).Methods("GET")

	// Interaction routes
	router.Handle("/api/v1/worlds/{world_id}/posts/{id}/like", jwtMiddleware.RequireAuth(http.HandlerFunc(interactionHandler.LikePost))).Methods("POST")
	router.Handle("/api/v1/worlds/{world_id}/posts/{id}/like", jwtMiddleware.RequireAuth(http.HandlerFunc(interactionHandler.UnlikePost))).Methods("DELETE")
	router.Handle("/api/v1/worlds/{world_id}/posts/{id}/comments", jwtMiddleware.RequireAuth(http.HandlerFunc(interactionHandler.AddComment))).Methods("POST")
	router.Handle("/api/v1/worlds/{world_id}/posts/{id}/comments", jwtMiddleware.Optional(http.HandlerFunc(interactionHandler.GetComments))).Methods("GET")
	router.Handle("/api/v1/worlds/{world_id}/posts/{id}/likes", jwtMiddleware.Optional(http.HandlerFunc(interactionHandler.GetLikes))).Methods("GET")

	// World routes - сначала конкретные маршруты, затем маршруты с параметрами
	router.Handle("/api/v1/worlds", jwtMiddleware.RequireAuth(http.HandlerFunc(worldHandler.GetWorlds))).Methods("GET")
	router.Handle("/api/v1/worlds", jwtMiddleware.RequireAuth(http.HandlerFunc(worldHandler.CreateWorld))).Methods("POST")
	router.Handle("/api/v1/worlds/{world_id}/join", jwtMiddleware.RequireAuth(http.HandlerFunc(worldHandler.JoinWorld))).Methods("POST")
	router.Handle("/api/v1/worlds/{world_id}/status", jwtMiddleware.RequireAuth(http.HandlerFunc(worldHandler.GetWorldStatus))).Methods("GET")
	router.HandleFunc("/api/v1/worlds/{world_id}/status/stream", worldHandler.StreamWorldStatus).Methods("GET")
	router.Handle("/api/v1/worlds/{world_id}", jwtMiddleware.RequireAuth(http.HandlerFunc(worldHandler.GetWorld))).Methods("GET")

	// Character routes
	router.Handle("/api/v1/worlds/{world_id}/characters", jwtMiddleware.RequireAuth(http.HandlerFunc(characterHandler.CreateCharacter))).Methods("POST")
	router.Handle("/api/v1/characters/{character_id}", jwtMiddleware.Optional(http.HandlerFunc(characterHandler.GetCharacter))).Methods("GET")
	router.Handle("/api/v1/worlds/{world_id}/users/{user_id}/characters", jwtMiddleware.Optional(http.HandlerFunc(characterHandler.GetUserCharactersInWorld))).Methods("GET")

	// Configure server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Service.Host, cfg.Service.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Logger.Info("Starting API Gateway", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Logger.Info("Server gracefully stopped")
}

func initGrpcClients(discoveryClient discovery.ServiceDiscovery) (*grpcClients, error) {
	// Shared gRPC options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				grpc_prometheus.UnaryClientInterceptor,
				grpc_zap.UnaryClientInterceptor(logger.Logger),
				otelgrpc.UnaryClientInterceptor(),
				grpc_retry.UnaryClientInterceptor(
					grpc_retry.WithMax(3),
					grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100*time.Millisecond)),
				),
			),
		),
	}

	// Resolve service addresses
	authAddr, err := discoveryClient.ResolveService("auth-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve auth service: %w", err)
	}

	postAddr, err := discoveryClient.ResolveService("post-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve post service: %w", err)
	}

	mediaAddr, err := discoveryClient.ResolveService("media-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve media service: %w", err)
	}

	interactionAddr, err := discoveryClient.ResolveService("interaction-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve interaction service: %w", err)
	}

	cacheAddr, err := discoveryClient.ResolveService("cache-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache service: %w", err)
	}

	cdnAddr, err := discoveryClient.ResolveService("cdn-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cdn service: %w", err)
	}

	worldAddr, err := discoveryClient.ResolveService("world-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve world service: %w", err)
	}

	characterAddr, err := discoveryClient.ResolveService("character-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve character service: %w", err)
	}

	// Create connections
	authConn, err := grpc.Dial(authAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	postConn, err := grpc.Dial(postAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to post service: %w", err)
	}

	mediaConn, err := grpc.Dial(mediaAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to media service: %w", err)
	}

	interactionConn, err := grpc.Dial(interactionAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to interaction service: %w", err)
	}

	cacheConn, err := grpc.Dial(cacheAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cache service: %w", err)
	}

	cdnConn, err := grpc.Dial(cdnAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cdn service: %w", err)
	}

	worldConn, err := grpc.Dial(worldAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to world service: %w", err)
	}

	characterConn, err := grpc.Dial(characterAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to character service: %w", err)
	}

	// Create clients
	return &grpcClients{
		authClient:        authpb.NewAuthServiceClient(authConn),
		postClient:        postpb.NewPostServiceClient(postConn),
		mediaClient:       mediapb.NewMediaServiceClient(mediaConn),
		interactionClient: interactionpb.NewInteractionServiceClient(interactionConn),
		cacheClient:       cachepb.NewCacheServiceClient(cacheConn),
		cdnClient:         cdnpb.NewCDNServiceClient(cdnConn),
		worldClient:       worldpb.NewWorldServiceClient(worldConn),
		characterClient:   characterpb.NewCharacterServiceClient(characterConn),
	}, nil
}
