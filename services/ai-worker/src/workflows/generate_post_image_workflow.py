from datetime import timedelta
from typing import Dict, Any, List

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.model_to_template import model_to_template
from ..prompts import POST_IMAGE_PROMPT
from ..schemas.post_image import PostImagePromptResponse


class GeneratePostImageInput(TaskInput):
    """Входные данные для workflow генерации изображения к посту"""
    world_id: str
    character_id: str
    post_detail: Dict[str, Any]
    character_detail: Dict[str, Any]
    post_index: int = 0
    character_index: int = 0


@workflow.defn
class GeneratePostImageWorkflow(BaseWorkflow):
    """
    Workflow для генерации изображения к посту и публикации поста
    """
    
    @workflow.run
    async def run(self, task_ref: TaskRef) -> WorkflowResult:
        """
        Основной метод выполнения workflow
        
        Args:
            task_ref: TaskRef с task_id для загрузки данных
            
        Returns:
            Результат выполнения workflow
        """
        try:
            # Загружаем данные задачи из MongoDB
            input = await self.load_task_data(task_ref, GeneratePostImageInput)
            
            character_name = input.character_detail.get("display_name", "Unknown")
            username = input.character_detail.get("username", "unknown")
            
            workflow.logger.info(f"Starting post image generation for character {character_name} ({input.character_id})")
            
            # Извлекаем данные для генерации изображения
            image_prompt = input.post_detail.get("image_prompt", "")
            image_style = input.post_detail.get("image_style", "")
            post_content = input.post_detail.get("content", "")
            hashtags = input.post_detail.get("hashtags", [])
            
            if not image_prompt:
                workflow.logger.warning(f"No image prompt for post by character {character_name}")
                return WorkflowResult(
                    success=True,
                    data={"message": "No image prompt provided"}
                )
            
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Формируем промпт для оптимизации изображения
            prompt = await self._build_post_image_prompt(input, world_params)
            
            # Увеличиваем счетчик LLM запросов для оптимизации промпта
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем оптимизированный промпт для изображения
            image_prompt_response = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "PostImagePromptResponse",
                    input.world_id,
                    workflow.info().workflow_id,
                    0.7,  # temperature
                    2048  # max_output_tokens
                ],
                task_queue="ai-worker-llm",
                start_to_close_timeout=timedelta(minutes=3),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=2),
                    maximum_interval=timedelta(minutes=1),
                    maximum_attempts=3
                )
            )
            
            optimized_image_prompt = image_prompt_response.get("prompt", image_prompt)
            
            workflow.logger.info(f"Generated optimized image prompt for post by {character_name}")
            # Увеличиваем счетчик image запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_images", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем изображение
            image_result = await workflow.execute_activity(
                "generate_image",
                args=[
                    optimized_image_prompt,
                    input.world_id,
                    "post_image",
                    True,  # enhance_prompt
                    input.character_id  # character_id
                ],
                task_queue="ai-worker-images",
                start_to_close_timeout=timedelta(minutes=5),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=5),
                    maximum_interval=timedelta(minutes=2),
                    maximum_attempts=3
                )
            )
            
            if not image_result or "media_id" not in image_result:
                raise ValueError("Failed to generate image for post")
            
            workflow.logger.info(f"Generated image for post by {character_name}")
            
            # Получаем URL изображения и медиа ID
            image_url = image_result.get("image_url")
            media_id = image_result.get("media_id")
            
            # Создаем пост через API
            post_data = {
                "content": post_content,
                "media_id": media_id,
                "hashtags": hashtags
            }
            post_id = await workflow.execute_activity(
                "create_post",
                args=[
                    post_data,
                    input.character_id,
                    input.world_id
                ],
                task_queue="ai-worker-services",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            if post_id:
                workflow.logger.info(f"Created post {post_id} with image for character {character_name}")
                
                # Увеличиваем счетчик созданных постов
                await workflow.execute_activity(
                    "increment_counter",
                    args=[input.world_id, "posts_created", 1],
                    task_queue="ai-worker-progress",
                    start_to_close_timeout=timedelta(seconds=30),
                    retry_policy=RetryPolicy(maximum_attempts=3)
                )
            
            workflow.logger.info(f"Successfully completed post image generation for character {character_name}")
            
            return WorkflowResult(
                success=True,
                data={
                    "character_id": input.character_id,
                    "character_name": character_name,
                    "username": username,
                    "image_url": image_url,
                    "media_id": media_id,
                    "optimized_prompt": optimized_image_prompt,
                    "post_id": post_id,
                    "content": post_content,
                    "hashtags": hashtags
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating post image: {str(e)}"
            workflow.logger.error(f"Workflow failed for character {input.character_id}: {error_msg}")
            raise 
            # return WorkflowResult(success=False, error=error_msg)
    
    async def _build_post_image_prompt(
        self, 
        input: GeneratePostImageInput, 
        world_params: Dict[str, Any]
    ) -> str:
        """
        Строит промпт для оптимизации промпта изображения поста
        
        Args:
            input: Входные данные
            world_params: Параметры мира
            
        Returns:
            Сформированный промпт
        """
        # Загружаем промпт из файла
        prompt_template = await workflow.execute_activity(
            "load_prompt",
            args=[POST_IMAGE_PROMPT],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Извлекаем данные
        character_name = input.character_detail.get("display_name", "Unknown")
        appearance = input.character_detail.get("appearance", "")
        avatar_description = input.character_detail.get("avatar_description", "")
        avatar_style = input.character_detail.get("avatar_style", "")
        image_prompt = input.post_detail.get("image_prompt", "")
        image_style = input.post_detail.get("image_style", "")
        post_content = input.post_detail.get("content", "")
        
        # Создаем описание структуры ответа из модели pydantic
        structure_description = model_to_template(PostImagePromptResponse)
        
        # Форматируем промпт с параметрами
        world_description = await workflow.execute_activity(
            "format_world_description",
            args=[world_params],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        prompt = prompt_template.format(
            world_description=world_description,
            character_name=character_name,
            appearance=appearance,
            avatar_description=avatar_description,
            avatar_style=avatar_style,
            image_prompt=image_prompt,
            image_style=image_style,
            post_content=post_content,
            structure_description=structure_description
        )
        
        return prompt