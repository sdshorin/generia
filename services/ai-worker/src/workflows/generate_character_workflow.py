from datetime import timedelta
from typing import Dict, Any, List
from temporalio.workflow import ParentClosePolicy

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..constants import GenerationStage, GenerationStatus
from ..utils.model_to_template import model_to_template
from ..prompts import CHARACTER_DETAIL_PROMPT
from ..schemas.character import CharacterDetailResponse


class GenerateCharacterInput(TaskInput):
    """Входные данные для workflow генерации персонажа"""
    world_id: str
    character_data: Dict[str, Any]
    posts_per_character: int = 5


@workflow.defn
class GenerateCharacterWorkflow(BaseWorkflow):
    """
    Workflow для генерации отдельного персонажа
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
            input = await self.load_task_data(task_ref, GenerateCharacterInput)
            
            character_concept = input.character_data.get("username", "")
            workflow.logger.info(f"Starting character generation for {character_concept} in world {input.world_id}")
            
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Формируем промпт для генерации детального описания персонажа
            prompt = await self._build_character_detail_prompt(input, world_params)
            
            # Увеличиваем счетчик LLM запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем детальное описание персонажа
            character_detail = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "CharacterDetailResponse",
                    input.world_id,
                    workflow.info().workflow_id,
                    0.8,  # temperature
                    4096  # max_output_tokens
                ],
                task_queue="ai-worker-llm",
                start_to_close_timeout=timedelta(minutes=5),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=2),
                    maximum_interval=timedelta(minutes=2),
                    maximum_attempts=3
                )
            )
            
            workflow.logger.info(f"Generated character detail for {character_detail['username']}")
            
            # Создаем персонажа в Character Service
            character_id = await workflow.execute_activity(
                "create_character",
                args=[character_detail, input.world_id],
                task_queue="ai-worker-services",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            workflow.logger.info(f"Created character {character_id} in Character Service")

            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "users_created", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Создаем и запускаем генерацию аватара персонажа (параллельно с постами)
            from .generate_character_avatar_workflow import GenerateCharacterAvatarInput
            
            avatar_input = GenerateCharacterAvatarInput(
                world_id=input.world_id,
                character_id=character_id,
                character_detail=character_detail
            )
            avatar_task_ref = await self.save_task_data(avatar_input, input.world_id)
            
            avatar_workflow_id = self.get_workflow_id(
                input.world_id,
                "generate-character-avatar",
                avatar_task_ref.task_id
            )
            
            await workflow.start_child_workflow(
                "GenerateCharacterAvatarWorkflow",
                avatar_task_ref,
                id=avatar_workflow_id,
                task_queue="ai-worker-main",
                parent_close_policy=ParentClosePolicy.ABANDON
            )
            
            # Создаем и запускаем генерацию постов для персонажа
            from .generate_post_batch_workflow import GeneratePostBatchInput
            
            posts_input = GeneratePostBatchInput(
                world_id=input.world_id,
                character_id=character_id,
                posts_count=input.character_data.get("posts_count", input.posts_per_character),
                character_detail=character_detail
            )
            posts_task_ref = await self.save_task_data(posts_input, input.world_id)
            
            posts_workflow_id = self.get_workflow_id(
                input.world_id,
                "generate-post-batch",
                posts_task_ref.task_id
            )
            
            await workflow.start_child_workflow(
                "GeneratePostBatchWorkflow", 
                posts_task_ref,
                id=posts_workflow_id,
                task_queue="ai-worker-main",
                parent_close_policy=ParentClosePolicy.ABANDON
            )
            
            workflow.logger.info(f"Started avatar and posts generation for character {character_id}")
            
            # Можем дождаться завершения генерации аватара (он быстрее)
            # avatar_result = await avatar_handle.result()
            
            workflow.logger.info(f"Successfully completed character generation for {character_concept}")
            
            return WorkflowResult(
                success=True,
                data={
                    "character_id": character_id,
                    "username": character_detail["username"],
                    "display_name": character_detail["display_name"],
                    "avatar_workflow_id": avatar_workflow_id,
                    "posts_workflow_id": posts_workflow_id
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating character: {str(e)}"
            workflow.logger.error(f"Workflow failed for character {character_concept}: {error_msg}")
            raise
            # return WorkflowResult(success=False, error=error_msg)
    
    async def _build_character_detail_prompt(
        self, 
        input: GenerateCharacterInput, 
        world_params: Dict[str, Any]
    ) -> str:
        """
        Строит промпт для генерации детального описания персонажа
        
        Args:
            input: Входные данные
            world_params: Параметры мира
            
        Returns:
            Сформированный промпт
        """
        # Загружаем промпт из файла
        prompt_template = await workflow.execute_activity(
            "load_prompt",
            args=[CHARACTER_DETAIL_PROMPT],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Извлекаем данные персонажа
        character_concept = input.character_data.get("username", "")
        role_in_world = input.character_data.get("role_in_world", "")
        personality_traits = input.character_data.get("personality_traits", [])
        interests = input.character_data.get("interests", [])
        
        # Преобразуем списки в строки для промпта
        personality_traits_str = ", ".join(personality_traits) if isinstance(personality_traits, list) else str(personality_traits)
        interests_str = ", ".join(interests) if isinstance(interests, list) else str(interests)
        
        # Генерируем описание структуры ответа
        structure_description = model_to_template(CharacterDetailResponse)
        
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
            character_concept=character_concept,
            role_in_world=role_in_world,
            personality_traits=personality_traits_str,
            interests=interests_str,
            structure_description=structure_description
        )
        
        return prompt