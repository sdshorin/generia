"""
Activity functions для Temporal с dependency injection
Следует best practices для resource sharing
"""

import asyncio
from typing import Any, Dict, Optional, Callable
from temporalio import activity
from datetime import datetime, timezone
import uuid


from ..constants import GenerationStatus, GenerationStage
from ..schemas.world_description import WorldDescription
from ..db.models import Task


def serialize_result(result):
    """
    Сериализует результат activity с правильной обработкой datetime объектов
    """
    if result is None:
        return result

    if hasattr(result, "model_dump"):
        data = result.model_dump()
    elif isinstance(result, dict):
        data = result.copy()
    else:
        return result

    # Рекурсивно конвертируем datetime объекты в ISO строки
    def convert_datetime(obj):
        if isinstance(obj, dict):
            return {key: convert_datetime(value) for key, value in obj.items()}
        elif isinstance(obj, list):
            return [convert_datetime(item) for item in obj]
        elif hasattr(obj, "isoformat"):
            return obj.isoformat()
        else:
            return obj

    return convert_datetime(data)


def create_activity_functions(resource_manager) -> Dict[str, Callable]:
    """
    Создает activity functions с injected resources
    Следует паттерну dependency injection из TEMPORAL_RESOURCES.md
    """

    # ==================== PROMPT ACTIVITIES ====================

    @activity.defn(name="load_prompt")
    async def load_prompt(filename: str) -> str:
        """
        Загружает промпт из файла (deterministic activity)
        """
        try:
            import os

            # Используем абсолютный путь для deterministic behavior
            current_dir = os.path.dirname(os.path.abspath(__file__))
            prompts_dir = os.path.join(os.path.dirname(current_dir), "prompts")
            file_path = os.path.join(prompts_dir, filename)

            with open(file_path, "r", encoding="utf-8") as f:
                prompt_content = f.read()

            # # # logger.info(f"Successfully loaded prompt: {filename}")
            return prompt_content

        except Exception as e:
            error_msg = f"Failed to load prompt {filename}: {str(e)}"
            # logger.error(error_msg)
            raise ValueError(error_msg)

    # ==================== LLM ACTIVITIES ====================

    @activity.defn(name="generate_structured_content")
    async def generate_structured_content(
        prompt: str,
        response_schema_name: str,
        world_id: str,
        task_id: str,
        temperature: float = 0.8,
        max_output_tokens: int = 4096,
    ) -> Dict[str, Any]:
        """
        Генерирует структурированный контент с помощью LLM
        """
        async with resource_manager.llm_semaphore:
            try:
                info = activity.info()
                # # # logger.info(f"Activity {info.activity_type} - Generating LLM content for world {world_id}")

                # Получаем схему ответа по имени
                response_schema = resource_manager.get_schema_by_name(
                    response_schema_name
                )
                if not response_schema:
                    raise ValueError(f"Unknown response schema: {response_schema_name}")

                # Генерируем контент
                result = await resource_manager.llm_client.generate_structured_content(
                    prompt=prompt,
                    response_schema=response_schema,
                    temperature=temperature,
                    max_output_tokens=max_output_tokens,
                    task_id=task_id,
                    world_id=world_id,
                )

                # # # logger.info(f"Successfully generated LLM content for world {world_id}")

                return serialize_result(result)

            except Exception as e:
                error_msg = f"Error generating LLM content: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="enhance_prompt")
    async def enhance_prompt(
        original_prompt: str, enhancement_context: str, world_id: str
    ) -> str:
        """
        Улучшает промпт с помощью LLM
        """
        async with resource_manager.llm_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Enhancing prompt for world {world_id}")

                enhanced_prompt = await resource_manager.llm_client.enhance_prompt(
                    prompt=original_prompt, context=enhancement_context
                )

                return enhanced_prompt

            except Exception as e:
                error_msg = f"Error enhancing prompt: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    # ==================== PROGRESS ACTIVITIES ====================

    @activity.defn(name="initialize_world_generation")
    async def initialize_world_generation(
        world_id: str,
        users_count: int,
        posts_count: int,
        user_prompt: str,
        api_call_limits_llm: int = 1000,
        api_call_limits_images: int = 500,
    ) -> Dict[str, Any]:
        """
        Инициализирует статус генерации мира
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Initializing world generation for {world_id}")

                status = await resource_manager.db_manager.initialize_world_generation_status(
                    world_id=world_id,
                    users_count=users_count,
                    posts_count=posts_count,
                    user_prompt=user_prompt,
                    api_call_limits_llm=api_call_limits_llm,
                    api_call_limits_images=api_call_limits_images,
                )

                return serialize_result(status)

            except Exception as e:
                error_msg = f"Error initializing world generation: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="update_stage")
    async def update_stage(world_id: str, stage: str, status: str) -> Dict[str, Any]:
        """
        Обновляет статус этапа генерации
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Updating stage {stage} to {status} for world {world_id}")

                updated_status = (
                    await resource_manager.db_manager.update_world_generation_stage(
                        world_id=world_id, stage=stage, status=status
                    )
                )

                return serialize_result(updated_status) if updated_status else {}

            except Exception as e:
                error_msg = f"Error updating stage: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="increment_counter")
    async def increment_counter(
        world_id: str, field: str, increment: int = 1
    ) -> Dict[str, Any]:
        """
        Увеличивает счетчик в статусе генерации
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                activity.logger.debug(
                    f"Activity {info.activity_type} - Incrementing {field} by {increment} for world {world_id}"
                )

                updated_status = await resource_manager.db_manager.increment_world_generation_counter(
                    world_id=world_id, field=field, increment=increment
                )

                return serialize_result(updated_status) if updated_status else {}

            except Exception as e:
                # error_msg = f"Error incrementing counter: {str(e)}"
                # # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="increment_cost")
    async def increment_cost(
        world_id: str, cost_type: str, cost: float
    ) -> Dict[str, Any]:
        """
        Увеличивает стоимость генерации
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                # # logger.debug(f"Activity {info.activity_type} - Incrementing {cost_type} cost by {cost} for world {world_id}")

                updated_status = (
                    await resource_manager.db_manager.increment_world_generation_cost(
                        world_id=world_id, cost_type=cost_type, cost=cost
                    )
                )

                return serialize_result(updated_status) if updated_status else {}

            except Exception as e:
                # error_msg = f"Error incrementing cost: {str(e)}"
                # # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="update_progress")
    async def update_progress(world_id: str, updates: Dict[str, Any]) -> Dict[str, Any]:
        """
        Обновляет несколько полей в статусе генерации
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                # logger.debug(f"Activity {info.activity_type} - Updating progress for world {world_id}")

                updated_status = (
                    await resource_manager.db_manager.update_world_generation_progress(
                        world_id=world_id, updates=updates
                    )
                )

                return serialize_result(updated_status) if updated_status else {}

            except Exception as e:
                error_msg = f"Error updating progress: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    # ==================== DATABASE ACTIVITIES ====================

    @activity.defn(name="save_world_parameters")
    async def save_world_parameters(
        world_data: Dict[str, Any], world_id: str
    ) -> Dict[str, Any]:
        """
        Сохраняет параметры мира в базу данных
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Saving world parameters for {world_id}")

                # Создаем объект WorldDescription - используем UTC время
                # В activities можно использовать обычное время, т.к. это не workflow context
                now = datetime.now(timezone.utc)

                world_description = WorldDescription(
                    **world_data, id=world_id, created_at=now, updated_at=now
                )

                await resource_manager.db_manager.save_world_parameters(
                    world_description
                )

                # Also store params in world-service
                await resource_manager.service_client.update_world_params(
                    world_id=world_id,
                    params=world_data,
                    task_id=info.task_token.decode() if info.task_token else None,
                )

                # # logger.info(f"Successfully saved world parameters for {world_id}")

                return {"world_id": world_id, "saved": True}

            except Exception as e:
                error_msg = f"Error saving world parameters: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="get_world_parameters")
    async def get_world_parameters(world_id: str) -> Dict[str, Any]:
        """
        Получает параметры мира из базы данных
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()
                # logger.debug(f"Activity {info.activity_type} - Getting world parameters for {world_id}")

                world_params = await resource_manager.db_manager.get_world_parameters(
                    world_id
                )

                if world_params:
                    return serialize_result(world_params)
                else:
                    raise ValueError(f"World parameters not found for world {world_id}")

            except Exception as e:
                error_msg = f"Error getting world parameters: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    # ==================== IMAGE ACTIVITIES ====================

    @activity.defn(name="generate_image")
    async def generate_image(
        prompt: str,
        world_id: str,
        image_type: str,
        enhance_prompt: bool = True,
        character_id: str = "",
    ) -> Dict[str, Any]:
        """
        Генерирует изображение по промпту
        """
        async with resource_manager.image_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Generating {image_type} for world {world_id}")

                # Определяем media_type_enum на основе image_type
                from ..constants import MediaType

                if image_type == "world_header":
                    media_type_enum = MediaType.WORLD_HEADER
                elif image_type == "world_icon":
                    media_type_enum = MediaType.WORLD_ICON
                elif image_type == "character_avatar":
                    media_type_enum = MediaType.CHARACTER_AVATAR
                elif image_type == "post_image":
                    media_type_enum = MediaType.POST_IMAGE
                else:
                    media_type_enum = MediaType.UNKNOWN

                # Генерируем изображение
                result = await resource_manager.image_generator.generate_image(
                    prompt=prompt,
                    world_id=world_id,
                    media_type_enum=media_type_enum,
                    character_id=character_id,
                    enhance_prompt=enhance_prompt,
                )

                # # logger.info(f"Successfully generated {image_type} for world {world_id}")

                return {
                    "image_url": result.get("image_url"),
                    "media_id": result.get("media_id"),
                    "prompt_used": prompt,  # Возвращаем использованный промпт
                }

            except Exception as e:
                error_msg = f"Error generating {image_type}: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="upload_image_to_media_service")
    async def upload_image_to_media_service(
        image_data: bytes, filename: str, world_id: str, content_type: str = "image/png"
    ) -> Dict[str, Any]:
        """
        Загружает изображение в Media Service
        """
        async with resource_manager.grpc_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Uploading image {filename} for world {world_id}")

                # Используем media uploader для загрузки
                from ..utils.media_uploader import MediaUploader

                uploader = MediaUploader(resource_manager.service_client)

                result = await uploader.upload_image(
                    image_data=image_data, filename=filename, content_type=content_type
                )

                # # logger.info(f"Successfully uploaded image {filename}")

                return result

            except Exception as e:
                error_msg = f"Error uploading image: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    # ==================== SERVICE ACTIVITIES ====================

    @activity.defn(name="create_character")
    async def create_character(character_data: Dict[str, Any], world_id: str) -> str:
        """
        Создает персонажа через Character Service
        """
        async with resource_manager.grpc_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Creating character for world {world_id}")

                character_id, _ = (
                    await resource_manager.service_client.create_character(
                        world_id=world_id,
                        display_name=character_data.get(
                            "display_name", character_data.get("username", "Unknown")
                        ),
                        meta=character_data,
                        avatar_media_id=character_data.get("avatar_media_id"),
                        task_id=info.activity_id,
                    )
                )

                # # logger.info(f"Successfully created character {character_id}")

                return character_id

            except Exception as e:
                error_msg = f"Error creating character: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="create_post")
    async def create_post(
        post_data: Dict[str, Any], character_id: str, world_id: str
    ) -> str:
        """
        Создает пост через Post Service
        """
        async with resource_manager.grpc_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Creating post for character {character_id}")

                media_id = post_data.get("media_id", "")

                result = await resource_manager.service_client.create_ai_post(
                    character_id=character_id,
                    caption=post_data.get("content", ""),
                    media_id=media_id,
                    world_id=world_id,
                    tags=post_data.get("hashtags", []),
                    task_id=info.activity_id,
                )
                post_id = result["post_id"]

                # # logger.info(f"Successfully created post {post_id}")

                return post_id

            except Exception as e:
                error_msg = f"Error creating post: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="update_character_avatar")
    async def update_character_avatar(
        character_id: str, avatar_media_id: str
    ) -> Dict[str, Any]:
        """
        Обновляет аватар персонажа
        """
        async with resource_manager.grpc_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Updating avatar for character {character_id}")

                result = await resource_manager.service_client.update_character(
                    character_id=character_id, avatar_media_id=avatar_media_id
                )

                # # logger.info(f"Successfully updated character avatar")

                return result

            except Exception as e:
                error_msg = f"Error updating character avatar: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="update_world_image")
    async def update_world_image(
        world_id: str, header_media_id: str, icon_media_id: str
    ) -> bool:
        """
        Обновляет изображения мира (header и icon)
        """
        async with resource_manager.grpc_semaphore:
            try:
                info = activity.info()
                # # logger.info(f"Activity {info.activity_type} - Updating world images for world {world_id}")

                result = await resource_manager.service_client.update_world_image(
                    world_id=world_id,
                    image_uuid=header_media_id,
                    icon_uuid=icon_media_id,
                    task_id=info.activity_id,
                )

                # # logger.info(f"Successfully updated world images")

                return result

            except Exception as e:
                error_msg = f"Error updating world images: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    # ==================== TASK STORAGE ACTIVITIES ====================

    @activity.defn(name="create_task")
    async def create_task(
        task_type: str,
        world_id: str,
        parameters: Dict[str, Any],
        task_id: Optional[str] = None,
    ) -> str:
        """
        Создает новую задачу в MongoDB и возвращает task_id

        Args:
            task_type: Тип задачи
            world_id: ID мира
            parameters: Параметры задачи
            task_id: Опциональный ID задачи (если не указан, будет сгенерирован)

        Returns:
            ID созданной задачи
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()

                # Генерируем task_id если не указан
                if task_id is None:
                    task_id = str(uuid.uuid4())

                # Создаем объект задачи
                now = datetime.now(timezone.utc)
                task = Task(
                    id=task_id,
                    type=task_type,
                    world_id=world_id,
                    status="pending",
                    parameters=parameters,
                    created_at=now,
                    updated_at=now,
                )

                # Сохраняем в MongoDB
                created_task_id = await resource_manager.db_manager.create_task(task)

                # logger.info(f"Created task {created_task_id} of type {task_type} for world {world_id}")

                return created_task_id

            except Exception as e:
                error_msg = f"Error creating task: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="get_task")
    async def get_task(task_id: str) -> Dict[str, Any]:
        """
        Загружает задачу из MongoDB по ID

        Args:
            task_id: ID задачи

        Returns:
            Данные задачи в виде словаря
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()

                # Загружаем задачу из MongoDB
                task = await resource_manager.db_manager.get_task(task_id)

                if task is None:
                    raise ValueError(f"Task with ID {task_id} not found")

                # logger.info(f"Retrieved task {task_id} of type {task.type}")

                # Возвращаем данные задачи в сериализованном виде
                return serialize_result(task)

            except Exception as e:
                error_msg = f"Error getting task: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    @activity.defn(name="update_task")
    async def update_task(task_id: str, updates: Dict[str, Any]) -> Dict[str, Any]:
        """
        Обновляет задачу в MongoDB

        Args:
            task_id: ID задачи
            updates: Словарь с обновлениями

        Returns:
            Результат обновления
        """
        async with resource_manager.db_semaphore:
            try:
                info = activity.info()

                # Обновляем задачу в MongoDB
                success = await resource_manager.db_manager.update_task(
                    task_id, updates
                )

                if not success:
                    raise ValueError(f"Failed to update task {task_id}")

                # logger.info(f"Updated task {task_id}")

                return {"task_id": task_id, "updated": True}

            except Exception as e:
                error_msg = f"Error updating task: {str(e)}"
                # logger.error(f"Activity failed: {error_msg}")
                raise

    # ==================== FORMATTING ACTIVITIES ====================

    @activity.defn(name="format_world_description")
    async def format_world_description(world_params: Dict[str, Any]) -> str:
        """
        Форматирует описание мира для использования в промптах

        Args:
            world_params: Параметры мира в виде словаря

        Returns:
            Отформатированное описание мира
        """
        try:
            info = activity.info()
            # logger.debug(f"Activity {info.activity_type} - Formatting world description")

            # Импортируем функцию форматирования
            from ..utils.format_world import format_world_description as format_func

            # Создаем объект WorldDescription из параметров
            from ..schemas.world_description import WorldDescription

            world_description = WorldDescription(**world_params)

            # Форматируем описание
            formatted_description = format_func(world_description)

            # logger.debug(f"Successfully formatted world description")

            return formatted_description

        except Exception as e:
            error_msg = f"Error formatting world description: {str(e)}"
            # logger.error(f"Activity failed: {error_msg}")
            raise

    return {
        "load_prompt": load_prompt,
        "generate_structured_content": generate_structured_content,
        "enhance_prompt": enhance_prompt,
        "initialize_world_generation": initialize_world_generation,
        "update_stage": update_stage,
        "increment_counter": increment_counter,
        "increment_cost": increment_cost,
        "update_progress": update_progress,
        "save_world_parameters": save_world_parameters,
        "get_world_parameters": get_world_parameters,
        "generate_image": generate_image,
        "upload_image_to_media_service": upload_image_to_media_service,
        "create_character": create_character,
        "create_post": create_post,
        "update_character_avatar": update_character_avatar,
        "update_world_image": update_world_image,
        "create_task": create_task,
        "get_task": get_task,
        "update_task": update_task,
        "format_world_description": format_world_description,
    }
