import uuid
import math
from typing import Dict, Any, List, Optional
from datetime import datetime, timezone

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..db.models import Task
from ..db import WorldParameters
from ..prompts import load_prompt, CHARACTER_BATCH_PROMPT, PREVIOUS_CHARACTERS_PROMPT, FIRST_BATCH_CHARACTERS_PROMPT
from ..schemas import CharacterBatchResponse

# Максимальное количество персонажей, генерируемых за один раз
MAX_CHARACTERS_PER_BATCH = 10

# Максимальная глубина рекурсии для генерации персонажей
MAX_CHARACTER_RECURSION_DEPTH = 50

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

        remaining_posts_count = int(self.task.parameters.get("remaining_posts_count", posts_count))

        total_users_count = int(self.task.parameters.get("total_users_count", users_count))

        # Получаем информацию о ранее сгенерированных персонажах
        generated_characters_description = self.task.parameters.get("generated_characters_description", "")
        generated_count = int(self.task.parameters.get("generated_count", 0))
        count_run = int(self.task.parameters.get("count_run", 0))

        recursion_depth = int(self.task.parameters.get("recursion_depth", 0))

        # Вычисляем максимально допустимую глубину рекурсии для данного количества персонажей
        max_allowed_depth = min(math.ceil(total_users_count / 8) + 1, MAX_CHARACTER_RECURSION_DEPTH)

        # Логируем входные параметры для отладки
        logger.debug(f"Character batch task parameters: users_count={users_count}, "
                    f"total_users_count={total_users_count}, generated_count={generated_count}, "
                    f"count_run={count_run}, recursion_depth={recursion_depth}, "
                    f"remaining_posts_count={remaining_posts_count}")

        # Проверяем, не превышена ли глубина рекурсии
        if recursion_depth >= max_allowed_depth:
            logger.warning(f"Maximum recursion depth reached ({recursion_depth}/{max_allowed_depth}) for world {world_id}. Stopping character generation.")
            return {
                "characters_count": 0,
                "total_characters_count": generated_count,
                "remaining_characters": users_count,
                "recursion_depth": recursion_depth,
                "max_allowed_depth": max_allowed_depth,
                "error": f"Maximum recursion depth reached ({recursion_depth}/{max_allowed_depth})"
            }

        # Ограничиваем количество персонажей для текущей генерации
        current_batch_size = min(users_count, MAX_CHARACTERS_PER_BATCH)

        # Получаем параметры мира
        world_params: WorldParameters = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(CHARACTER_BATCH_PROMPT)

        # Формируем информацию о персонажах
        previous_characters_info = ""

        # Вычисляем количество персонажей, которые будут сгенерированы в будущих запусках
        future_users_count = users_count - current_batch_size

        if generated_count > 0:
            # Если уже есть сгенерированные персонажи, используем промпт для последующих генераций
            previous_characters_template = load_prompt(PREVIOUS_CHARACTERS_PROMPT)
            previous_characters_info = previous_characters_template.format(
                count_run=count_run,
                count=generated_count,
                total_users_count=total_users_count,
                current_batch_size=current_batch_size,
                future_users_count=future_users_count,
                description=generated_characters_description
            )
        elif users_count > current_batch_size:
            # Если это первая генерация, но будут еще генерации, используем промпт для первой генерации
            first_batch_template = load_prompt(FIRST_BATCH_CHARACTERS_PROMPT)
            previous_characters_info = first_batch_template.format(
                total_users_count=total_users_count,
                current_batch_size=current_batch_size,
                future_users_count=future_users_count
            )

        # Генерируем описание структуры ответа
        structure_description = model_to_template(CharacterBatchResponse)

        # Calculate the number of posts for the current batch proportionally to the number of characters
        posts_count_for_batch = int(remaining_posts_count * (current_batch_size / users_count))
        # Ensure that each character gets at least 1 post
        posts_count_for_batch = max(posts_count_for_batch, current_batch_size)

        logger.debug(f"Posts count for current batch: {posts_count_for_batch} (remaining: {remaining_posts_count}, batch size: {current_batch_size}, total users: {users_count})")

        # Format the prompt with parameters
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            users_count=current_batch_size,
            posts_count=posts_count_for_batch,
            world_description=world_description,
            structure_description=structure_description,
            previous_characters_info=previous_characters_info
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

            actual_characters_count = len(character_batch.characters)
            logger.info(f"Generated character batch for world {world_id} with {actual_characters_count} characters (requested {current_batch_size})")

            # Проверяем, что LLM вернула хотя бы одного персонажа
            if actual_characters_count == 0:
                logger.warning(f"LLM returned 0 characters for world {world_id} (requested {current_batch_size})")
                return {
                    "characters_count": 0,
                    "total_characters_count": generated_count,
                    "remaining_characters": users_count,
                    "error": "LLM returned 0 characters",
                    "recursion_depth": recursion_depth,
                    "max_allowed_depth": max_allowed_depth
                }

            # Если LLM вернула меньше персонажей, чем запрошено, корректируем remaining_users
            if actual_characters_count < current_batch_size:
                logger.warning(f"LLM returned fewer characters than requested: {actual_characters_count} < {current_batch_size}")

            self._adjust_posts_count(character_batch, posts_count_for_batch)

            # Создаем задачи для генерации каждого персонажа
            tasks_to_create = []
            now = datetime.now(timezone.utc)

            # Обновляем счетчики сгенерированных персонажей
            new_generated_count = generated_count + len(character_batch.characters)
            new_count_run = count_run + 1

            # Создаем краткое описание сгенерированных персонажей
            character_descriptions = []
            for character in character_batch.characters:
                desc = f"{character.concept_short} Роль: {character.role_in_world}. Черты: {', '.join(character.personality_traits)}."
                character_descriptions.append(desc)

            # Объединяем с предыдущим описанием
            new_description = generated_characters_description
            if character_descriptions:
                if new_description:
                    new_description += "\n\n"
                new_description += "\n".join(character_descriptions)

            # Создаем задачи для генерации каждого персонажа
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
                        "character_index": i + generated_count,
                    },
                    created_at=now,
                    updated_at=now,
                    attempt_count=0
                )
                tasks_to_create.append({"task": character_task})

            remaining_users = users_count - actual_characters_count

            if remaining_users > 0:
                posts_allocated = sum(character.posts_count for character in character_batch.characters)
                new_remaining_posts = remaining_posts_count - posts_allocated

                # Проверяем, что осталось достаточно постов для оставшихся персонажей
                if new_remaining_posts < remaining_users:
                    logger.warning(f"Not enough posts left for remaining characters. Adjusting to ensure at least 1 post per character.")
                    new_remaining_posts = remaining_users

                new_recursion_depth = recursion_depth + 1

                if new_recursion_depth >= max_allowed_depth:
                    logger.warning(f"Would exceed maximum recursion depth ({new_recursion_depth}/{max_allowed_depth}) for world {world_id}. Stopping further character generation.")
                else:
                    next_batch_task_id = str(uuid.uuid4())
                    next_batch_task = Task(
                        _id=next_batch_task_id,
                        type=TaskType.GENERATE_CHARACTER_BATCH,
                        world_id=world_id,
                        status="pending",
                        worker_id=None,
                        parameters={
                            "users_count": remaining_users,
                            "posts_count": posts_count,
                            "remaining_posts_count": new_remaining_posts,
                            "generated_characters_description": new_description,
                            "generated_count": new_generated_count,
                            "count_run": new_count_run,
                            "recursion_depth": new_recursion_depth,
                            "total_users_count": total_users_count
                        },
                        created_at=now,
                        updated_at=now,
                        attempt_count=0
                    )

                    logger.debug(f"Next character batch task parameters: users_count={remaining_users}, "
                                f"total_users_count={total_users_count}, generated_count={new_generated_count}, "
                                f"count_run={new_count_run}, recursion_depth={new_recursion_depth}, "
                                f"remaining_posts_count={new_remaining_posts}")
                    tasks_to_create.append({"task": next_batch_task})

                    logger.info(f"Creating recursive task to generate {remaining_users} more characters for world {world_id} (recursion depth: {new_recursion_depth}/{max_allowed_depth})")

            created_task_ids = await self.create_next_tasks(tasks_to_create)


            # Calculate posts statistics
            posts_allocated = sum(character.posts_count for character in character_batch.characters)
            new_remaining_posts = remaining_posts_count - posts_allocated

            # Log detailed post distribution information
            logger.info(f"Post distribution for batch {count_run + 1}: allocated {posts_allocated} posts to {len(character_batch.characters)} characters")
            logger.info(f"Remaining posts: {new_remaining_posts} for {remaining_users} characters in future batches")

            # Log individual character post counts for debugging
            for i, character in enumerate(character_batch.characters):
                logger.debug(f"Character {i+1}: {character.concept_short} - {character.posts_count} posts")

            # Передаем структуру персонажей в результат
            return {
                "characters_count": len(character_batch.characters),
                "total_characters_count": new_generated_count,
                "remaining_characters": remaining_users,
                "world_interpretation": character_batch.world_interpretation,
                "character_connections": [conn.model_dump() for conn in character_batch.character_connections],
                "next_tasks": created_task_ids,
                "recursion_depth": recursion_depth,
                "max_allowed_depth": max_allowed_depth,
                "posts_allocated": posts_allocated,
                "remaining_posts_count": new_remaining_posts
            }

        except Exception as e:
            logger.error(f"Error generating character batch for world {world_id}: {str(e)}")
            raise

    def _adjust_posts_count(self, character_batch: CharacterBatchResponse, posts_count_for_batch: int) -> None:
        """
        Checks and adjusts the number of posts for each character,
        so that the sum matches the target value for the current batch

        Args:
            character_batch: Пакет персонажей
            target_posts_count: Целевое количество постов
        """
        if not character_batch.characters:
            return

        actual_target = posts_count_for_batch

        current_sum = sum(character.posts_count for character in character_batch.characters)

        if current_sum == actual_target:
            logger.info(f"Posts count already matches target: {current_sum}")
            return

        logger.info(f"Adjusting posts count from {current_sum} to {actual_target}")


        for character in character_batch.characters:
            if character.posts_count < 1:
                character.posts_count = 1

        current_sum = sum(character.posts_count for character in character_batch.characters)

        if current_sum < actual_target:
            diff = actual_target - current_sum

            total_weight = sum(character.posts_count for character in character_batch.characters)

            if len(set(character.posts_count for character in character_batch.characters)) <= 1:
                posts_per_character = diff // len(character_batch.characters)
                remainder = diff % len(character_batch.characters)

                for i, character in enumerate(character_batch.characters):
                    character.posts_count += posts_per_character
                    if i < remainder:
                        character.posts_count += 1
            else:
                remaining_to_add = diff

                for character in character_batch.characters:
                    weight = character.posts_count / total_weight
                    posts_to_add = int(diff * weight)
                    character.posts_count += posts_to_add
                    remaining_to_add -= posts_to_add

                if remaining_to_add > 0:
                    sorted_characters = sorted(
                        character_batch.characters,
                        key=lambda c: c.posts_count / total_weight,
                        reverse=True
                    )

                    for i in range(int(remaining_to_add)):
                        sorted_characters[i % len(sorted_characters)].posts_count += 1

        elif current_sum > actual_target:
            diff = current_sum - actual_target

            sorted_characters = sorted(
                character_batch.characters,
                key=lambda c: c.posts_count,
                reverse=True
            )

            remaining_to_remove = diff
            total_posts = current_sum

            for character in sorted_characters:
                if remaining_to_remove <= 0:
                    break

                weight = character.posts_count / total_posts
                posts_to_remove = min(int(diff * weight), character.posts_count - 1)

                if posts_to_remove > 0:
                    character.posts_count -= posts_to_remove
                    remaining_to_remove -= posts_to_remove

            if remaining_to_remove > 0:
                sorted_characters = sorted(
                    character_batch.characters,
                    key=lambda c: c.posts_count,
                    reverse=True
                )

                for character in sorted_characters:
                    if remaining_to_remove <= 0:
                        break

                    if character.posts_count > 1:
                        character.posts_count -= 1
                        remaining_to_remove -= 1

        # Проверяем результат
        new_sum = sum(character.posts_count for character in character_batch.characters)
        logger.info(f"Adjusted posts count: {new_sum} (target: {actual_target})")

        if new_sum != actual_target:
            logger.warning(f"Failed to adjust posts count exactly: {new_sum} != {actual_target}")

        for character in character_batch.characters:
            if character.posts_count < 1:
                logger.warning(f"Character has less than 1 post after adjustment, setting to 1")
                character.posts_count = 1

        final_sum = sum(character.posts_count for character in character_batch.characters)
        if final_sum != new_sum:
            logger.info(f"Final adjustment changed posts count from {new_sum} to {final_sum}")

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        total_count = result.get('total_characters_count', result.get('characters_count', 0))
        remaining = result.get('remaining_characters', 0)
        recursion_depth = result.get('recursion_depth', 0)
        max_allowed_depth = result.get('max_allowed_depth', 0)
        posts_allocated = result.get('posts_allocated', 0)
        remaining_posts = result.get('remaining_posts_count', 0)

        avg_posts_per_character = posts_allocated / result.get('characters_count') if result.get('characters_count') > 0 else 0
        avg_remaining_posts = remaining_posts / remaining if remaining > 0 else 0

        if remaining > 0:
            logger.info(
                f"Successfully generated character batch for world {self.task.world_id} "
                f"with {result.get('characters_count')} characters. "
                f"Total generated: {total_count}. Remaining to generate: {remaining}. "
                f"Recursion depth: {recursion_depth}/{max_allowed_depth}. "
                f"Posts allocated: {posts_allocated} (avg {avg_posts_per_character:.1f} per character), "
                f"remaining: {remaining_posts} (avg {avg_remaining_posts:.1f} per remaining character)"
            )
        else:
            logger.info(
                f"Successfully generated character batch for world {self.task.world_id} "
                f"with {result.get('characters_count')} characters. "
                f"Total generated: {total_count}. All characters generated. "
                f"Recursion depth: {recursion_depth}/{max_allowed_depth}. "
                f"Posts allocated: {posts_allocated} (avg {avg_posts_per_character:.1f} per character), "
                f"remaining: {remaining_posts}"
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