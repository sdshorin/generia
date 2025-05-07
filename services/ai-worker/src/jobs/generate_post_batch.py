import uuid
import math
from typing import Dict, Any, List
from datetime import datetime, timezone

from ..core.base_job import BaseJob
from ..constants import TaskType
from ..utils.logger import logger
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..db.models import Task
from ..prompts import load_prompt, POST_BATCH_PROMPT, PREVIOUS_POSTS_PROMPT, FIRST_BATCH_POSTS_PROMPT
from ..schemas import PostBatchResponse

# Максимальное количество постов, генерируемых за один раз
MAX_POSTS_PER_BATCH = 10

# Максимальная глубина рекурсии для генерации постов
MAX_POST_RECURSION_DEPTH = 30

class GeneratePostBatchJob(BaseJob):
    """
    Задание для генерации пакета постов для персонажа
    """

    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации пакета постов

        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        character_name = self.task.parameters.get("character_name", "")
        character_description = self.task.parameters.get("character_description", {})
        posts_count = int(self.task.parameters.get("posts_count", 5))
        username = character_description.get("username", "")
        character_id = self.task.parameters.get("character_id", "")
        character_index = character_description.get("character_index", 0)


        generated_posts_count = int(self.task.parameters.get("generated_posts_count", 0))

        recursion_depth = int(self.task.parameters.get("recursion_depth", 0))

        total_posts_count = int(self.task.parameters.get("total_posts_count", posts_count))

        generated_posts_description = self.task.parameters.get("generated_posts_description", "")

        count_run = int(self.task.parameters.get("count_run", 0))

        max_allowed_depth = min(math.ceil(total_posts_count / 8) + 1, MAX_POST_RECURSION_DEPTH)

        logger.debug(f"Post batch task parameters: posts_count={posts_count}, "
                    f"total_posts_count={total_posts_count}, generated_posts_count={generated_posts_count}, "
                    f"count_run={count_run}, recursion_depth={recursion_depth}")

        if recursion_depth >= max_allowed_depth:
            logger.warning(f"Maximum recursion depth reached ({recursion_depth}/{max_allowed_depth}) for character {character_name} in world {world_id}. Stopping post generation.")
            return {
                "posts_count": 0,
                "total_posts_count": generated_posts_count,
                "remaining_posts": posts_count - generated_posts_count,
                "recursion_depth": recursion_depth,
                "max_allowed_depth": max_allowed_depth,
                "character_id": character_id,
                "character_name": character_name,
                "username": username,
                "error": f"Maximum recursion depth reached ({recursion_depth}/{max_allowed_depth})"
            }

        current_batch_size = min(posts_count - generated_posts_count, MAX_POSTS_PER_BATCH)

        if not character_id:
            raise ValueError("Character ID is required to generate posts")

        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")

        # Загружаем промпт из файла
        prompt_template = load_prompt(POST_BATCH_PROMPT)

        # Преобразуем списки в строки для промпта
        interests = character_description.get("interests", [])
        interests_str = ", ".join(interests) if isinstance(interests, list) else interests

        common_topics = character_description.get("common_topics", [])
        common_topics_str = ", ".join(common_topics) if isinstance(common_topics, list) else common_topics

        previous_posts_info = ""

        future_posts_count = posts_count - generated_posts_count - current_batch_size

        if generated_posts_count > 0:
            previous_posts_template = load_prompt(PREVIOUS_POSTS_PROMPT)
            previous_posts_info = previous_posts_template.format(
                count_run=count_run,
                count=generated_posts_count,
                total_posts_count=total_posts_count,
                current_batch_size=current_batch_size,
                future_posts_count=future_posts_count,
                description=generated_posts_description
            )
        elif posts_count > current_batch_size:
            first_batch_template = load_prompt(FIRST_BATCH_POSTS_PROMPT)
            previous_posts_info = first_batch_template.format(
                total_posts_count=total_posts_count,
                current_batch_size=current_batch_size,
                future_posts_count=future_posts_count
            )

        # Генерируем описание структуры ответа
        structure_description = model_to_template(PostBatchResponse)

        # Форматируем промпт с параметрами
        world_description = format_world_description(world_params)
        prompt = prompt_template.format(
            world_description=world_description,
            character_name=character_name,
            character_description=f"{character_description.get('bio', '')} {character_description.get('personality', '')}",
            interests=interests_str,
            speaking_style=character_description.get("speaking_style", ""),
            common_topics=common_topics_str,
            appearance=character_description.get("appearance", ""),
            secret=character_description.get("secret", ""),
            daily_routine=character_description.get("daily_routine", ""),
            avatar_description=character_description.get("avatar_description", ""),
            avatar_style=character_description.get("avatar_style", ""),
            posts_count=current_batch_size,
            structure_description=structure_description,
            previous_posts_info=previous_posts_info
        )

        # Генерируем пакет постов с помощью LLM
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )

        try:
            # Генерация структурированного контента
            post_batch = await self.llm_client.generate_structured_content(
                prompt=prompt,
                response_schema=PostBatchResponse,
                temperature=0.8,
                max_output_tokens=6144,  # Большой лимит для множества постов
                task_id=self.task.id,
                world_id=world_id
            )

            actual_posts_count = len(post_batch.posts)
            logger.info(f"Generated post batch for character {character_name} in world {world_id} with {actual_posts_count} posts (requested {current_batch_size})")

            if actual_posts_count == 0:
                logger.warning(f"LLM returned 0 posts for character {character_name} in world {world_id} (requested {current_batch_size})")
                return {
                    "posts_count": 0,
                    "total_posts_count": generated_posts_count,
                    "remaining_posts": posts_count - generated_posts_count,
                    "character_id": character_id,
                    "character_name": character_name,
                    "username": username,
                    "error": "LLM returned 0 posts",
                    "recursion_depth": recursion_depth,
                    "max_allowed_depth": max_allowed_depth
                }

            if actual_posts_count < current_batch_size:
                logger.warning(f"LLM returned fewer posts than requested: {actual_posts_count} < {current_batch_size}")

            # Создаем задачи для генерации каждого поста
            tasks_to_create = []
            now = datetime.now(timezone.utc)

            # Обновляем счетчик сгенерированных постов
            new_generated_posts_count = generated_posts_count + actual_posts_count
            new_count_run = count_run + 1

            post_descriptions = []
            for post in post_batch.posts:
                desc = f"Тема: {post.topic}. Краткое содержание: {post.content_brief}. Тон: {post.emotional_tone}. Тип: {post.post_type}."
                post_descriptions.append(desc)

            # Объединяем с предыдущим описанием
            new_description = generated_posts_description
            if post_descriptions:
                if new_description:
                    new_description += "\n\n"
                new_description += "\n".join(post_descriptions)

            for i, post in enumerate(post_batch.posts):
                post_task_id = str(uuid.uuid4())
                post_task = Task(
                    _id=post_task_id,
                    type=TaskType.GENERATE_POST,
                    world_id=world_id,
                    status="pending",
                    worker_id=None,
                    parameters={
                        "post_topic": post.topic,
                        "post_brief": post.content_brief,
                        "emotional_tone": post.emotional_tone,
                        "post_type": post.post_type,
                        "relevance_to_character": post.relevance_to_character,
                        "character_name": character_name,
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
                        },
                        "character_id": character_id,
                        "username": username,
                        "character_index": character_index,
                        "post_index": i + generated_posts_count
                    },
                    created_at=now,
                    updated_at=now,
                    attempt_count=0
                )
                tasks_to_create.append({"task": post_task})


            remaining_posts = posts_count - new_generated_posts_count

            if remaining_posts > 0:
                new_recursion_depth = recursion_depth + 1

                if new_recursion_depth >= max_allowed_depth:
                    logger.warning(f"Would exceed maximum recursion depth ({new_recursion_depth}/{max_allowed_depth}) for character {character_name} in world {world_id}. Stopping further post generation.")
                else:
                    next_batch_task_id = str(uuid.uuid4())
                    next_batch_task = Task(
                        _id=next_batch_task_id,
                        type=TaskType.GENERATE_POST_BATCH,
                        world_id=world_id,
                        status="pending",
                        worker_id=None,
                        parameters={
                            "character_name": character_name,
                            "character_description": character_description,
                            "character_id": character_id,
                            "username": username,
                            "character_index": character_index,
                            "posts_count": posts_count,
                            "generated_posts_count": new_generated_posts_count,
                            "recursion_depth": new_recursion_depth,
                            "total_posts_count": total_posts_count,
                            "generated_posts_description": new_description,
                            "count_run": new_count_run
                        },
                        created_at=now,
                        updated_at=now,
                        attempt_count=0
                    )

                    logger.debug(f"Next post batch task parameters: posts_count={posts_count}, "
                                f"total_posts_count={total_posts_count}, generated_posts_count={new_generated_posts_count}, "
                                f"count_run={new_count_run}, recursion_depth={new_recursion_depth}")
                    tasks_to_create.append({"task": next_batch_task})

                    logger.info(f"Creating recursive task to generate {remaining_posts} more posts for character {character_name} in world {world_id} (recursion depth: {new_recursion_depth}/{max_allowed_depth})")

            created_task_ids = await self.create_next_tasks(tasks_to_create)

            # Передаем информацию о пакете постов в результат
            return {
                "character_id": character_id,
                "character_name": character_name,
                "username": username,
                "posts_count": len(post_batch.posts),
                "total_posts_count": new_generated_posts_count,
                "remaining_posts": remaining_posts,
                "narrative_arc": post_batch.narrative_arc,
                "character_development": post_batch.character_development,
                "recurring_themes": post_batch.recurring_themes,
                "next_tasks": created_task_ids,
                "recursion_depth": recursion_depth,
                "max_allowed_depth": max_allowed_depth
            }

        except Exception as e:
            logger.error(f"Error generating post batch for character {character_name} in world {world_id}: {str(e)}")
            raise

    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        total_count = result.get('total_posts_count', result.get('posts_count', 0))
        remaining = result.get('remaining_posts', 0)
        recursion_depth = result.get('recursion_depth', 0)
        max_allowed_depth = result.get('max_allowed_depth', 0)

        if remaining > 0:
            logger.info(
                f"Successfully generated post batch for character {result.get('character_name')} "
                f"(@{result.get('username')}) in world {self.task.world_id} "
                f"with {result.get('posts_count')} posts. "
                f"Total generated: {total_count}. Remaining to generate: {remaining}. "
                f"Recursion depth: {recursion_depth}/{max_allowed_depth}."
            )
        else:
            logger.info(
                f"Successfully generated post batch for character {result.get('character_name')} "
                f"(@{result.get('username')}) in world {self.task.world_id} "
                f"with {result.get('posts_count')} posts. "
                f"Total generated: {total_count}. All posts generated. "
                f"Recursion depth: {recursion_depth}/{max_allowed_depth}."
            )

    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate post batch: {str(error)}")