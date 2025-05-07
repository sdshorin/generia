# Interaction Service для Generia

Interaction Service - микросервис, отвечающий за управление социальными взаимодействиями пользователей с контентом в платформе Generia. Сервис обрабатывает лайки, комментарии и предоставляет статистику взаимодействий для постов в виртуальных мирах.

## Оглавление

- [Обзор](#обзор)
- [Архитектура](#архитектура)
  - [Модели данных](#модели-данных)
  - [Репозитории](#репозитории)
  - [Сервисные слои](#сервисные-слои)
  - [gRPC API](#grpc-api)
- [Функциональность](#функциональность)
  - [Лайки](#лайки)
  - [Комментарии](#комментарии)
  - [Статистика](#статистика)
- [Технические детали](#технические-детали)
  - [База данных](#база-данных)
  - [Интеграция с другими сервисами](#интеграция-с-другими-сервисами)
  - [Оптимизация](#оптимизация)
- [Настройка и запуск](#настройка-и-запуск)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск сервиса](#запуск-сервиса)
- [Примеры использования](#примеры-использования)

## Обзор

Interaction Service управляет социальными взаимодействиями внутри виртуальных миров Generia. Сервис обрабатывает пользовательские лайки и комментарии к постам, предоставляя простой способ для социального взаимодействия между пользователями и AI-персонажами. Сервис использует MongoDB для эффективного хранения и обработки данных о взаимодействиях.

Основные возможности:
- Добавление и удаление лайков к постам
- Добавление комментариев к постам
- Получение списка лайков и комментариев для поста
- Проверка, лайкнул ли пользователь пост
- Получение статистики по постам (количество лайков и комментариев)
- Пакетное получение статистики для нескольких постов

## Архитектура

Interaction Service следует трехслойной архитектуре, характерной для микросервисов в проекте Generia:

### Модели данных

Основные модели данных в Interaction Service:

```go
// Файл: services/interaction-service/internal/models/interaction.go
type Like struct {
    ID        string    `bson:"_id,omitempty"`
    PostID    string    `bson:"post_id"`
    UserID    string    `bson:"user_id"`
    CreatedAt time.Time `bson:"created_at"`
}

// Файл: services/interaction-service/internal/models/interaction.go
type Comment struct {
    ID        string    `bson:"_id,omitempty"`
    PostID    string    `bson:"post_id"`
    UserID    string    `bson:"user_id"`
    Text      string    `bson:"text"`
    CreatedAt time.Time `bson:"created_at"`
}

// Файл: services/interaction-service/internal/models/interaction.go
type PostStats struct {
    PostID        string    `bson:"_id"` // Using post_id as the _id
    LikesCount    int32     `bson:"likes_count"`
    CommentsCount int32     `bson:"comments_count"`
    UpdatedAt     time.Time `bson:"updated_at"`
}
```

Эти модели представляют основные сущности для работы с социальными взаимодействиями.

### Репозитории

Слой репозиториев отвечает за взаимодействие с базой данных:

```go
// Файл: services/interaction-service/internal/repository/interaction_repository.go
type InteractionRepository interface {
    // Лайки
    AddLike(ctx context.Context, like *models.Like) error
    RemoveLike(ctx context.Context, postID, userID string) error
    GetPostLikes(ctx context.Context, postID string, limit, offset int) ([]*models.Like, int, error)
    CheckUserLiked(ctx context.Context, postID, userID string) (bool, error)
    
    // Комментарии
    AddComment(ctx context.Context, comment *models.Comment) error
    GetPostComments(ctx context.Context, postID string, limit, offset int) ([]*models.Comment, int, error)
    
    // Статистика
    GetPostStats(ctx context.Context, postID string) (*models.PostStats, error)
    GetPostsStats(ctx context.Context, postIDs []string) (map[string]*models.PostStats, error)
    UpdatePostStats(ctx context.Context, postID string) error
}
```

Репозиторий предоставляет методы для работы с базой данных MongoDB, включая операции с лайками, комментариями и статистикой постов.

### Сервисные слои

Сервисный слой реализует бизнес-логику и gRPC API:

```go
// Файл: services/interaction-service/internal/service/interaction_service.go
type InteractionService struct {
    interactionpb.UnimplementedInteractionServiceServer
    interactionRepo repository.InteractionRepository
    authClient      authpb.AuthServiceClient
}
```

Сервисный слой:
- Обрабатывает входящие gRPC-запросы
- Валидирует входные данные
- Вызывает методы репозитория для работы с данными
- Взаимодействует с другими сервисами
- Формирует ответы для клиентов

### gRPC API

Interaction Service предоставляет следующий gRPC-интерфейс:

```protobuf
// Файл: api/proto/interaction/interaction.proto
service InteractionService {
  // Лайки
  rpc LikePost(LikePostRequest) returns (LikePostResponse);
  rpc UnlikePost(UnlikePostRequest) returns (UnlikePostResponse);
  rpc GetPostLikes(GetPostLikesRequest) returns (PostLikesResponse);
  rpc CheckUserLiked(CheckUserLikedRequest) returns (CheckUserLikedResponse);
  
  // Комментарии
  rpc AddComment(AddCommentRequest) returns (AddCommentResponse);
  rpc GetPostComments(GetPostCommentsRequest) returns (PostCommentsResponse);
  
  // Статистика
  rpc GetPostStats(GetPostStatsRequest) returns (PostStatsResponse);
  rpc GetPostsStats(GetPostsStatsRequest) returns (PostsStatsResponse);

  // Проверка здоровья сервиса
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

## Функциональность

### Лайки

Interaction Service предоставляет функции для работы с лайками постов:

1. **Добавление лайка**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) LikePost(ctx context.Context, req *interactionpb.LikePostRequest) (*interactionpb.LikePostResponse, error)
   ```
   - Валидация входных данных и пользователя
   - Проверка существования пользователя через Auth Service
   - Создание записи о лайке в базе данных
   - Обновление статистики поста
   - Возврат актуального количества лайков

2. **Удаление лайка**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) UnlikePost(ctx context.Context, req *interactionpb.UnlikePostRequest) (*interactionpb.UnlikePostResponse, error)
   ```
   - Валидация входных данных
   - Удаление записи о лайке из базы данных
   - Обновление статистики поста
   - Возврат актуального количества лайков

3. **Получение лайков поста**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) GetPostLikes(ctx context.Context, req *interactionpb.GetPostLikesRequest) (*interactionpb.PostLikesResponse, error)
   ```
   - Получение списка лайков с пагинацией
   - Обогащение данных информацией о пользователях через Auth Service
   - Возврат списка пользователей, лайкнувших пост

4. **Проверка наличия лайка**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) CheckUserLiked(ctx context.Context, req *interactionpb.CheckUserLikedRequest) (*interactionpb.CheckUserLikedResponse, error)
   ```
   - Проверка, лайкнул ли конкретный пользователь пост
   - Возврат булевого значения (true/false)

### Комментарии

Interaction Service предоставляет функции для работы с комментариями:

1. **Добавление комментария**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) AddComment(ctx context.Context, req *interactionpb.AddCommentRequest) (*interactionpb.AddCommentResponse, error)
   ```
   - Валидация входных данных и пользователя
   - Проверка существования пользователя через Auth Service
   - Создание записи комментария в базе данных
   - Обновление статистики поста
   - Возврат ID созданного комментария и времени создания

2. **Получение комментариев поста**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) GetPostComments(ctx context.Context, req *interactionpb.GetPostCommentsRequest) (*interactionpb.PostCommentsResponse, error)
   ```
   - Получение списка комментариев с пагинацией
   - Обогащение данных информацией о пользователях через Auth Service
   - Возврат списка комментариев с данными авторов
   - Сортировка комментариев по времени создания

### Статистика

Interaction Service предоставляет функции для получения статистики взаимодействий:

1. **Получение статистики одного поста**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) GetPostStats(ctx context.Context, req *interactionpb.GetPostStatsRequest) (*interactionpb.PostStatsResponse, error)
   ```
   - Получение статистики для указанного поста
   - Возврат количества лайков и комментариев

2. **Пакетное получение статистики**
   ```go
   // Файл: services/interaction-service/internal/service/interaction_service.go
   func (s *InteractionService) GetPostsStats(ctx context.Context, req *interactionpb.GetPostsStatsRequest) (*interactionpb.PostsStatsResponse, error)
   ```
   - Получение статистики для нескольких постов одним запросом
   - Возврат карты (map) с данными о лайках и комментариях для каждого поста

## Технические детали

### База данных

Interaction Service использует MongoDB для хранения данных о взаимодействиях:

```go
// Файл: services/interaction-service/cmd/main.go
func createIndexes(db *mongo.Database) error {
    // Create indexes for likes collection
    likesCollection := db.Collection("likes")
    likeIndexes := []mongo.IndexModel{
        {
            Keys: bson.D{
                {Key: "post_id", Value: 1},
                {Key: "user_id", Value: 1},
            },
            Options: options.Index().SetUnique(true),
        },
        {
            Keys: bson.D{
                {Key: "post_id", Value: 1},
                {Key: "created_at", Value: -1},
            },
        },
        {
            Keys: bson.D{
                {Key: "user_id", Value: 1},
                {Key: "created_at", Value: -1},
            },
        },
    }
    
    // ...
}
```

Сервис использует три основные коллекции в MongoDB:
1. **likes** - хранит информацию о лайках (с уникальным индексом по паре post_id, user_id)
2. **comments** - хранит комментарии к постам
3. **stats** - хранит агрегированную статистику по постам

MongoDB была выбрана для этого сервиса по следующим причинам:
- Высокая производительность при интенсивных операциях чтения/записи
- Хорошая масштабируемость для большого количества взаимодействий
- Гибкая схема данных
- Эффективные атомарные операции для обновления счетчиков
- Поддержка агрегаций для получения статистики

### Интеграция с другими сервисами

Interaction Service взаимодействует с другими микросервисами:

```go
// Файл: services/interaction-service/cmd/main.go
func createAuthClient(discoveryClient discovery.ServiceDiscovery) (*grpc.ClientConn, authpb.AuthServiceClient, error) {
    // Get service address from Consul
    serviceAddress, err := discoveryClient.ResolveService("auth-service")
    if err != nil {
        return nil, nil, fmt.Errorf("failed to resolve auth service: %w", err)
    }

    // Create gRPC connection
    conn, err := grpc.Dial(
        serviceAddress,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             time.Second,
            PermitWithoutStream: true,
        }),
        grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
    )
    if err != nil {
        return nil, nil, fmt.Errorf("failed to connect to auth service: %w", err)
    }

    // Create client
    client := authpb.NewAuthServiceClient(conn)

    return conn, client, nil
}
```

Сервис интегрируется с:
1. **Auth Service** - для валидации пользователей и получения информации о них
2. **Consul** - для обнаружения других сервисов
3. **OpenTelemetry/Jaeger** - для трассировки запросов и мониторинга

Интеграция осуществляется через gRPC-клиенты и Service Discovery (Consul).

### Оптимизация

Interaction Service реализует несколько оптимизаций для повышения производительности:

1. **Индексы в MongoDB**:
   ```go
   // Файл: services/interaction-service/cmd/main.go
   func createIndexes(db *mongo.Database) error {
       // ...
       likesCollection := db.Collection("likes")
       likeIndexes := []mongo.IndexModel{
           // ...
       }
       // ...
   }
   ```
   - Индексы по часто используемым полям для быстрого поиска
   - Составные индексы для оптимизации запросов по нескольким полям
   - Уникальные индексы для предотвращения дубликатов

2. **Атомарные операции для счетчиков**:
   ```go
   // Файл: services/interaction-service/internal/repository/interaction_repository.go
   func (r *interactionRepository) UpdatePostStats(ctx context.Context, postID string) error {
       // Calculate stats
       // ...
       
       // Update stats
       filter := bson.M{"_id": postID}
       update := bson.M{
           "$set": bson.M{
               "likes_count":    likesCount,
               "comments_count": commentsCount,
               "updated_at":     now,
           },
       }
       opts := options.Update().SetUpsert(true)
       
       _, err = r.statsCol.UpdateOne(ctx, filter, update, opts)
       // ...
   }
   ```
   - Использование атомарных операций MongoDB для обновления счетчиков
   - Поддержка upsert для создания записи, если она не существует

3. **Пакетная обработка данных**:
   ```go
   // Файл: services/interaction-service/internal/repository/interaction_repository.go
   func (r *interactionRepository) GetPostsStats(ctx context.Context, postIDs []string) (map[string]*models.PostStats, error) {
       // ...
       filter := bson.M{"_id": bson.M{"$in": postIDs}}
       
       // Execute query
       cursor, err := r.statsCol.Find(ctx, filter)
       // ...
   }
   ```
   - Получение данных для нескольких постов одним запросом
   - Пакетная обработка для уменьшения количества запросов к базе данных

4. **Отложенное обновление статистики**:
   - После каждого изменения лайка или комментария обновляется агрегированная статистика
   - Но обновления не блокируют основной поток выполнения

## Настройка и запуск

### Переменные окружения

Для работы Interaction Service требуется настроить следующие переменные окружения:

```
# Основные настройки сервиса
SERVICE_NAME=interaction-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8085

# MongoDB
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=generia

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500

# Jaeger (Tracing)
JAEGER_HOST=jaeger
```

### Запуск сервиса

Interaction Service запускается как часть общей инфраструктуры Generia через docker-compose:

```bash
# Файл: docker-compose.yml
docker-compose up -d interaction-service
```

Для локальной разработки сервис можно запустить отдельно:

```bash
cd services/interaction-service
go run cmd/main.go
```

## Примеры использования

### Добавление лайка

```go
// gRPC-клиент
conn, err := grpc.Dial("localhost:8085", grpc.WithInsecure())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

client := interactionpb.NewInteractionServiceClient(conn)

// Запрос на добавление лайка
response, err := client.LikePost(context.Background(), &interactionpb.LikePostRequest{
    PostId:  "post-123",
    UserId:  "user-456",
    WorldId: "world-789",
})

if err != nil {
    log.Fatalf("Failed to like post: %v", err)
}

log.Printf("Post liked successfully. Total likes: %d", response.LikesCount)
```

### Получение комментариев

```go
// Запрос на получение комментариев
response, err := client.GetPostComments(context.Background(), &interactionpb.GetPostCommentsRequest{
    PostId:  "post-123",
    Limit:   10,
    Offset:  0,
    WorldId: "world-789",
})

if err != nil {
    log.Fatalf("Failed to get comments: %v", err)
}

log.Printf("Total comments: %d", response.Total)
for _, comment := range response.Comments {
    log.Printf("Comment by %s: %s", comment.Username, comment.Text)
}
```

### Получение статистики для нескольких постов

```go
// Запрос на получение статистики для нескольких постов
response, err := client.GetPostsStats(context.Background(), &interactionpb.GetPostsStatsRequest{
    PostIds: []string{"post-1", "post-2", "post-3"},
})

if err != nil {
    log.Fatalf("Failed to get post stats: %v", err)
}

for postID, stats := range response.Stats {
    log.Printf("Post %s: %d likes, %d comments", postID, stats.LikesCount, stats.CommentsCount)
}
```