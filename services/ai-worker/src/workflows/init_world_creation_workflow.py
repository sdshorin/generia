from datetime import timedelta
from typing import Dict, Any, Optional
from temporalio.workflow import ParentClosePolicy

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput
from ..constants import GenerationStage, GenerationStatus, DEFAULT_VALUES


class InitWorldCreationInput(TaskInput):
    """Входные данные для workflow инициализации создания мира"""
    world_id: str
    world_name: str
    world_prompt: str  # Изменено с user_prompt на world_prompt
    characters_count: int  # Изменено с users_count на characters_count
    posts_count: int
    api_call_limits_llm: Optional[int] = None
    api_call_limits_images: Optional[int] = None
    
    def model_post_init(self, __context):
        """Устанавливает значения по умолчанию (Pydantic v2)"""
        if self.api_call_limits_llm is None:
            self.api_call_limits_llm = DEFAULT_VALUES["api_call_limits_LLM"]
        if self.api_call_limits_images is None:
            self.api_call_limits_images = DEFAULT_VALUES["api_call_limits_images"]


@workflow.defn
class InitWorldCreationWorkflow(BaseWorkflow):
    """
    Workflow для инициализации создания мира
    Это главный workflow, который запускает всю цепочку генерации
    """
    
    @workflow.run
    async def run(self, input: InitWorldCreationInput) -> WorkflowResult:
        """
        Основной метод выполнения workflow
        
        Args:
            input: Входные данные для инициализации
            
        Returns:
            Результат выполнения workflow
        """
        try:
            workflow.logger.info(f"Initializing world creation for world {input.world_id}")
            
            # Проверяем валидность промпта
            if not input.world_prompt:
                raise ValueError("World prompt is required")
            
            # Инициализируем запись о статусе генерации мира
            init_result = await workflow.execute_activity(
                "initialize_world_generation",
                args=[
                    input.world_id,
                    input.characters_count,
                    input.posts_count,
                    input.world_prompt,
                    input.api_call_limits_llm,
                    input.api_call_limits_images
                ],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Обновляем статус этапа инициализации на "Завершен"
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.INITIALIZING, GenerationStatus.COMPLETED],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Создаем задачу и запускаем child workflow для генерации описания мира
            from .generate_world_description_workflow import GenerateWorldDescriptionInput
            
            description_input = GenerateWorldDescriptionInput(
                world_id=input.world_id,
                user_prompt=input.world_prompt,
                users_count=input.characters_count,
                posts_count=input.posts_count
            )
            
            # Сохраняем задачу в MongoDB и получаем TaskRef
            description_task_ref = await self.save_task_data(description_input, input.world_id)
            
            world_description_workflow_id = self.get_workflow_id(
                input.world_id,
                "generate-world-description",
                description_task_ref.task_id
            )
            
            # Запускаем child workflow с TaskRef
            await workflow.start_child_workflow(
                "GenerateWorldDescriptionWorkflow",
                description_task_ref,
                id=world_description_workflow_id,
                task_queue="ai-worker-main",
                parent_close_policy=ParentClosePolicy.ABANDON
            )
            
            # Обновляем статус следующего этапа
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.WORLD_DESCRIPTION, GenerationStatus.IN_PROGRESS],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Ждем результата child workflow (опционально)
            # child_result = await child_handle.result()
            
            workflow.logger.info(f"Successfully initialized world creation for world {input.world_id}")
            
            return WorkflowResult(
                success=True,
                data={
                    "message": "World generation initialized successfully",
                    "world_id": input.world_id,
                    "child_workflow_id": world_description_workflow_id,
                    "users_count": input.characters_count,
                    "posts_count": input.posts_count
                }
            )
            
        except Exception as e:
            error_msg = f"Error initializing world creation: {str(e)}"
            workflow.logger.error(f"Workflow failed for world {input.world_id}: {error_msg}")
            raise
            # # Обновляем статус инициализации на "Ошибка"
            # try:
            #     await workflow.execute_activity(
            #         "update_stage",
            #         args=[input.world_id, GenerationStage.INITIALIZING, GenerationStatus.FAILED],
            #         task_queue="ai-worker-progress",
            #         start_to_close_timeout=timedelta(seconds=30),
            #         retry_policy=RetryPolicy(maximum_attempts=3)
            #     )
                
            #     # Устанавливаем общий статус генерации как неудачный
            #     await workflow.execute_activity(
            #         "update_progress",
            #         args=[input.world_id, {"status": GenerationStatus.FAILED}],
            #         task_queue="ai-worker-progress",
            #         start_to_close_timeout=timedelta(seconds=30),
            #         retry_policy=RetryPolicy(maximum_attempts=3)
            #     )
            # except Exception as update_error:
            #     workflow.logger.error(f"Failed to update failure status: {str(update_error)}")
            
            # return WorkflowResult(success=False, error=error_msg)