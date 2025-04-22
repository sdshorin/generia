# Проделанные изменения в системе Generia

## 1. Изменения в базе данных

### Создание таблицы world_user_characters
Добавлена новая таблица `world_user_characters` в схему базы данных с UUID в качестве первичного ключа. Эта таблица связывает реальных пользователей с персонажами в мирах, а также хранит AI-персонажей:

```sql
CREATE TABLE IF NOT EXISTS world_user_characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    real_user_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- NULL => AI-NPC
    is_ai BOOLEAN GENERATED ALWAYS AS (real_user_id IS NULL) STORED,
    display_name TEXT NOT NULL,
    avatar_media_id UUID,
    meta JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Обновление таблицы posts
Таблица `posts` была обновлена для использования `character_id` вместо `user_id`:

```sql
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
```

### Обновление таблицы media
Таблица `media` была обновлена для использования `character_id` вместо `user_id`:

```sql
CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL,
    world_id UUID REFERENCES worlds(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    bucket TEXT NOT NULL,
    object_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## 2. Создание сервиса Character

Создан новый микросервис `character-service` для управления персонажами пользователей (как реальных, так и AI) в разных мирах:

### Структура сервиса
- `/services/character-service/cmd/main.go` - точка входа сервиса
- `/services/character-service/internal/models/character.go` - модель данных персонажа
- `/services/character-service/internal/repository/character_repository.go` - слой для работы с базой данных
- `/services/character-service/internal/service/character_service.go` - реализация логики сервиса
- `/services/character-service/Dockerfile` - конфигурация контейнера

### API сервиса (Proto)
Создан файл `/api/proto/character/character.proto` с определением API:

```proto
service CharacterService {
  // Создание профиля персонажа для реального пользователя или AI
  rpc CreateCharacter(CreateCharacterRequest) returns (Character);
  
  // Получение персонажа по ID
  rpc GetCharacter(GetCharacterRequest) returns (Character);
  
  // Получение персонажей пользователя в конкретном мире
  rpc GetUserCharactersInWorld(GetUserCharactersInWorldRequest) returns (CharacterList);
  
  // Проверка здоровья сервиса
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

## 3. Обновление Media сервиса

### API изменения
Обновлен протокол для работы с персонажами вместо пользователей:

```proto
message MediaMetadata {
  string character_id = 1;
  string world_id = 2;
  string filename = 3;
  string content_type = 4;
  int64 size = 5;
}

message Media {
  string media_id = 1;
  string character_id = 2;
  string world_id = 3;
  string filename = 4;
  string content_type = 5;
  int64 size = 6;
  repeated MediaVariant variants = 7;
  string created_at = 8;
}
```

## 4. Обновление Post сервиса

### API изменения
- Добавлен метод для создания AI-постов:
```proto
// Создание AI поста (внутренний метод для AI генератора)
rpc CreateAIPost(CreateAIPostRequest) returns (CreatePostResponse);
```

- Добавлен метод для получения постов по character_id:
```proto
// Получение постов по character_id
rpc GetCharacterPosts(GetCharacterPostsRequest) returns (PostList);
```

### Обновление полей в структурах
- Изменение `user_id` на `character_id` в структурах постов
- Добавление флага `is_ai` для определения источника поста

## 5. Интеграции между сервисами

- Post сервис теперь взаимодействует с Character сервисом для получения и проверки персонажей
- Изменена логика проверки принадлежности медиа - теперь проверяется соответствие character_id
- Добавлена автоматическая проверка, что AI-посты создаются только для AI-персонажей