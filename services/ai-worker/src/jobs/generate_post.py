import uuid
from typing import Dict, Any, List, Optional
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..utils.logger import logger
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
        has_image = self.task.parameters.get("has_image", False)
        emotional_tone = self.task.parameters.get("emotional_tone", "")
        post_type = self.task.parameters.get("post_type", "")
        relevance_to_character = self.task.parameters.get("relevance_to_character", "")
        character_name = self.task.parameters.get("character_name", "")
        character_description = self.task.parameters.get("character_description", {})
        username = self.task.parameters.get("username", "")
        character_index = self.task.parameters.get("character_index", 0)
        post_index = self.task.parameters.get("post_index", 0)
        
        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")
        
        # Загружаем промпт из файла
        prompt_template = load_prompt(POST_DETAIL_PROMPT)
        
        # Форматируем промпт с параметрами
        prompt = prompt_template.format(
            world_name=world_params.name,
            world_theme=world_params.theme,
            world_culture=world_params.culture,
            character_name=character_name,
            character_description=character_description.get("personality", ""),
            speaking_style=character_description.get("speaking_style", ""),
            post_topic=post_topic,
            post_brief=post_brief,
            has_image=has_image,
            emotional_tone=emotional_tone,
            post_type=post_type,
            relevance_to_character=relevance_to_character
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
            now = datetime.utcnow()
            image_task_id = None
            
            if has_image and post_detail.image_prompt:
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
                        "character_index": character_index,
                        "post_index": post_index
                    },
                    created_at=now,
                    updated_at=now,
                    attempt_count=0
                )
                next_tasks.append({"task": image_task})
            
            # Определяем, нужно ли создавать пост через API
            should_create_post = not has_image  # Если нет изображения, создаем пост сразу
            user_id = None
            post_id = None
            
            # Если у сервисного клиента есть доступ к API, и пост не требует изображения,
            # сразу создаем пост через сервисный клиент
            if should_create_post and self.service_client:
                try:
                    # Сначала получаем ID пользователя
                    user_result = await self.service_client.create_ai_user(
                        world_id=world_id,
                        username=username,
                        display_name=character_name,
                        bio=character_description.get("bio", ""),
                        task_id=self.task.id
                    )
                    
                    user_id = user_result.get("id")
                    
                    if user_id:
                        # Создаем пост
                        post_result = await self.service_client.create_post(
                            world_id=world_id,
                            user_id=user_id,
                            content=post_detail.content,
                            task_id=self.task.id
                        )
                        
                        post_id = post_result.get("id")
                        
                        if post_id:
                            logger.info(f"Created post {post_id} for user {username} in world {world_id}")
                            
                            # Увеличиваем счетчик созданных постов
                            if self.progress_manager:
                                await self.progress_manager.increment_task_counter(
                                    world_id=world_id,
                                    field="posts_created"
                                )
                
                except Exception as e:
                    logger.error(f"Error creating post via API: {str(e)}")
                    # Продолжаем выполнение даже в случае ошибки
            
            # Создаем следующие задачи, если есть
            created_task_ids = []
            if next_tasks:
                created_task_ids = await self.create_next_tasks(next_tasks)
            
            # Передаем информацию о посте в результат
            return {
                "character_name": character_name,
                "username": username,
                "content": post_detail.content,
                "has_image": has_image,
                "image_prompt": post_detail.image_prompt,
                "hashtags": post_detail.hashtags,
                "mood": post_detail.mood,
                "context": post_detail.context,
                "next_tasks": created_task_ids,
                "post_id": post_id,
                "user_id": user_id
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