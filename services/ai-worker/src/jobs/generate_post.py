import uuid
from typing import Dict, Any, List, Optional
from datetime import datetime, timezone

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..db.models import Task
from ..prompts import load_prompt, POST_DETAIL_PROMPT
from ..schemas import PostDetailResponse

class GeneratePostJob(BaseJob):
    """
    Задание для генерации поста
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации поста

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        post_topic = self.task.parameters.get("post_topic", "")
        post_brief = self.task.parameters.get("post_brief", "")
        emotional_tone = self.task.parameters.get("emotional_tone", "")
        post_type = self.task.parameters.get("post_type", "")
        relevance_to_character = self.task.parameters.get("relevance_to_character", "")
        character_name = self.task.parameters.get("character_name", "")
        character_description = self.task.parameters.get("character_description", {})
        username = character_description.get("username", "")
        character_id = self.task.parameters.get("character_id", "")
        character_index = self.task.parameters.get("character_index", 0)
        post_index = self.task.parameters.get("post_index", 0)

        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(POST_DETAIL_PROMPT)

        # Генерируем описание структуры ответа
        structure_description = model_to_template(PostDetailResponse)

        # Форматируем промпт с параметрами
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            world_description=world_description,
            character_name=character_name,
            character_description=character_description.get("personality", ""),
            speaking_style=character_description.get("speaking_style", ""),
            appearance=character_description.get("appearance", ""),
            secret=character_description.get("secret", ""),
            daily_routine=character_description.get("daily_routine", ""),
            avatar_description=character_description.get("avatar_description", ""),
            avatar_style=character_description.get("avatar_style", ""),
            post_topic=post_topic,
            post_brief=post_brief,
            emotional_tone=emotional_tone,
            post_type=post_type,
            relevance_to_character=relevance_to_character,
            structure_description=structure_description
        )

        # Генерируем пост с помощью LLM
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного контента
            post_detail = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=PostDetailResponse,
                temperature=0.8,
                max_output_tokens=4096,
                task_id=self.task.id,
                world_id=world_id
            )

            logger.info(f"Generated post for character {character_name} in world {world_id}")

            # Создаем задачу для генерации изображения, если пост должен иметь изображение
            next_tasks = []
            now = datetime.now(timezone.utc)
            image_task_id = None

            if post_detail.image_prompt:
                image_task_id = str(uuid.uuid4())
                image_task = Task(
                    _id=image_task_id,
                    type=TaskType.GENERATE_POST_IMAGE,
                    world_id=world_id,
                    status="pending",
                    worker_id=None,
                    parameters={
                        "image_prompt": post_detail.image_prompt,
                        "image_style": post_detail.image_style,
                        "post_content": post_detail.content,
                        "character_name": character_name,
                        "username": username,
                        "character_id": character_id,
                        "character_index": character_index,
                        "post_index": post_index,
                        "tags": post_detail.hashtags,
                        "character_description": {
                            "username": username,
                            "bio": character_description.get("bio", ""),
                            "personality": character_description.get("personality", ""),
                            "interests": character_description.get("interests", []),
                            "speaking_style": character_description.get("speaking_style", ""),
                            "common_topics": character_description.get("common_topics", []),
                            "appearance": character_description.get("appearance", ""),
                            "secret": character_description.get("secret", ""),
                            "daily_routine": character_description.get("daily_routine", ""),
                            "avatar_description": character_description.get("avatar_description", ""),
                            "avatar_style": character_description.get("avatar_style", ""),
                            "character_index": character_index
                        }
                    },
                    created_at=now,
                    updated_at=now,
                    attempt_count=0
                )
                next_tasks.append({"task": image_task})

            # Создаем следующие задачи, если есть
            created_task_ids = []
            if next_tasks:
                created_task_ids = await self.create_next_tasks(next_tasks)

            # Передаем информацию о посте в результат
            return {
                "character_id": character_id,
                "character_name": character_name,
                "username": username,
                "content": post_detail.content,
                "image_prompt": post_detail.image_prompt,
                "hashtags": post_detail.hashtags,
                "mood": post_detail.mood,
                "context": post_detail.context,
                "next_tasks": created_task_ids
            }

        except Exception as e:
            logger.error(f"Error generating post for character {character_name} in world {world_id}: {str(e)}")
            raise

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated post for character {result.get('character_name')} "
            f"(@{result.get('username')}) in world {self.task.world_id}"
        )

        # Пост уже создан, если у него нет изображения
        # Если у поста есть изображение, он будет создан в задаче generate_post_image

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate post: {str(error)}")