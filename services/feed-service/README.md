# Feed Service для Generia

Feed Service - микросервис, отвечающий за формирование персонализированных лент контента в платформе Generia. Этот сервис собирает посты из разных источников, применяет алгоритмы ранжирования и предоставляет оптимизированные ленты контента для пользователей в виртуальных мирах.

## Оглавление

- [Обзор](#обзор)
- [Архитектура](#архитектура)
  - [Модели данных](#модели-данных)
  - [Репозитории](#репозитории)
  - [Сервисные слои](#сервисные-слои)
  - [gRPC API](#grpc-api)
- [Функциональность](#функциональность)
  - [Лента мира](#лента-мира)
  - [Персонализированная лента](#персонализированная-лента)
  - [Кэширование](#кэширование)
  - [Обновление ленты](#обновление-ленты)
- [Технические детали](#технические-детали)
  - [База данных](#база-данных)
  - [Интеграция с другими сервисами](#интеграция-с-другими-сервисами)
  - [Алгоритмы ранжирования](#алгоритмы-ранжирования)
- [Настройка и запуск](#настройка-и-запуск)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск сервиса](#запуск-сервиса)
- [Примеры использования](#примеры-использования)

## Обзор

Feed Service является центральным компонентом для формирования лент контента в виртуальных мирах Generia. Служба агрегирует посты из разных источников, применяет алгоритмы ранжирования и персонализации, оптимизирует производительность с помощью кэширования и предоставляет готовые ленты контента для пользователей.

Основные возможности:
- Формирование ленты постов для виртуального мира
- Персонализация контента на основе предпочтений пользователя
- Кэширование часто запрашиваемых лент для повышения производительности
- Инкрементальные обновления для оптимизации загрузки ленты
- Обработка больших объемов данных с эффективной пагинацией
- Интеграция с другими сервисами для получения полной информации о постах

## Архитектура

Feed Service следует трехслойной архитектуре, типичной для микросервисов:

### Модели данных

Основные модели данных в Feed Service:

```go
// Модель элемента ленты
type FeedItem struct {
    ID        string    `json:"id" redis:"id"`
    PostID    string    `json:"post_id" redis:"post_id"`
    WorldID   string    `json:"world_id" redis:"world_id"`
    UserID    string    `json:"user_id" redis:"user_id"`
    Score     float64   `json:"score" redis:"score"`
    CreatedAt time.Time `json:"created_at" redis:"created_at"`
    UpdatedAt time.Time `json:"updated_at" redis:"updated_at"`
}

// Модель ленты с курсором
type Feed struct {
    Items      []*FeedItem `json:"items"`
    NextCursor string      `json:"next_cursor"`
    HasMore    bool        `json:"has_more"`
}

// Расширенная модель поста для ленты
type FeedPost struct {
    ID            string    `json:"id"`
    WorldID       string    `json:"world_id"`
    UserID        string    `json:"user_id"`
    CharacterID   string    `json:"character_id"`
    Username      string    `json:"username"`
    DisplayName   string    `json:"display_name"`
    AvatarURL     string    `json:"avatar_url"`
    Caption       string    `json:"caption"`
    MediaURL      string    `json:"media_url"`
    LikesCount    int       `json:"likes_count"`
    CommentsCount int       `json:"comments_count"`
    UserLiked     bool      `json:"user_liked"`
    IsAI          bool      `json:"is_ai"`
    CreatedAt     time.Time `json:"created_at"`
}
```

### Репозитории

Слой репозиториев отвечает за взаимодействие с Redis:

- `FeedRepository` - Работа с лентами контента
- `CacheRepository` - Кэширование данных

```go
type FeedRepository interface {
    SaveFeedItem(ctx context.Context, worldID string, item *FeedItem) error
    GetWorldFeed(ctx context.Context, worldID string, limit int, cursor string) ([]*FeedItem, string, bool, error)
    GetPersonalizedFeed(ctx context.Context, userID, worldID string, limit int, cursor string) ([]*FeedItem, string, bool, error)
    RemoveFeedItem(ctx context.Context, worldID, postID string) error
    UpdateFeedItemScore(ctx context.Context, worldID, postID string, score float64) error
}

type CacheRepository interface {
    SavePostToCache(ctx context.Context, post *FeedPost) error
    GetPostFromCache(ctx context.Context, postID string) (*FeedPost, error)
    InvalidatePostCache(ctx context.Context, postID string) error
    SaveFeedToCache(ctx context.Context, cacheKey string, feed *Feed) error
    GetFeedFromCache(ctx context.Context, cacheKey string) (*Feed, error)
    InvalidateFeedCache(ctx context.Context, worldID, userID string) error
}
```

### Сервисные слои

Сервисные слои реализуют бизнес-логику:

- `FeedService` - Формирование лент контента
- `RankingService` - Ранжирование и персонализация

```go
type FeedService interface {
    GetWorldFeed(ctx context.Context, worldID string, limit int, cursor string, currentUserID string) (*Feed, error)
    GetPersonalizedFeed(ctx context.Context, userID, worldID string, limit int, cursor string) (*Feed, error)
    InvalidateCache(ctx context.Context, worldID, postID string) error
    ProcessNewPost(ctx context.Context, worldID, postID string) error
}

type RankingService interface {
    CalculatePostScore(ctx context.Context, post *FeedPost, userID string) float64
    RankFeedItems(ctx context.Context, items []*FeedItem, userID string) []*FeedItem
    GetPersonalizationData(ctx context.Context, userID string) (*UserPreferences, error)
}
```

### gRPC API

Feed Service предоставляет gRPC API для других сервисов:

```protobuf
service FeedService {
    rpc GetWorldFeed(GetWorldFeedRequest) returns (FeedResponse);
    rpc GetPersonalizedFeed(GetPersonalizedFeedRequest) returns (FeedResponse);
    rpc InvalidateCache(InvalidateCacheRequest) returns (EmptyResponse);
    rpc ProcessNewPost(ProcessNewPostRequest) returns (EmptyResponse);
}
```

## Функциональность

### Лента мира

Feed Service предоставляет основную функцию для получения ленты постов мира:

1. Проверка кэша для быстрого ответа
2. При отсутствии в кэше - получение элементов ленты из Redis
3. Получение детальной информации о постах из Post Service
4. Агрегация данных из других сервисов (Character, Media, Interaction)
5. Сохранение результата в кэш для будущих запросов
6. Возврат ленты с курсором для пагинации

```go
func (s *feedService) GetWorldFeed(ctx context.Context, worldID string, limit int, cursor string, currentUserID string) (*Feed, error) {
    // Проверка кэша
    cacheKey := fmt.Sprintf("feed:world:%s:limit:%d:cursor:%s:user:%s", worldID, limit, cursor, currentUserID)
    cachedFeed, err := s.cacheRepo.GetFeedFromCache(ctx, cacheKey)
    if err == nil && cachedFeed != nil {
        return cachedFeed, nil
    }
    
    // Получение элементов ленты из Redis
    feedItems, nextCursor, hasMore, err := s.feedRepo.GetWorldFeed(ctx, worldID, limit, cursor)
    if err != nil {
        return nil, err
    }
    
    if len(feedItems) == 0 {
        return &Feed{
            Items:      []*FeedPost{},
            NextCursor: "",
            HasMore:    false,
        }, nil
    }
    
    // Получение информации о постах
    postIDs := make([]string, 0, len(feedItems))
    for _, item := range feedItems {
        postIDs = append(postIDs, item.PostID)
    }
    
    posts, err := s.getPostsByIDs(ctx, postIDs, currentUserID)
    if err != nil {
        return nil, err
    }
    
    // Формирование ленты в правильном порядке
    feedPosts := make([]*FeedPost, 0, len(postIDs))
    for _, item := range feedItems {
        if post, ok := posts[item.PostID]; ok {
            feedPosts = append(feedPosts, post)
        }
    }
    
    feed := &Feed{
        Items:      feedPosts,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }
    
    // Сохранение в кэш
    go func() {
        cacheTTL := 5 * time.Minute
        if err := s.cacheRepo.SaveFeedToCache(context.Background(), cacheKey, feed, cacheTTL); err != nil {
            s.logger.Error("Failed to save feed to cache", zap.Error(err))
        }
    }()
    
    return feed, nil
}
```

### Персонализированная лента

Feed Service также поддерживает персонализированные ленты для пользователей:

1. Получение базовой ленты мира
2. Получение данных о предпочтениях пользователя
3. Применение алгоритма ранжирования с учетом предпочтений
4. Фильтрация контента на основе интересов пользователя
5. Возврат персонализированной ленты

```go
func (s *feedService) GetPersonalizedFeed(ctx context.Context, userID, worldID string, limit int, cursor string) (*Feed, error) {
    // Проверка кэша
    cacheKey := fmt.Sprintf("feed:personalized:user:%s:world:%s:limit:%d:cursor:%s", userID, worldID, limit, cursor)
    cachedFeed, err := s.cacheRepo.GetFeedFromCache(ctx, cacheKey)
    if err == nil && cachedFeed != nil {
        return cachedFeed, nil
    }
    
    // Получение предпочтений пользователя
    preferences, err := s.rankingService.GetPersonalizationData(ctx, userID)
    if err != nil {
        // В случае ошибки возвращаем обычную ленту
        return s.GetWorldFeed(ctx, worldID, limit, cursor, userID)
    }
    
    // Получение элементов ленты с персонализацией
    feedItems, nextCursor, hasMore, err := s.feedRepo.GetPersonalizedFeed(ctx, userID, worldID, limit*2, cursor)
    if err != nil {
        return nil, err
    }
    
    if len(feedItems) == 0 {
        return &Feed{
            Items:      []*FeedPost{},
            NextCursor: "",
            HasMore:    false,
        }, nil
    }
    
    // Ранжирование элементов на основе предпочтений пользователя
    rankedItems := s.rankingService.RankFeedItems(ctx, feedItems, userID, preferences)
    
    // Ограничение до запрошенного количества
    if len(rankedItems) > limit {
        rankedItems = rankedItems[:limit]
    }
    
    // Получение информации о постах
    postIDs := make([]string, 0, len(rankedItems))
    for _, item := range rankedItems {
        postIDs = append(postIDs, item.PostID)
    }
    
    posts, err := s.getPostsByIDs(ctx, postIDs, userID)
    if err != nil {
        return nil, err
    }
    
    // Формирование ленты
    feedPosts := make([]*FeedPost, 0, len(postIDs))
    for _, item := range rankedItems {
        if post, ok := posts[item.PostID]; ok {
            feedPosts = append(feedPosts, post)
        }
    }
    
    feed := &Feed{
        Items:      feedPosts,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }
    
    // Сохранение в кэш
    go func() {
        cacheTTL := 5 * time.Minute
        if err := s.cacheRepo.SaveFeedToCache(context.Background(), cacheKey, feed, cacheTTL); err != nil {
            s.logger.Error("Failed to save personalized feed to cache", zap.Error(err))
        }
    }()
    
    return feed, nil
}
```

### Кэширование

Feed Service использует многоуровневое кэширование для оптимизации производительности:

1. **Кэш постов** - Кэширование информации о постах
2. **Кэш лент** - Кэширование готовых лент для конкретных запросов
3. **Инвалидация кэша** - Механизм сброса кэша при изменении данных

```go
func (r *redisCache) SaveFeedToCache(ctx context.Context, cacheKey string, feed *Feed, ttl time.Duration) error {
    // Сериализация ленты в JSON
    feedJSON, err := json.Marshal(feed)
    if err != nil {
        return err
    }
    
    // Сохранение в Redis с TTL
    return r.client.Set(ctx, cacheKey, feedJSON, ttl).Err()
}

func (r *redisCache) InvalidateFeedCache(ctx context.Context, worldID, userID string) error {
    // Удаление кэшей лент мира
    worldPattern := fmt.Sprintf("feed:world:%s:*", worldID)
    worldKeys, err := r.client.Keys(ctx, worldPattern).Result()
    if err != nil {
        return err
    }
    
    if len(worldKeys) > 0 {
        if err := r.client.Del(ctx, worldKeys...).Err(); err != nil {
            return err
        }
    }
    
    // Если указан пользователь, удаляем его персонализированные ленты
    if userID != "" {
        userPattern := fmt.Sprintf("feed:personalized:user:%s:world:%s:*", userID, worldID)
        userKeys, err := r.client.Keys(ctx, userPattern).Result()
        if err != nil {
            return err
        }
        
        if len(userKeys) > 0 {
            if err := r.client.Del(ctx, userKeys...).Err(); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

### Обновление ленты

Feed Service обрабатывает события создания новых постов для обновления лент:

1. Получение информации о новом посте
2. Расчет начального рейтинга поста
3. Добавление поста в ленту мира
4. Инвалидация кэша лент для отображения новых данных

```go
func (s *feedService) ProcessNewPost(ctx context.Context, worldID, postID string) error {
    // Получение информации о посте
    post, err := s.postClient.GetPost(ctx, &postpb.GetPostRequest{
        PostId: postID,
    })
    if err != nil {
        return err
    }
    
    // Расчет начального рейтинга
    initialScore := s.calculateInitialScore(post.Post.CreatedAt)
    
    // Создание элемента ленты
    feedItem := &FeedItem{
        ID:        uuid.New().String(),
        PostID:    postID,
        WorldID:   worldID,
        UserID:    post.Post.UserId,
        Score:     initialScore,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    // Сохранение в ленту мира
    if err := s.feedRepo.SaveFeedItem(ctx, worldID, feedItem); err != nil {
        return err
    }
    
    // Инвалидация кэша
    if err := s.InvalidateCache(ctx, worldID, ""); err != nil {
        s.logger.Error("Failed to invalidate cache", zap.Error(err))
    }
    
    return nil
}

func (s *feedService) calculateInitialScore(postCreatedAt string) float64 {
    // Конвертация строки времени
    createdTime, err := time.Parse(time.RFC3339, postCreatedAt)
    if err != nil {
        return 0
    }
    
    // Базовый рейтинг на основе времени создания
    // Более новые посты получают высший рейтинг
    now := time.Now()
    ageInHours := now.Sub(createdTime).Hours()
    
    // Формула рейтинга: recency_score = 1 / (1 + age_in_hours)
    // Это даст значение близкое к 1 для новых постов и близкое к 0 для старых
    recencyScore := 1.0 / (1.0 + ageInHours)
    
    return recencyScore
}
```

## Технические детали

### База данных

Feed Service использует Redis для хранения и управления лентами:

1. **Отсортированные множества** (Sorted Sets) - Для хранения элементов ленты с рейтингами
2. **Хеши** (Hashes) - Для кэширования данных о постах
3. **Строки** (Strings) - Для хранения кэшированных лент

```
// Схема использования Redis:

// Лента мира (sorted set)
feed:world:{world_id} -> [
    {post_id_1, score_1},
    {post_id_2, score_2},
    ...
]

// Персонализированная лента (sorted set)
feed:user:{user_id}:world:{world_id} -> [
    {post_id_1, personalized_score_1},
    {post_id_2, personalized_score_2},
    ...
]

// Кэш поста (hash)
post:{post_id} -> {
    id: "post_id",
    world_id: "world_id",
    user_id: "user_id",
    caption: "Post text",
    ...
}

// Кэш ленты (string/json)
feed:world:{world_id}:limit:{limit}:cursor:{cursor}:user:{user_id} -> JSON
feed:personalized:user:{user_id}:world:{world_id}:limit:{limit}:cursor:{cursor} -> JSON
```

### Интеграция с другими сервисами

Feed Service взаимодействует с другими микросервисами:

1. **Post Service** - Получение информации о постах
2. **Character Service** - Получение информации о персонажах-авторах
3. **Interaction Service** - Получение статистики взаимодействий
4. **Media Service** - Получение URL медиа-контента

```go
// Пример интеграции с Post Service для получения информации о постах
func (s *feedService) getPostsByIDs(ctx context.Context, postIDs []string, currentUserID string) (map[string]*FeedPost, error) {
    // Проверка кэша
    cachedPosts := make(map[string]*FeedPost)
    missingPostIDs := make([]string, 0, len(postIDs))
    
    for _, postID := range postIDs {
        post, err := s.cacheRepo.GetPostFromCache(ctx, postID)
        if err == nil && post != nil {
            cachedPosts[postID] = post
        } else {
            missingPostIDs = append(missingPostIDs, postID)
        }
    }
    
    // Если все посты найдены в кэше, возвращаем их
    if len(missingPostIDs) == 0 {
        return cachedPosts, nil
    }
    
    // Запрос недостающих постов из Post Service
    resp, err := s.postClient.GetPostsBatch(ctx, &postpb.GetPostsBatchRequest{
        PostIds:       missingPostIDs,
        CurrentUserId: currentUserID,
    })
    if err != nil {
        return nil, err
    }
    
    // Объединение результатов и кэширование новых постов
    result := cachedPosts
    for _, post := range resp.Posts {
        feedPost := &FeedPost{
            ID:            post.Id,
            WorldID:       post.WorldId,
            UserID:        post.UserId,
            CharacterID:   post.CharacterId,
            Username:      post.Username,
            DisplayName:   post.DisplayName,
            AvatarURL:     post.AvatarUrl,
            Caption:       post.Caption,
            MediaURL:      post.MediaUrl,
            LikesCount:    int(post.LikesCount),
            CommentsCount: int(post.CommentsCount),
            UserLiked:     post.UserLiked,
            IsAI:          post.IsAi,
            CreatedAt:     time.Now(), // Преобразование из строки
        }
        
        result[post.Id] = feedPost
        
        // Асинхронное кэширование
        go func(p *FeedPost) {
            if err := s.cacheRepo.SavePostToCache(context.Background(), p); err != nil {
                s.logger.Error("Failed to cache post", zap.Error(err), zap.String("postID", p.ID))
            }
        }(feedPost)
    }
    
    return result, nil
}
```

### Алгоритмы ранжирования

Feed Service использует несколько алгоритмов для ранжирования контента:

1. **Базовое ранжирование** - На основе времени создания и популярности
2. **Персонализированное ранжирование** - С учетом предпочтений пользователя
3. **Интеллектуальное смешивание** - Добавление разнообразия в ленту

```go
func (s *rankingService) CalculatePostScore(ctx context.Context, post *FeedPost, userID string, preferences *UserPreferences) float64 {
    // Базовые компоненты рейтинга
    const (
        recencyWeight    = 0.5  // Вес новизны
        popularityWeight = 0.3  // Вес популярности
        personalWeight   = 0.2  // Вес персонализации
    )
    
    // Расчет компонента новизны (более новые посты имеют более высокий рейтинг)
    ageInHours := time.Now().Sub(post.CreatedAt).Hours()
    recencyScore := 1.0 / (1.0 + ageInHours)
    
    // Расчет компонента популярности
    engagementCount := float64(post.LikesCount + post.CommentsCount)
    popularityScore := math.Log1p(engagementCount) // Логарифмическая шкала
    
    // Нормализация популярности (0 до 1)
    maxPopularity := 10.0 // Предполагаемое максимальное значение
    popularityScore = math.Min(popularityScore/maxPopularity, 1.0)
    
    // Персонализированный компонент, если есть предпочтения
    personalScore := 0.5 // Значение по умолчанию
    if preferences != nil {
        // Изучение истории взаимодействий пользователя
        if post.IsAI && preferences.AIContentPreference > 0 {
            // Пользователь предпочитает AI-контент
            personalScore = preferences.AIContentPreference
        }
        
        // Проверка на совпадение интересов
        if post.CharacterID != "" && preferences.PreferredCharacters[post.CharacterID] {
            personalScore = 0.8 // Предпочтительный персонаж
        }
        
        // Проверка на контекстную релевантность
        // ...
    }
    
    // Расчет итогового рейтинга
    finalScore := (recencyWeight * recencyScore) + 
                 (popularityWeight * popularityScore) + 
                 (personalWeight * personalScore)
    
    return finalScore
}
```

## Настройка и запуск

### Переменные окружения

Для работы Feed Service требуется настроить следующие переменные окружения:

```
# Основные настройки
SERVICE_NAME=feed-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=9070

# Redis
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=
REDIS_DB=0

# Кэширование
CACHE_TTL=300 # 5 минут
POST_CACHE_TTL=600 # 10 минут

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500
CONSUL_HEALTH_CHECK_INTERVAL=10s

# Tracing
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=feed-service

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Запуск сервиса

Feed Service является частью общей инфраструктуры Generia и запускается через docker-compose:

```bash
docker-compose up -d feed-service
```

Для локальной разработки можно запустить сервис отдельно:

```bash
cd services/feed-service
go run cmd/main.go
```

## Примеры использования

### Получение ленты мира

```go
// gRPC-клиент
client := pb.NewFeedServiceClient(conn)

// Запрос на получение ленты мира
response, err := client.GetWorldFeed(context.Background(), &pb.GetWorldFeedRequest{
    WorldId:      "world-id-123",
    Limit:        10,
    Cursor:       "", // Пустой курсор для первой страницы
    CurrentUserId: "user-id-456", // Для проверки, лайкнул ли пользователь посты
})

if err != nil {
    log.Fatalf("Failed to get world feed: %v", err)
}

log.Printf("Received %d posts, next cursor: %s", len(response.Posts), response.NextCursor)
for _, post := range response.Posts {
    log.Printf("- Post by %s: %s", post.DisplayName, post.Caption)
    log.Printf("  Likes: %d, Comments: %d", post.LikesCount, post.CommentsCount)
    if post.UserLiked {
        log.Printf("  You liked this post")
    }
}
```

### Получение персонализированной ленты

```go
// Запрос на получение персонализированной ленты
response, err := client.GetPersonalizedFeed(context.Background(), &pb.GetPersonalizedFeedRequest{
    UserId:  "user-id-456",
    WorldId: "world-id-123",
    Limit:   10,
    Cursor:  "",
})

if err != nil {
    log.Fatalf("Failed to get personalized feed: %v", err)
}

log.Printf("Received %d personalized posts", len(response.Posts))
for i, post := range response.Posts {
    log.Printf("%d. Post by %s: %s", i+1, post.DisplayName, post.Caption)
}
```

### Обработка нового поста

```go
// Обработка нового поста для добавления в ленту
_, err := client.ProcessNewPost(context.Background(), &pb.ProcessNewPostRequest{
    WorldId: "world-id-123",
    PostId:  "post-id-789",
})

if err != nil {
    log.Fatalf("Failed to process new post: %v", err)
}

log.Printf("Post processed and added to feed")
```