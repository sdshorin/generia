# Media Service для Generia

Media Service - микросервис, отвечающий за загрузку, хранение и доступ к медиа-файлам в проекте Generia. Сервис обрабатывает изображения для постов, аватары персонажей и другие медиа-файлы, используемые в виртуальных мирах.

## Оглавление

- [Обзор](#обзор)
- [Архитектура](#архитектура)
  - [Модели данных](#модели-данных)
  - [Репозитории](#репозитории)
  - [Сервисные слои](#сервисные-слои)
  - [gRPC API](#grpc-api)
- [Функциональность](#функциональность)
  - [Загрузка медиа-файлов](#загрузка-медиа-файлов)
  - [Получение медиа](#получение-медиа)
  - [Генерация вариантов изображений](#генерация-вариантов-изображений)
- [Технические детали](#технические-детали)
  - [База данных](#база-данных)
  - [Интеграция с MinIO](#интеграция-с-minio)
  - [Безопасность](#безопасность)
- [Настройка и запуск](#настройка-и-запуск)
  - [Переменные окружения](#переменные-окружения)
  - [Запуск сервиса](#запуск-сервиса)
- [Примеры использования](#примеры-использования)

## Обзор

Media Service обеспечивает управление всеми медиа-файлами в платформе Generia. Сервис отвечает за загрузку, хранение и доступ к изображениям, используемым в виртуальных мирах. Media Service интегрируется с MinIO для хранения файлов и предоставляет gRPC API для других сервисов и клиентских приложений.

Основные возможности:
- Предоставление предподписанных URL для прямой загрузки файлов в хранилище
- Подтверждение успешной загрузки файлов
- Генерация вариантов изображений разных размеров
- Предоставление URL для доступа к медиа-файлам
- Оптимизация изображений

## Архитектура

Media Service следует трехслойной архитектуре, характерной для микросервисов в проекте Generia:

### Модели данных

Основные модели данных в Media Service:

```go
// Файл: services/media-service/internal/models/media.go
type Media struct {
    ID          string    `db:"id" json:"id"`
    CharacterId string    `db:"character_id" json:"character_id"`
    Filename    string    `db:"filename" json:"filename"`
    ContentType string    `db:"content_type" json:"content_type"`
    Size        int64     `db:"size" json:"size"`
    BucketName  string    `db:"bucket" json:"bucket"`
    ObjectName  string    `db:"object_name" json:"object_name"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Файл: services/media-service/internal/models/media.go
type MediaVariant struct {
    ID        string    `db:"id" json:"id"`
    MediaID   string    `db:"media_id" json:"media_id"`
    Name      string    `db:"name" json:"name"`
    URL       string    `db:"url" json:"url"`
    Width     int32     `db:"width" json:"width"`
    Height    int32     `db:"height" json:"height"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}
```

### Репозитории

Слой репозиториев отвечает за взаимодействие с базой данных:

```go
// Файл: services/media-service/internal/repository/media_repository.go
type MediaRepository interface {
    CreateMedia(ctx context.Context, media *models.Media) error
    GetMediaByID(ctx context.Context, id string) (*models.Media, error)
    GetMediaVariants(ctx context.Context, mediaID string) ([]*models.MediaVariant, error)
    CreateMediaVariant(ctx context.Context, variant *models.MediaVariant) error
}
```

Репозиторий предоставляет методы для:
- Создания новых записей о медиа-файлах
- Получения информации о файле по ID
- Работы с вариантами изображений

### Сервисные слои

Сервисный слой реализует бизнес-логику:

```go
// Файл: services/media-service/internal/service/media_service.go
type MediaService struct {
    repo        repository.MediaRepository
    minioClient *minio.Client
    bucket      string
    logger      *zap.Logger
}
```

Сервисный слой предоставляет методы для:
- Создания записей о медиа-файлах
- Генерации предподписанных URL для загрузки
- Подтверждения загрузки файлов
- Получения информации о медиа-файлах
- Генерации вариантов изображений
- Получения URL для доступа к файлам

### gRPC API

Media Service предоставляет следующий gRPC-интерфейс:

```protobuf
// Файл: api/proto/media/media.proto
service MediaService {  
  // Получение предподписанного URL для прямой загрузки в хранилище
  rpc GetPresignedUploadURL(GetPresignedUploadURLRequest) returns (GetPresignedUploadURLResponse);
  
  // Подтверждение загрузки файла через предподписанный URL
  rpc ConfirmUpload(ConfirmUploadRequest) returns (ConfirmUploadResponse);
  
  // Получение информации о медиафайле
  rpc GetMedia(GetMediaRequest) returns (Media);
  
  // Получение URL для доступа к медиафайлу
  rpc GetMediaURL(GetMediaURLRequest) returns (GetMediaURLResponse);

  // Оптимизация изображения
  rpc OptimizeImage(OptimizeImageRequest) returns (OptimizeImageResponse);

  // Проверка здоровья сервиса
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

## Функциональность

### Загрузка медиа-файлов

Media Service реализует двухэтапный процесс загрузки файлов:

1. **Получение предподписанного URL для загрузки**:
   ```go
   // Файл: services/media-service/internal/service/media_service.go
   func (s *MediaService) GeneratePresignedPutURL(ctx context.Context, characterID, filename, contentType string, size int64) (*models.Media, string, time.Time, error)
   ```
   - Клиент запрашивает предподписанный URL для загрузки файла
   - Сервис создает запись о медиа-файле в базе данных
   - Генерирует предподписанный URL для загрузки в MinIO
   - Возвращает URL и идентификатор созданной записи

2. **Подтверждение загрузки файла**:
   ```go
   // Файл: services/media-service/internal/service/media_service.go
   func (s *MediaService) ConfirmMediaUpload(ctx context.Context, mediaID, characterID string) error
   ```
   - После загрузки файла клиент подтверждает успешную загрузку
   - Сервис проверяет наличие файла в хранилище
   - Верифицирует привязку к персонажу
   - Генерирует варианты изображений при необходимости

Преимущества такого подхода:
- Файлы загружаются напрямую в хранилище, минуя сервер приложения
- Снижается нагрузка на сервер
- Ускоряется процесс загрузки для пользователя

Пример взаимодействия при загрузке файла:

```go
// Файл: services/media-service/cmd/main.go
func (s *MediaService) GetPresignedUploadURL(ctx context.Context, req *mediapb.GetPresignedUploadURLRequest) (*mediapb.GetPresignedUploadURLResponse, error) {
    // Генерация предподписанного URL для загрузки
    media, presignedURL, expiresAt, err := mediaService.GeneratePresignedPutURL(
        ctx,
        req.CharacterId,
        req.Filename,
        req.ContentType,
        req.Size,
    )
    
    return &mediapb.GetPresignedUploadURLResponse{
        MediaId:   media.ID,
        UploadUrl: presignedURL,
        ExpiresAt: expiresAt.Unix(),
    }, nil
}
```

### Получение медиа

Media Service предоставляет два основных метода для получения информации о медиа-файлах:

1. **Получение метаданных файла**:
   ```go
   // Файл: services/media-service/cmd/main.go
   func (s *MediaService) GetMedia(ctx context.Context, req *mediapb.GetMediaRequest) (*mediapb.Media, error)
   ```
   - Получение основной информации о файле и его вариантах
   - Возвращает ID файла, данные о его размере, типе содержимого и т.д.
   - Предоставляет список доступных вариантов изображения

2. **Получение URL для доступа к файлу**:
   ```go
   // Файл: services/media-service/cmd/main.go
   func (s *MediaService) GetMediaURL(ctx context.Context, req *mediapb.GetMediaURLRequest) (*mediapb.GetMediaURLResponse, error)
   ```
   - Генерация временного URL для доступа к файлу
   - Поддерживает указание конкретного варианта изображения (original, thumbnail, medium)
   - Позволяет указать время жизни URL

Пример генерации URL:

```go
// Файл: services/media-service/internal/service/media_service.go
func (s *MediaService) GetPresignedURL(ctx context.Context, media *models.Media, variant string, expiresIn time.Duration) (string, time.Time, error) {
    // Определение объекта в зависимости от варианта
    var objectName string
    if variant == "original" {
        objectName = media.ObjectName
    } else {
        // Формирование имени для варианта изображения
        objectName = fmt.Sprintf("%s/%s_%s%s", media.CharacterId, media.ID, variant, filepath.Ext(media.Filename))
    }
    
    // Генерация предподписанного URL для получения файла
    url, err := s.minioClient.PresignedGetObject(ctx, media.BucketName, objectName, expiresIn, reqParams)
    
    expiresAt := time.Now().Add(expiresIn)
    return url.String(), expiresAt, nil
}
```

### Генерация вариантов изображений

Media Service поддерживает создание различных вариантов изображений для оптимизации отображения:

```go
// Файл: services/media-service/cmd/main.go
func (s *MediaService) OptimizeImage(ctx context.Context, req *mediapb.OptimizeImageRequest) (*mediapb.OptimizeImageResponse, error)
```

Варианты изображений включают:
- **Thumbnail** - маленькая миниатюра для списков и превью
- **Medium** - средний размер для отображения в ленте и профилях
- **Original** - оригинальное изображение без изменений

В реальной имплементации сервис бы:
1. Загружал оригинальное изображение из хранилища
2. Обрабатывал изображение для создания вариантов разных размеров
3. Загружал варианты обратно в хранилище
4. Сохранял информацию о вариантах в базе данных

Процесс генерации вариантов может быть запущен:
- Синхронно - при подтверждении загрузки файла
- Асинхронно - через отдельный API-метод или очередь задач

## Технические детали

### База данных

Media Service использует PostgreSQL для хранения метаданных о медиа-файлах:

```sql
-- Файл: scripts/schema.sql
CREATE TABLE media (
    id UUID PRIMARY KEY,
    character_id UUID NOT NULL,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    bucket TEXT NOT NULL,
    object_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE media_variants (
    id UUID PRIMARY KEY,
    media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT,
    width INTEGER,
    height INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_media_character_id ON media(character_id);
CREATE INDEX idx_media_variants_media_id ON media_variants(media_id);
```

### Интеграция с MinIO

Media Service использует MinIO для хранения самих файлов:

```go
// Файл: services/media-service/cmd/main.go
// Инициализация MinIO клиента
minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
    Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
    Secure: cfg.Minio.UseSSL,
})

// Проверка/создание бакета
exists, err := minioClient.BucketExists(ctx, cfg.Minio.Bucket)
if !exists {
    err = minioClient.MakeBucket(ctx, cfg.Minio.Bucket, minio.MakeBucketOptions{})
}
```

Взаимодействие с MinIO включает:
- Создание и проверку существования бакетов
- Генерацию предподписанных URL для загрузки и скачивания
- Проверку наличия файлов
- Работу с метаданными файлов

### Безопасность

Media Service реализует несколько уровней безопасности:

1. **Верификация персонажа**:
   ```go
   // Файл: services/media-service/internal/service/media_service.go
   if media.CharacterId != characterID {
       return fmt.Errorf("character ID mismatch")
   }
   ```
   - Проверка, что файл принадлежит указанному персонажу

2. **Случайные имена файлов**:
   ```go
   // Файл: services/media-service/internal/service/media_service.go
   func GenerateID() (string, error) {
       bytes := make([]byte, 16)
       if _, err := rand.Read(bytes); err != nil {
           return "", err
       }
       return hex.EncodeToString(bytes), nil
   }
   ```
   - Использование случайных идентификаторов для предотвращения перебора

3. **Временные URL**:
   - Предподписанные URL имеют ограниченное время действия
   - По истечении времени доступ по URL прекращается

4. **Проверка существования файлов**:
   ```go
   // Файл: services/media-service/internal/service/media_service.go
   _, err = s.minioClient.StatObject(ctx, media.BucketName, media.ObjectName, minio.StatObjectOptions{})
   if err != nil {
       return fmt.Errorf("failed to verify media in storage: %w", err)
   }
   ```
   - Проверка наличия файла в хранилище перед подтверждением загрузки

## Настройка и запуск

### Переменные окружения

Для работы Media Service требуется настроить следующие переменные окружения:

```
# Основные настройки сервиса
SERVICE_NAME=media-service
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8084

# База данных
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=generia
POSTGRES_SSL_MODE=disable

# MinIO конфигурация
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=generia-media
MINIO_USE_SSL=false

# Consul (Service Discovery)
CONSUL_ADDRESS=consul:8500

# Телеметрия
OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317
OTEL_SERVICE_NAME=media-service
```

### Запуск сервиса

Media Service запускается как часть общей инфраструктуры Generia через docker-compose:

```bash
# Файл: docker-compose.yml
docker-compose up -d media-service
```

Для локальной разработки сервис можно запустить отдельно:

```bash
cd services/media-service
go run cmd/main.go
```

## Примеры использования

### Загрузка файла через предподписанный URL

```go
// gRPC-клиент
conn, err := grpc.Dial("localhost:8084", grpc.WithInsecure())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer conn.Close()

client := mediapb.NewMediaServiceClient(conn)

// Шаг 1: Получить предподписанный URL для загрузки
urlResp, err := client.GetPresignedUploadURL(context.Background(), &mediapb.GetPresignedUploadURLRequest{
    CharacterId: "character-123",
    Filename:    "avatar.jpg",
    ContentType: "image/jpeg",
    Size:        1024 * 1024, // 1 MB
})

if err != nil {
    log.Fatalf("Failed to get upload URL: %v", err)
}

mediaID := urlResp.MediaId
uploadURL := urlResp.UploadUrl

log.Printf("Got upload URL: %s for media ID: %s", uploadURL, mediaID)

// Шаг 2: Загрузить файл напрямую по URL (клиентский код)
file, _ := os.Open("avatar.jpg")
defer file.Close()

req, _ := http.NewRequest("PUT", uploadURL, file)
req.Header.Set("Content-Type", "image/jpeg")

resp, err := http.DefaultClient.Do(req)
if err != nil {
    log.Fatalf("Failed to upload file: %v", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    log.Fatalf("Upload failed with status: %d", resp.StatusCode)
}

// Шаг 3: Подтвердить загрузку
confirmResp, err := client.ConfirmUpload(context.Background(), &mediapb.ConfirmUploadRequest{
    MediaId:     mediaID,
    CharacterId: "character-123",
})

if err != nil {
    log.Fatalf("Failed to confirm upload: %v", err)
}

log.Printf("Upload confirmed successfully")
for _, variant := range confirmResp.Variants {
    log.Printf("Variant: %s, URL: %s", variant.Name, variant.Url)
}
```

### Получение URL изображения

```go
// Получение URL для доступа к изображению
urlResp, err := client.GetMediaURL(context.Background(), &mediapb.GetMediaURLRequest{
    MediaId:   "media-123",
    Variant:   "thumbnail", // thumbnail, medium, original
    ExpiresIn: 3600, // URL действителен 1 час
})

if err != nil {
    log.Fatalf("Failed to get media URL: %v", err)
}

imageURL := urlResp.Url
log.Printf("Image URL: %s (expires at: %d)", imageURL, urlResp.ExpiresAt)
```