## Полная архитектура поддержки «моих миров» и «моих постов»


---

### 1. Новые сущности и таблицы

| Название | Цель |
|----------|------|
| **`world_user_profiles`** |  Связывает глобальный аккаунт с его «персонажем» в конкретном мире либо описывает AI‑NPC (у таких строк `real_user_id IS NULL`). |
| **`world_memberships`** | Явно фиксирует факт «приглашён / вступил» + роль внутри мира. |
| *(уже есть)* **`posts`** |  Добавляем колонку `author_world_user_id` (FK → `world_user_profiles.id`). |

#### 1.1  `world_user_profiles`

```sql
CREATE TABLE world_user_profiles (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    world_id        UUID NOT NULL,
    real_user_id    UUID,                -- NULL ⇒ AI‑NPC
    is_ai           BOOLEAN GENERATED ALWAYS AS (real_user_id IS NULL) STORED,
    display_name    TEXT NOT NULL,
    avatar_media_id UUID,
    meta            JSONB,
    created_at      TIMESTAMPTZ DEFAULT now(),
    -- UNIQUE (world_id, real_user_id)  -- один персонаж / мир (нет, в будущем смогу существовать несколько первонажей  в одном мире)
);
```

**Индексы**

```sql
CREATE INDEX ON world_user_profiles (real_user_id, world_id);
CREATE INDEX ON world_user_profiles (world_id);
```

**Механика**

- Создаем новый сервис - character-service, который будет отвечать за аккаунты в виртуальных мирах
- Реальный пользователь может иметь 0…N персонажей в каждом мире.
- Реальный пользователь может участвовать в нескольких мирах

В микросервисе character-service:
- метод "получить или создать-и-получить профиль персонажа" для реальных пользователей
- метод "создать профиль ai-персонажа"
(больше методов пока нет)

Оба метода доступны только для gRPC внутри кластера.
- получение профиля - нужно для post сервиса, когда реальный пользователь хочет написать новый пост в мир
- создать профиль ai-персонажа - вызывается из ai-worker при создании персонажа (будет сделано позже)

service CharacterService {
  // создание профиля персонажа для реального пользователя или AI 
  rpc CreateCharacter(CreateCharacterRequest) returns (Character);
  
  // Получение персонажа по ID
  rpc GetCharacter(GetCharacterRequest) returns (Character);
  
  // Получение персонажей пользователя в конкретном мире
  rpc GetUserCharactersInWorld(GetUserCharactersInWorldRequest) returns (CharacterList);
  
  // Проверка здоровья сервиса
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}


При создании нового поста реальный пользователь может указать опциональное поле character_id - если его нет, то загружается любой персонаж этого пользователя. Если персонажей у пользователя нет - создается новый с ником == username
( в post service нужно проверять что character_id принадлежит этому пользователю )


