# API проекта Generia

## REST API (внешний интерфейс)

API Gateway предоставляет REST интерфейс для внешних клиентов.

### Auth API

#### Регистрация нового пользователя
**Запрос**: `POST /api/v1/auth/register`

**Тело запроса**:
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Успешный ответ** (200 OK):
```json
{
  "id": "user123",
  "username": "john_doe",
  "email": "john@example.com",
  "created_at": "2023-09-01T12:00:00Z"
}
```

#### Вход пользователя
**Запрос**: `POST /api/v1/auth/login`

**Тело запроса**:
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Успешный ответ** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

#### Получение информации о текущем пользователе
**Запрос**: `GET /api/v1/auth/me`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Успешный ответ** (200 OK):
```json
{
  "id": "user123",
  "username": "john_doe",
  "email": "john@example.com",
  "profile_image": "https://cdn.generia.com/images/profile/user123.jpg",
  "bio": "Enthusiastic photographer",
  "created_at": "2023-09-01T12:00:00Z"
}
```

#### Обновление токена доступа
**Запрос**: `POST /api/v1/auth/refresh`

**Тело запроса**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Успешный ответ** (200 OK):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

### Posts API

#### Создание нового поста
**Запрос**: `POST /api/v1/posts`

**Заголовки**:
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**Тело запроса**:
```
caption: "Beautiful sunset"
media: [файл изображения]
```

**Успешный ответ** (201 Created):
```json
{
  "id": "post123",
  "user_id": "user123",
  "caption": "Beautiful sunset",
  "media_urls": ["https://cdn.generia.com/images/posts/post123.jpg"],
  "created_at": "2023-09-01T18:30:00Z"
}
```

#### Получение поста по ID
**Запрос**: `GET /api/v1/posts/{id}`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Успешный ответ** (200 OK):
```json
{
  "id": "post123",
  "user": {
    "id": "user123",
    "username": "john_doe",
    "profile_image": "https://cdn.generia.com/images/profile/user123.jpg"
  },
  "caption": "Beautiful sunset",
  "media_urls": ["https://cdn.generia.com/images/posts/post123.jpg"],
  "likes_count": 42,
  "comments_count": 5,
  "created_at": "2023-09-01T18:30:00Z"
}
```

#### Получение глобальной ленты
**Запрос**: `GET /api/v1/feed`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Параметры запроса**:
```
page=1
limit=10
```

**Успешный ответ** (200 OK):
```json
{
  "posts": [
    {
      "id": "post123",
      "user": {
        "id": "user123",
        "username": "john_doe",
        "profile_image": "https://cdn.generia.com/images/profile/user123.jpg"
      },
      "caption": "Beautiful sunset",
      "media_urls": ["https://cdn.generia.com/images/posts/post123.jpg"],
      "likes_count": 42,
      "comments_count": 5,
      "created_at": "2023-09-01T18:30:00Z"
    },
    // ... другие посты
  ],
  "page": 1,
  "limit": 10,
  "total": 128
}
```

#### Получение постов пользователя
**Запрос**: `GET /api/v1/users/{user_id}/posts`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Параметры запроса**:
```
page=1
limit=10
```

**Успешный ответ** (200 OK):
```json
{
  "posts": [
    {
      "id": "post123",
      "caption": "Beautiful sunset",
      "media_urls": ["https://cdn.generia.com/images/posts/post123.jpg"],
      "likes_count": 42,
      "comments_count": 5,
      "created_at": "2023-09-01T18:30:00Z"
    },
    // ... другие посты
  ],
  "page": 1,
  "limit": 10,
  "total": 25
}
```

### Media API

#### Загрузка медиафайла
**Запрос**: `POST /api/v1/media/upload`

**Заголовки**:
```
Authorization: Bearer <access_token>
Content-Type: multipart/form-data
```

**Тело запроса**:
```
file: [файл изображения]
type: "post_image"
```

**Успешный ответ** (201 Created):
```json
{
  "id": "media123",
  "url": "https://cdn.generia.com/images/uploads/media123.jpg",
  "thumbnail_url": "https://cdn.generia.com/images/uploads/thumbnails/media123.jpg",
  "type": "post_image",
  "created_at": "2023-09-01T18:30:00Z"
}
```

#### Получение URL медиафайла
**Запрос**: `GET /api/v1/media/{id}`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Успешный ответ** (200 OK):
```json
{
  "id": "media123",
  "url": "https://cdn.generia.com/images/uploads/media123.jpg",
  "thumbnail_url": "https://cdn.generia.com/images/uploads/thumbnails/media123.jpg",
  "type": "post_image",
  "created_at": "2023-09-01T18:30:00Z"
}
```

### Interactions API

#### Лайк поста
**Запрос**: `POST /api/v1/posts/{id}/like`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Успешный ответ** (200 OK):
```json
{
  "success": true,
  "likes_count": 43
}
```

