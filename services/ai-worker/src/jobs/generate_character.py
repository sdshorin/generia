import uuid
import json
from typing import Dict, Any, List, Optional
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..db.models import Task
from ..prompts import load_prompt, CHARACTER_DETAIL_PROMPT
from ..schemas import CharacterDetailResponse

class GenerateCharacterJob(BaseJob):
    """
    Задание для генерации персонажа и создания его в character-service
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации персонажа

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        character_concept = self.task.parameters.get("character_concept", "")
        role_in_world = self.task.parameters.get("role_in_world", "")
        personality_traits = self.task.parameters.get("personality_traits", [])
        interests = self.task.parameters.get("interests", [])
        posts_count = int(self.task.parameters.get("posts_count", 5))
        character_index = int(self.task.parameters.get("character_index", 0))

        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(CHARACTER_DETAIL_PROMPT)

        # Преобразуем списки в строки для промпта
        personality_traits_str = ", ".join(personality_traits) if isinstance(personality_traits, list) else personality_traits
        interests_str = ", ".join(interests) if isinstance(interests, list) else interests

        # Генерируем описание структуры ответа
        structure_description = model_to_template(CharacterDetailResponse)

        # Форматируем промпт с параметрами
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            world_description=world_description,
            character_concept=character_concept,
            role_in_world=role_in_world,
            personality_traits=personality_traits_str,
            interests=interests_str,
            structure_description=structure_description
        )

        # Генерируем детальное описание персонажа с помощью LLM
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного контента
            character_detail = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=CharacterDetailResponse,
                temperature=0.8,
                max_output_tokens=4096,
                task_id=self.task.id,
                world_id=world_id
            )

            logger.info(f"Generated character detail for world {world_id}: {character_detail.username}")

            # Создаем персонажа через Character Service
            character_id = await self._create_character(character_detail, world_id)

            if not character_id:
                raise ValueError("Failed to create character in Character Service")

            logger.info(f"Created character in Character Service: {character_id}")

            # Генерируем аватар
            # avatar_info = await self._generate_avatar(character_detail, character_id, world_id)

            # Создаем задачи для пакета постов
            tasks_to_create = []
            now = datetime.utcnow()

            # Задача для генерации пакета постов
            post_batch_task_id = str(uuid.uuid4())
            post_batch_task = Task(
                _id=post_batch_task_id,
                type=TaskType.GENERATE_POST_BATCH,
                world_id=world_id,
                status="pending",
                worker_id=None,
                parameters={
                    "character_id": character_id,
                    "character_name": character_detail.display_name,
                    "character_description": {
                        "username": character_detail.username,
                        "display_name": character_detail.display_name,
                        "bio": character_detail.bio,
                        "background_story": character_detail.background_story,
                        "personality": character_detail.personality,
                        "speaking_style": character_detail.speaking_style,
                        "common_topics": character_detail.common_topics,
                        "character_index": character_index
                    },
                    "posts_count": posts_count
                },
                created_at=now,
                updated_at=now,
                attempt_count=0
            )
            tasks_to_create.append({"task": post_batch_task})

            created_task_ids = await self.create_next_tasks(tasks_to_create)

            # Увеличиваем счетчик созданных пользователей
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="users_created"
                )

            # Передаем информацию о персонаже в результат
            return {
                "character_id": character_id,
                "username": character_detail.username,
                "display_name": character_detail.display_name,
                "bio": character_detail.bio,
                # "avatar_media_id":  avatar_info.get("media_id"),
                # "avatar_url": avatar_info.get("image_url"),
                "speaking_style": character_detail.speaking_style,
                "next_tasks": created_task_ids
            }

        except Exception as e:
            logger.error(f"Error generating character detail for world {world_id}: {str(e)}")
            raise

    async def _create_character(self, character_detail: CharacterDetailResponse, world_id: str) -> Optional[str]:
        """
        Создает персонажа через Character Service

        Args:
            character_detail: Детали персонажа
            world_id: ID мира

        Returns:
            ID созданного персонажа или None в случае ошибки
        """
        if not self.service_client:
            logger.warning("Service client not available, cannot create character")
            return None

        try:
            # Подготовка метаданных персонажа
            meta = {
                "bio": character_detail.bio,
                "background_story": character_detail.background_story,
                "personality": character_detail.personality,
                "interests": character_detail.interests,
                "speaking_style": character_detail.speaking_style,
                "appearance": character_detail.appearance,
                "common_topics": character_detail.common_topics,
                "username": character_detail.username
            }

            # Создаем персонажа
            character_id, character_response = await self.service_client.create_character(
                world_id=world_id,
                display_name=character_detail.display_name,
                meta=meta,
                task_id=self.task.id
            )

            if not character_id:
                logger.error(f"Failed to create character: {character_response}")
                return None

            return character_id

        except Exception as e:
            logger.error(f"Error creating character: {str(e)}")
            return None

    async def _generate_avatar(self, character_detail: CharacterDetailResponse, character_id: str, world_id: str) -> Dict[str, Any]:
        """
        Генерирует аватар для персонажа

        Args:
            character_detail: Детали персонажа
            character_id: ID персонажа
            world_id: ID мира

        Returns:
            Информация о сгенерированном аватаре
        """
        if not self.image_generator:
            logger.warning("Image generator not available, cannot generate avatar")
            return {}

        try:
            # Формируем промпт для аватара
            avatar_prompt = f"Portrait of {character_detail.display_name}. {character_detail.avatar_description}. Style: {character_detail.avatar_style if hasattr(character_detail, 'avatar_style') else 'photorealistic'}. {character_detail.appearance}."

            # Генерируем аватар
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_images"
                )

            avatar_result = await self.image_generator.generate_image(
                prompt=avatar_prompt,
                width=512,
                height=512,
                character_id=character_id,
                world_id=world_id,
                task_id=self.task.id,
                filename=f"avatar_{world_id}_{character_detail.username}.png",
                # enhance_prompt=True
            )

            if not avatar_result or "media_id" not in avatar_result:
                logger.error(f"Failed to generate avatar: {avatar_result}")
                return {}

            # Обновляем персонажа с новым аватаром
            if self.service_client and "media_id" in avatar_result:
                try:
                    # Логика обновления аватара в character-service может быть добавлена здесь
                    # в зависимости от API сервиса
                    logger.info(f"Avatar generated for character {character_id}: {avatar_result['media_id']}")
                except Exception as e:
                    logger.error(f"Error updating character with avatar: {str(e)}")

            return avatar_result

        except Exception as e:
            logger.error(f"Error generating avatar: {str(e)}")
            return {}

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated character for world {self.task.world_id}: "
            f"{result.get('display_name')} (@{result.get('username')}), "
            f"ID: {result.get('character_id')}"
        )

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate character: {str(error)}")

        # В случае ошибки генерации персонажа, мы не обновляем статус этапа,
        # так как это не критическая ошибка для всего процесса генерации.