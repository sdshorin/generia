from datetime import timedelta
from typing import Dict, Any

import asyncio
from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..schemas.task_base import TaskInput, TaskRef
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.format_world import format_world_description
from ..utils.model_to_template import model_to_template
from ..prompts import WORLD_IMAGE_PROMPT
from ..schemas.image_prompts import ImagePromptResponse


class GenerateWorldImageInput(TaskInput):
    """Входные данные для workflow генерации изображения мира"""
    world_id: str


@workflow.defn
class GenerateWorldImageWorkflow(BaseWorkflow):
    """
    Workflow для генерации изображения мира
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
            input = await self.load_task_data(task_ref, GenerateWorldImageInput)
            
            workflow.logger.info(f"Starting world image generation for world {input.world_id}")
            
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Загружаем промпт через activity
            prompt_template = await workflow.execute_activity(
                "load_prompt",
                args=[WORLD_IMAGE_PROMPT],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Формируем промпт для генерации промптов изображений
            structure_description = model_to_template(ImagePromptResponse)
            
            # Подготавливаем переменные для промпта
            world_description = format_world_description(world_params)
            prompt = prompt_template.format(
                world_description=world_description,
                structure_description=structure_description
            )
            
            # Увеличиваем счетчик LLM запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем промпты для изображений с помощью LLM
            image_prompts = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "ImagePromptResponse",
                    input.world_id,
                    workflow.info().workflow_id,
                    0.7,  # temperature
                    2048  # max_output_tokens
                ],
                task_queue="ai-worker-llm",  # Используем специализированный LLM worker
                start_to_close_timeout=timedelta(minutes=3),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=2),
                    maximum_interval=timedelta(minutes=1),
                    maximum_attempts=3
                )
            )
            
            workflow.logger.info(f"Generated image prompts for world {input.world_id}")
            
            # Увеличиваем счетчик image запросов (2 изображения)
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_images", 2],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем фоновое изображение и иконку параллельно
            header_task = workflow.execute_activity(
                "generate_image",
                args=[
                    image_prompts["header_prompt"],
                    input.world_id,
                    "world_header",
                    True  # enhance_prompt
                ],
                task_queue="ai-worker-images",
                start_to_close_timeout=timedelta(minutes=5),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=5),
                    maximum_interval=timedelta(minutes=2),
                    maximum_attempts=3
                )
            )
            
            icon_task = workflow.execute_activity(
                "generate_image",
                args=[
                    image_prompts["icon_prompt"],
                    input.world_id,
                    "world_icon",
                    True  # enhance_prompt
                ],
                task_queue="ai-worker-images",
                start_to_close_timeout=timedelta(minutes=5),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=5),
                    maximum_interval=timedelta(minutes=2),
                    maximum_attempts=3
                )
            )
            
            # Ждем завершения обеих задач
            header_image, icon_image = await asyncio.gather(header_task, icon_task)
            
            workflow.logger.info(f"Generated both images for world {input.world_id}")
            
            # Обновляем статус этапа на "Завершен"
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.WORLD_IMAGE, GenerationStatus.COMPLETED],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            
            workflow.logger.info(f"Successfully completed world image generation for world {input.world_id}")
            
            return WorkflowResult(
                success=True,
                data={
                    "header_prompt": image_prompts["header_prompt"],
                    "icon_prompt": image_prompts["icon_prompt"],
                    "header_url": header_image.get("image_url"),
                    "header_id": header_image.get("media_id"),
                    "icon_url": icon_image.get("image_url"),
                    "icon_id": icon_image.get("media_id"),
                    "style_reference": image_prompts.get("style_reference"),
                    "visual_elements": image_prompts.get("visual_elements"),
                    "mood": image_prompts.get("mood"),
                    "color_palette": image_prompts.get("color_palette")
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating world images: {str(e)}"
            workflow.logger.error(f"Workflow failed for world {input.world_id}: {error_msg}")
            
            # Обновляем статус этапа на "Ошибка"
            try:
                await workflow.execute_activity(
                    "update_stage",
                    args=[input.world_id, GenerationStage.WORLD_IMAGE, GenerationStatus.FAILED],
                    task_queue="ai-worker-progress",
                    start_to_close_timeout=timedelta(seconds=30),
                    retry_policy=RetryPolicy(maximum_attempts=3)
                )
            except Exception as update_error:
                workflow.logger.error(f"Failed to update failure status: {str(update_error)}")
            
            return WorkflowResult(success=False, error=error_msg)