#### Отмена лайка поста
**Запрос**: `DELETE /api/v1/posts/{id}/like`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Успешный ответ** (200 OK):
```json
{
  "success": true,
  "likes_count": 42
}
```

#### Добавление комментария к посту
**Запрос**: `POST /api/v1/posts/{id}/comments`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Тело запроса**:
```json
{
  "text": "Great photo!"
}
```

**Успешный ответ** (201 Created):
```json
{
  "id": "comment123",
  "post_id": "post123",
  "user": {
    "id": "user456",
    "username": "jane_smith",
    "profile_image": "https://cdn.generia.com/images/profile/user456.jpg"
  },
  "text": "Great photo!",
  "created_at": "2023-09-02T10:15:00Z"
}
```

#### Получение комментариев к посту
**Запрос**: `GET /api/v1/posts/{id}/comments`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Параметры запроса**:
```
page=1
limit=10
```

**Успешный ответ** (200 OK):
```json
{
  "comments": [
    {
      "id": "comment123",
      "post_id": "post123",
      "user": {
        "id": "user456",
        "username": "jane_smith",
        "profile_image": "https://cdn.generia.com/images/profile/user456.jpg"
      },
      "text": "Great photo!",
      "created_at": "2023-09-02T10:15:00Z"
    },
    // ... другие комментарии
  ],
  "page": 1,
  "limit": 10,
  "total": 5
}
```

#### Получение лайков поста
**Запрос**: `GET /api/v1/posts/{id}/likes`

**Заголовки**:
```
Authorization: Bearer <access_token>
```

**Параметры запроса**:
```
page=1
limit=10
```

**Успешный ответ** (200 OK):
```json
{
  "likes": [
    {
      "user": {
        "id": "user456",
        "username": "jane_smith",
        "profile_image": "https://cdn.generia.com/images/profile/user456.jpg"
      },
      "created_at": "2023-09-02T09:30:00Z"
    },
    // ... другие лайки
  ],
  "page": 1,
  "limit": 10,
  "total": 42
}
```

## gRPC API (внутренний интерфейс)

Сервисы взаимодействуют между собой с использованием gRPC и Protocol Buffers.

### Auth Service

```protobuf
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUserInfo(GetUserInfoRequest) returns (GetUserInfoResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Post Service

```protobuf
service PostService {
  rpc CreatePost(CreatePostRequest) returns (CreatePostResponse);
  rpc GetPost(GetPostRequest) returns (GetPostResponse);
  rpc GetUserPosts(GetUserPostsRequest) returns (GetUserPostsResponse);
  rpc GetPostsByIds(GetPostsByIdsRequest) returns (GetPostsByIdsResponse);
  rpc GetGlobalFeed(GetGlobalFeedRequest) returns (GetGlobalFeedResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Media Service

```protobuf
service MediaService {
  rpc UploadMedia(stream UploadMediaRequest) returns (UploadMediaResponse);
  rpc GetMedia(GetMediaRequest) returns (Media);
  rpc GetMediaURL(GetMediaURLRequest) returns (GetMediaURLResponse);
  rpc OptimizeImage(OptimizeImageRequest) returns (OptimizeImageResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Interaction Service

```protobuf
service InteractionService {
  rpc LikePost(LikePostRequest) returns (LikePostResponse);
  rpc UnlikePost(UnlikePostRequest) returns (UnlikePostResponse);
  rpc CommentPost(CommentPostRequest) returns (CommentPostResponse);
  rpc GetPostComments(GetPostCommentsRequest) returns (GetPostCommentsResponse);
  rpc GetPostLikes(GetPostLikesRequest) returns (GetPostLikesResponse);
  rpc GetPostInteractionCounts(GetPostInteractionCountsRequest) returns (GetPostInteractionCountsResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Feed Service

```protobuf
service FeedService {
  rpc GetUserFeed(GetUserFeedRequest) returns (GetUserFeedResponse);
  rpc GetGlobalFeed(GetGlobalFeedRequest) returns (GetGlobalFeedResponse);
  rpc GetExploreContent(GetExploreContentRequest) returns (GetExploreContentResponse);
  rpc InvalidateUserFeedCache(InvalidateUserFeedCacheRequest) returns (InvalidateUserFeedCacheResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### Cache Service

```protobuf
service CacheService {
  rpc Set(SetRequest) returns (SetResponse);
  rpc Get(GetRequest) returns (GetResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  rpc Exists(ExistsRequest) returns (ExistsResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

### CDN Service

```protobuf
service CDNService {
  rpc GenerateSignedUrl(GenerateSignedUrlRequest) returns (GenerateSignedUrlResponse);
  rpc GetCDNInfo(GetCDNInfoRequest) returns (GetCDNInfoResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```
