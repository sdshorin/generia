import uuid
from typing import Dict, Any, List
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.logger import logger
from ..prompts import load_prompt, CHARACTER_AVATAR_PROMPT

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
        
        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")
        
        # Загружаем промпт из файла
        prompt_template = load_prompt(CHARACTER_AVATAR_PROMPT)
        
        # Форматируем промпт с параметрами
        prompt = prompt_template.format(
            world_name=world_params.name,
            world_visual_style=world_params.visual_style,
            character_name=character_name,
            appearance_description=appearance_description,
            avatar_description=avatar_description,
            avatar_style=avatar_style
        )
        
        # Генерируем оптимизированный промпт для создания аватара
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )
        
        try:
            # Генерация оптимизированного промпта для аватара
            avatar_prompt_result = await self.llm_client.generate_content(
                prompt=prompt,
                temperature=0.7,
                task_id=self.task.id,
                world_id=world_id
            )
            
            optimized_avatar_prompt = avatar_prompt_result["text"]
            
            logger.info(f"Generated avatar prompt for character {character_name} in world {world_id}")
            
            # Генерируем изображение аватара
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_images"
                )
            
            # Генерируем изображение аватара
            avatar_image = await self.image_generator.generate_image(
                prompt=optimized_avatar_prompt,
                width=512,
                height=512,
                task_id=self.task.id,
                world_id=world_id,
                filename=f"avatar_{world_id}_{username}_{character_index}.png",
                media_type="image/png"
            )
            
            logger.info(f"Generated avatar image for character {character_name} in world {world_id}")
            
            # Получаем URL аватара
            avatar_url = avatar_image.get("image_url")
            avatar_id = avatar_image.get("media_id")
            
            # Создаем пользователя через Auth Service
            if self.service_client:
                user_result = await self.service_client.create_ai_user(
                    world_id=world_id,
                    username=username,
                    display_name=character_name,
                    avatar_url=avatar_url,
                    bio=self.task.parameters.get("bio", ""),
                    task_id=self.task.id
                )
                
                user_id = user_result.get("id")
                logger.info(f"Created AI user {username} ({user_id}) in world {world_id}")
                
                # Сохраняем ID пользователя в результат
                return {
                    "username": username,
                    "display_name": character_name,
                    "avatar_url": avatar_url,
                    "avatar_id": avatar_id,
                    "user_id": user_id,
                    "avatar_prompt": optimized_avatar_prompt
                }
            
            # Если нет сервисного клиента, просто возвращаем данные об аватаре
            return {
                "username": username,
                "display_name": character_name,
                "avatar_url": avatar_url,
                "avatar_id": avatar_id,
                "avatar_prompt": optimized_avatar_prompt
            }
            
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