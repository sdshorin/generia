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
	"instagram-clone/pkg/config"
	"instagram-clone/pkg/discovery"
	"instagram-clone/pkg/logger"
	"instagram-clone/services/api-gateway/handlers"
	"instagram-clone/services/api-gateway/middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	authpb "instagram-clone/api/grpc/auth"
	cachepb "instagram-clone/api/grpc/cache"
	cdnpb "instagram-clone/api/grpc/cdn"
	feedpb "instagram-clone/api/grpc/feed"
	interactionpb "instagram-clone/api/grpc/interaction"
	mediapb "instagram-clone/api/grpc/media"
	postpb "instagram-clone/api/grpc/post"
)

// grpcClients contains all gRPC clients for interacting with microservices
type grpcClients struct {
	authClient        authpb.AuthServiceClient
	postClient        postpb.PostServiceClient
	mediaClient       mediapb.MediaServiceClient
	interactionClient interactionpb.InteractionServiceClient
	feedClient        feedpb.FeedServiceClient
	cacheClient       cachepb.CacheServiceClient
	cdnClient         cdnpb.CDNServiceClient
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

	// Initialize tracer
	tp, err := initTracer(cfg.Jaeger.Host)
	if err != nil {
		logger.Logger.Warn("Failed to initialize tracer, continuing without tracing", zap.Error(err))
		// Create a no-op tracer provider instead of failing
		tp = tracesdk.NewTracerProvider()
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()
	tracer := tp.Tracer("api-gateway")

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
	feedHandler := handlers.NewFeedHandler(clients.feedClient, tracer)

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
	router.Handle("/api/v1/posts", jwtMiddleware.RequireAuth(http.HandlerFunc(postHandler.CreatePost))).Methods("POST")
	router.Handle("/api/v1/posts/{id}", jwtMiddleware.Optional(http.HandlerFunc(postHandler.GetPost))).Methods("GET")
	router.Handle("/api/v1/feed", jwtMiddleware.Optional(http.HandlerFunc(feedHandler.GetGlobalFeed))).Methods("GET")
	router.Handle("/api/v1/users/{user_id}/posts", jwtMiddleware.Optional(http.HandlerFunc(postHandler.GetUserPosts))).Methods("GET")

	// Media routes
	router.Handle("/api/v1/media/upload", jwtMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.UploadMedia))).Methods("POST")
	router.HandleFunc("/api/v1/media/{id}", mediaHandler.GetMediaURLs).Methods("GET")

	// Interaction routes
	router.Handle("/api/v1/posts/{id}/like", jwtMiddleware.RequireAuth(http.HandlerFunc(interactionHandler.LikePost))).Methods("POST")
	router.Handle("/api/v1/posts/{id}/like", jwtMiddleware.RequireAuth(http.HandlerFunc(interactionHandler.UnlikePost))).Methods("DELETE")
	router.Handle("/api/v1/posts/{id}/comments", jwtMiddleware.RequireAuth(http.HandlerFunc(interactionHandler.AddComment))).Methods("POST")
	router.Handle("/api/v1/posts/{id}/comments", jwtMiddleware.Optional(http.HandlerFunc(interactionHandler.GetComments))).Methods("GET")
	router.Handle("/api/v1/posts/{id}/likes", jwtMiddleware.Optional(http.HandlerFunc(interactionHandler.GetLikes))).Methods("GET")

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

func initTracer(jaegerHost string) (*tracesdk.TracerProvider, error) {
	// Create Jaeger exporter with explicit port configuration
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(jaegerHost),
		jaeger.WithAgentPort("6831"), // Explicit port specification to avoid UDP connection issues
	))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("api-gateway"),
		)),
		// Add sampling configuration to reduce trace volume in production
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(0.5))),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
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

	feedAddr, err := discoveryClient.ResolveService("feed-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve feed service: %w", err)
	}

	cacheAddr, err := discoveryClient.ResolveService("cache-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache service: %w", err)
	}

	cdnAddr, err := discoveryClient.ResolveService("cdn-service")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cdn service: %w", err)
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

	feedConn, err := grpc.Dial(feedAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to feed service: %w", err)
	}

	cacheConn, err := grpc.Dial(cacheAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cache service: %w", err)
	}

	cdnConn, err := grpc.Dial(cdnAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cdn service: %w", err)
	}

	// Create clients
	return &grpcClients{
		authClient:        authpb.NewAuthServiceClient(authConn),
		postClient:        postpb.NewPostServiceClient(postConn),
		mediaClient:       mediapb.NewMediaServiceClient(mediaConn),
		interactionClient: interactionpb.NewInteractionServiceClient(interactionConn),
		feedClient:        feedpb.NewFeedServiceClient(feedConn),
		cacheClient:       cachepb.NewCacheServiceClient(cacheConn),
		cdnClient:         cdnpb.NewCDNServiceClient(cdnConn),
	}, nil
}