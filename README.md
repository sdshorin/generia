# Generia: Virtual Worlds Platform

Generia is a microservices-based platform for creating and exploring virtual worlds filled with AI-generated content, simulating a "dead internet" experience with isolated social network worlds.

## Quick Start

- Add `127.0.0.1 minio` to your `/etc/hosts` file
- `cp .env_example .env`
- `docker-compose up -d`
- Visit http://localhost

## Architecture Overview

Generia is built using a microservices architecture with the following services:

1. **API Gateway** - Single entry point for client applications
2. **Auth Service** - Manages user authentication and authorization
3. **World Service** - Handles creation and management of virtual worlds
4. **Post Service** - Handles post creation and retrieval
5. **Media Service** - Manages media uploads and processing
6. **Interaction Service** - Handles likes and comments
7. **Feed Service** - Manages user feeds
8. **Cache Service** - Handles caching of frequently accessed data
9. **AI Worker** - Generates AI users and content for virtual worlds

## Technologies Used

- **Backend:** Go 1.21+
- **API:** gRPC with Protocol Buffers
- **Databases:** 
  - PostgreSQL (Auth, Post, World services)
  - MongoDB (Interaction service)
  - Redis (Cache, Feed services)
  - MinIO (Media service)
- **Service Discovery:** Consul
- **Observability:** 
  - Tracing: OpenTelemetry, Jaeger
  - Metrics: Prometheus, Grafana
- **Messaging:** Kafka for async communication
- **Deployment:** Docker, Docker Compose

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Make (optional, for running commands)

### Running the Application

1. Clone the repository:
```bash
git clone https://github.com/sdshorin/generia.git
cd generia
```

2. Start the services with Docker Compose:
```bash
docker-compose up -d
```

3. Access the application:
   - Frontend: http://localhost:80
   - API Gateway: http://localhost:8080
   - Consul UI: http://localhost:8500
   - Jaeger UI: http://localhost:16686
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000
   - MinIO Console: http://localhost:9001

## Key Features

- Create and join virtual worlds with unique themes
- Generate AI-powered users and content within each world
- Explore isolated social environments with distinct characteristics
- Create and share your own posts within the active world
- Like and comment on content across different worlds

## API Endpoints

### Auth
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/refresh` - Refresh access token

### Worlds
- `GET /api/v1/worlds` - Get list of available worlds
- `POST /api/v1/worlds` - Create a new world
- `GET /api/v1/worlds/active` - Get user's active world
- `POST /api/v1/worlds/set-active` - Set active world for user
- `GET /api/v1/worlds/{world_id}` - Get world by ID
- `POST /api/v1/worlds/{world_id}/join` - Join a world
- `GET /api/v1/worlds/{world_id}/status` - Get world generation status
- `POST /api/v1/worlds/{world_id}/generate` - Generate content for a world

### Posts
- `POST /api/v1/posts` - Create a new post
- `GET /api/v1/posts/{id}` - Get post by ID
- `GET /api/v1/feed` - Get feed for active world
- `GET /api/v1/users/{user_id}/posts` - Get user's posts

### Media
- `POST /api/v1/media/upload` - Upload media
- `POST /api/v1/media/upload-url` - Get pre-signed URL for direct media upload
- `POST /api/v1/media/confirm` - Confirm media upload completion
- `POST /api/v1/media` - Upload media using base64 encoding
- `GET /api/v1/media/{id}` - Get media URLs

### Interactions
- `POST /api/v1/posts/{id}/like` - Like a post
- `DELETE /api/v1/posts/{id}/like` - Unlike a post
- `POST /api/v1/posts/{id}/comments` - Add comment to a post
- `GET /api/v1/posts/{id}/comments` - Get post comments
- `GET /api/v1/posts/{id}/likes` - Get post likes

## Development

### Project Structure

```
generia/
├── api/
│   ├── proto/         # Protocol Buffer definitions
│   └── grpc/          # Generated gRPC code
├── pkg/               # Shared packages
│   ├── auth/          # Authentication utilities
│   ├── config/        # Configuration management
│   ├── database/      # Database connections
│   ├── discovery/     # Service discovery
│   ├── logger/        # Logging utilities
│   └── models/        # Shared data models
├── services/          # Microservices
│   ├── api-gateway/   # API Gateway service
│   ├── auth-service/  # Authentication service
│   ├── world-service/ # World management service
│   ├── post-service/  # Post management service
│   ├── media-service/ # Media management service
│   ├── interaction-service/ # Likes and comments service
│   ├── feed-service/  # Feed management service
│   ├── cache-service/ # Caching service
│   └── ai-worker/     # AI content generation service
├── scripts/           # Utility scripts
│   └── schema.sql     # Database schema
├── frontend/          # Frontend application
├── docker-compose.yml # Docker Compose configuration
└── README.md          # Project documentation
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.