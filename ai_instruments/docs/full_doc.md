# Полная документация проекта Generia

*Автоматически сгенерированный документ*

## Содержание

1. [Общая архитектура](#общая-архитектура)
2. [Микросервисы](#микросервисы)
3. [API](#api)
4. [Фронтенд](#фронтенд)
5. [Базы данных](#базы-данных)
6. [Инфраструктура](#инфраструктура)


## Общая Архитектура проекта Generia {#общая-архитектура}

## Обзор

Generia представляет собой клон Instagram, построенный на микросервисной архитектуре. Проект разработан с использованием современных технологий и паттернов проектирования для обеспечения масштабируемости, отказоустойчивости и удобства разработки.

## Архитектурный паттерн

- **Микросервисная архитектура** с четким разделением ответственности между сервисами
- **API Gateway** как единая точка входа для всех внешних запросов
- **gRPC** для внутренней коммуникации между сервисами
- **REST API** для взаимодействия с внешними клиентами

## Инфраструктура

- **Service Discovery** с использованием Consul для регистрации и обнаружения сервисов
- **Распределенная трассировка** с использованием Jaeger
- **Мониторинг** с использованием Prometheus и Grafana
- **Обмен событиями** с использованием Kafka
- **Контейнеризация** с использованием Docker и Docker Compose

## Базы данных

- **PostgreSQL** для хранения данных пользователей, постов и медиафайлов
- **Redis** для кэширования и управления лентой новостей
- **MongoDB** для хранения данных взаимодействий пользователей
- **MinIO** для хранения пользовательских медиафайлов

## Диаграмма архитектуры

```
+-------------------+
|                   |
|     Клиенты       |
|  (Web, Mobile)    |
|                   |
+--------+----------+
         |
         | HTTP/REST
         v
+--------+----------+
|                   |
|   API Gateway     |
|                   |
+--------+----------+
         |
         | gRPC
         v
+--------+----------+       +-------------------+       +-------------------+
|                   |       |                   |       |                   |
|  Auth Service     +------>+  Post Service     +------>+  Media Service    |
|                   |       |                   |       |                   |
+-------------------+       +--------+----------+       +-------------------+
                                     |
                                     |
                                     v
                            +--------+----------+       +-------------------+
                            |                   |       |                   |
                            | Interaction       +------>+  Feed Service     |
                            | Service           |       |                   |
                            +--------+----------+       +-------------------+
                                     |
                                     |
                                     v
                            +--------+----------+       +-------------------+
                            |                   |       |                   |
                            |  Cache Service    +------>+  CDN Service      |
                            |                   |       |                   |
                            +-------------------+       +-------------------+
```

## Масштабируемость и отказоустойчивость

- Каждый сервис может быть масштабирован горизонтально независимо от других
- Использование Service Discovery позволяет динамически обнаруживать экземпляры сервисов
- Применение паттерна Circuit Breaker для предотвращения каскадных отказов
- Кэширование для снижения нагрузки на базу данных и повышения производительности

## Технологический стек

- **Языки программирования**: Go (бэкенд), TypeScript/React (фронтенд)
- **Коммуникация**: gRPC, REST
- **Базы данных**: PostgreSQL, Redis
- **Контейнеризация**: Docker, Docker Compose
- **Мониторинг и трассировка**: Prometheus, Jaeger
- **Service Discovery**: Consul

---

## Микросервисы проекта Generia {#микросервисы-проекта}

## Обзор

Generia реализована как набор взаимодействующих микросервисов, каждый из которых выполняет определенную функцию. Взаимодействие между сервисами осуществляется через gRPC, а внешний интерфейс предоставляется через REST API.

## API Gateway (Порт 8080)

**Описание**: Служит единой точкой входа для всех клиентских запросов.

**Функциональность**:
- Маршрутизация запросов к соответствующим сервисам
- Аутентификация и авторизация запросов
- Реализация REST API для внешних клиентов
- Трансляция REST запросов в gRPC вызовы

**Технологии**:
- Go
- gorilla/mux для HTTP маршрутизации
- JWT для аутентификации
- Middleware для логирования, восстановления и CORS

**Файлы**:
- `/services/api-gateway/cmd/main.go` - Точка входа
- `/services/api-gateway/handlers/` - Обработчики запросов
- `/services/api-gateway/middleware/` - Промежуточные слои обработки

## Auth Service (Порт 8081)

**Описание**: Отвечает за аутентификацию и авторизацию пользователей.

**Функциональность**:
- Регистрация и аутентификация пользователей
- Управление JWT-токенами
- Хранение учетных данных пользователей

**Технологии**:
- Go
- gRPC
- PostgreSQL
- Библиотека jwt-go для работы с JWT-токенами

**Файлы**:
- `/services/auth-service/cmd/main.go` - Точка входа
- `/services/auth-service/internal/models/user.go` - Модель пользователя
- `/services/auth-service/internal/repository/user_repository.go` - Работа с БД
- `/services/auth-service/internal/service/auth_service.go` - Бизнес-логика

## Post Service (Порт 8082)

**Описание**: Управляет созданием и получением постов.

**Функциональность**:
- Создание и получение постов
- Хранение метаданных постов
- Взаимодействие с Media Service для работы с медиафайлами

**Технологии**:
- Go
- gRPC
- PostgreSQL

**Файлы**:
- `/services/post-service/cmd/main.go` - Точка входа
- `/services/post-service/internal/models/post.go` - Модель поста
- `/services/post-service/internal/repository/post_repository.go` - Работа с БД
- `/services/post-service/internal/service/post_service.go` - Бизнес-логика

## Media Service (Порт 8083)

**Описание**: Отвечает за управление медиафайлами.

**Функциональность**:
- Загрузка и обработка медиафайлов
- Генерация различных вариантов размеров изображений
- Хранение медиафайлов и их метаданных

**Технологии**:
- Go
- gRPC
- PostgreSQL для метаданных
- MinIO для хранения медиафайлов

**Файлы**:
- `/services/media-service/cmd/main.go` - Точка входа
- `/services/media-service/internal/models/media.go` - Модель медиафайла
- `/services/media-service/internal/repository/media_repository.go` - Работа с БД
- `/services/media-service/internal/service/media_service.go` - Бизнес-логика

## Interaction Service (Порт 8084)

**Описание**: Управляет взаимодействиями пользователей с контентом.

**Функциональность**:
- Управление лайками и комментариями
- Хранение данных взаимодействия
- Предоставление API для получения статистики взаимодействий

**Технологии**:
- Go
- gRPC
- MongoDB
- Kafka для событийной коммуникации

**Файлы**:
- `/services/interaction-service/cmd/main.go` - Точка входа
- `/services/interaction-service/internal/models/interaction.go` - Модель взаимодействия
- `/services/interaction-service/internal/repository/interaction_repository.go` - Работа с БД
- `/services/interaction-service/internal/service/interaction_service.go` - Бизнес-логика

## Feed Service (Порт 8085)

**Описание**: Формирует ленты новостей для пользователей.

**Функциональность**:
- Формирование ленты новостей для пользователей
- Кэширование данных ленты
- Оптимизация запросов для быстрой загрузки ленты

**Технологии**:
- Go
- gRPC
- Redis для кэширования

**Файлы**:
- `/services/feed-service/cmd/main.go` - Точка входа
- `/services/feed-service/internal/models/` - Модели данных
- `/services/feed-service/internal/repository/` - Работа с БД и кэшем
- `/services/feed-service/internal/service/` - Бизнес-логика

## Cache Service (Порт 8086)

**Описание**: Обеспечивает централизованное кэширование данных.

**Функциональность**:
- Централизованное кэширование данных
- Использование Redis для хранения кэша
- Предоставление API для работы с кэшем

**Технологии**:
- Go
- gRPC
- Redis

**Файлы**:
- `/services/cache-service/cmd/main.go` - Точка входа
- `/services/cache-service/internal/models/` - Модели данных
- `/services/cache-service/internal/repository/` - Работа с Redis
- `/services/cache-service/internal/service/` - Бизнес-логика

## CDN Service (Порт 8087)

**Описание**: Оптимизирует доставку медиаконтента.

**Функциональность**:
- Оптимизация доставки медиаконтента
- Генерация защищенных URL для доступа к медиа
- Управление TTL для кэширования контента

**Технологии**:
- Go
- gRPC

**Файлы**:
- `/services/cdn-service/cmd/main.go` - Точка входа
- `/services/cdn-service/internal/models/` - Модели данных
- `/services/cdn-service/internal/repository/` - Работа с хранилищем
- `/services/cdn-service/internal/service/` - Бизнес-логика

---

## API проекта Generia {#api-проекта}

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

---

## Фронтенд проекта Generia {#фронтенд-проекта}

## Обзор

Фронтенд Generia разработан с использованием React и TypeScript. Он обеспечивает удобный пользовательский интерфейс для взаимодействия с бэкенд-сервисами.

## Технологии

- **React 18+** - JavaScript-библиотека для создания пользовательских интерфейсов
- **TypeScript** - типизированный JavaScript для повышения надежности кода
- **React Router** - библиотека для маршрутизации в React-приложениях
- **Axios** - HTTP-клиент для выполнения запросов к API
- **Context API** - API React для управления глобальным состоянием приложения

## Структура фронтенда

```
/frontend
  /src
    /api
      - axios.ts         # Конфигурация HTTP-клиента
    /components
      - CreatePost.tsx   # Компонент создания поста
      - Feed.tsx         # Компонент отображения ленты
      - Login.tsx        # Компонент входа в систему
      - Navbar.tsx       # Компонент навигационной панели
      - Register.tsx     # Компонент регистрации
    /context
      - AuthContext.tsx  # Контекст для управления аутентификацией
    - App.tsx            # Корневой компонент приложения
    - index.tsx          # Точка входа в приложение
    - types.ts           # Типы TypeScript
  - package.json         # Зависимости и скрипты
  - tsconfig.json        # Конфигурация TypeScript
  - Dockerfile           # Инструкции для создания Docker-образа
  - nginx.conf           # Конфигурация Nginx
```

## Маршрутизация

В приложении реализованы следующие маршруты:

- `/` - Главная страница с лентой постов
- `/login` - Страница входа в систему
- `/register` - Страница регистрации
- `/create` - Страница создания поста

## Аутентификация

Аутентификация реализована с использованием JWT-токенов.

**AuthContext.tsx** управляет состоянием аутентификации и предоставляет следующие возможности:

- Вход пользователя
- Регистрация нового пользователя
- Выход из системы
- Проверка аутентификации пользователя
- Автоматическое добавление токена в заголовки запросов

## Компоненты

### Navbar.tsx

Навигационная панель, которая отображается на всех страницах приложения. Содержит ссылки на основные разделы и кнопку выхода из системы.

### Login.tsx

Форма входа в систему с валидацией полей и обработкой ошибок.

### Register.tsx

Форма регистрации нового пользователя с валидацией полей и обработкой ошибок.

### Feed.tsx

Компонент для отображения ленты постов. Обрабатывает пагинацию и подгрузку новых постов при прокрутке.

### CreatePost.tsx

Форма создания нового поста с возможностью загрузки изображений, добавления описания и отправки на сервер.

## Взаимодействие с API

Взаимодействие с бэкендом осуществляется через axios.ts, который настраивает HTTP-клиент Axios для работы с REST API.

```typescript
// axios.ts
import axios from 'axios';

// Use the API Gateway's address - using relative URL for better compatibility with proxy
const API_URL = '/api/v1';

const axiosInstance = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  // Add timeout to prevent hanging requests
  timeout: 10000,
});

// Intercept requests to add authorization token
axiosInstance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Add response interceptor to handle common errors
axiosInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Log errors for debugging
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

export default axiosInstance;
```

## Основные типы данных

В файле `types.ts` определены основные типы данных, используемые в приложении:

```typescript
// types.ts
export interface User {
  id: string;
  username: string;
  email: string;
  profile_image?: string;
  bio?: string;
  created_at: string;
}

export interface Post {
  id: string;
  user_id: string;
  user?: User;
  caption: string;
  media_urls: string[];
  likes_count: number;
  comments_count: number;
  created_at: string;
}

export interface Comment {
  id: string;
  post_id: string;
  user: User;
  text: string;
  created_at: string;
}

export interface Like {
  user: User;
  created_at: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  loading: boolean;
  error: string | null;
}
```

## Сборка и запуск

Фронтенд запускается в Docker-контейнере с использованием Nginx в качестве веб-сервера.

```dockerfile
# Dockerfile
FROM node:16-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

Nginx настроен для проксирования запросов к API на соответствующие бэкенд-сервисы:

```nginx
# nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/v1/ {
        proxy_pass http://api-gateway:8080/api/v1/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Базы данных проекта Generia {#базы-данных}

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

---

## Инфраструктура проекта Generia {#инфраструктура-проекта}

## Обзор

Инфраструктура проекта Generia построена на базе Docker и Docker Compose, что обеспечивает легкость развертывания и масштабирования. Все компоненты системы упакованы в Docker-контейнеры и оркестрируются с помощью Docker Compose.

## Docker Compose

Все сервисы и зависимости управляются через Docker Compose. Основной файл конфигурации находится в корне проекта - `docker-compose.yml`.

```yaml
version: '3.8'

services:
  # База данных PostgreSQL
  postgres:
    image: postgres:14-alpine
    container_name: generia-postgres
    restart: always
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=generia
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/schema.sql:/docker-entrypoint-initdb.d/schema.sql
    networks:
      - generia_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Redis для кэширования
  redis:
    image: redis:alpine
    container_name: generia-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - generia_network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
      
  # MongoDB для хранения данных взаимодействий
  mongodb:
    image: mongo:latest
    container_name: generia-mongodb
    restart: always
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
    volumes:
      - mongo_data:/data/db
    networks:
      - generia_network
    healthcheck:
      test: ["CMD", "mongosh", "--quiet", "--eval", "db.runCommand('ping').ok"]
      interval: 10s
      timeout: 10s
      retries: 5
      
  # MinIO для хранения медиафайлов
  minio:
    image: minio/minio
    container_name: generia-minio
    restart: always
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
    networks:
      - generia_network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
      
  # Kafka для обмена событиями
  kafka:
    image: bitnami/kafka:latest
    container_name: generia-kafka
    restart: always
    ports:
      - "9092:9092"
    environment:
      - KAFKA_CFG_NODE_ID=1
      - KAFKA_CFG_PROCESS_ROLES=controller,broker
      - KAFKA_CFG_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093
      - KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      - KAFKA_CFG_CONTROLLER_QUORUM_VOTERS=1@kafka:9093
      - KAFKA_CFG_CONTROLLER_LISTENER_NAMES=CONTROLLER
      - ALLOW_PLAINTEXT_LISTENER=yes
    volumes:
      - kafka_data:/bitnami/kafka
    networks:
      - generia_network

  # Service Discovery with Consul
  consul:
    image: consul:1.14
    ports:
      - "8500:8500"
    volumes:
      - consul_data:/consul/data
    command: "agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0"

  # Трассировка с Jaeger
  jaeger:
    image: jaegertracing/all-in-one:1.40
    ports:
      - "6831:6831/udp"
      - "16686:16686"

  # Prometheus для мониторинга
  prometheus:
    image: prom/prometheus:latest
    container_name: generia-prometheus
    restart: always
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - generia_network
      
  # Grafana для визуализации метрик
  grafana:
    image: grafana/grafana:latest
    container_name: generia-grafana
    restart: always
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - generia_network
    depends_on:
      - prometheus

  # API Gateway
  api-gateway:
    build:
      context: ./services/api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - PORT=8080

  # Auth Service
  auth-service:
    build:
      context: ./services/auth-service
    depends_on:
      - postgres
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - POSTGRES_URI=postgresql://generia:password@postgres:5432/generia?sslmode=disable
      - PORT=8081

  # Post Service
  post-service:
    build:
      context: ./services/post-service
    depends_on:
      - postgres
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - POSTGRES_URI=postgresql://generia:password@postgres:5432/generia?sslmode=disable
      - PORT=8082

  # Media Service
  media-service:
    build:
      context: ./services/media-service
    depends_on:
      - postgres
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - POSTGRES_URI=postgresql://generia:password@postgres:5432/generia?sslmode=disable
      - PORT=8083

  # Interaction Service
  interaction-service:
    build:
      context: ./services/interaction-service
    depends_on:
      - mongo
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - MONGO_URI=mongodb://mongo:27017/generia
      - PORT=8084

  # Feed Service
  feed-service:
    build:
      context: ./services/feed-service
    depends_on:
      - redis
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - REDIS_ADDR=redis:6379
      - PORT=8085

  # Cache Service
  cache-service:
    build:
      context: ./services/cache-service
    depends_on:
      - redis
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - REDIS_ADDR=redis:6379
      - PORT=8086

  # CDN Service
  cdn-service:
    build:
      context: ./services/cdn-service
    depends_on:
      - consul
      - jaeger
    environment:
      - CONSUL_ADDR=consul:8500
      - JAEGER_ADDR=jaeger:6831
      - PORT=8087

  # Frontend application
  frontend:
    build:
      context: ./frontend
    ports:
      - "80:80"
    depends_on:
      - api-gateway

volumes:
  postgres_data:
  redis_data:
  mongo_data:
  minio_data:
  kafka_data:
  grafana_data:
```

## Docker

Каждый сервис упакован в Docker-контейнер с использованием соответствующего Dockerfile.

### API Gateway Dockerfile

```dockerfile
FROM golang:1.21-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
```

### Frontend Dockerfile

```dockerfile
FROM node:16-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Service Discovery with Consul

Consul используется для регистрации и обнаружения сервисов, что позволяет микросервисам находить друг друга без необходимости знать их точные адреса.

```go
// pkg/discovery/consul.go
package discovery

import (
    "github.com/hashicorp/consul/api"
)

type ServiceDiscovery interface {
    Register(name, host string, port int, tags []string) error
    Deregister() error
    GetService(name string) (string, error)
}

type ConsulClient struct {
    client *api.Client
    serviceID string
}

func NewConsulClient(address string) (*ConsulClient, error) {
    config := api.DefaultConfig()
    config.Address = address
    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }
    return &ConsulClient{client: client}, nil
}

// Реализация методов...
```

## Мониторинг с Prometheus

Prometheus используется для сбора и хранения метрик о работе сервисов. Конфигурация Prometheus находится в файле `configs/prometheus.yml`.

```yaml
# configs/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'api-gateway'
    scrape_interval: 5s
    static_configs:
      - targets: ['api-gateway:8080']

  - job_name: 'auth-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['auth-service:8081']

  - job_name: 'post-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['post-service:8082']

  - job_name: 'media-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['media-service:8083']

  - job_name: 'interaction-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['interaction-service:8084']

  - job_name: 'feed-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['feed-service:8085']

  - job_name: 'cache-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['cache-service:8086']

  - job_name: 'cdn-service'
    scrape_interval: 5s
    static_configs:
      - targets: ['cdn-service:8087']
```

## Трассировка с Jaeger

Jaeger используется для распределенной трассировки запросов, что помогает понять поток запросов через различные микросервисы и обнаружить узкие места.

```go
// pkg/tracing/jaeger.go
package tracing

import (
    "io"

    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

// InitTracer создает новый трассировщик Jaeger
func InitTracer(serviceName, agentHostPort string) (opentracing.Tracer, io.Closer, error) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: agentHostPort,
        },
    }
    tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
    if err != nil {
        return nil, nil, err
    }
    opentracing.SetGlobalTracer(tracer)
    return tracer, closer, nil
}
```

## Логирование

Для централизованного логирования используется пакет zap от Uber.

```go
// pkg/logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Logger обертка над zap.Logger
type Logger struct {
    *zap.Logger
}

// NewLogger создает новый логгер
func NewLogger(serviceName string, debug bool) (*Logger, error) {
    var config zap.Config
    if debug {
        config = zap.NewDevelopmentConfig()
    } else {
        config = zap.NewProductionConfig()
    }

    logger, err := config.Build()
    if err != nil {
        return nil, err
    }

    logger = logger.With(zap.String("service", serviceName))
    return &Logger{logger}, nil
}

// Sync синхронизирует буферы логгера
func (l *Logger) Sync() error {
    return l.Logger.Sync()
}
```

## Запуск и остановка

Для запуска всех сервисов используется команда:

```bash
docker-compose up -d
```

Для остановки всех сервисов:

```bash
docker-compose down
```

Для просмотра логов конкретного сервиса:

```bash
docker-compose logs -f <service-name>
```

## Масштабирование

Микросервисы могут быть масштабированы горизонтально с помощью Docker Compose:

```bash
docker-compose up -d --scale auth-service=3 --scale post-service=3
```

В реальном производственном окружении для оркестрации контейнеров лучше использовать Kubernetes или аналогичные инструменты.

---
