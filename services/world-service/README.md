# World Service for Generia

The World Service is a core microservice in the Generia project responsible for creating and managing virtual worlds. This service handles world creation, content generation orchestration, and user-world relationships.

## Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Data Models](#data-models)
  - [Repository Layer](#repository-layer)
  - [Service Layer](#service-layer)
  - [gRPC API](#grpc-api)
- [Core Functionality](#core-functionality)
  - [World Creation](#world-creation)
  - [World Management](#world-management)
  - [Content Generation](#content-generation)
  - [User-World Relationships](#user-world-relationships)
- [Technical Details](#technical-details)
  - [Database Schema](#database-schema)
  - [AI Worker Integration](#ai-worker-integration)
  - [Service Dependencies](#service-dependencies)
  - [Monitoring and Metrics](#monitoring-and-metrics)
- [Configuration and Deployment](#configuration-and-deployment)
  - [Environment Variables](#environment-variables)
  - [Running the Service](#running-the-service)
- [API Usage Examples](#api-usage-examples)

## Overview

The World Service serves as a central component of the Generia platform, managing the virtual worlds that form the foundation of the experience. Each world represents an isolated social environment with unique characteristics, theme, and content. This service coordinates with the AI Worker for content generation and interacts with other services for user management and content access.

Key capabilities:
- Create new virtual worlds based on user prompts
- Retrieve information about existing worlds
- Initiate and track AI content generation for worlds
- Manage user access to worlds
- Provide world statistics and information

## Architecture

The World Service follows a clean layered architecture:

### Data Models

The service uses two primary data models:

**World**
```go
type World struct {
    ID               string    `db:"id"`
    Name             string    `db:"name"`
    Description      string    `db:"description"`
    Prompt           string    `db:"prompt"`
    CreatorID        string    `db:"creator_id"`
    Status           string    `db:"status"`
    GenerationStatus string    `db:"generation_status"`
    CreatedAt        time.Time `db:"created_at"`
    UpdatedAt        time.Time `db:"updated_at"`
}
```

**UserWorld** (represents a user's access to a world)
```go
type UserWorld struct {
    ID        string    `db:"id"`
    UserID    string    `db:"user_id"`
    WorldID   string    `db:"world_id"`
    CreatedAt time.Time `db:"created_at"`
}
```

Reference: [internal/models/world.go](internal/models/world.go)

### Repository Layer

The repository layer handles database operations through the following interface:

```go
type WorldRepository interface {
    // World operations
    Create(ctx context.Context, world *models.World) error
    GetByID(ctx context.Context, id string) (*models.World, error)
    GetAll(ctx context.Context, limit, offset int, status string) ([]*models.World, int, error)
    GetByUser(ctx context.Context, userID string, limit, offset int, status string) ([]*models.World, int, error)
    UpdateStatus(ctx context.Context, id, status string) error

    // User world operations
    AddUserToWorld(ctx context.Context, userID, worldID string) error
    RemoveUserFromWorld(ctx context.Context, userID, worldID string) error
    GetUserWorlds(ctx context.Context, userID string) ([]*models.UserWorld, error)
    CheckUserWorld(ctx context.Context, userID, worldID string) (bool, error)

    
}
```

This interface abstracts database operations, making the service more testable and maintainable.

Reference: [internal/repository/world_repository.go](internal/repository/world_repository.go)

### Service Layer

The service layer implements the business logic for world management:

```go
type WorldService struct {
    worldpb.UnimplementedWorldServiceServer
    worldRepo     repository.WorldRepository
    authClient    authpb.AuthServiceClient
    postClient    postpb.PostServiceClient
    kafkaProducer *kafka.Producer
}
```

The service handles operations like world creation, retrieval, joining, and sending generation tasks to the AI Worker.

Reference: [internal/service/world_service.go](internal/service/world_service.go)

### gRPC API

The World Service exposes the following gRPC API:

```protobuf
service WorldService {
    rpc CreateWorld(CreateWorldRequest) returns (WorldResponse);
    rpc GetWorld(GetWorldRequest) returns (WorldResponse);
    rpc GetWorlds(GetWorldsRequest) returns (WorldsResponse);
    rpc JoinWorld(JoinWorldRequest) returns (JoinWorldResponse);
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

This API is used by the API Gateway and other services to interact with worlds.

## Core Functionality

### World Creation

The world creation process:

1. Validate input parameters (name, description, prompt, user ID)
2. Verify the user exists using the Auth Service
3. Create a new world record in the database
4. Automatically add the creator to the world
5. Initiate content generation by creating a task in MongoDB and sending it to Kafka
6. Return the created world information

The created world has default status values: `active` for the world status and an empty string for the generation status.

Reference: [internal/service/world_service.go:CreateWorld](internal/service/world_service.go)

### World Management

The service provides several operations for managing worlds:

1. **Get World by ID**: Retrieve detailed information about a specific world, including whether the requesting user has joined it
2. **Get Worlds**: Retrieve a paginated list of worlds a user has access to, with filtering by status
3. **Update World Status**: Change a world's status (active/archived) - currently implemented as an internal method



Reference: [internal/service/world_service.go:GetWorld](internal/service/world_service.go), [internal/service/world_service.go:GetWorlds](internal/service/world_service.go)

### Content Generation

World Service initiates AI content generation through:

1. **Task Creation**: When a world is created, a generation task is automatically created
2. **MongoDB Storage**: The task is stored in MongoDB with a status of "pending"
3. **Kafka Message**: A message is sent to the "generia-tasks" Kafka topic with task details
4. **Task Format**:
   ```json
   {
     "event_type": "task_created",
     "task_id": "unique-uuid",
     "task_type": "init_world_creation",
     "world_id": "world-uuid",
     "parameters": {
       "user_prompt": "User's world description",
       "users_count": 20,
       "posts_count": 100,
       "created_at": "ISO timestamp"
     }
   }
   ```

While the service can initiate generation, the status tracking endpoint (`GetWorldGenerationStatus`) is currently commented out in the implementation.

Reference: [internal/service/world_service.go:createInitialGenerationTasks](internal/service/world_service.go)

### User-World Relationships

The service manages the relationship between users and worlds:

1. **Join World**: 
   - Validate both user and world exist
   - Check if the user has already joined the world
   - Create a record connecting the user to the world if not
   - Return success status

2. **Check User Access**: Verify whether a user has joined a particular world

The implementation is simpler than what's described in the older documentation - it lacks the concept of "active" world that a user is currently interacting with.

Reference: [internal/service/world_service.go:JoinWorld](internal/service/world_service.go)

## Technical Details

### Database Schema

The World Service uses two databases:

1. **PostgreSQL** for storing world and user-world relationship data:

```sql
-- World table
CREATE TABLE worlds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    prompt TEXT NOT NULL,
    params JSONB,
    creator_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    generation_status VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User-World relationships table
CREATE TABLE user_worlds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, world_id)
);

-- Indexes
CREATE INDEX idx_worlds_creator_id ON worlds(creator_id);
CREATE INDEX idx_user_worlds_user_id ON user_worlds(user_id);
CREATE INDEX idx_user_worlds_world_id ON user_worlds(world_id);
```

2. **MongoDB** for storing generation tasks and status information:

```javascript
// Tasks collection
{
    _id: "task-uuid",
    world_id: "world-uuid",
    type: "init_world_creation",
    status: "pending", // pending, in_progress, completed, failed
    parameters: {
        user_prompt: "User's world description",
        users_count: 20,
        posts_count: 100,
        created_at: ISODate("...")
    },
    created_at: ISODate("..."),
    updated_at: ISODate("..."),
    attempt_count: 0,
    worker_id: null,
    result: null,
    error: null
}
```

### AI Worker Integration

The World Service integrates with the AI Worker service through:

1. **MongoDB**: Creating task documents in the `tasks` collection
2. **Kafka**: Sending task notifications to the `generia-tasks` topic

The integration follows these steps:
- Create a unique task ID
- Store task details in MongoDB
- Send a notification message to Kafka
- AI Worker listens to the Kafka topic and processes tasks asynchronously

While the documentation mentions a status feedback mechanism through the `generia-progress` topic, the current implementation doesn't include code to consume these messages.

Reference: [internal/service/world_service.go:createMongoDBTask](internal/service/world_service.go)

### Service Dependencies

The World Service relies on several other services:

1. **Auth Service**: Validates user existence during world creation and joining
2. **Post Service**: Referenced in main.go but not actively used in current implementation
3. **Consul**: For service discovery and registration
4. **MongoDB**: For storing task information
5. **Kafka**: For sending task notifications
6. **Prometheus**: For metrics collection

These dependencies are initialized in the main function and passed to service components as needed.

Reference: [cmd/main.go](cmd/main.go)

### Monitoring and Metrics

The service includes several monitoring features:

1. **Prometheus Metrics**: Exposed on a separate port (service port + 10000)
2. **gRPC Health Checks**: Standard gRPC health service for load balancers
3. **Structured Logging**: Using Zap logger with contextual information
4. **Tracing**: OpenTelemetry integration for distributed tracing

Reference: [cmd/main.go](cmd/main.go)

## Configuration and Deployment

### Environment Variables

The service requires the following environment variables:

```
# Core service settings
SERVICE_NAME=world-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=9020

# Database
DATABASE_URL=postgres://user:password@postgres:5432/generia?sslmode=disable

# MongoDB
MONGODB_URI=mongodb://admin:password@mongodb:27017

# Kafka
KAFKA_BROKERS=kafka:9092

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500

# Telemetry
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=world-service

# Logging
LOG_LEVEL=info
```

Reference: [cmd/main.go](cmd/main.go)

### Running the Service

The World Service can be run as part of the Generia infrastructure using Docker Compose:

```bash
docker-compose up -d world-service
```

For local development:

```bash
cd services/world-service
go run cmd/main.go
```

The service registers itself with Consul for service discovery and handles graceful shutdown with SIGINT and SIGTERM signals.

## API Usage Examples

### Creating a World

```go
client := worldpb.NewWorldServiceClient(conn)

resp, err := client.CreateWorld(ctx, &worldpb.CreateWorldRequest{
    Name:        "Cyberpunk Metropolis",
    Description: "A dystopian future dominated by megacorporations",
    Prompt:      "Create a cyberpunk world with neon-lit streets, advanced technology, and corporate control",
    UserId:      "user-123"
})

if err != nil {
    // Handle error
}

fmt.Printf("Created world: %s\n", resp.Id)
```

### Retrieving a World

```go
resp, err := client.GetWorld(ctx, &worldpb.GetWorldRequest{
    WorldId: "world-123",
    UserId:  "user-456"  // Optional, used to check if user has joined
})

if err != nil {
    // Handle error
}

fmt.Printf("World: %s\n", resp.Name)
fmt.Printf("Creator: %s\n", resp.CreatorId)
fmt.Printf("User has joined: %t\n", resp.IsJoined)
```

### Joining a World

```go
resp, err := client.JoinWorld(ctx, &worldpb.JoinWorldRequest{
    WorldId: "world-123",
    UserId:  "user-456"
})

if err != nil {
    // Handle error
}

fmt.Printf("Success: %t\n", resp.Success)
fmt.Printf("Message: %s\n", resp.Message)
```

### Retrieving User's Worlds

```go
resp, err := client.GetWorlds(ctx, &worldpb.GetWorldsRequest{
    UserId: "user-123",
    Limit:  10,
    Offset: 0,
    Status: "active"  // Optional filter: "active", "archived", or empty for all
})

if err != nil {
    // Handle error
}

fmt.Printf("Total worlds: %d\n", resp.Total)
for _, world := range resp.Worlds {
    fmt.Printf("World: %s (%s)\n", world.Name, world.Id)
}
```