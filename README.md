# Instagram Clone (Microservices Architecture)

This project is a microservices-based Instagram clone, designed as an MVP with basic functionality for users, posts, likes, and comments.

## Architecture Overview

The application is built using a microservices architecture with the following services:

1. **API Gateway** - Single entry point for client applications
2. **Auth Service** - Manages user authentication and authorization
3. **Post Service** - Handles post creation and retrieval
4. **Media Service** - Manages media uploads and processing
5. **Interaction Service** - Handles likes and comments
6. **Feed Service** - Manages user feeds
7. **Cache Service** - Handles caching of frequently accessed data
8. **CDN Service** - Manages content delivery

## Technologies Used

- **Backend:** Go 1.21+
- **API:** gRPC with Protocol Buffers
- **Databases:** 
  - PostgreSQL (Auth, Post services)
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
git clone https://github.com/yourusername/github.com/sdshorin/generia.git
cd github.com/sdshorin/generia
```

2. Start the services with Docker Compose:
```bash
docker-compose up -d
```

This will start all the required services, including infrastructure (PostgreSQL, MongoDB, Redis, etc.) and application services.

3. Access the application:
   - Frontend: http://localhost:80
   - API Gateway: http://localhost:8080
   - Consul UI: http://localhost:8500
   - Jaeger UI: http://localhost:16686
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000
   - MinIO Console: http://localhost:9001

## API Endpoints

### Auth
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/refresh` - Refresh access token

### Posts
- `POST /api/v1/posts` - Create a new post
- `GET /api/v1/posts/{id}` - Get post by ID
- `GET /api/v1/feed` - Get global feed
- `GET /api/v1/users/{user_id}/posts` - Get user's posts

### Media
- `POST /api/v1/media/upload` - Upload media
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
github.com/sdshorin/generia/
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
│   ├── post-service/  # Post management service
│   ├── media-service/ # Media management service
│   ├── interaction-service/ # Likes and comments service
│   ├── feed-service/  # Feed management service
│   ├── cache-service/ # Caching service
│   └── cdn-service/   # Content delivery service
├── scripts/           # Utility scripts
│   └── schema.sql     # Database schema
├── frontend/          # Frontend application
├── docker-compose.yml # Docker Compose configuration
└── README.md          # Project documentation
```

### Adding a New Feature

1. Define the API in Protocol Buffers
2. Generate the gRPC code
3. Implement the service
4. Update the API Gateway
5. Test and deploy

## License

This project is licensed under the MIT License - see the LICENSE file for details.