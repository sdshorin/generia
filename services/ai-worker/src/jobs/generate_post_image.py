import uuid
from typing import Dict, Any, List, Optional
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..prompts import load_prompt, POST_IMAGE_PROMPT
from ..schemas.post_image import PostImagePromptResponse

class GeneratePostImageJob(BaseJob):
    """
    Задание для генерации изображения к посту и публикации поста
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации изображения к посту и публикации

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        image_prompt = self.task.parameters.get("image_prompt", "")
        image_style = self.task.parameters.get("image_style", "")
        post_content = self.task.parameters.get("post_content", "")
        character_name = self.task.parameters.get("character_name", "")
        username = self.task.parameters.get("username", "")
        character_id = self.task.parameters.get("character_id", "")
        character_index = self.task.parameters.get("character_index", 0)
        post_index = self.task.parameters.get("post_index", 0)
        tags = self.task.parameters.get("tags", [])
        character_description = self.task.parameters.get("character_description", {})

        if not character_id:
            raise ValueError("Character ID is required to generate post image and publish post")

        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(POST_IMAGE_PROMPT)

        # Создаем описание структуры ответа из модели pydantic
        structure_description = model_to_template(PostImagePromptResponse)

        # Форматируем промпт с параметрами
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            world_description=world_description,
            character_name=character_name,
            appearance=character_description.get("appearance", ""),
            avatar_description=character_description.get("avatar_description", ""),
            avatar_style=character_description.get("avatar_style", ""),
            image_prompt=image_prompt,
            image_style=image_style,
            post_content=post_content,
            structure_description=structure_description
        )

        try:
            # Генерация оптимизированного промпта для изображения
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_LLM"
                )

            # Генерация структурированного ответа для промпта изображения
            image_prompt_response = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=PostImagePromptResponse,
                temperature=0.7,
                task_id=self.task.id,
                world_id=world_id
            )

            optimized_image_prompt = image_prompt_response.prompt

            logger.info(f"Generated optimized image prompt for post by {character_name} (char_id: {character_id}) in world {world_id}")

            # Генерируем изображение для поста
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_images"
                )

            # Генерируем изображение
            image_result = await self.image_generator.generate_image(
                prompt=optimized_image_prompt,
                world_id=world_id,
                media_type_enum=MediaType.POST_IMAGE,
                character_id=character_id,
                width=512,
                height=512,
                task_id=self.task.id,
                filename=f"post_{world_id}_{character_id}_{post_index}.png",
                media_type="image/png",
                # enhance_prompt=True
            )

            if not image_result or "media_id" not in image_result:
                raise ValueError("Failed to generate image for post")

            logger.info(f"Generated image for post by {character_name} (ID: {character_id}) in world {world_id}")

            # Получаем URL изображения и медиа ID
            image_url = image_result.get("image_url")
            media_id = image_result.get("media_id")

            # Создаем пост через API
            post_id = await self._create_post(
                character_id=character_id,
                media_id=media_id,
                post_content=post_content,
                world_id=world_id,
                tags=tags
            )

            if post_id:
                logger.info(f"Created post {post_id} with image for character {character_id} in world {world_id}")

                # Увеличиваем счетчик созданных постов
                if self.progress_manager:
                    await self.progress_manager.increment_task_counter(
                        world_id=world_id,
                        field="posts_created"
                    )

            return {
                "character_id": character_id,
                "character_name": character_name,
                "username": username,
                "image_url": image_url,
                "media_id": media_id,
                "optimized_prompt": optimized_image_prompt,
                "post_id": post_id
            }

        except Exception as e:
            logger.error(f"Error generating image for post by {character_name} in world {world_id}: {str(e)}")
            raise

    async def _create_post(
        self,
        character_id: str,
        media_id: str,
        post_content: str,
        world_id: str,
        tags: List[str] = None
    ) -> Optional[str]:
        """
        Создает пост через Post Service

        Args:
            character_id: ID персонажа
            media_id: ID медиа
            post_content: Текст поста
            world_id: ID мира
            tags: Теги поста

        Returns:
            ID созданного поста или None в случае ошибки
        """
        if not self.service_client:
            logger.warning("Service client not available, cannot create post")
            return None

        try:
            # Создаем пост с изображением
            post_result = await self.service_client.create_ai_post(
                character_id=character_id,
                caption=post_content,
                media_id=media_id,
                world_id=world_id,
                tags=tags or [],
                task_id=self.task.id
            )

            if not post_result or "post_id" not in post_result:
                logger.error(f"Failed to create post: {post_result}")
                return None

            return post_result["post_id"]

        except Exception as e:
            logger.error(f"Error creating post: {str(e)}")
            return None

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated image and published post by {result.get('character_name')} "
            f"(ID: {result.get('character_id')}) in world {self.task.world_id}"
        )

        # Проверяем, завершена ли генерация мира
        if self.progress_manager:
            # Получаем текущий статус
            world_status = await self.db_manager.world_generation_status_collection.find_one({"_id": self.task.world_id})

            if world_status:
                # Проверяем, все ли этапы завершены
                all_stages_completed = True
                for stage in world_status.get("stages", []):
                    if stage["status"] != GenerationStatus.COMPLETED:
                        all_stages_completed = False
                        break

                # Проверяем, достигнуто ли предсказанное количество постов
                posts_created = world_status.get("posts_created", 0)
                posts_predicted = world_status.get("posts_predicted", 0)

                # Если все этапы завершены или создано достаточно постов, обновляем общий статус
                if all_stages_completed or (posts_predicted > 0 and posts_created >= posts_predicted):
                    await self.progress_manager.update_progress(
                        world_id=self.task.world_id,
                        updates={
                            "status": GenerationStatus.COMPLETED,
                            "current_stage": GenerationStage.FINISHING
                        }
                    )

                    # Обновляем статус этапа
                    await self.progress_manager.update_stage(
                        world_id=self.task.world_id,
                        stage=GenerationStage.POSTS,
                        status=GenerationStatus.COMPLETED
                    )

                    await self.progress_manager.update_stage(
                        world_id=self.task.world_id,
                        stage=GenerationStage.FINISHING,
                        status=GenerationStatus.COMPLETED
                    )


    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate post image: {str(error)}")