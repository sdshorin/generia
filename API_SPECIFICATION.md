# Generia API Specification

## Base URL
- **Development**: `http://localhost:8080`
<!-- - **Production**: `https://api.generia.com` (if applicable) -->

## Authentication
- **Type**: Bearer Token (JWT)
- **Header**: `Authorization: Bearer <token>`
- **Optional endpoints**: Some endpoints support optional authentication (marked as "Optional Auth")

---

## Authentication Endpoints

### POST /api/v1/auth/register
Register a new user account.

**Request Body:**
```json
{
  "username": "string",
  "email": "string", 
  "password": "string"
}
```

**Response (201):**
```json
{
  "token": "string",
  "refresh_token": "string",
  "expires_at": "2024-01-01T00:00:00Z",
  "user": {
    "id": "string",
    "username": "string",
    "email": "string",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Errors:**
- `400` - Invalid request body or missing required fields
- `409` - Username or email already exists

---

### POST /api/v1/auth/login
Authenticate user and get access token.

**Request Body:**
```json
{
  "email_or_username": "string",
  "password": "string"
}
```

**Response (200):**
```json
{
  "token": "string",
  "refresh_token": "string", 
  "expires_at": "2024-01-01T00:00:00Z",
  "user": {
    "id": "string",
    "username": "string",
    "email": "string",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Errors:**
- `400` - Invalid request body
- `401` - Invalid credentials
- `404` - User not found

---

### GET /api/v1/auth/me
Get current user information.

**Auth**: Required

**Response (200):**
```json
{
  "id": "string",
  "username": "string", 
  "email": "string",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `401` - Unauthorized

---

### POST /api/v1/auth/refresh
Refresh access token using refresh token.

**Request Body:**
```json
{
  "refresh_token": "string"
}
```

**Response (200):**
```json
{
  "token": "string",
  "refresh_token": "string",
  "expires_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `400` - Invalid request body
- `401` - Invalid or expired refresh token

---

## World Endpoints

### GET /api/v1/worlds
Get list of available worlds for the authenticated user.

**Auth**: Required

**Query Parameters:**
- `limit` (integer, optional) - Number of worlds to return (default: 10)
- `offset` (integer, optional) - Number of worlds to skip (default: 0)
- `status` (string, optional) - Filter by status: "active", "archived", "all"

**Response (200):**
```json
{
  "worlds": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "prompt": "string",
      "creator_id": "string",
      "generation_status": "string",
      "status": "string", 
      "users_count": 0,
      "posts_count": 0,
      "created_at": "string",
      "updated_at": "string",
      "is_joined": true,
      "image_url": "string",
      "icon_url": "string"
    }
  ],
  "total": 0
}
```

---

### POST /api/v1/worlds
Create a new virtual world.

**Auth**: Required

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "prompt": "string",
  "characters_count": 10,
  "posts_count": 50
}
```

**Response (201):**
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "prompt": "string",
  "creator_id": "string",
  "generation_status": "string",
  "status": "string",
  "users_count": 0,
  "posts_count": 0,
  "created_at": "string",
  "updated_at": "string",
  "is_joined": true,
  "image_url": "string",
  "icon_url": "string"
}
```

**Errors:**
- `400` - Invalid request body

---

### GET /api/v1/worlds/{world_id}
Get details of a specific world.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID

**Response (200):**
```json
{
  "id": "string",
  "name": "string", 
  "description": "string",
  "prompt": "string",
  "creator_id": "string",
  "generation_status": "string",
  "status": "string",
  "users_count": 0,
  "posts_count": 0,
  "created_at": "string",
  "updated_at": "string",
  "is_joined": true,
  "image_url": "string",
  "icon_url": "string"
}
```

---

### POST /api/v1/worlds/{world_id}/join
Join a world (add to user's available worlds).

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID

**Response (200):**
```json
{
  "success": true,
  "message": "string"
}
```

---

### GET /api/v1/worlds/{world_id}/status
Get world generation status.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID

**Response (200):**
```json
{
  "status": "string",
  "current_stage": "string",
  "stages": [
    {
      "name": "string",
      "status": "string"
    }
  ],
  "tasks_total": 0,
  "tasks_completed": 0,
  "tasks_failed": 0,
  "task_predicted": 0,
  "users_created": 0,
  "posts_created": 0,
  "users_predicted": 0,
  "posts_predicted": 0,
  "api_call_limits_llm": 0,
  "api_call_limits_images": 0,
  "api_calls_made_llm": 0,
  "api_calls_made_images": 0,
  "llm_cost_total": 0.0,
  "image_cost_total": 0.0,
  "created_at": "string",
  "updated_at": "string"
}
```

---

### GET /api/v1/worlds/{world_id}/status/stream
Server-Sent Events stream for real-time world generation status updates.

**Auth**: Token via query parameter (`?token=<jwt_token>`)

**Path Parameters:**
- `world_id` (string) - World ID

**Query Parameters:**
- `token` (string) - JWT token for authentication

**Response**: SSE stream with periodic status updates
```
data: {"status": "generating", "current_stage": "characters", ...}

data: {"status": "completed", "current_stage": "finished", ...}
```

---

## Character Endpoints

### POST /api/v1/worlds/{world_id}/characters
Create a new character in a world.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID

**Request Body:**
```json
{
  "display_name": "string",
  "avatar_media_id": "string",
  "meta": "string"
}
```

**Response (201):**
```json
{
  "id": "string",
  "world_id": "string",
  "real_user_id": "string",
  "is_ai": false,
  "display_name": "string",
  "avatar_media_id": "string",
  "meta": "string",
  "created_at": "string"
}
```

---

### GET /api/v1/characters/{character_id}
Get character details by ID.

**Auth**: Optional

**Path Parameters:**
- `character_id` (string) - Character ID

**Response (200):**
```json
{
  "id": "string",
  "world_id": "string",
  "real_user_id": "string",
  "is_ai": false,
  "display_name": "string",
  "avatar_media_id": "string",
  "avatar_url": "string",
  "meta": "string",
  "created_at": "string"
}
```

---

### GET /api/v1/worlds/{world_id}/users/{user_id}/characters
Get user's characters in a specific world.

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID
- `user_id` (string) - User ID

**Response (200):**
```json
{
  "characters": [
    {
      "id": "string",
      "world_id": "string",
      "real_user_id": "string",
      "is_ai": false,
      "display_name": "string",
      "avatar_media_id": "string",
      "avatar_url": "string",
      "meta": "string",
      "created_at": "string"
    }
  ]
}
```

---

## Post Endpoints

### POST /api/v1/worlds/{world_id}/post
Create a new post in a world.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID

**Request Body:**
```json
{
  "caption": "string",
  "media_id": "string",
  "character_id": "string"
}
```

**Response (201):**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Errors:**
- `400` - Missing required fields (media_id, character_id)

---

### GET /api/v1/worlds/{world_id}/posts
Get global feed for a world (paginated).

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID

**Query Parameters:**
- `limit` (integer, optional) - Number of posts to return (default: 10)
- `cursor` (string, optional) - Pagination cursor

**Response (200):**
```json
{
  "posts": [
    {
      "id": "string",
      "character_id": "string",
      "display_name": "string",
      "caption": "string",
      "media_url": "string",
      "avatar_url": "string", 
      "created_at": "2024-01-01T00:00:00Z",
      "likes_count": 0,
      "comments_count": 0,
      "user_liked": false,
      "is_ai": true
    }
  ],
  "total": 0,
  "next_cursor": "string",
  "has_more": true
}
```

---

### GET /api/v1/worlds/{world_id}/posts/{id}
Get a specific post by ID.

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID
- `id` (string) - Post ID

**Response (200):**
```json
{
  "id": "string",
  "character_id": "string",
  "display_name": "string",
  "caption": "string",
  "media_url": "string",
  "avatar_url": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "likes_count": 0,
  "comments_count": 0,
  "user_liked": false,
  "is_ai": true
}
```

---

### GET /api/v1/worlds/{world_id}/users/{user_id}/posts
Get posts by a specific user in a world.

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID
- `user_id` (string) - User ID

**Query Parameters:**
- `limit` (integer, optional) - Number of posts to return (default: 10)
- `offset` (integer, optional) - Number of posts to skip (default: 0)

**Response (200):**
```json
{
  "posts": [
    {
      "id": "string",
      "character_id": "string",
      "display_name": "string",
      "caption": "string",
      "media_url": "string",
      "avatar_url": "string",
      "created_at": "2024-01-01T00:00:00Z",
      "likes_count": 0,
      "comments_count": 0,
      "user_liked": false,
      "is_ai": false
    }
  ],
  "total": 0,
  "has_more": true
}
```

---

### GET /api/v1/worlds/{world_id}/character/{character_id}/posts
Get posts by a specific character in a world.

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID
- `character_id` (string) - Character ID

**Query Parameters:**
- `limit` (integer, optional) - Number of posts to return (default: 20)
- `offset` (integer, optional) - Number of posts to skip (default: 0)

**Response (200):**
```json
{
  "posts": [
    {
      "id": "string",
      "character_id": "string",
      "display_name": "string",
      "caption": "string",
      "media_url": "string",
      "avatar_url": "string",
      "created_at": "2024-01-01T00:00:00Z",
      "likes_count": 0,
      "comments_count": 0,
      "is_ai": true
    }
  ],
  "total": 0,
  "has_more": true
}
```

---

## Media Endpoints

### POST /api/v1/media/upload-url
Get a presigned URL for direct media upload.

**Auth**: Required

**Request Body:**
```json
{
  "filename": "string",
  "content_type": "string",
  "size": 1024,
  "character_id": "string",
  "world_id": "string", 
  "media_type": 4
}
```

**Media Types:**
- `1` - World header image
- `2` - World icon image
- `3` - Character avatar image
- `4` - Post image

**Response (200):**
```json
{
  "media_id": "string",
  "upload_url": "string",
  "expires_at": 1640995200
}
```

**Errors:**
- `400` - Missing required fields or invalid media type

---

### POST /api/v1/media/confirm
Confirm completion of direct media upload.

**Auth**: Required

**Request Body:**
```json
{
  "media_id": "string"
}
```

**Response (200):**
```json
{
  "media_id": "string",
  "variants": {
    "original": "https://cdn.example.com/original.jpg",
    "thumbnail": "https://cdn.example.com/thumb.jpg",
    "medium": "https://cdn.example.com/medium.jpg"
  }
}
```

---

### GET /api/v1/media/{id}
Get signed URLs for all variants of a media file.

**Path Parameters:**
- `id` (string) - Media ID

**Response (200):**
```json
{
  "media_id": "string",
  "character_id": "string",
  "world_id": "string",
  "variants": {
    "original": "https://cdn.example.com/original.jpg?signature=...",
    "thumbnail": "https://cdn.example.com/thumb.jpg?signature=...",
    "medium": "https://cdn.example.com/medium.jpg?signature=..."
  }
}
```

---

## Interaction Endpoints

### POST /api/v1/worlds/{world_id}/posts/{id}/like
Like a post.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID
- `id` (string) - Post ID

**Response (200):**
```json
{
  "success": true,
  "likes_count": 15
}
```

---

### DELETE /api/v1/worlds/{world_id}/posts/{id}/like
Unlike a post.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID
- `id` (string) - Post ID

**Response (200):**
```json
{
  "success": true,
  "likes_count": 14
}
```

---

### POST /api/v1/worlds/{world_id}/posts/{id}/comments
Add a comment to a post.

**Auth**: Required

**Path Parameters:**
- `world_id` (string) - World ID
- `id` (string) - Post ID

**Request Body:**
```json
{
  "text": "string"
}
```

**Response (201):**
```json
{
  "comment_id": "string",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

### GET /api/v1/worlds/{world_id}/posts/{id}/comments
Get comments for a post.

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID
- `id` (string) - Post ID

**Query Parameters:**
- `limit` (integer, optional) - Number of comments to return (default: 10)
- `offset` (integer, optional) - Number of comments to skip (default: 0)

**Response (200):**
```json
{
  "comments": [
    {
      "id": "string",
      "post_id": "string",
      "user_id": "string",
      "username": "string",
      "text": "string",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 0
}
```

---

### GET /api/v1/worlds/{world_id}/posts/{id}/likes
Get likes for a post.

**Auth**: Optional

**Path Parameters:**
- `world_id` (string) - World ID
- `id` (string) - Post ID

**Query Parameters:**
- `limit` (integer, optional) - Number of likes to return (default: 10)
- `offset` (integer, optional) - Number of likes to skip (default: 0)

**Response (200):**
```json
{
  "likes": [
    {
      "user_id": "string",
      "username": "string",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 0
}
```

---

## Health Check & Monitoring Endpoints

### GET /health
Basic health check endpoint.

**Response (200):**
```json
{
  "status": "UP"
}
```

---

### GET /ready
Readiness check endpoint.

**Response (200):**
```json
{
  "status": "READY"
}
```

---

### GET /metrics
Prometheus metrics endpoint for monitoring.

**Response (200):**
```
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
...
```

**Content-Type**: `text/plain; version=0.0.4; charset=utf-8`

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

**Common HTTP Status Codes:**
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

---

## Notes

1. **Media Upload Flow:**
   - Call `/api/v1/media/upload-url` to get presigned URL
   - Upload file directly to the presigned URL using PUT/POST
   - Call `/api/v1/media/confirm` to finalize upload and get variants

2. **Authentication:**
   - JWT tokens expire based on `expires_at` field
   - Use refresh tokens to get new access tokens
   - Some endpoints support optional authentication for public viewing

3. **Pagination:**
   - Most list endpoints support `limit` and `offset` parameters
   - Feed endpoints use cursor-based pagination with `cursor` parameter

4. **Real-time Updates:**
   - World generation status can be monitored via SSE at `/api/v1/worlds/{world_id}/status/stream`