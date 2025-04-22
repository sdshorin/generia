import uuid
from typing import Dict, Any, List
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..db.models import Task
from ..prompts import load_prompt, POST_BATCH_PROMPT
from ..schemas import PostBatchResponse

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
        
        # Форматируем промпт с параметрами
        prompt = prompt_template.format(
            world_name=world_params.name,
            world_description=world_params.description,
            world_theme=world_params.theme,
            world_culture=world_params.culture,
            character_name=character_name,
            character_description=f"{character_description.get('bio', '')} {character_description.get('personality', '')}",
            interests=interests_str,
            speaking_style=character_description.get("speaking_style", ""),
            common_topics=common_topics_str,
            posts_count=posts_count
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
            
            logger.info(f"Generated post batch for character {character_name} in world {world_id} with {len(post_batch.posts)} posts")
            
            # Создаем задачи для генерации каждого поста
            tasks_to_create = []
            now = datetime.utcnow()
            
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
                        "character_description": character_description,
                        "character_id": character_id,
                        "username": username,
                        "character_index": character_index,
                        "post_index": i
                    },
                    created_at=now,
                    updated_at=now,
                    attempt_count=0
                )
                tasks_to_create.append({"task": post_task})
            
            created_task_ids = await self.create_next_tasks(tasks_to_create)
            
            # Передаем информацию о пакете постов в результат
            return {
                "character_id": character_id,
                "character_name": character_name,
                "username": username,
                "posts_count": len(post_batch.posts),
                "narrative_arc": post_batch.narrative_arc,
                "character_development": post_batch.character_development,
                "recurring_themes": post_batch.recurring_themes,
                "next_tasks": created_task_ids
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
        logger.info(
            f"Successfully generated post batch for character {result.get('character_name')} "
            f"(@{result.get('username')}) in world {self.task.world_id} "
            f"with {result.get('posts_count')} posts"
        )
    
    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания
        
        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate post batch: {str(error)}")