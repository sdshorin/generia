# Character Service для Generia

Character Service - микросервис, отвечающий за управление персонажами в виртуальных мирах проекта Generia. Сервис обеспечивает создание, хранение и получение информации о персонажах, включая как реальных пользователей, так и AI-сгенерированных персонажей.

## Оглавление

- [Обзор](#обзор)
- [Архитектура](#архитектура)
  - [Модели данных](#модели-данных)
  - [Репозитории](#репозитории)
  - [Сервисные слои](#сервисные-слои)
  - [gRPC API](#grpc-api)
- [Функциональность](#функциональность)
  - [Создание персонажей](#создание-персонажей)
  - [Получение информации о персонажах](#получение-информации-о-персонажах)
  - [AI-персонажи](#ai-персонажи)
- [Технические детали](#технические-детали)
  - [База данных](#база-данных)
  - [Интеграция с другими сервисами](#интеграция-с-другими-сервисами)
  - [Обработка метаданных](#обработка-метаданных)
- [Настройка и запуск](#настройка-и-запуск)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск сервиса](#запуск-сервиса)
- [Примеры использования](#примеры-использования)

## Обзор

Character Service управляет персонажами в виртуальных мирах Generia. Персонаж может представлять как реального пользователя платформы, так и AI-сгенерированный профиль. Сервис обеспечивает хранение и получение данных о персонажах внутри каждого мира.

Основные возможности:
- Создание персонажей для реальных пользователей в конкретных мирах
- Хранение и управление AI-сгенерированными персонажами
- Получение информации о персонажах по ID
- Получение персонажей пользователя в конкретном мире
- Хранение метаданных персонажей (внешность, характер, интересы и т.д.)

## Архитектура

Character Service следует трехслойной архитектуре, типичной для микросервисов в проекте Generia:

### Модели данных

Основная модель данных в Character Service - структура `Character`:

```go
// Файл: services/character-service/internal/models/character.go
type Character struct {
    ID            string
    WorldID       string
    RealUserID    sql.NullString // NULL для AI-персонажей
    IsAI          bool           // Вычисляемое поле на основе RealUserID
    DisplayName   string
    AvatarMediaID sql.NullString
    Meta          json.RawMessage
    CreatedAt     time.Time
}
```

Для создания персонажа используется структура:

```go
// Файл: services/character-service/internal/models/character.go
type CreateCharacterParams struct {
    WorldID       string
    RealUserID    sql.NullString
    DisplayName   string
    AvatarMediaID sql.NullString
    Meta          json.RawMessage
}
```

### Репозитории

Слой репозиториев отвечает за взаимодействие с базой данных:

```go
// Файл: services/character-service/internal/repository/character_repository.go
type CharacterRepository interface {
    CreateCharacter(ctx context.Context, params models.CreateCharacterParams) (*models.Character, error)
    GetCharacter(ctx context.Context, id string) (*models.Character, error)
    GetUserCharactersInWorld(ctx context.Context, userID, worldID string) ([]*models.Character, error)
}
```

Репозиторий предоставляет методы для:
- Создания новых персонажей
- Получения персонажа по ID
- Получения всех персонажей пользователя в конкретном мире

### Сервисные слои

Сервисный слой реализует бизнес-логику и gRPC-интерфейс:

```go
// Файл: services/character-service/internal/service/character_service.go
type CharacterService struct {
    pb.UnimplementedCharacterServiceServer
    repo repository.CharacterRepository
}

func (s *CharacterService) CreateCharacter(ctx context.Context, req *pb.CreateCharacterRequest) (*pb.Character, error)
func (s *CharacterService) GetCharacter(ctx context.Context, req *pb.GetCharacterRequest) (*pb.Character, error)
func (s *CharacterService) GetUserCharactersInWorld(ctx context.Context, req *pb.GetUserCharactersInWorldRequest) (*pb.CharacterList, error)
func (s *CharacterService) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error)
```

Сервисный слой обрабатывает gRPC-запросы, вызывает соответствующие методы репозитория и преобразует данные между внутренними моделями и протобаферами.

### gRPC API

Character Service предоставляет следующий gRPC-интерфейс:

```protobuf
// Файл: api/proto/character/character.proto
service CharacterService {
  // Создать персонажа для реального пользователя или AI
  rpc CreateCharacter(CreateCharacterRequest) returns (Character);
  
  // Получить персонажа по ID
  rpc GetCharacter(GetCharacterRequest) returns (Character);
  
  // Получить персонажей пользователя в конкретном мире
  rpc GetUserCharactersInWorld(GetUserCharactersInWorldRequest) returns (CharacterList);
  
  // Проверка здоровья сервиса
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

## Функциональность

### Создание персонажей

Процесс создания нового персонажа:

1. Клиент передает информацию о персонаже через gRPC-запрос `CreateCharacterRequest`
2. Сервис подготавливает параметры для создания персонажа, преобразуя строковые поля в нужные типы
3. Если поле `real_user_id` не указано, создается AI-персонаж
4. Преобразование метаданных из JSON-строки в формат для хранения
5. Вызов метода репозитория для записи данных в БД
6. Возврат созданного персонажа в формате protobuf

Пример кода сервиса для создания персонажа:

```go
// Файл: services/character-service/internal/service/character_service.go
func (s *CharacterService) CreateCharacter(ctx context.Context, req *pb.CreateCharacterRequest) (*pb.Character, error) {
    logger.Logger.Info("Creating character",
        zap.String("world_id", req.WorldId),
        zap.String("display_name", req.DisplayName))

    var meta json.RawMessage = []byte("{}")
    if req.Meta != nil && *req.Meta != "" {
        meta = json.RawMessage(*req.Meta)
    }

    var realUserID sql.NullString
    if req.RealUserId != nil && *req.RealUserId != "" {
        realUserID = sql.NullString{String: *req.RealUserId, Valid: true}
    }

    var avatarMediaID sql.NullString
    if req.AvatarMediaId != nil && *req.AvatarMediaId != "" {
        avatarMediaID = sql.NullString{String: *req.AvatarMediaId, Valid: true}
    }

    params := models.CreateCharacterParams{
        WorldID:       req.WorldId,
        RealUserID:    realUserID,
        DisplayName:   req.DisplayName,
        AvatarMediaID: avatarMediaID,
        Meta:          meta,
    }

    character, err := s.repo.CreateCharacter(ctx, params)
    if err != nil {
        logger.Logger.Error("Failed to create character", zap.Error(err))
        return nil, status.Error(codes.Internal, "Failed to create character")
    }

    return characterModelToProto(character), nil
}
```

### Получение информации о персонажах

Character Service предоставляет два основных метода для получения информации о персонажах:

1. Получение конкретного персонажа по его ID:
   ```go
   // Файл: services/character-service/internal/service/character_service.go
   func (s *CharacterService) GetCharacter(ctx context.Context, req *pb.GetCharacterRequest) (*pb.Character, error)
   ```

2. Получение всех персонажей пользователя в конкретном мире:
   ```go
   // Файл: services/character-service/internal/service/character_service.go
   func (s *CharacterService) GetUserCharactersInWorld(ctx context.Context, req *pb.GetUserCharactersInWorldRequest) (*pb.CharacterList, error)
   ```

Оба метода вызывают соответствующие функции репозитория для получения данных из базы и преобразуют результаты в protobuf-сообщения для возврата клиенту.

### AI-персонажи

Character Service поддерживает работу с AI-сгенерированными персонажами:

1. AI-персонаж определяется отсутствием связи с реальным пользователем (поле `real_user_id` равно NULL)
2. Флаг `is_ai` автоматически генерируется на уровне базы данных
3. Метаданные AI-персонажей хранятся в JSONB-поле `meta` и могут содержать информацию о внешности, характере, интересах и т.д.

В отличие от персонажей реальных пользователей, AI-персонажи имеют более богатые метаданные, которые используются для генерации контента. Структура метаданных не фиксирована и может адаптироваться под разные типы персонажей и миров.

## Технические детали

### База данных

Character Service использует PostgreSQL для хранения данных о персонажах. Основная таблица:

```sql
-- Файл: scripts/schema.sql
CREATE TABLE IF NOT EXISTS world_user_characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    real_user_id UUID REFERENCES users(id) ON DELETE SET NULL,    -- NULL => AI-NPC
    is_ai BOOLEAN GENERATED ALWAYS AS (real_user_id IS NULL) STORED,
    display_name TEXT NOT NULL,
    avatar_media_id UUID,
    meta JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_world_user_characters_real_user_id ON world_user_characters(real_user_id);
CREATE INDEX IF NOT EXISTS idx_world_user_characters_world_id ON world_user_characters(world_id);
CREATE INDEX IF NOT EXISTS idx_world_user_characters_real_user_id_world_id ON world_user_characters(real_user_id, world_id);
CREATE INDEX IF NOT EXISTS idx_world_user_characters_is_ai ON world_user_characters(is_ai);
```

Ключевые особенности схемы:
- `real_user_id` - NULL для AI-персонажей
- `is_ai` - вычисляемое поле, TRUE если `real_user_id` равно NULL
- `meta` - JSONB-поле для хранения дополнительных метаданных персонажа
- Индексы для ускорения запросов по миру, пользователю, и типу персонажа (AI/реальный)

### Интеграция с другими сервисами

Character Service интегрируется с другими микросервисами:

1. **World Service** - Проверка существования мира
2. **Media Service** - Работа с аватарами персонажей
3. **Auth Service** - Проверка существования пользователя

Интеграция осуществляется через gRPC-клиенты. В текущей реализации прямая проверка связи с World Service отсутствует, но в будущем может быть добавлена для проверки доступа пользователя к миру.

### Обработка метаданных

Character Service использует гибкий подход к хранению метаданных персонажей через JSONB:

1. Метаданные передаются в формате JSON-строки через gRPC API
2. Перед сохранением в БД выполняется валидация JSON
3. Метаданные хранятся в нативном JSONB формате PostgreSQL
4. При получении персонажа метаданные конвертируются обратно в JSON-строку

Этот подход позволяет хранить различные наборы атрибутов для разных типов персонажей, не изменяя схему базы данных.

## Настройка и запуск

### Переменные окружения

Для работы Character Service требуется настроить следующие переменные окружения:

```
# Общие настройки
SERVICE_NAME=character-service

# База данных
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=generia
POSTGRES_SSL_MODE=disable

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500

# Логирование
LOG_LEVEL=info
```

### Запуск сервиса

Character Service запускается как часть общей инфраструктуры Generia через docker-compose:

```bash
# Файл: docker-compose.yml
docker-compose up -d character-service
```

Для локальной разработки сервис можно запустить отдельно:

```bash
cd services/character-service
go run cmd/main.go
```

Сервис по умолчанию слушает порт 8089 для gRPC-запросов.

## Примеры использования

### Создание персонажа

```go
// gRPC-клиент
conn, err := grpc.Dial("localhost:8089", grpc.WithInsecure())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

client := pb.NewCharacterServiceClient(conn)

// Пример метаданных
metadata := map[string]interface{}{
    "interests":      []string{"AI", "Virtual Worlds", "Gaming"},
    "personality":    "Outgoing and friendly",
    "speaking_style": "Casual with tech jargon",
}
metadataBytes, _ := json.Marshal(metadata)
metadataStr := string(metadataBytes)

// Создание персонажа для реального пользователя
userID := "user-id-123"
response, err := client.CreateCharacter(context.Background(), &pb.CreateCharacterRequest{
    WorldId:     "world-id-456",
    RealUserId:  &userID,
    DisplayName: "Tech Explorer",
    Meta:        &metadataStr,
})

if err != nil {
    log.Fatalf("Failed to create character: %v", err)
}

log.Printf("Character created: ID=%s, DisplayName=%s", response.Id, response.DisplayName)
```

### Получение персонажей пользователя в мире

```go
// Запрос на получение персонажей пользователя в мире
response, err := client.GetUserCharactersInWorld(context.Background(), &pb.GetUserCharactersInWorldRequest{
    UserId:  "user-id-123",
    WorldId: "world-id-456",
})

if err != nil {
    log.Fatalf("Failed to get user characters: %v", err)
}

log.Printf("Found %d characters for user", len(response.Characters))
for _, character := range response.Characters {
    log.Printf("- %s (AI: %t)", character.DisplayName, character.IsAi)
}
```