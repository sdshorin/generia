# Базы данных проекта Generia

## Обзор

Проект Generia использует несколько баз данных для хранения различных типов данных. Каждый микросервис взаимодействует со своей собственной базой данных, что обеспечивает изоляцию данных и независимость сервисов.

## PostgreSQL

PostgreSQL используется для хранения структурированных данных, таких как информация о пользователях, постах и медиафайлах.

### Таблица users (Auth Service)

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(30) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Таблица refresh_tokens (Auth Service)

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Таблица posts (Post Service)

```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    caption TEXT,
    media_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Таблица media (Media Service)

```sql
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    bucket TEXT NOT NULL,
    path TEXT NOT NULL,
    variants JSONB NOT NULL, -- { "original": "path", "thumbnail": "path", "medium": "path" }
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## MongoDB

MongoDB используется для хранения данных о взаимодействиях пользователей, таких как лайки и комментарии.

### Коллекция Likes (Interaction Service)

```json
{
  "_id": ObjectId("...."),
  "user_id": "user123",
  "post_id": "post456",
  "created_at": ISODate("2023-09-01T12:00:00Z")
}
```

### Коллекция Comments (Interaction Service)

```json
{
  "_id": ObjectId("...."),
  "user_id": "user123",
  "post_id": "post456",
  "text": "Отличное фото!",
  "created_at": ISODate("2023-09-01T12:05:00Z"),
  "updated_at": ISODate("2023-09-01T12:05:00Z")
}
```

## Redis

Redis используется для кэширования часто запрашиваемых данных и хранения временных данных, таких как лента новостей.

### Кэш для ленты новостей (Feed Service)

```
# Ключ: user:feed:{user_id}:{page}
# Значение: JSON-массив с идентификаторами постов
user:feed:user123:1 -> ["post1", "post2", "post3", ...]
```

### Кэш для счетчиков взаимодействий (Interaction Service)

```
# Ключ: post:likes:{post_id}
# Значение: количество лайков
post:likes:post456 -> 42

# Ключ: post:comments:{post_id}
# Значение: количество комментариев
post:comments:post456 -> 7
```

### Сессии пользователей (Auth Service)

```
# Ключ: session:{session_id}
# Значение: JSON с данными сессии
session:abc123 -> {"user_id": "user123", "expires_at": 1630500000}
```

## MinIO (S3-совместимое хранилище)

MinIO используется для хранения медиафайлов, таких как изображения и видео.

### Бакеты

- `profile-images` - Аватары пользователей
- `post-images` - Изображения постов
- `post-videos` - Видео постов
- `thumbnails` - Миниатюры изображений и видео

## Схема данных

### User

```json
{
  "id": "user123",
  "username": "john_doe",
  "email": "john@example.com",
  "password_hash": "$2a$10$...",
  "created_at": "2023-09-01T12:00:00Z",
  "updated_at": "2023-09-01T12:00:00Z"
}
```

### Post

```json
{
  "id": "post456",
  "user_id": "user123",
  "caption": "Beautiful sunset",
  "media_id": "media789",
  "created_at": "2023-09-01T18:30:00Z",
  "updated_at": "2023-09-01T18:30:00Z"
}
```

### Media

```json
{
  "id": "media789",
  "user_id": "user123",
  "filename": "image1.jpg",
  "content_type": "image/jpeg",
  "size": 1024000,
  "bucket": "generia-images",
  "path": "user123/posts/image1.jpg",
  "variants": {
    "original": "user123/posts/image1.jpg",
    "thumbnail": "user123/posts/thumbnails/image1.jpg",
    "medium": "user123/posts/medium/image1.jpg"
  },
  "created_at": "2023-09-01T18:30:00Z"
}
```

### Like

```json
{
  "id": "like101",
  "user_id": "user123",
  "post_id": "post456",
  "created_at": "2023-09-01T19:00:00Z"
}
```

### Comment

```json
{
  "id": "comment202",
  "user_id": "user123",
  "post_id": "post456",
  "text": "Отличное фото!",
  "created_at": "2023-09-01T19:05:00Z",
  "updated_at": "2023-09-01T19:05:00Z"
}
```

## Миграции

Миграции баз данных осуществляются с помощью инструмента golang-migrate, который обеспечивает версионирование схемы базы данных и возможность обратной миграции. Миграционные скрипты хранятся в директории `/scripts/migrations/` для каждого сервиса.

## Доступ к базам данных

Каждый микросервис имеет свой собственный репозиторий для работы с базой данных:

- Auth Service: `/services/auth-service/internal/repository/user_repository.go`
- Post Service: `/services/post-service/internal/repository/post_repository.go`
- Media Service: `/services/media-service/internal/repository/media_repository.go`
- Interaction Service: `/services/interaction-service/internal/repository/interaction_repository.go`

Репозитории инкапсулируют логику доступа к базе данных и предоставляют абстрактный интерфейс для сервисов.
