## Разделение реальных и вымышленных 


---

### 1. Новые сущности и таблицы

| Название | Хранится где | Цель |
|----------|--------------|------|
| **`world_user_profiles`** | PostgreSQL, шардируется по `world_id` | Связывает глобальный аккаунт с его «персонажем» в конкретном мире либо описывает AI‑NPC (у таких строк `real_user_id IS NULL`). |
| **`world_memberships`** | та же партиция, что и `world_user_profiles` | Явно фиксирует факт «приглашён / вступил» + роль внутри мира. |
| *(уже есть)* **`posts`** | партиционирована по `world_id` | Добавляем колонку `author_world_user_id` (FK → `world_user_profiles.id`). |

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
    UNIQUE (world_id, real_user_id)  -- один персонаж / мир
);
```

**Индексы**

```sql
CREATE INDEX ON world_user_profiles (real_user_id, world_id);
CREATE INDEX ON world_user_profiles (world_id);
```

#### 1.2  `world_memberships`

```sql
CREATE TABLE world_memberships (
    world_id      UUID,
    real_user_id  UUID,
    role          TEXT CHECK (role IN ('owner','moderator','member')),
    joined_at     TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (world_id, real_user_id)
);
CREATE INDEX ON world_memberships (real_user_id);
```

---

### 2. Граф запросов

#### 2.1  «Мои миры»

```sql
WITH my_profiles AS (
    SELECT world_id
    FROM world_user_profiles
    WHERE real_user_id = :me
)
SELECT w.world_id,
       w.title,
       COALESCE(m.role, 'guest') AS my_role,   -- guest = только писал
       GREATEST(
            MAX(p.created_at),                 -- моя последняя активность
            MAX(COALESCE(m.joined_at, 'epoch'))
       ) AS last_activity
FROM worlds w
LEFT JOIN world_memberships m
       ON m.world_id = w.world_id AND m.real_user_id = :me
LEFT JOIN posts p
       ON p.world_id = w.world_id
      AND p.author_world_user_id IN (
              SELECT id
              FROM world_user_profiles
              WHERE real_user_id = :me
            )
WHERE w.world_id IN (
      SELECT world_id FROM my_profiles
      UNION
      SELECT world_id FROM world_memberships WHERE real_user_id = :me
)
GROUP BY w.world_id, w.title, my_role
ORDER BY last_activity DESC
LIMIT :limit OFFSET :offset;
```

* **Зачем `UNION`** — объединяем «меня пригласили» и «я что‑то уже писал, но официально не вступал».  
* **Скорость** — все условия бьют по индексам `real_user_id`, `author_world_user_id`, поэтому запрос O(log N).

#### 2.2  «Мои посты в мире X»

```sql
SELECT *
FROM posts
WHERE world_id = :world
  AND author_world_user_id IN (
        SELECT id
        FROM world_user_profiles
        WHERE real_user_id = :me
          AND world_id     = :world
      )
ORDER BY created_at DESC, id DESC
LIMIT :limit OFFSET :offset;
```

---

### 3. API‑контракты

#### 3.1  REST

```http
GET /api/v1/me/worlds
  → 200 OK  [{world}, …]

GET /api/v1/worlds/{world_id}/my-posts?cursor=...
  → 200 OK  {posts:[…], next_cursor:"..."}
```

#### 3.2  gRPC / protobuf

```proto
service MeService {
  rpc ListMyWorlds(google.protobuf.Empty)
      returns (ListMyWorldsResp);
}

message ListMyWorldsResp { repeated WorldBrief worlds = 1; }

service PostService {
  rpc ListUserPostsInWorld(ListUserPostsReq)
      returns (PostPage);
}

message ListUserPostsReq {
  string  world_id = 1;
  string  cursor   = 2;  // created_at__id
  uint32  limit    = 3;  // <=100
}

message PostPage {
  repeated Post posts = 1;
  string next_cursor  = 2;
}
```

* `WorldBrief` = {id, slug, title, my_role, last_activity}.

---

### 4. Сервисы и потоки событий

| Компонент | Задача | Вход | Выход |
|-----------|--------|------|-------|
| **ACL‑service** | Проверяет доступ, кеширует правила | JWT + world_id | PERMIT / DENY |
| **Feed‑service** | Личная лента | Kafka `post.created` | Redis‑ключи `user:{id}:worlds`, `feed:{user}` |
| **Search‑index** | full‑text | Kafka `post.created`, `post.deleted` | OpenSearch |
| **Wallet‑service** | кредитный счёт | Stripe‑Webhook → Kafka | txn‑journal |

**Kafka topics**

```
world_membership.created
world_membership.deleted
post.created
post.visibility_changed
```

ACL‑service и Feed‑service на них инвалидируют кеш.

---

### 5. Кеш‑стратегия

| Ключ | TTL / инвалидация | Содержимое |
|------|-------------------|------------|
| `user:{id}:worlds` | event‑driven | список world_id |
| `world:{id}:profiles:{user}` | 10 мин | `profile_id` текущего пользователя в мире |
| `feed:{user}` | 30 сек, LRU | первые N постов ленты |

---

### 6. UI/UX wireframes

1. **/my-worlds**  
   - Grid карточек (обложка, название, роль‑бейдж, «последняя активность N h ago»).  
   - Фильтр «Показывать: все / только публичные / только приватные / гостевые».

2. **/world/{slug}**  
   - Переключатель «👥 Все» / «👤 Мои».  
   - Если у пользователя >1 персонажа, выпадающий список «От лица кого смотреть».

3. **Empty state**  
   - Если «Мои миры» пустой → CTA «Создай мир» / «Прими приглашение по ссылке».

---

### 7. Безопасность и производительность

* **Изоляция данных**  
  - Auth‑DB (реальные пользователи) живёт в отдельном VPC; только Auth‑service имеет соединение.  
  - Контент‑БД (worlds, posts) — шард‑кластер, доступ через Post‑, World‑ и Interaction‑service.

* **Партиционирование**  
  - `world_id` — routing‑key для соединений и Kafka‑partition; все запросы внутри одной партиции узла снимают нагрузку с межузловых JOIN‑ов.

* **Stale‑read‑replicas**  
  - End‑поинт `GET /me/worlds` допускает 100–200 ms задержку: читаем из hot‑standby реплики.

---

### 8. Дальнейшие улучшения

| Идея | Польза |
|------|--------|
| **GraphQL‑шлюз** | Один round‑trip, клиенты сами выбирают поля |
| **Realtime WebSocket** | push‑уведомления о новых постах в «моём» мире |
| **Edge‑кеш** CDN | для public‑картинок и обложек миров — экономия исходящего трафика |
| **Темпоральная БД** | хранить историю членства (когда вступил/вышел), удобно для RP‑логов |
| **А/B‑эксперименты** | отслеживай, как часто заходят на /my-worlds → метрики монетизации |

---

## Результат

*Пользователь* всегда видит:

1. **Полный список миров**, в которых он **приглашён или хоть раз публиковался**.  
2. Внутри любого мира — **фильтр своих постов**, работающий без тяжёлых JOIN‑ов.  
3. Всё кешируется, а права доступа проверяются одним RPC‑запросом в ACL‑service.  

Дизайн остаётся совместимым с текущей микросервисной схемой и легко масштабируется до десятков миллионов миров и пользователей.