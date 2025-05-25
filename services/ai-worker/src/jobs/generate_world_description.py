import uuid
from typing import Dict, Any, List
from datetime import datetime, timezone

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..utils.model_to_template import model_to_template
from ..db.models import Task
from ..prompts import load_prompt, WORLD_DESCRIPTION_PROMPT
from ..schemas.world_description import WorldDescription, WorldDescriptionResponse

class GenerateWorldDescriptionJob(BaseJob):
    """
    Задание для генерации описания мира
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации описания мира

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        user_prompt = self.task.parameters.get("user_prompt", "")

        # Загружаем промпт из файла
        prompt_template = load_prompt(WORLD_DESCRIPTION_PROMPT)

        structure_description = model_to_template(WorldDescriptionResponse)

        # Форматируем промпт с параметрами
        prompt = prompt_template.format(user_prompt=user_prompt, structure_description=structure_description)

        # Генерируем описание мира с помощью LLM
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного контента
            world_description_response = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=WorldDescriptionResponse,
                temperature=0.8,
                max_output_tokens=4096,
                task_id=self.task.id,
                world_id=world_id
            )

            # Сохраняем параметры мира в БД
            now = datetime.now(timezone.utc)

            world_description = WorldDescription(
                **world_description_response.model_dump(),
                id=world_id,
                created_at=now,
                updated_at=now
            )

            await self.db_manager.save_world_parameters(world_description)

            logger.info(f"Generated and saved world description for world {world_id}")

            # Создаем следующие задачи
            tasks_to_create = []

            # Задача для генерации изображения мира
            world_image_task_id = str(uuid.uuid4())
            world_image_task = Task(
                _id=world_image_task_id,
                type=TaskType.GENERATE_WORLD_IMAGE,
                world_id=world_id,
                status="pending",
                worker_id=None,
                parameters={},
                created_at=now,
                updated_at=now,
                attempt_count=0
            )
            tasks_to_create.append({"task": world_image_task})

            # Задача для генерации пакета персонажей
            users_count = self.task.parameters.get("users_count", 10)
            posts_count = self.task.parameters.get("posts_count", 50)
            character_batch_task_id = str(uuid.uuid4())
            character_batch_task = Task(
                _id=character_batch_task_id,
                type=TaskType.GENERATE_CHARACTER_BATCH,
                world_id=world_id,
                status="pending",
                worker_id=None,
                parameters={
                    "users_count": users_count,
                    "posts_count": posts_count
                },
                created_at=now,
                updated_at=now,
                attempt_count=0
            )
            tasks_to_create.append({"task": character_batch_task})

            created_task_ids = await self.create_next_tasks(tasks_to_create)

            # Обновляем статус этапа
            if self.progress_manager:
                await self.progress_manager.update_stage(
                    world_id=world_id,
                    stage=GenerationStage.WORLD_DESCRIPTION,
                    status=GenerationStatus.COMPLETED
                )

                # Обновляем статус следующих этапов
                await self.progress_manager.update_stage(
                    world_id=world_id,
                    stage=GenerationStage.WORLD_IMAGE,
                    status=GenerationStatus.IN_PROGRESS
                )

                await self.progress_manager.update_stage(
                    world_id=world_id,
                    stage=GenerationStage.CHARACTERS,
                    status=GenerationStatus.IN_PROGRESS
                )

            return {
                "world_name": world_description.name,
                "world_description": world_description.description,
                "world_theme": world_description.theme,
                # "next_tasks": created_task_ids
            }

        except Exception as e:
            logger.error(f"Error generating world description for world {world_id}: {str(e)}")
            raise

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated world description for world {self.task.world_id}: "
            f"{result.get('world_name')} - {result.get('world_theme')}"
        )

        # # Обновляем информацию о мире в World Service, если необходимо
        # try:
        #     if self.service_client:
        #         # Передаем информацию о сгенерированном мире в World Service
        #         await self.service_client.update_world_status(
        #             world_id=self.task.world_id,
        #             status="description_generated",
        #             task_id=self.task.id
        #         )
        # except Exception as e:
        #     logger.error(f"Failed to update world service with description: {str(e)}")

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate world description: {str(error)}")

        # Обновляем статус этапа
        if self.progress_manager:
            await self.progress_manager.update_stage(
                world_id=self.task.world_id,
                stage=GenerationStage.WORLD_DESCRIPTION,
                status=GenerationStatus.FAILED
            )

            # Устанавливаем общий статус генерации как неудачный
            await self.progress_manager.update_progress(
                world_id=self.task.world_id,
                updates={"status": GenerationStatus.FAILED}
            )