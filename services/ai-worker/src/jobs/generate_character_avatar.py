import uuid
from typing import Dict, Any, List
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..prompts import load_prompt, CHARACTER_AVATAR_PROMPT
from ..schemas.character_avatar import CharacterAvatarPromptResponse

class GenerateCharacterAvatarJob(BaseJob):
    """
    Задание для генерации аватара персонажа
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации аватара персонажа

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        character_name = self.task.parameters.get("character_name", "")
        appearance_description = self.task.parameters.get("appearance_description", "")
        avatar_description = self.task.parameters.get("avatar_description", "")
        avatar_style = self.task.parameters.get("avatar_style", "")
        username = self.task.parameters.get("username", "")
        character_index = self.task.parameters.get("character_index", 0)
        character_id = self.task.parameters.get("character_id")

        if not character_id:
            raise ValueError("Character ID is required to generate avatar")

        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(CHARACTER_AVATAR_PROMPT)

        structure_description = model_to_template(CharacterAvatarPromptResponse)

        # Форматируем промпт с параметрами
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            world_description=world_description,
            character_name=character_name,
            appearance_description=appearance_description,
            avatar_description=avatar_description,
            avatar_style=avatar_style,
            structure_description=structure_description
        )

        # Генерируем оптимизированный промпт для создания аватара
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного ответа для аватара
            avatar_response = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=CharacterAvatarPromptResponse,
                temperature=0.7,
                task_id=self.task.id,
                world_id=world_id
            )

            optimized_avatar_prompt = avatar_response.prompt

            logger.info(f"Generated avatar prompt for character {character_name} in world {world_id}: {optimized_avatar_prompt}")

            # Генерируем изображение аватара
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_images"
                )

            # Получаем ID персонажа из параметров задачи
            character_id = self.task.parameters.get("character_id")

            # Генерируем изображение аватара
            avatar_image = await self.image_generator.generate_image(
                prompt=optimized_avatar_prompt,
                width=512,
                height=512,
                task_id=self.task.id,
                world_id=world_id,
                character_id=character_id,
                filename=f"avatar_{world_id}_{username}_{character_index}.png",
                media_type="image/png"
            )

            logger.info(f"Generated avatar image for character {character_name} in world {world_id}")

            # Получаем URL аватара и ID медиа
            avatar_url = avatar_image.get("image_url")
            avatar_id = avatar_image.get("media_id")

            try:
                # Обновляем персонажа с новым аватаром
                character_result = await self.service_client.update_character(
                    character_id=character_id,
                    avatar_media_id=avatar_id,
                    task_id=self.task.id
                )

                logger.info(f"Updated character {character_id} with avatar {avatar_id}")

                # Сохраняем результат обновления персонажа
                return {
                    "username": username,
                    "display_name": character_name,
                    "avatar_url": avatar_url,
                    "avatar_id": avatar_id,
                    "character_id": character_id,
                    "avatar_prompt": optimized_avatar_prompt
                }
            except Exception as e:
                logger.error(f"Error updating character with avatar: {str(e)}")

        except Exception as e:
            logger.error(f"Error generating avatar for character {character_name} in world {world_id}: {str(e)}")
            raise

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated avatar for character {result.get('display_name')} "
            f"(@{result.get('username')}) in world {self.task.world_id}"
        )

        # В реальной реализации здесь можно было бы обновить статус персонажа в БД

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate character avatar: {str(error)}")