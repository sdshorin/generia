# Микросервисы проекта Generia

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
