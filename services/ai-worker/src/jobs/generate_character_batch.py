import uuid
from typing import Dict, Any, List
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..db.models import Task, WorldParameters
from ..prompts import load_prompt, CHARACTER_BATCH_PROMPT
from ..schemas import CharacterBatchResponse

class GenerateCharacterBatchJob(BaseJob):
    """
    Задание для генерации пакета персонажей
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации пакета персонажей

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        users_count = int(self.task.parameters.get("users_count", 10))
        posts_count = int(self.task.parameters.get("posts_count", 50))

        # Получаем параметры мира
        world_params: WorldParameters = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(CHARACTER_BATCH_PROMPT)

        # Генерируем описание структуры ответа
        structure_description = model_to_template(CharacterBatchResponse)

        # Форматируем промпт с параметрами
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            users_count=users_count,
            posts_count=posts_count,
            world_description=world_description,
            structure_description=structure_description
        )

        # Генерируем пакет персонажей с помощью LLM
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного контента
            character_batch = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=CharacterBatchResponse,
                temperature=0.8,
                max_output_tokens=8192,  # Увеличиваем лимит токенов для большого пакета персонажей
                task_id=self.task.id,
                world_id=world_id
            )

            logger.info(f"Generated character batch for world {world_id} with {len(character_batch.characters)} characters")

            # Проверяем и корректируем posts_count
            self._adjust_posts_count(character_batch, posts_count)

            # Создаем задачи для генерации каждого персонажа
            tasks_to_create = []
            now = datetime.utcnow()

            for i, character in enumerate(character_batch.characters):
                character_task_id = str(uuid.uuid4())
                character_task = Task(
                    _id=character_task_id,
                    type=TaskType.GENERATE_CHARACTER,
                    world_id=world_id,
                    status="pending",
                    worker_id=None,
                    parameters={
                        "character_concept": character.concept,
                        "role_in_world": character.role_in_world,
                        "posts_count": character.posts_count,
                        "personality_traits": character.personality_traits,
                        "interests": character.interests,
                        "character_index": i,  # Добавляем индекс для отслеживания
                    },
                    created_at=now,
                    updated_at=now,
                    attempt_count=0
                )
                tasks_to_create.append({"task": character_task})

            created_task_ids = await self.create_next_tasks(tasks_to_create)

            # Обновляем информацию о прогрессе
            if self.progress_manager:
                # Обновляем предсказанное количество пользователей
                await self.progress_manager.update_progress(
                    world_id=world_id,
                    updates={"users_predicted": len(character_batch.characters)}
                )

            # Передаем структуру персонажей в результат
            return {
                "characters_count": len(character_batch.characters),
                "world_interpretation": character_batch.world_interpretation,
                "character_connections": [conn.model_dump() for conn in character_batch.character_connections],
                "next_tasks": created_task_ids,
            }

        except Exception as e:
            logger.error(f"Error generating character batch for world {world_id}: {str(e)}")
            raise

    def _adjust_posts_count(self, character_batch: CharacterBatchResponse, target_posts_count: int) -> None:
        """
        Проверяет и корректирует количество постов для каждого персонажа,
        чтобы сумма соответствовала целевому значению

        Args:
            character_batch: Пакет персонажей
            target_posts_count: Целевое количество постов
        """
        if not character_batch.characters:
            return

        # Вычисляем текущую сумму постов
        current_sum = sum(character.posts_count for character in character_batch.characters)

        # Если сумма уже равна целевому значению, ничего не делаем
        if current_sum == target_posts_count:
            logger.info(f"Posts count already matches target: {current_sum}")
            return

        logger.info(f"Adjusting posts count from {current_sum} to {target_posts_count}")

        # Если сумма меньше целевого, добавляем посты
        if current_sum < target_posts_count:
            diff = target_posts_count - current_sum
            # Распределяем разницу между персонажами
            posts_per_character = diff // len(character_batch.characters)
            remainder = diff % len(character_batch.characters)

            for i, character in enumerate(character_batch.characters):
                # Добавляем базовое количество постов
                character.posts_count += posts_per_character
                # Распределяем остаток
                if i < remainder:
                    character.posts_count += 1

        # Если сумма больше целевого, убираем посты
        else:
            diff = current_sum - target_posts_count
            # Сортируем персонажей по количеству постов (по убыванию)
            sorted_characters = sorted(
                character_batch.characters,
                key=lambda c: c.posts_count,
                reverse=True
            )

            # Убираем посты, начиная с персонажей с наибольшим количеством
            for character in sorted_characters:
                if diff <= 0:
                    break

                # Убираем посты, но не меньше 1
                posts_to_remove = max(0, min(diff, character.posts_count - 1))
                character.posts_count -= posts_to_remove
                diff -= posts_to_remove

        # Проверяем результат
        new_sum = sum(character.posts_count for character in character_batch.characters)
        logger.info(f"Adjusted posts count: {new_sum} (target: {target_posts_count})")

        if new_sum != target_posts_count:
            logger.warning(f"Failed to adjust posts count exactly: {new_sum} != {target_posts_count}")

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated character batch for world {self.task.world_id} "
            f"with {result.get('characters_count')} characters"
        )

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate character batch: {str(error)}")

        # Обновляем статус этапа
        if self.progress_manager:
            await self.progress_manager.update_stage(
                world_id=self.task.world_id,
                stage=GenerationStage.CHARACTERS,
                status=GenerationStatus.FAILED
            )