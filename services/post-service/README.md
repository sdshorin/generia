# Post Service для Generia

Post Service - микросервис, отвечающий за управление постами в платформе Generia. Сервис обеспечивает создание, хранение и получение постов внутри виртуальных миров, а также формирование лент контента для пользователей.

## Оглавление

- [Обзор](#обзор)
- [Архитектура](#архитектура)
  - [Модели данных](#модели-данных)
  - [Репозитории](#репозитории)
  - [Сервисные слои](#сервисные-слои)
  - [gRPC API](#grpc-api)
- [Функциональность](#функциональность)
  - [Создание постов](#создание-постов)
  - [Получение постов](#получение-постов)
  - [Глобальная лента](#глобальная-лента)
  - [AI-посты](#ai-посты)
- [Технические детали](#технические-детали)
  - [База данных](#база-данных)
  - [Интеграция с другими сервисами](#интеграция-с-другими-сервисами)
  - [Пагинация](#пагинация)
- [Настройка и запуск](#настройка-и-запуск)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск сервиса](#запуск-сервиса)
- [Примеры использования](#примеры-использования)

## Обзор

Post Service является ключевым компонентом платформы Generia, обеспечивающим функциональность социальной сети внутри виртуальных миров. Сервис управляет постами, создаваемыми как реальными пользователями, так и AI-персонажами, формирует ленты контента для пользователей и обеспечивает взаимодействие с другими сервисами для отображения полной информации о постах.

Основные возможности:
- Создание новых постов с текстом и медиа-контентом
- Получение постов по ID, по автору или персонажу
- Формирование глобальной ленты постов мира с пагинацией
- Поддержка AI-сгенерированных постов через специальный API
- Интеграция с другими сервисами для получения дополнительной информации (данные персонажей, медиа URL, статистика взаимодействий)

## Архитектура

Post Service следует трехслойной архитектуре, характерной для микросервисов в проекте Generia:

### Модели данных

Основная модель данных - структура `Post`:

```go
// Файл: services/post-service/internal/models/post.go
type Post struct {
    ID           string    `db:"id"`
    CharacterID  string    `db:"character_id"`
    IsAI         bool      `db:"is_ai"`
    WorldID      string    `db:"world_id"`
    Caption      string    `db:"caption"`
    MediaID      string    `db:"media_id"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
    DisplayName  string    // Не хранится в БД, заполняется из Character service
    MediaURL     string    // Не хранится в БД, заполняется из Media service
    LikesCount   int32     // Не хранится в БД, заполняется из Interaction service
    CommentsCount int32    // Не хранится в БД, заполняется из Interaction service
}
```

### Репозитории

Слой репозиториев отвечает за взаимодействие с базой данных:

```go
// Файл: services/post-service/internal/repository/post_repository.go
type PostRepository interface {
    Create(ctx context.Context, post *models.Post) error
    GetByID(ctx context.Context, id string) (*models.Post, error)
    GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Post, int, error)
    GetByCharacterID(ctx context.Context, characterID string, limit, offset int) ([]*models.Post, int, error)
    GetGlobalFeed(ctx context.Context, limit int, cursor string, worldID string) ([]*models.Post, string, error)
    GetByIDs(ctx context.Context, ids []string) ([]*models.Post, error)
}
```

Репозиторий предоставляет методы для:
- Создания новых постов
- Получения поста по ID
- Получения постов пользователя
- Получения постов персонажа
- Получения глобальной ленты с пагинацией
- Пакетного получения постов по массиву ID

### Сервисные слои

Сервисный слой реализует бизнес-логику и gRPC API:

```go
// Файл: services/post-service/internal/service/post_service.go
type PostService struct {
    postpb.UnimplementedPostServiceServer
    postRepo          repository.PostRepository
    authClient        authpb.AuthServiceClient
    mediaClient       mediapb.MediaServiceClient
    interactionClient interactionpb.InteractionServiceClient
    characterClient   characterpb.CharacterServiceClient
}
```

Сервисный слой:
- Обрабатывает gRPC-запросы
- Вызывает соответствующие методы репозитория
- Дополняет данные из других сервисов (Character, Media, Interaction)
- Формирует полные ответы для клиентов

### gRPC API

Post Service предоставляет следующий gRPC-интерфейс:

```protobuf
// Файл: api/proto/post/post.proto
service PostService {
  // Создание поста
  rpc CreatePost(CreatePostRequest) returns (CreatePostResponse);
  
  // Создание AI поста (внутренний метод для AI генератора)
  rpc CreateAIPost(CreateAIPostRequest) returns (CreatePostResponse);
  
  // Получение поста по ID
  rpc GetPost(GetPostRequest) returns (Post);
  
  // Получение постов пользователя
  rpc GetUserPosts(GetUserPostsRequest) returns (PostList);
  
  // Получение постов по character_id
  rpc GetCharacterPosts(GetCharacterPostsRequest) returns (PostList);
  
  // Получение нескольких постов по ID
  rpc GetPostsByIds(GetPostsByIdsRequest) returns (PostList);

  // Получение постов для глобальной ленты
  rpc GetGlobalFeed(GetGlobalFeedRequest) returns (PostList);

  // Проверка здоровья сервиса
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

## Функциональность

### Создание постов

Процесс создания нового поста:

1. **Валидация входных данных**
   - Проверка обязательных полей (user_id, character_id, media_id, world_id)
   - Проверка принадлежности персонажа пользователю через Character Service
   - Проверка принадлежности медиа персонажу через Media Service

2. **Запись в базу данных**
   - Формирование объекта Post
   - Сохранение поста в базе данных

3. **Возврат информации**
   - Возврат ID созданного поста и времени создания

```go
// Файл: services/post-service/internal/service/post_service.go
func (s *PostService) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*postpb.CreatePostResponse, error) {
    // Валидация входных данных и проверка прав доступа
    // ...

    // Создание поста
    post := &models.Post{
        CharacterID: req.CharacterId,
        IsAI:        req.IsAi,
        WorldID:     req.WorldId,
        Caption:     req.Caption,
        MediaID:     req.MediaId,
    }

    err = s.postRepo.Create(ctx, post)
    // ...

    return &postpb.CreatePostResponse{
        PostId:    post.ID,
        CreatedAt: post.CreatedAt.Format(time.RFC3339),
    }, nil
}
```

### Получение постов

Post Service предоставляет несколько методов для получения постов:

1. **Получение поста по ID**
   ```go
   // Файл: services/post-service/internal/service/post_service.go
   func (s *PostService) GetPost(ctx context.Context, req *postpb.GetPostRequest) (*postpb.Post, error)
   ```
   - Получение базовой информации о посте из БД
   - Дополнение данными о персонаже из Character Service
   - Получение URL медиа из Media Service
   - Получение статистики взаимодействий (лайки/комментарии) из Interaction Service

2. **Получение постов пользователя**
   ```go
   // Файл: services/post-service/internal/service/post_service.go
   func (s *PostService) GetUserPosts(ctx context.Context, req *postpb.GetUserPostsRequest) (*postpb.PostList, error)
   ```
   - Получение постов всех персонажей пользователя с пагинацией
   - Обогащение данных информацией из других сервисов

3. **Получение постов персонажа**
   ```go
   // Файл: services/post-service/internal/service/post_service.go
   func (s *PostService) GetCharacterPosts(ctx context.Context, req *postpb.GetCharacterPostsRequest) (*postpb.PostList, error) 
   ```
   - Получение постов конкретного персонажа с пагинацией
   - Обогащение данных информацией из других сервисов

4. **Пакетное получение постов по ID**
   ```go
   // Файл: services/post-service/internal/service/post_service.go
   func (s *PostService) GetPostsByIds(ctx context.Context, req *postpb.GetPostsByIdsRequest) (*postpb.PostList, error)
   ```
   - Эффективное получение множества постов за один запрос
   - Пакетное получение дополнительных данных из других сервисов

### Глобальная лента

Глобальная лента постов - ключевая функциональность для отображения контента в мире:

```go
// Файл: services/post-service/internal/service/post_service.go
func (s *PostService) GetGlobalFeed(ctx context.Context, req *postpb.GetGlobalFeedRequest) (*postpb.PostList, error)
```

Особенности реализации:
- Курсор-пагинация для эффективной загрузки ленты
- Фильтрация постов по конкретному миру
- Сортировка от новых к старым
- Обогащение данных информацией из Character, Media и Interaction сервисов

### AI-посты

Post Service поддерживает создание постов от имени AI-персонажей через специальный API:

```go
// Файл: services/post-service/internal/service/post_service.go
func (s *PostService) CreateAIPost(ctx context.Context, req *postpb.CreateAIPostRequest) (*postpb.CreatePostResponse, error)
```

Особенности:
- Не требует аутентификации пользователя (вызывается из AI Worker)
- Проверяет, что персонаж действительно является AI
- Валидирует связь медиа с персонажем
- Устанавливает флаг `isAI = true` для отслеживания AI-сгенерированного контента

## Технические детали

### База данных

Post Service использует PostgreSQL для хранения данных о постах:

```sql
-- Файл: scripts/schema.sql
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID REFERENCES world_user_characters(id) ON DELETE SET NULL,
    is_ai BOOLEAN NOT NULL DEFAULT FALSE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    caption TEXT,
    media_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_posts_character_id ON posts(character_id);
CREATE INDEX IF NOT EXISTS idx_posts_world_id ON posts(world_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
CREATE INDEX IF NOT EXISTS idx_posts_is_ai ON posts(is_ai);
```

Ключевые особенности:
- Использование UUID для идентификаторов
- Внешние ключи на таблицы персонажей и миров
- Оптимизация запросов через индексы
- Поле `is_ai` для разделения обычных и AI-сгенерированных постов
- Поле `caption` для текстового содержимого поста
- Связь с медиа через media_id

### Интеграция с другими сервисами

Post Service активно взаимодействует с другими микросервисами:

1. **Character Service**
   ```go
   // Файл: services/post-service/cmd/main.go
   func createCharacterClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, characterpb.CharacterServiceClient, error)
   ```
   - Получение информации о персонажах (имя, аватар)
   - Проверка принадлежности персонажа пользователю
   - Проверка существования персонажа
   - Валидация AI-персонажей

2. **Media Service**
   ```go
   // Файл: services/post-service/cmd/main.go
   func createMediaClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, mediapb.MediaServiceClient, error)
   ```
   - Получение URL медиа-контента
   - Проверка существования медиа
   - Валидация принадлежности медиа персонажу

3. **Interaction Service**
   ```go
   // Файл: services/post-service/cmd/main.go
   func createInteractionClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, interactionpb.InteractionServiceClient, error)
   ```
   - Получение статистики взаимодействий (лайки, комментарии)
   - Эффективное получение статистики для множества постов

4. **Auth Service**
   ```go
   // Файл: services/post-service/cmd/main.go
   func createAuthClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, authpb.AuthServiceClient, error)
   ```
   - Проверка аутентификации и авторизации пользователей

Все взаимодействия осуществляются через gRPC с использованием Service Discovery (Consul) для определения адресов сервисов.

### Пагинация

Post Service использует два подхода к пагинации:

1. **Offset-пагинация** (для персональных лент и списков)
   ```go
   // Файл: services/post-service/internal/repository/post_repository.go
   func (r *postRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Post, int, error)
   ```
   - Использование LIMIT и OFFSET в SQL-запросах
   - Возврат общего количества записей
   - Подходит для относительно стабильных данных

2. **Курсор-пагинация** (для глобальной ленты)
   ```go
   // Файл: services/post-service/internal/repository/post_repository.go
   func (r *postRepository) GetGlobalFeed(ctx context.Context, limit int, cursor string, worldID string) ([]*models.Post, string, error)
   ```
   - Использование ID последнего поста как курсора
   - Не требует подсчета общего количества записей
   - Более эффективна для больших наборов данных
   - Стабильна при добавлении новых записей

## Настройка и запуск

### Переменные окружения

Для работы Post Service требуется настроить следующие переменные окружения:

```
# Основные настройки сервиса
SERVICE_NAME=post-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8083

# База данных
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=generia
POSTGRES_SSL_MODE=disable

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500

# Телеметрия
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=post-service

# Логирование
LOG_LEVEL=info
LOG_FORMAT=json
```

### Запуск сервиса

Post Service запускается как часть общей инфраструктуры Generia через docker-compose:

```bash
# Файл: docker-compose.yml
docker-compose up -d post-service
```

Для локальной разработки сервис можно запустить отдельно:

```bash
cd services/post-service
go run cmd/main.go
```

## Примеры использования

### Создание поста

```go
// gRPC-клиент
conn, err := grpc.Dial("localhost:8083", grpc.WithInsecure())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

client := postpb.NewPostServiceClient(conn)

// Запрос на создание поста
response, err := client.CreatePost(context.Background(), &postpb.CreatePostRequest{
    UserId:      "user-123",
    CharacterId: "character-456",
    Caption:     "Мой первый пост в этом удивительном мире!",
    MediaId:     "media-789",
    WorldId:     "world-101",
    IsAi:        false,
})

if err != nil {
    log.Fatalf("Failed to create post: %v", err)
}

log.Printf("Post created: ID=%s, CreatedAt=%s", response.PostId, response.CreatedAt)
```

### Получение глобальной ленты

```go
// Запрос на получение глобальной ленты
response, err := client.GetGlobalFeed(context.Background(), &postpb.GetGlobalFeedRequest{
    Limit:    10,
    Cursor:   "", // Пустой курсор для первой страницы
    WorldId:  "world-101",
})

if err != nil {
    log.Fatalf("Failed to get global feed: %v", err)
}

log.Printf("Received %d posts, next cursor: %s", len(response.Posts), response.NextCursor)
for _, post := range response.Posts {
    log.Printf("- %s: %s", post.DisplayName, post.Caption)
    log.Printf("  Likes: %d, Comments: %d", post.LikesCount, post.CommentsCount)
}

// Загрузка следующей страницы
if response.NextCursor != "" {
    nextPageResponse, err := client.GetGlobalFeed(context.Background(), &postpb.GetGlobalFeedRequest{
        Limit:    10,
        Cursor:   response.NextCursor,
        WorldId:  "world-101",
    })
    // ...
}
```

### Создание AI-поста

```go
// Запрос на создание AI-поста (используется только из AI Worker)
response, err := client.CreateAIPost(context.Background(), &postpb.CreateAIPostRequest{
    CharacterId: "ai-character-789",
    Caption:     "Исследую новые технологии виртуальной реальности!",
    MediaId:     "ai-generated-media-456",
    WorldId:     "world-101",
})

if err != nil {
    log.Fatalf("Failed to create AI post: %v", err)
}

log.Printf("AI post created: ID=%s", response.PostId)
```