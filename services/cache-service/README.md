# Cache Service для Generia

Cache Service - микросервис, отвечающий за централизованное кэширование данных в проекте Generia. Этот сервис обеспечивает эффективное хранение и получение часто запрашиваемых данных, снижая нагрузку на основные сервисы и базы данных.

## Оглавление

- [Обзор](#обзор)
- [Архитектура](#архитектура)
  - [Модели данных](#модели-данных)
  - [Репозитории](#репозитории)
  - [Сервисные слои](#сервисные-слои)
  - [gRPC API](#grpc-api)
- [Функциональность](#функциональность)
  - [Управление кэшем](#управление-кэшем)
  - [Типы данных](#типы-данных)
  - [Политики кэширования](#политики-кэширования)
- [Технические детали](#технические-детали)
  - [Redis как хранилище](#redis-как-хранилище)
  - [Сериализация данных](#сериализация-данных)
  - [Управление TTL](#управление-ttl)
- [Настройка и запуск](#настройка-и-запуск)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск сервиса](#запуск-сервиса)
- [Примеры использования](#примеры-использования)

## Обзор

Cache Service предоставляет унифицированный интерфейс для кэширования разнообразных данных в платформе Generia. Сервис использует Redis в качестве хранилища и поддерживает различные типы данных, политики инвалидации кэша и механизмы оптимизации. Централизованный подход к кэшированию позволяет эффективно управлять ресурсами и обеспечивать высокую производительность платформы в целом.

Основные возможности:
- Хранение и получение кэшированных данных различных типов
- Настраиваемые политики кэширования с разным TTL (время жизни)
- Поддержка атомарных операций с кэшем
- Эффективная инвалидация кэша по ключам или паттернам
- Сериализация и десериализация структурированных данных
- Мониторинг состояния кэша и статистики использования

## Архитектура

Cache Service следует трехслойной архитектуре, типичной для микросервисов:

### Модели данных

Основные модели данных в Cache Service:

```go
// Модель кэшированного значения
type CacheEntry struct {
    Key       string    `json:"key"`
    Value     []byte    `json:"value"`
    TTL       int       `json:"ttl"` // в секундах
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Статистика использования кэша
type CacheStats struct {
    Hits      int64 `json:"hits"`
    Misses    int64 `json:"misses"`
    KeysCount int64 `json:"keys_count"`
    MemUsage  int64 `json:"mem_usage"` // в байтах
}
```

### Репозитории

Слой репозиториев отвечает за взаимодействие с Redis:

- `CacheRepository` - Работа с кэшем

```go
type CacheRepository interface {
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Get(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    DeleteByPattern(ctx context.Context, pattern string) error
    Exists(ctx context.Context, key string) (bool, error)
    SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
    Increment(ctx context.Context, key string, delta int64) (int64, error)
    GetStats(ctx context.Context) (*CacheStats, error)
}
```

### Сервисные слои

Сервисные слои реализуют бизнес-логику:

- `CacheService` - Управление кэшем

```go
type CacheService interface {
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Get(ctx context.Context, key string) ([]byte, error)
    Delete(ctx context.Context, key string) error
    DeleteByPattern(ctx context.Context, pattern string) error
    Exists(ctx context.Context, key string) (bool, error)
    SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
    Increment(ctx context.Context, key string, delta int64) (int64, error)
    GetStats(ctx context.Context) (*CacheStats, error)
}
```

### gRPC API

Cache Service предоставляет gRPC API для других сервисов:

```protobuf
service CacheService {
    rpc Set(SetRequest) returns (EmptyResponse);
    rpc Get(GetRequest) returns (GetResponse);
    rpc Delete(DeleteRequest) returns (EmptyResponse);
    rpc DeleteByPattern(DeleteByPatternRequest) returns (DeleteByPatternResponse);
    rpc Exists(ExistsRequest) returns (ExistsResponse);
    rpc SetNX(SetNXRequest) returns (SetNXResponse);
    rpc Increment(IncrementRequest) returns (IncrementResponse);
    rpc GetStats(EmptyRequest) returns (StatsResponse);
}
```

## Функциональность

### Управление кэшем

Cache Service предоставляет полный набор операций для работы с кэшем:

1. **Установка значения** - Сохранение данных в кэш с указанным TTL:

```go
func (s *cacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    if err := s.validateKey(key); err != nil {
        return err
    }
    
    if len(value) > s.config.MaxValueSize {
        return ErrValueTooLarge
    }
    
    // Применение префикса к ключу для изоляции
    prefixedKey := s.applyPrefix(key)
    
    // Логирование операции
    s.logger.Debug("Setting cache value", 
        zap.String("key", key),
        zap.Int("value_size", len(value)),
        zap.Duration("ttl", ttl))
    
    // Сохранение в Redis
    return s.repo.Set(ctx, prefixedKey, value, ttl)
}
```

2. **Получение значения** - Получение данных из кэша по ключу:

```go
func (s *cacheService) Get(ctx context.Context, key string) ([]byte, error) {
    if err := s.validateKey(key); err != nil {
        return nil, err
    }
    
    // Применение префикса к ключу
    prefixedKey := s.applyPrefix(key)
    
    // Получение из Redis
    value, err := s.repo.Get(ctx, prefixedKey)
    if err != nil {
        if errors.Is(err, redis.Nil) {
            // Кэш-промах
            s.metrics.IncrementMisses()
            return nil, ErrCacheMiss
        }
        return nil, err
    }
    
    // Кэш-попадание
    s.metrics.IncrementHits()
    
    return value, nil
}
```

3. **Удаление значения** - Удаление данных из кэша по ключу:

```go
func (s *cacheService) Delete(ctx context.Context, key string) error {
    if err := s.validateKey(key); err != nil {
        return err
    }
    
    // Применение префикса к ключу
    prefixedKey := s.applyPrefix(key)
    
    // Удаление из Redis
    return s.repo.Delete(ctx, prefixedKey)
}
```

4. **Удаление по паттерну** - Удаление группы данных, соответствующих паттерну:

```go
func (s *cacheService) DeleteByPattern(ctx context.Context, pattern string) error {
    if pattern == "" {
        return ErrInvalidPattern
    }
    
    // Применение префикса к паттерну
    prefixedPattern := s.applyPrefix(pattern)
    
    // Удаление по паттерну из Redis
    count, err := s.repo.DeleteByPattern(ctx, prefixedPattern)
    if err != nil {
        return err
    }
    
    s.logger.Debug("Deleted cache entries by pattern",
        zap.String("pattern", pattern),
        zap.Int64("count", count))
    
    return nil
}
```

5. **Атомарные операции** - Поддержка атомарных операций (SetNX, Increment):

```go
func (s *cacheService) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
    if err := s.validateKey(key); err != nil {
        return false, err
    }
    
    if len(value) > s.config.MaxValueSize {
        return false, ErrValueTooLarge
    }
    
    // Применение префикса к ключу
    prefixedKey := s.applyPrefix(key)
    
    // Атомарная операция SetNX в Redis
    return s.repo.SetNX(ctx, prefixedKey, value, ttl)
}

func (s *cacheService) Increment(ctx context.Context, key string, delta int64) (int64, error) {
    if err := s.validateKey(key); err != nil {
        return 0, err
    }
    
    // Применение префикса к ключу
    prefixedKey := s.applyPrefix(key)
    
    // Атомарная операция Increment в Redis
    return s.repo.Increment(ctx, prefixedKey, delta)
}
```

### Типы данных

Cache Service поддерживает кэширование различных типов данных:

1. **Простые типы** - строки, числа, boolean-значения
2. **Сложные структуры** - JSON-представления объектов
3. **Бинарные данные** - сериализованные объекты, протобуферы и т.д.

```go
// Пример сериализации/десериализации JSON
func Serialize(value interface{}) ([]byte, error) {
    return json.Marshal(value)
}

func Deserialize(data []byte, target interface{}) error {
    return json.Unmarshal(data, target)
}

// Пример использования в сервисе
func (s *cacheService) SetObject(ctx context.Context, key string, obj interface{}, ttl time.Duration) error {
    data, err := Serialize(obj)
    if err != nil {
        return err
    }
    return s.Set(ctx, key, data, ttl)
}

func (s *cacheService) GetObject(ctx context.Context, key string, target interface{}) error {
    data, err := s.Get(ctx, key)
    if err != nil {
        return err
    }
    return Deserialize(data, target)
}
```

### Политики кэширования

Cache Service поддерживает различные политики кэширования:

1. **TTL-based** - Автоматическое удаление данных по истечении времени жизни
2. **LRU (Least Recently Used)** - Redis автоматически удаляет наименее используемые данные при достижении лимита памяти
3. **Prefixed keys** - Использование префиксов для группировки и изоляции данных различных сервисов

```go
// Примеры префиксов для разных типов данных
const (
    PrefixUser      = "user:"
    PrefixPost      = "post:"
    PrefixCharacter = "character:"
    PrefixWorld     = "world:"
    PrefixFeed      = "feed:"
)

// Примеры TTL для разных типов данных
const (
    TTLShort  = 5 * time.Minute
    TTLMedium = 1 * time.Hour
    TTLLong   = 24 * time.Hour
)
```

## Технические детали

### Redis как хранилище

Cache Service использует Redis в качестве основного хранилища данных:

1. **Эффективность** - Redis хранит данные в памяти, обеспечивая высокую скорость доступа
2. **Поддержка TTL** - Встроенный механизм автоматического удаления устаревших данных
3. **Атомарные операции** - Поддержка атомарных операций для безопасной работы в распределенной среде
4. **Поддержка паттернов** - Возможность работы с группами ключей по паттернам

```go
// Подключение к Redis
func NewRedisClient(config *config.Config) (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     config.RedisAddress,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
        PoolSize: config.RedisPoolSize,
    })
    
    // Проверка соединения
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }
    
    return client, nil
}
```

### Сериализация данных

Cache Service поддерживает различные форматы сериализации:

1. **JSON** - Для структурированных данных с хорошей читаемостью
2. **Protocol Buffers** - Для компактного и эффективного представления данных
3. **Binary** - Для произвольных бинарных данных

```go
// Интерфейс сериализатора
type Serializer interface {
    Serialize(value interface{}) ([]byte, error)
    Deserialize(data []byte, target interface{}) error
}

// JSON сериализатор
type JSONSerializer struct{}

func (s *JSONSerializer) Serialize(value interface{}) ([]byte, error) {
    return json.Marshal(value)
}

func (s *JSONSerializer) Deserialize(data []byte, target interface{}) error {
    return json.Unmarshal(data, target)
}

// Protocol Buffers сериализатор
type ProtobufSerializer struct{}

func (s *ProtobufSerializer) Serialize(value interface{}) ([]byte, error) {
    if m, ok := value.(proto.Message); ok {
        return proto.Marshal(m)
    }
    return nil, errors.New("value is not a proto.Message")
}

func (s *ProtobufSerializer) Deserialize(data []byte, target interface{}) error {
    if m, ok := target.(proto.Message); ok {
        return proto.Unmarshal(data, m)
    }
    return errors.New("target is not a proto.Message")
}
```

### Управление TTL

Cache Service предоставляет гибкие механизмы управления временем жизни кэша:

1. **Динамическое TTL** - Установка времени жизни на основе типа данных
2. **Обновление TTL** - Продление времени жизни при обращении к данным
3. **Default TTL** - Значения по умолчанию для разных категорий данных

```go
// Пример функции для определения оптимального TTL
func (s *cacheService) determineTTL(key string, requestedTTL time.Duration) time.Duration {
    // Если TTL задано явно, используем его
    if requestedTTL > 0 {
        // Ограничиваем максимальным значением
        if requestedTTL > s.config.MaxTTL {
            return s.config.MaxTTL
        }
        return requestedTTL
    }
    
    // Иначе определяем TTL на основе типа данных
    switch {
    case strings.HasPrefix(key, "user:"):
        return s.config.DefaultUserTTL
    case strings.HasPrefix(key, "post:"):
        return s.config.DefaultPostTTL
    case strings.HasPrefix(key, "feed:"):
        return s.config.DefaultFeedTTL
    default:
        return s.config.DefaultTTL
    }
}
```

## Настройка и запуск

### Переменные окружения

Для работы Cache Service требуется настроить следующие переменные окружения:

```
# Основные настройки
SERVICE_NAME=cache-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=9090

# Redis
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=100

# Кэширование
DEFAULT_TTL=300 # 5 минут
MAX_TTL=86400 # 24 часа
MAX_VALUE_SIZE=10485760 # 10 MB

# Префиксы
KEY_PREFIX=generia:

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500
CONSUL_HEALTH_CHECK_INTERVAL=10s

# Tracing
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=cache-service

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Запуск сервиса

Cache Service является частью общей инфраструктуры Generia и запускается через docker-compose:

```bash
docker-compose up -d cache-service
```

Для локальной разработки можно запустить сервис отдельно:

```bash
cd services/cache-service
go run cmd/main.go
```

## Примеры использования

### Кэширование данных

```go
// gRPC-клиент
client := pb.NewCacheServiceClient(conn)

// Кэширование данных пользователя
userData := map[string]interface{}{
    "id":        "user-123",
    "username":  "john_doe",
    "email":     "john@example.com",
    "created_at": "2023-01-01T00:00:00Z",
}

userDataJSON, _ := json.Marshal(userData)

// Установка значения в кэш
_, err := client.Set(context.Background(), &pb.SetRequest{
    Key:   "user:user-123",
    Value: userDataJSON,
    Ttl:   3600, // 1 час в секундах
})

if err != nil {
    log.Fatalf("Failed to set cache: %v", err)
}

log.Printf("User data cached successfully")
```

### Получение данных из кэша

```go
// Получение данных из кэша
response, err := client.Get(context.Background(), &pb.GetRequest{
    Key: "user:user-123",
})

if err != nil {
    log.Fatalf("Failed to get from cache: %v", err)
}

// Десериализация данных
var userData map[string]interface{}
if err := json.Unmarshal(response.Value, &userData); err != nil {
    log.Fatalf("Failed to unmarshal user data: %v", err)
}

log.Printf("Retrieved user from cache: %s", userData["username"])
```

### Инвалидация кэша

```go
// Удаление конкретного ключа
_, err := client.Delete(context.Background(), &pb.DeleteRequest{
    Key: "user:user-123",
})

if err != nil {
    log.Fatalf("Failed to delete from cache: %v", err)
}

// Удаление группы ключей по паттерну
response, err := client.DeleteByPattern(context.Background(), &pb.DeleteByPatternRequest{
    Pattern: "user:*",
})

if err != nil {
    log.Fatalf("Failed to delete by pattern: %v", err)
}

log.Printf("Deleted %d cache entries by pattern", response.Count)
```

### Атомарные операции

```go
// Атомарная установка значения (только если ключа не существует)
response, err := client.SetNX(context.Background(), &pb.SetNXRequest{
    Key:   "lock:resource-123",
    Value: []byte("locked"),
    Ttl:   60, // 1 минута
})

if err != nil {
    log.Fatalf("Failed to set lock: %v", err)
}

if response.Success {
    log.Printf("Lock acquired")
    
    // Работа с защищенным ресурсом
    
    // Освобождение блокировки
    _, err := client.Delete(context.Background(), &pb.DeleteRequest{
        Key: "lock:resource-123",
    })
    
    if err != nil {
        log.Printf("Failed to release lock: %v", err)
    }
} else {
    log.Printf("Resource is already locked")
}
```