# API Gateway Service for Generia

The API Gateway service provides a unified entry point for all client applications in the Generia project. This service acts as an intermediary between the frontend application and internal microservices, handling request routing, authentication, and authorization.

## Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Request Routing](#request-routing)
  - [Middleware Components](#middleware-components)
  - [Handlers](#handlers)
- [API Endpoints](#api-endpoints)
- [Technical Details](#technical-details)
  - [gRPC Clients](#grpc-clients)
  - [Security and Protection](#security-and-protection)
  - [Error Handling](#error-handling)
  - [Monitoring and Telemetry](#monitoring-and-telemetry)
- [Configuration and Deployment](#configuration-and-deployment)
  - [Environment Variables](#environment-variables)
  - [Running the Service](#running-the-service)

## Overview

API Gateway processes HTTP requests from client applications, transforms them into gRPC calls for the appropriate microservices, and returns results to clients. The service implements a RESTful API on top of gRPC communications between services, allowing clients to work with a unified API without worrying about the internal system architecture.

Key functions:
- Routing HTTP requests to corresponding microservices
- Verification and validation of JWT tokens
- Application of CORS policy
- Request and response logging
- Error handling and standardized response formatting
- Load optimization for backend services
- Health checking and readiness probes

## Architecture

### Request Routing

API Gateway receives HTTP requests and routes them to the appropriate handlers based on the request path. Each resource type (auth, worlds, posts, etc.) has its own set of handlers that call the corresponding microservice through gRPC.

The routing is implemented using the [gorilla/mux](https://github.com/gorilla/mux) router, with routes defined in the [main.go](cmd/main.go) file.

Example route setup:
```go
// Auth routes
router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")
router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
router.Handle("/api/v1/auth/me", jwtMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me))).Methods("GET")
router.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshToken).Methods("POST")
```

### Middleware Components

API Gateway uses several middleware components to process requests:

1. **JWT Middleware** [middleware/jwt.go](middleware/jwt.go) - Checks the validity of JWT tokens and adds user information to the request context. Provides both required and optional authentication modes.

2. **CORS Middleware** [middleware/cors.go](middleware/cors.go) - Handles Cross-Origin Resource Sharing for requests from client applications.

3. **Logging Middleware** [middleware/logging.go](middleware/logging.go) - Logs all incoming requests and outgoing responses with structured information.

4. **Recovery Middleware** [middleware/recovery.go](middleware/recovery.go) - Recovers service operation in case of panic in a handler.

5. **Tracing Middleware** [middleware/tracing.go](middleware/tracing.go) - Adds request tracing to track execution across various services using OpenTelemetry.

### Handlers

Handlers in API Gateway are organized by domain areas:

1. **AuthHandler** [handlers/auth.go](handlers/auth.go) - Authentication and authorization handling.
2. **WorldHandler** [handlers/world.go](handlers/world.go) - Virtual world management.
3. **CharacterHandler** [handlers/character.go](handlers/character.go) - Working with characters in worlds.
4. **PostHandler** [handlers/post.go](handlers/post.go) - Post management.
5. **MediaHandler** [handlers/media.go](handlers/media.go) - Media content upload and retrieval.
6. **InteractionHandler** [handlers/interaction.go](handlers/interaction.go) - Processing likes and comments.
7. **FeedHandler** [handlers/feed.go](handlers/feed.go) - Retrieving post feeds.
8. **HealthHandler** [handlers/health.go](handlers/health.go) - Service availability checking.

Each handler has its own gRPC client for interacting with the corresponding microservice.

## API Endpoints

### Auth
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Authenticate a user
- `GET /api/v1/auth/me` - Get current user information (requires authentication)
- `POST /api/v1/auth/refresh` - Refresh access token

### Worlds
- `GET /api/v1/worlds` - Get list of available worlds (requires authentication)
- `POST /api/v1/worlds` - Create a new world (requires authentication)
- `GET /api/v1/worlds/{world_id}` - Get information about a world (requires authentication)
- `POST /api/v1/worlds/{world_id}/join` - Join a world (requires authentication)

### Characters
- `POST /api/v1/worlds/{world_id}/characters` - Create a new character in a world (requires authentication)
- `GET /api/v1/characters/{character_id}` - Get information about a character (optional authentication)
- `GET /api/v1/worlds/{world_id}/users/{user_id}/characters` - Get a user's characters in a world (optional authentication)

### Posts
- `POST /api/v1/worlds/{world_id}/posts` - Create a new post (requires authentication)
- `GET /api/v1/worlds/{world_id}/posts/{id}` - Get a post by ID (optional authentication)
- `GET /api/v1/worlds/{world_id}/feed` - Get post feed for a specific world (optional authentication)
- `GET /api/v1/worlds/{world_id}/users/{user_id}/posts` - Get a user's posts in a specific world (optional authentication)
- `GET /api/v1/worlds/{world_id}/character/{character_id}/posts` - Get character's posts in a specific world (optional authentication)

### Media
- `POST /api/v1/media/upload-url` - Get pre-signed URL for direct media upload (requires authentication)
- `POST /api/v1/media/confirm` - Confirm completion of a media upload (requires authentication)
- `GET /api/v1/media/{id}` - Get media URLs

### Interactions
- `POST /api/v1/worlds/{world_id}/posts/{id}/like` - Like a post (requires authentication)
- `DELETE /api/v1/worlds/{world_id}/posts/{id}/like` - Unlike a post (requires authentication)
- `POST /api/v1/worlds/{world_id}/posts/{id}/comments` - Add a comment to a post (requires authentication)
- `GET /api/v1/worlds/{world_id}/posts/{id}/comments` - Get comments for a post (optional authentication)
- `GET /api/v1/worlds/{world_id}/posts/{id}/likes` - Get likes for a post (optional authentication)

### Monitoring and Health
- `GET /metrics` - Prometheus metrics endpoint
- `GET /health` - Service availability check
- `GET /ready` - Service readiness check

## Technical Details

### gRPC Clients

The API Gateway uses gRPC clients to communicate with microservices. These clients are initialized at service startup with retry logic to handle temporary connection failures to other services.

```go
// Shared gRPC options
opts := []grpc.DialOption{
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             time.Second,
        PermitWithoutStream: true,
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
```

The service uses [Consul](https://www.consul.io/) for service discovery, automatically resolving the addresses of other microservices during initialization.

Reference: [cmd/main.go](cmd/main.go)

### Security and Protection

API Gateway implements several layers of protection:

1. **JWT Token Verification** - Protected endpoints check for JWT token presence and validity
2. **CORS Policy** - Defines which client applications are allowed to access the API
3. **Authentication Modes**:
   - Required authentication: Endpoints that need a valid JWT token
   - Optional authentication: Endpoints that work with or without authentication
4. **Resource Access Checking** - Verifies that the user has access to the requested resource

Reference: [middleware/jwt.go](middleware/jwt.go)

### Error Handling

The API Gateway implements a standardized error handling approach:

1. **Structured Error Responses** - All errors are returned in a consistent JSON format
2. **Status Code Mapping** - gRPC error codes are mapped to appropriate HTTP status codes
3. **Context-Aware Errors** - Errors include relevant context without exposing internal details
4. **Request Tracing** - Error responses include trace IDs for debugging across services
5. **Graceful Recovery** - Recovery middleware prevents service crashes on panics

### Monitoring and Telemetry

API Gateway supports the following tools for monitoring and debugging:

1. **OpenTelemetry** - Request tracing across different services with comprehensive context
2. **Prometheus** - Performance metrics collection via the `/metrics` endpoint
3. **Structured Logging** - JSON-formatted logs for analysis and debugging with contextual information
4. **Health Checks** - Endpoints for checking service availability (`/health`) and readiness (`/ready`)
5. **Circuit Breaking** - Basic circuit breaking functionality for gRPC client calls with retries

Reference: [middleware/tracing.go](middleware/tracing.go)

## Configuration and Deployment

### Environment Variables

API Gateway requires the following environment variables:

```
# Core settings
SERVICE_NAME=api-gateway
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8080

# JWT
JWT_SECRET=your_jwt_secret
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500
CONSUL_HEALTH_CHECK_INTERVAL=10s

# Tracing
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=api-gateway

# CORS
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Running the Service

API Gateway is part of the overall Generia infrastructure and is launched via docker-compose:

```bash
docker-compose up -d api-gateway
```

For local development, you can run the service separately:

```bash
cd services/api-gateway
go run cmd/main.go
```

The service will be available at http://localhost:8080

The server uses the following timeout settings:
- Read timeout: 15 seconds
- Write timeout: 15 seconds
- Idle timeout: 60 seconds

Reference: [cmd/main.go](cmd/main.go)