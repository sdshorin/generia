
У меня есть проект - аналог instagram на микросервисах.
Для него мне нужно реализовать загрузку медиа с помощью Direct Upload в s3/MinIO и создание новых постов от пользователей.

README.md - описание проекта

docker-compose.yml - тут можно посмотреть запуск minio и других микросервисов

services/media-service - тут находится прототип media-service
services/post-service - тут находится прототип post-service
scripts/schema.sql - схема бд (прототип)
api/proto - прото-схемы (прототип)
api/proto/media/pedia.proto (прототип)
api/proto/post/post.proto (прототип)

frontend/src/components/CreatePost.tsx - создание постов

ai_instruments/prompts/media_info.md - тут описаны сценарии, как в итоге должна выглядеть создание постов.
(Api gateway и система авторизации пользователей уже реализована)

напиши код, чтобы загрузка и получение медиа были реализованы как описано в ai_instruments/prompts/media_info.md



