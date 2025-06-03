from datetime import timedelta
from typing import Dict, Any
from temporalio.workflow import ParentClosePolicy

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..constants import TaskType, GenerationStage, GenerationStatus
from ..prompts import WORLD_DESCRIPTION_PROMPT
from ..utils.model_to_template import model_to_template
from ..schemas.world_description import WorldDescriptionResponse


class GenerateWorldDescriptionInput(TaskInput):
    """Входные данные для workflow генерации описания мира"""
    world_id: str
    user_prompt: str
    users_count: int = 10
    posts_count: int = 50


@workflow.defn
class GenerateWorldDescriptionWorkflow(BaseWorkflow):
    """
    Workflow для генерации описания мира
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
            input = await self.load_task_data(task_ref, GenerateWorldDescriptionInput)
            
            workflow.logger.info(f"Starting world description generation for world {input.world_id}")
            
            # Обновляем статус этапа на "В процессе"
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.WORLD_DESCRIPTION, GenerationStatus.IN_PROGRESS],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Загружаем промпт через activity
            prompt_template = await workflow.execute_activity(
                "load_prompt",
                args=[WORLD_DESCRIPTION_PROMPT],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Подготавливаем промпт для LLM
            structure_description = model_to_template(WorldDescriptionResponse)
            prompt = prompt_template.format(
                user_prompt=input.user_prompt, 
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
            
            # Генерируем описание мира с помощью LLM
            llm_result = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "WorldDescriptionResponse",
                    input.world_id,
                    workflow.info().workflow_id,
                    0.8,  # temperature
                    4096  # max_output_tokens
                ],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(minutes=5),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=2),
                    maximum_interval=timedelta(minutes=2),
                    maximum_attempts=5
                )
            )
            
            # Сохраняем параметры мира в БД
            save_result = await workflow.execute_activity(
                "save_world_parameters",
                args=[llm_result, input.world_id, input.users_count, input.posts_count ],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # save_result теперь уже содержит результат или вызовет exception
            
            # Обновляем статус этапа на "Завершен"
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.WORLD_DESCRIPTION, GenerationStatus.COMPLETED],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Запускаем следующие workflows
            await self._start_next_workflows(input)
            
            workflow.logger.info(f"Successfully completed world description generation for world {input.world_id}")
            
            return WorkflowResult(
                success=True,
                data={
                    "world_name": llm_result.get("name"),
                    "world_description": llm_result.get("description"),
                    "world_theme": llm_result.get("theme"),
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating world description: {str(e)}"
            workflow.logger.error(f"Workflow failed for world {input.world_id}: {error_msg}")
            raise
            # # Обновляем статус этапа на "Ошибка"
            # try:
            #     await workflow.execute_activity(
            #         "update_stage",
            #         args=[input.world_id, GenerationStage.WORLD_DESCRIPTION, GenerationStatus.FAILED],
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
    
    async def _start_next_workflows(self, input: GenerateWorldDescriptionInput):
        """
        Запускает следующие workflows для генерации изображения и персонажей
        
        Args:
            input: Входные данные исходного workflow
        """
        try:
            # Создаем и запускаем workflow для генерации изображения мира
            from .generate_world_image_workflow import GenerateWorldImageInput
            
            image_input = GenerateWorldImageInput(world_id=input.world_id)
            image_task_ref = await self.save_task_data(image_input, input.world_id)
            
            world_image_workflow_id = self.get_workflow_id(
                input.world_id, 
                "generate-world-image",
                image_task_ref.task_id
            )
            
            await workflow.start_child_workflow(
                "GenerateWorldImageWorkflow",
                image_task_ref,
                id=world_image_workflow_id,
                task_queue="ai-worker-main",
                parent_close_policy=ParentClosePolicy.ABANDON
            )
            
            # Создаем и запускаем workflow для генерации персонажей
            from .generate_character_batch_workflow import GenerateCharacterBatchInput
            
            character_batch_input = GenerateCharacterBatchInput(
                world_id=input.world_id,
                users_count=input.users_count,
                posts_count=input.posts_count
            )
            character_task_ref = await self.save_task_data(character_batch_input, input.world_id)
            
            character_batch_workflow_id = self.get_workflow_id(
                input.world_id, 
                "generate-character-batch",
                character_task_ref.task_id
            )
            
            await workflow.start_child_workflow(
                "GenerateCharacterBatchWorkflow",
                character_task_ref,
                id=character_batch_workflow_id,
                task_queue="ai-worker-main",
                parent_close_policy=ParentClosePolicy.ABANDON
            )
            
            # Обновляем статусы следующих этапов
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.WORLD_IMAGE, GenerationStatus.IN_PROGRESS],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            await workflow.execute_activity(
                "update_stage",
                args=[input.world_id, GenerationStage.CHARACTERS, GenerationStatus.IN_PROGRESS],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            workflow.logger.info(f"Started child workflows for world {input.world_id}")
            
        except Exception as e:
            workflow.logger.error(f"Failed to start child workflows for world {input.world_id}: {str(e)}")
            # Не прерываем основной workflow из-за ошибки запуска дочерних