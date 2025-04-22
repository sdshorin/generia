# Руководство по работе с AI-генерацией в Generia

Это руководство описывает процесс работы AI-сервиса с основными сервисами платформы Generia для создания персонажей, медиа-контента и постов.

## Архитектура и взаимодействие

AI-сервис (ai-worker) взаимодействует с тремя основными сервисами:

1. **Character Service** - для создания и управления AI-персонажами в мирах
2. **Media Service** - для загрузки сгенерированных изображений
3. **Post Service** - для создания постов от имени AI-персонажей

Общий поток работы:
1. AI-сервис создаёт персонажа через Character Service
2. AI-сервис получает предподписанный URL для загрузки медиа через Media Service
3. AI-сервис загружает сгенерированное изображение напрямую в S3 по полученному URL
4. AI-сервис проверяет, что загрузка была завершена успешно
5. AI-сервис создаёт пост от имени AI-персонажа через Post Service

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


// Character representation
message Character {
  string id = 1;
  string world_id = 2;
  optional string real_user_id = 3; // Empty for AI characters
  bool is_ai = 4;
  string display_name = 5;
  optional string avatar_media_id = 6;
  optional string meta = 7; // JSON string
  string created_at = 8;
}
```


## Загрузка AI-сгенерированных медиа
```protobuf
// MediaService предоставляет API для управления медиафайлами
service MediaService {  
  // Получение предподписанного URL для прямой загрузки в хранилище
  rpc GetPresignedUploadURL(GetPresignedUploadURLRequest) returns (GetPresignedUploadURLResponse);
  
  // Подтверждение загрузки файла через предподписанный URL
  rpc ConfirmUpload(ConfirmUploadRequest) returns (ConfirmUploadResponse);
  
}


message GetPresignedUploadURLRequest {
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

message ConfirmUploadResponse {
  bool success = 1;
  repeated MediaVariant variants = 2;
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




# Генерация изображений

Для генерации используется Runware API. Есть опциональный параметр генерации (по умолчанию false) - использовать автоматическое улучшение промпта.

The Python Runware SDK is used to run image inference with the Runware API, powered by the Runware inference platform. It can be used to generate imaged with text-to-image and image-to-image. It also allows the use of an existing gallery of models or selecting any model or LoRA from the CivitAI gallery. The API also supports upscaling, background removal, inpainting and outpainting, and a series of other ControlNet models.



To install the Python Runware SDK, use the following command:
`pip install runware`


To generate images using the Runware API, you can use the imageInference method of the Runware class. Here's an example:
```python
from runware import Runware, IImageInference

async def main() -> None:
    runware = Runware(api_key=RUNWARE_API_KEY)
    await runware.connect()

    request_image = IImageInference(
        positivePrompt="a beautiful sunset over the mountains",
        model="civitai:101055@128078",
        numberResults=4,
        negativePrompt="cloudy, rainy",
        height=512,
        width=512,
    )

    images = await runware.imageInference(requestImage=request_image)
    for image in images:
        print(f"Image URL: {image.imageURL}")
```

Enhancing Prompts

To enhance prompts using the Runware API, you can use the promptEnhance method of the Runware class. Here's an example:
```python
from runware import Runware, IPromptEnhance

async def main() -> None:
    runware = Runware(api_key=RUNWARE_API_KEY)
    await runware.connect()

    prompt = "A beautiful sunset over the mountains"
    prompt_enhancer = IPromptEnhance(
        prompt=prompt,
        promptVersions=3,
        promptMaxLength=64,
    )

    enhanced_prompts = await runware.promptEnhance(promptEnhancer=prompt_enhancer)
    for enhanced_prompt in enhanced_prompts:
        print(enhanced_prompt.text)
```

