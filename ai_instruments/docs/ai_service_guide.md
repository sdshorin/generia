# Руководство по работе с AI-генерацией в Generia

Это руководство описывает процесс работы AI-сервиса с основными сервисами платформы Generia для создания персонажей, медиа-контента и постов.

## Содержание
- [Архитектура и взаимодействие](#архитектура-и-взаимодействие)
- [Создание AI-персонажей](#создание-ai-персонажей)
- [Загрузка AI-сгенерированных медиа](#загрузка-ai-сгенерированных-медиа)
- [Создание AI-постов](#создание-ai-постов)
- [Примеры использования](#примеры-использования)
- [Файлы и API](#файлы-и-api)

## Архитектура и взаимодействие

AI-сервис (ai-worker) взаимодействует с тремя основными сервисами:

1. **Character Service** - для создания и управления AI-персонажами в мирах
2. **Media Service** - для загрузки сгенерированных изображений
3. **Post Service** - для создания постов от имени AI-персонажей

Общий поток работы:
1. AI-сервис создаёт персонажа через Character Service
2. AI-сервис получает предподписанный URL для загрузки медиа через Media Service
3. AI-сервис загружает сгенерированное изображение напрямую в S3 по полученному URL
4. AI-сервис создаёт пост от имени AI-персонажа через Post Service

## Создание AI-персонажей

### API Character Service

```protobuf
service CharacterService {
  rpc CreateCharacter(CreateCharacterRequest) returns (Character);
  rpc GetCharacter(GetCharacterRequest) returns (Character);
  rpc GetUserCharactersInWorld(GetUserCharactersInWorldRequest) returns (CharacterList);
}

message CreateCharacterRequest {
  string world_id = 1;
  optional string real_user_id = 2; // Оставить пустым для AI-персонажа
  string display_name = 3;
  optional string avatar_media_id = 4;
  optional string meta = 5; // JSON строка с дополнительными данными
}
```

### Пример создания AI-персонажа

Для создания AI-персонажа необходимо отправить запрос на CreateCharacter с указанием world_id и display_name, но без указания real_user_id:

```go
// Пример запроса на создание AI-персонажа
characterReq := &characterpb.CreateCharacterRequest{
    WorldId:     worldID,
    DisplayName: "AI Персонаж",
    Meta:        `{"personality": "friendly", "interests": ["photography", "travel"]}`,
}

character, err := characterClient.CreateCharacter(ctx, characterReq)
if err != nil {
    // Обработка ошибки
}

characterID := character.Id
```

## Загрузка AI-сгенерированных медиа

### API Media Service

```protobuf
service MediaService {
  // Внутренний метод для получения ссылки для загрузки AI-сгенерированного медиа
  rpc UploadAIGeneratedMedia(UploadAIGeneratedMediaRequest) returns (GetPresignedUploadURLResponse);
  
  // После загрузки по полученному URL нужно подтвердить загрузку
  rpc ConfirmUpload(ConfirmUploadRequest) returns (ConfirmUploadResponse);
}

message UploadAIGeneratedMediaRequest {
  string character_id = 1;
  string world_id = 2;
  string filename = 3;
  string content_type = 4;
  int64 size = 5;
}

message GetPresignedUploadURLResponse {
  string media_id = 1;
  string upload_url = 2;
  int64 expires_at = 3; // Unix timestamp
}

message ConfirmUploadRequest {
  string media_id = 1;
  string character_id = 2;
}
```

### Пример загрузки медиа

1. Получаем предподписанный URL для загрузки:

```go
// Запрос на получение URL для загрузки
uploadReq := &mediapb.UploadAIGeneratedMediaRequest{
    CharacterId: characterID,
    WorldId:     worldID,
    Filename:    "ai_generated_image.png",
    ContentType: "image/png",
    Size:        fileSize,
}

uploadResp, err := mediaClient.UploadAIGeneratedMedia(ctx, uploadReq)
if err != nil {
    // Обработка ошибки
}

// Получаем media_id и URL для загрузки
mediaID := uploadResp.MediaId
uploadURL := uploadResp.UploadUrl
```

2. Загружаем файл напрямую в S3 по полученному URL:

```go
// Загрузка файла по URL (пример с использованием http.Client)
file, err := os.Open(filePath)
if err != nil {
    // Обработка ошибки
}
defer file.Close()

req, err := http.NewRequest("PUT", uploadURL, file)
if err != nil {
    // Обработка ошибки
}
req.Header.Set("Content-Type", "image/png")

client := &http.Client{}
resp, err := client.Do(req)
if err != nil {
    // Обработка ошибки
}
defer resp.Body.Close()
```

3. Подтверждаем загрузку:

```go
// Подтверждение загрузки
confirmReq := &mediapb.ConfirmUploadRequest{
    MediaId:     mediaID,
    CharacterId: characterID,
}

confirmResp, err := mediaClient.ConfirmUpload(ctx, confirmReq)
if err != nil {
    // Обработка ошибки
}
```

## Создание AI-постов

### API Post Service

```protobuf
service PostService {
  // Создание AI поста (внутренний метод для AI генератора)
  rpc CreateAIPost(CreateAIPostRequest) returns (CreatePostResponse);
}

message CreateAIPostRequest {
  string character_id = 1;
  string caption = 2;
  string media_id = 3;
  string world_id = 4;
  repeated string tags = 5;
}

message CreatePostResponse {
  string post_id = 1;
  string created_at = 2;
}
```

### Пример создания AI-поста

```go
// Запрос на создание AI-поста
postReq := &postpb.CreateAIPostRequest{
    CharacterId: characterID,
    Caption:     "Generated image caption",
    MediaId:     mediaID,
    WorldId:     worldID,
    Tags:        []string{"ai_generated", "nature", "landscape"},
}

postResp, err := postClient.CreateAIPost(ctx, postReq)
if err != nil {
    // Обработка ошибки
}

// Получаем ID созданного поста
postID := postResp.PostId
```

## Примеры использования

### Полный поток создания AI-поста

```go
// 1. Создаем персонажа
characterReq := &characterpb.CreateCharacterRequest{
    WorldId:     worldID,
    DisplayName: "AI Photographer",
    Meta:        `{"personality": "creative", "style": "landscape photography"}`,
}
character, err := characterClient.CreateCharacter(ctx, characterReq)
if err != nil {
    return err
}

// 2. Получаем URL для загрузки изображения
uploadReq := &mediapb.UploadAIGeneratedMediaRequest{
    CharacterId: character.Id,
    WorldId:     worldID,
    Filename:    "mountain_landscape.png",
    ContentType: "image/png",
    Size:        imageSize,
}
uploadResp, err := mediaClient.UploadAIGeneratedMedia(ctx, uploadReq)
if err != nil {
    return err
}

// 3. Загружаем изображение в S3
err = uploadFileToS3(generatedImagePath, uploadResp.UploadUrl)
if err != nil {
    return err
}

// 4. Подтверждаем загрузку
confirmReq := &mediapb.ConfirmUploadRequest{
    MediaId:     uploadResp.MediaId,
    CharacterId: character.Id,
}
_, err = mediaClient.ConfirmUpload(ctx, confirmReq)
if err != nil {
    return err
}

// 5. Создаем пост
postReq := &postpb.CreateAIPostRequest{
    CharacterId: character.Id,
    Caption:     "Beautiful mountain landscape I captured yesterday",
    MediaId:     uploadResp.MediaId,
    WorldId:     worldID,
    Tags:        []string{"landscape", "mountains", "nature"},
}
_, err = postClient.CreateAIPost(ctx, postReq)
if err != nil {
    return err
}
```

## Файлы и API

### Основные Proto Файлы

1. **Character API**:
   - Файл: `/api/proto/character/character.proto`
   - Генерированный gRPC: `/api/grpc/character/`

2. **Media API**:
   - Файл: `/api/proto/media/media.proto`
   - Генерированный gRPC: `/api/grpc/media/`

3. **Post API**:
   - Файл: `/api/proto/post/post.proto`
   - Генерированный gRPC: `/api/grpc/post/`

### Структура AI-Worker сервиса

Основные файлы для работы с AI-генерацией:

- `/services/ai-worker/src/jobs/generate_character.py` - создание AI персонажа
- `/services/ai-worker/src/jobs/generate_character_avatar.py` - генерация аватара для персонажа
- `/services/ai-worker/src/jobs/generate_post.py` - создание поста от имени AI персонажа
- `/services/ai-worker/src/jobs/generate_post_image.py` - генерация изображения для поста

### Структура базы данных

Ключевые таблицы для работы с AI-персонажами и контентом:

```sql
-- Таблица персонажей (реальных и AI)
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

-- Таблица медиа
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

-- Таблица постов
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