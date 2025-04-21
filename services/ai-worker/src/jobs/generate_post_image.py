import uuid
from typing import Dict, Any, List, Optional
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.logger import logger
from ..prompts import load_prompt, POST_IMAGE_PROMPT

class GeneratePostImageJob(BaseJob):
    """
    Задание для генерации изображения к посту
    """
    
    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по генерации изображения к посту
        
        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        image_prompt = self.task.parameters.get("image_prompt", "")
        image_style = self.task.parameters.get("image_style", "")
        post_content = self.task.parameters.get("post_content", "")
        character_name = self.task.parameters.get("character_name", "")
        username = self.task.parameters.get("username", "")
        character_index = self.task.parameters.get("character_index", 0)
        post_index = self.task.parameters.get("post_index", 0)
        
        # Получаем параметры мира
        world_params = await self.get_world_parameters(world_id)
        if not world_params:
            raise ValueError(f"Cannot find world parameters for world {world_id}")
        
        # Загружаем промпт из файла
        prompt_template = load_prompt(POST_IMAGE_PROMPT)
        
        # Форматируем промпт с параметрами
        prompt = prompt_template.format(
            world_name=world_params.name,
            world_visual_style=world_params.visual_style,
            character_name=character_name,
            image_prompt=image_prompt,
            image_style=image_style,
            post_content=post_content
        )
        
        # Генерируем оптимизированный промпт для создания изображения
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="api_calls_made_LLM"
            )
        
        try:
            # Генерация оптимизированного промпта для изображения
            optimized_prompt_result = await self.llm_client.generate_content(
                prompt=prompt,
                temperature=0.7,
                task_id=self.task.id,
                world_id=world_id
            )
            
            optimized_image_prompt = optimized_prompt_result["text"]
            
            logger.info(f"Generated optimized image prompt for post by {character_name} in world {world_id}")
            
            # Генерируем изображение для поста
            if self.progress_manager:
                await self.progress_manager.increment_task_counter(
                    world_id=world_id,
                    field="api_calls_made_images"
                )
            
            # Генерируем изображение
            post_image = await self.image_generator.generate_image(
                prompt=optimized_image_prompt,
                width=800,
                height=600,
                task_id=self.task.id,
                world_id=world_id,
                filename=f"post_{world_id}_{username}_{character_index}_{post_index}.png",
                media_type="image/png"
            )
            
            logger.info(f"Generated image for post by {character_name} in world {world_id}")
            
            # Получаем URL изображения
            image_url = post_image.get("image_url")
            image_id = post_image.get("media_id")
            
            # Создаем пост через API
            user_id = None
            post_id = None
            
            if self.service_client:
                try:
                    # Сначала получаем ID пользователя
                    user_result = await self.service_client.create_ai_user(
                        world_id=world_id,
                        username=username,
                        display_name=character_name,
                        task_id=self.task.id
                    )
                    
                    user_id = user_result.get("id")
                    
                    if user_id and image_id:
                        # Создаем пост с изображением
                        post_result = await self.service_client.create_post(
                            world_id=world_id,
                            user_id=user_id,
                            content=post_content,
                            media_ids=[image_id],
                            task_id=self.task.id
                        )
                        
                        post_id = post_result.get("id")
                        
                        if post_id:
                            logger.info(f"Created post {post_id} with image for user {username} in world {world_id}")
                            
                            # Увеличиваем счетчик созданных постов
                            if self.progress_manager:
                                await self.progress_manager.increment_task_counter(
                                    world_id=world_id,
                                    field="posts_created"
                                )
                
                except Exception as e:
                    logger.error(f"Error creating post with image via API: {str(e)}")
                    # Продолжаем выполнение даже в случае ошибки
            
            return {
                "character_name": character_name,
                "username": username,
                "image_url": image_url,
                "image_id": image_id,
                "optimized_prompt": optimized_image_prompt,
                "post_id": post_id,
                "user_id": user_id
            }
            
        except Exception as e:
            logger.error(f"Error generating image for post by {character_name} in world {world_id}: {str(e)}")
            raise
    
    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания
        
        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"Successfully generated image for post by {result.get('character_name')} "
            f"(@{result.get('username')}) in world {self.task.world_id}"
        )
        
        # Проверяем, завершена ли генерация мира
        if self.progress_manager:
            # Получаем текущий статус
            world_status = await self.db_manager.world_generation_status_collection.find_one({"_id": self.task.world_id})
            
            if world_status:
                # Проверяем, все ли этапы завершены
                all_stages_completed = True
                for stage in world_status.get("stages", []):
                    if stage["status"] != GenerationStatus.COMPLETED:
                        all_stages_completed = False
                        break
                
                # Проверяем, достигнуто ли предсказанное количество постов
                posts_created = world_status.get("posts_created", 0)
                posts_predicted = world_status.get("posts_predicted", 0)
                
                # Если все этапы завершены или создано достаточно постов, обновляем общий статус
                if all_stages_completed or (posts_predicted > 0 and posts_created >= posts_predicted):
                    await self.progress_manager.update_progress(
                        world_id=self.task.world_id,
                        updates={
                            "status": GenerationStatus.COMPLETED,
                            "current_stage": GenerationStage.FINISHING
                        }
                    )
                    
                    # Обновляем статус этапа
                    await self.progress_manager.update_stage(
                        world_id=self.task.world_id,
                        stage=GenerationStage.POSTS,
                        status=GenerationStatus.COMPLETED
                    )
                    
                    await self.progress_manager.update_stage(
                        world_id=self.task.world_id,
                        stage=GenerationStage.FINISHING,
                        status=GenerationStatus.COMPLETED
                    )
                    
                    # Обновляем статус мира в World Service
                    if self.service_client:
                        try:
                            await self.service_client.update_world_status(
                                world_id=self.task.world_id,
                                status="completed",
                                task_id=self.task.id
                            )
                            logger.info(f"Updated world status to completed for world {self.task.world_id}")
                        except Exception as e:
                            logger.error(f"Error updating world status: {str(e)}")
    
    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания
        
        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to generate post image: {str(error)}")