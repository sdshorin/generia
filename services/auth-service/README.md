# Auth Service for Generia

The Auth Service is the central security component responsible for user authentication and authorization in the Generia platform. It manages user registration, login processes, and JWT token lifecycle.

## Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Data Models](#data-models)
  - [Repository Layer](#repository-layer)
  - [Service Layer](#service-layer)
  - [gRPC API](#grpc-api)
- [Core Functionality](#core-functionality)
  - [User Registration](#user-registration)
  - [Authentication](#authentication)
  - [Token Management](#token-management)
  - [User Information](#user-information)
- [Technical Details](#technical-details)
  - [Database Schema](#database-schema)
  - [Security Implementation](#security-implementation)
  - [Error Handling](#error-handling)
  - [Token Retry Mechanism](#token-retry-mechanism)
  - [Monitoring and Metrics](#monitoring-and-metrics)
- [Configuration and Deployment](#configuration-and-deployment)
  - [Environment Variables](#environment-variables)
  - [Running the Service](#running-the-service)
- [Usage Examples](#usage-examples)

## Overview

The Auth Service provides a robust authentication and authorization infrastructure for the Generia platform. As a central security component, it interacts with other microservices via gRPC, providing user identification and authentication services.

Key features:
- User registration with unique username and email validation
- Secure password handling with bcrypt hashing
- JWT-based authentication with access and refresh tokens
- Token validation and refresh mechanisms
- User information retrieval

## Architecture

The Auth Service follows a clean, layered architecture:

### Data Models

The service uses two primary data models:

**User**
```go
type User struct {
    ID           string    `db:"id"`           // UUID, primary key
    Username     string    `db:"username"`     // Unique username
    Email        string    `db:"email"`        // Unique email
    PasswordHash string    `db:"password_hash"` // Bcrypt hashed password
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}
```

**RefreshToken**
```go
type RefreshToken struct {
    ID        string    `db:"id"`         // UUID, primary key
    UserID    string    `db:"user_id"`    // Foreign key to users.id
    TokenHash string    `db:"token_hash"` // SHA-256 hash of the token
    ExpiresAt time.Time `db:"expires_at"`
    CreatedAt time.Time `db:"created_at"`
}
```

Reference: [internal/models/user.go](internal/models/user.go)

### Repository Layer

The repository layer handles database operations through the following interface:

```go
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetByUsername(ctx context.Context, username string) (*models.User, error)
    SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error
    GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
    DeleteRefreshToken(ctx context.Context, tokenHash string) error
}
```

This interface abstracts database operations, making the service more testable and maintainable.

Reference: [internal/repository/user_repository.go](internal/repository/user_repository.go)

### Service Layer

The service layer implements the business logic and exposes the gRPC interface:

```go
type AuthService struct {
    authpb.UnimplementedAuthServiceServer
    userRepo       repository.UserRepository
    jwtSecret      string
    jwtExpiration  time.Duration
}
```

The service is responsible for handling registration, login, token validation, and user information requests.

Reference: [internal/service/auth_service.go](internal/service/auth_service.go)

### gRPC API

The Auth Service exposes the following gRPC API:

```protobuf
service AuthService {
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
    rpc GetUserInfo(GetUserInfoRequest) returns (UserInfo);
    rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

This API is used by the API Gateway and other services to perform authentication and authorization operations.

## Core Functionality

### User Registration

The registration process:

1. Validate input data (username, email, password)
2. Check for existing users with the same email or username
3. Hash the password using bcrypt
4. Create a new user record with a UUID
5. Generate access and refresh tokens
6. Store the refresh token hash in the database
7. Return tokens and user information

Reference: [internal/service/auth_service.go:Register](internal/service/auth_service.go)

### Authentication

The authentication (login) process:

1. Accept login via email or username (automatically detected)
2. Find the user in the database
3. Verify the password against the stored hash
4. Generate access and refresh tokens
5. Store the refresh token hash in the database
6. Return tokens and user information

Reference: [internal/service/auth_service.go:Login](internal/service/auth_service.go)

### Token Management

Token handling includes:

1. **Access Tokens**:
   - Short-lived JWT tokens (duration configured via environment)
   - Signed with a secret key
   - Contain user ID, issue time, expiration, and issuer claims

2. **Refresh Tokens**:
   - Longer-lived tokens (typically 30x the access token duration)
   - Stored as SHA-256 hashes in the database
   - Used to generate new access tokens without re-authentication

3. **Token Validation**:
   - Verification of JWT signature
   - Checking token expiration
   - Confirming user existence

4. **Token Refresh**:
   - Validating the refresh token against stored hash
   - Checking for token expiration
   - Generating new access and refresh tokens
   - Deleting the old refresh token

References:
- [internal/service/auth_service.go:ValidateToken](internal/service/auth_service.go)
- [internal/service/auth_service.go:RefreshToken](internal/service/auth_service.go)

### User Information

The service provides a method to retrieve user information by ID:

1. Validate the user ID
2. Retrieve user data from the database
3. Return non-sensitive user information (excluding password hash)

Reference: [internal/service/auth_service.go:GetUserInfo](internal/service/auth_service.go)

## Technical Details

### Database Schema

The Auth Service uses PostgreSQL with the following schema:

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance optimization
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

### Security Implementation

The Auth Service implements several security measures:

1. **Password Hashing**: Passwords are hashed using bcrypt with appropriate cost factors
2. **JWT Token Security**:
   - Signed with a secret key
   - Include expiration timestamps
   - Contain minimal necessary claims
3. **Token Storage**:
   - Refresh tokens are stored as SHA-256 hashes, not plaintext
   - Access tokens are never stored server-side
4. **Database Protection**:
   - Tokens include an `ON CONFLICT DO NOTHING` clause to prevent race conditions
   - Expired token cleanup during operations
5. **Error Handling**: Avoids leaking sensitive information in error messages

Reference: [internal/service/auth_service.go](internal/service/auth_service.go)

### Error Handling

The service returns standardized gRPC error codes:

- `INVALID_ARGUMENT`: Missing or invalid input data
- `ALREADY_EXISTS`: Username or email already in use
- `NOT_FOUND`: User or token not found
- `UNAUTHENTICATED`: Invalid credentials or token
- `INTERNAL`: Server-side errors

Each error is logged with appropriate context information without exposing sensitive data.

### Token Retry Mechanism

The service implements a sophisticated retry mechanism for token operations:

1. When saving a refresh token, up to 3 attempts are made in case of database contention
2. Between retries, a new token is generated to avoid unique constraint violations
3. Exponential backoff is used between retries
4. Comprehensive error logging for each attempt

This ensures robustness in high-concurrency environments.

Reference: [internal/service/auth_service.go:Register](internal/service/auth_service.go), [internal/service/auth_service.go:Login](internal/service/auth_service.go), [internal/service/auth_service.go:RefreshToken](internal/service/auth_service.go)

### Monitoring and Metrics

The Auth Service exposes:

1. **Prometheus Metrics**: Available on a separate port (service port + 10000)
2. **gRPC Health Checks**: Standard gRPC health service for load balancers
3. **Structured Logging**: Detailed logs using Zap logger
4. **Tracing**: OpenTelemetry integration for distributed tracing

Reference: [cmd/main.go](cmd/main.go)

## Configuration and Deployment

### Environment Variables

The service requires the following environment variables:

```
# Core settings
SERVICE_NAME=auth-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=9000

# Database
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=password
DATABASE_NAME=generia_auth
DATABASE_SSL_MODE=disable

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=24h

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500

# Telemetry
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=auth-service

# Logging
LOG_LEVEL=info
```

Reference: [cmd/main.go](cmd/main.go)

### Running the Service

The Auth Service can be run as part of the Generia infrastructure using Docker Compose:

```bash
docker-compose up -d auth-service
```

For local development:

```bash
cd services/auth-service
go run cmd/main.go
```

The service registers itself with Consul for service discovery and handles graceful shutdown.

## Usage Examples

### Registration

```go
client := authpb.NewAuthServiceClient(conn)

resp, err := client.Register(ctx, &authpb.RegisterRequest{
    Username: "john_doe",
    Email:    "john@example.com",
    Password: "secure_password",
})

if err != nil {
    // Handle error
}

// Use the tokens
fmt.Printf("User registered with ID: %s\n", resp.UserId)
fmt.Printf("Access Token: %s\n", resp.AccessToken)
fmt.Printf("Refresh Token: %s\n", resp.RefreshToken)
```

### Authentication

```go
resp, err := client.Login(ctx, &authpb.LoginRequest{
    EmailOrUsername: "john@example.com", // Can also use username
    Password:        "secure_password",
})

if err != nil {
    // Handle error
}

// Store tokens
fmt.Printf("User authenticated. Token expires at: %v\n", 
    time.Unix(resp.ExpiresAt, 0))
```

### Token Validation

```go
resp, err := client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
    Token: accessToken,
})

if err != nil {
    // Handle error
}

if resp.Valid {
    fmt.Printf("Token is valid for user: %s\n", resp.UserId)
} else {
    fmt.Println("Token is invalid")
}
```

### Retrieving User Information

```go
resp, err := client.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
    UserId: userId,
})

if err != nil {
    // Handle error
}

fmt.Printf("Username: %s\n", resp.Username)
fmt.Printf("Email: %s\n", resp.Email)
```

### Refreshing Tokens

```go
resp, err := client.RefreshToken(ctx, &authpb.RefreshTokenRequest{
    RefreshToken: refreshToken,
})

if err != nil {
    // Handle error
}

// Update stored tokens
fmt.Printf("New access token: %s\n", resp.AccessToken)
fmt.Printf("New refresh token: %s\n", resp.RefreshToken)
```