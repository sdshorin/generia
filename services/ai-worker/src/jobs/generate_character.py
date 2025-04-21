import uuid
from typing import Dict, Any, List
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..db.models import Task
from ..prompts import load_prompt, CHARACTER_DETAIL_PROMPT
from ..schemas import CharacterDetailResponse

class GenerateCharacterJob(BaseJob):
    """
    Задание для генерации персонажа
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
        
        # Форматируем промпт с параметрами
        prompt = prompt_template.format(
            world_name=world_params.name,
            world_description=world_params.description,
            world_theme=world_params.theme,
            world_technology_level=world_params.technology_level,
            world_social_structure=world_params.social_structure,
            world_culture=world_params.culture,
            world_geography=world_params.geography,
            world_visual_style=world_params.visual_style,
            character_concept=character_concept,
            role_in_world=role_in_world,
            personality_traits=personality_traits_str,
            interests=interests_str
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
            
            # Создаем задачи для аватара и пакета постов
            tasks_to_create = []
            now = datetime.utcnow()
            
            # # Задача для генерации аватара
            # avatar_task_id = str(uuid.uuid4())
            # avatar_task = Task(
            #     _id=avatar_task_id,
            #     type=TaskType.GENERATE_CHARACTER_AVATAR,
            #     world_id=world_id,
            #     status="pending",
            #     worker_id=None,
            #     parameters={
            #         "character_name": character_detail.display_name,
            #         "appearance_description": character_detail.appearance,
            #         "avatar_description": character_detail.avatar_description,
            #         "avatar_style": character_detail.avatar_style,
            #         "username": character_detail.username,
            #         "character_index": character_index
            #     },
            #     created_at=now,
            #     updated_at=now,
            #     attempt_count=0
            # )
            # tasks_to_create.append({"task": avatar_task})
            
            # Задача для генерации пакета постов
            post_batch_task_id = str(uuid.uuid4())
            post_batch_task = Task(
                _id=post_batch_task_id,
                type=TaskType.GENERATE_POST_BATCH,
                world_id=world_id,
                status="pending",
                worker_id=None,
                parameters={
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
                "username": character_detail.username,
                "display_name": character_detail.display_name,
                "bio": character_detail.bio,
                "avatar_description": character_detail.avatar_description,
                "speaking_style": character_detail.speaking_style,
                "next_tasks": created_task_ids
            }
            
        except Exception as e:
            logger.error(f"Error generating character detail for world {world_id}: {str(e)}")
            raise
    
    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания
        
        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated character detail for world {self.task.world_id}: "
            f"{result.get('display_name')} (@{result.get('username')})"
        )
    
    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания
        
        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate character detail: {str(error)}")
        
        # В случае ошибки генерации персонажа, мы не обновляем статус этапа,
        # так как это не критическая ошибка для всего процесса генерации.