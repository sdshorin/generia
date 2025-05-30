import uuid
from typing import Dict, Any, List
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..prompts import load_prompt, WORLD_IMAGE_PROMPT
from ..schemas import ImagePromptResponse

class GenerateWorldImageJob(BaseJob):
    """
    Задание для генерации изображения мира
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации изображения мира

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id

        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Формируем промпт для генерации промптов изображений
        prompt_template = load_prompt(WORLD_IMAGE_PROMPT)

        # Генерируем описание структуры ответа
        structure_description = model_to_template(ImagePromptResponse)

        # Подготавливаем переменные для промпта
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            world_description=world_description,
            structure_description=structure_description
        )

        # Генерируем промпты для изображений с помощью LLM
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного контента для промптов изображений
            image_prompts = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=ImagePromptResponse,
                temperature=0.7,
                task_id=self.task.id,
                world_id=world_id
            )

            logger.info(f"Generated image prompts for world {world_id}")

            # Генерируем изображения
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_images",
                    increment=2  # Два изображения: хэдер и иконка
                )

            # Генерируем фоновое изображение
            header_image = await self.image_generator.generate_image(
                prompt=image_prompts.header_prompt,
                world_id=world_id,
                media_type_enum=MediaType.WORLD_HEADER,
                character_id=None,  # No character for world-level images
                width=1024,
                height=512,
                task_id=self.task.id,
                filename=f"world_{world_id}_header.png",
                media_type="image/png"
            )

            logger.info(f"Generated header image for world {world_id}")

            # Генерируем иконку
            icon_image = await self.image_generator.generate_image(
                prompt=image_prompts.icon_prompt,
                world_id=world_id,
                media_type_enum=MediaType.WORLD_ICON,
                character_id=None,  # No character for world-level images
                width=512,
                height=512,
                task_id=self.task.id,
                filename=f"world_{world_id}_icon.png",
                media_type="image/png"
            )

            logger.info(f"Generated icon image for world {world_id}")

            # Получаем URL-адреса сгенерированных изображений
            header_url = header_image.get("image_url")
            header_id = header_image.get("media_id")
            icon_url = icon_image.get("image_url")
            icon_id = icon_image.get("media_id")

            # Обновляем статус этапа
            if self.progress_manager:
                await self.progress_manager.update_stage(
                    world_id=world_id,
                    stage=GenerationStage.WORLD_IMAGE,
                    status=GenerationStatus.COMPLETED
                )

            # Обновляем информацию о мире в World Service
            if self.service_client:
                await self.service_client.update_world_image(
                    world_id=world_id,
                    image_uuid=header_id,
                    icon_uuid=icon_id,
                    task_id=self.task.id
                )

                logger.info(f"Updated world images in World Service for world {world_id}, header_id: {header_id}, icon_id: {icon_id}")

            return {
                "header_prompt": image_prompts.header_prompt,
                "icon_prompt": image_prompts.icon_prompt,
                "header_url": header_url,
                "header_id": header_id,
                "icon_url": icon_url,
                "icon_id": icon_id,
                "style_reference": image_prompts.style_reference,
                "visual_elements": image_prompts.visual_elements,
                "mood": image_prompts.mood,
                "color_palette": image_prompts.color_palette
            }

        except Exception as e:
            logger.error(f"Error generating world images for world {world_id}: {str(e)}")
            raise

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated world images for world {self.task.world_id}. "
            f"Header URL: {result.get('header_url')}, Icon URL: {result.get('icon_url')}"
        )

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate world images: {str(error)}")

        # Обновляем статус этапа
        if self.progress_manager:
            await self.progress_manager.update_stage(
                world_id=self.task.world_id,
                stage=GenerationStage.WORLD_IMAGE,
                status=GenerationStatus.FAILED
            )