from datetime import timedelta
from typing import Dict, Any

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..constants import GenerationStage, GenerationStatus, MediaType
from ..utils.model_to_template import model_to_template
from ..prompts import CHARACTER_AVATAR_PROMPT
from ..schemas.character_avatar import CharacterAvatarPromptResponse


class GenerateCharacterAvatarInput(TaskInput):
    """Входные данные для workflow генерации аватара персонажа"""
    world_id: str
    character_id: str
    character_detail: Dict[str, Any]
    character_index: int = 0


@workflow.defn
class GenerateCharacterAvatarWorkflow(BaseWorkflow):
    """
    Workflow для генерации аватара персонажа
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
            input = await self.load_task_data(task_ref, GenerateCharacterAvatarInput)
            
            character_name = input.character_detail.get("display_name", "Unknown")
            username = input.character_detail.get("username", "unknown")
            
            workflow.logger.info(f"Starting avatar generation for character {character_name} ({input.character_id})")
            
            # Получаем описания аватара из детального описания персонажа
            appearance_description = input.character_detail.get("appearance", "")
            avatar_description = input.character_detail.get("avatar_description", "")
            avatar_style = input.character_detail.get("avatar_style", "")
            
            if not avatar_description:
                workflow.logger.warning(f"No avatar description for character {character_name}")
                return WorkflowResult(
                    success=True,
                    data={"message": "No avatar description provided"}
                )
            
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Формируем промпт для оптимизации промпта аватара
            prompt = await self._build_avatar_prompt(
                input, world_params, character_name, appearance_description, avatar_description, avatar_style
            )
            
            # Увеличиваем счетчик LLM запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем оптимизированный промпт для создания аватара
            avatar_response = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "CharacterAvatarPromptResponse",
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
            
            optimized_avatar_prompt = avatar_response.get("prompt", avatar_description)
            
            workflow.logger.info(f"Generated avatar prompt for character {character_name}: {optimized_avatar_prompt}")
            
            # Увеличиваем счетчик image запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_images", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем изображение аватара
            avatar_image = await workflow.execute_activity(
                "generate_image",
                args=[
                    optimized_avatar_prompt,
                    input.world_id,
                    "character_avatar",
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
            
            workflow.logger.info(f"Generated avatar image for character {character_name}")
            
            # Получаем URL аватара и ID медиа
            avatar_url = avatar_image.get("image_url")
            avatar_id = avatar_image.get("media_id")
            
            # Обновляем персонажа с новым аватаром
            character_result = await workflow.execute_activity(
                "update_character_avatar",
                args=[input.character_id, avatar_id],
                task_queue="ai-worker-services",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            workflow.logger.info(f"Successfully generated avatar for character {character_name} ({input.character_id})")
            
            return WorkflowResult(
                success=True,
                data={
                    "username": username,
                    "display_name": character_name,
                    "avatar_url": avatar_url,
                    "avatar_id": avatar_id,
                    "character_id": input.character_id,
                    "avatar_prompt": optimized_avatar_prompt
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating character avatar: {str(e)}"
            workflow.logger.error(f"Workflow failed for character {input.character_id}: {error_msg}")
            raise
            # return WorkflowResult(success=False, error=error_msg)
    
    async def _build_avatar_prompt(
        self, 
        input: GenerateCharacterAvatarInput, 
        world_params: Dict[str, Any],
        character_name: str,
        appearance_description: str,
        avatar_description: str,
        avatar_style: str
    ) -> str:
        """
        Строит промпт для оптимизации промпта аватара
        
        Args:
            input: Входные данные
            world_params: Параметры мира
            character_name: Имя персонажа
            appearance_description: Описание внешности
            avatar_description: Описание аватара
            avatar_style: Стиль аватара
            
        Returns:
            Сформированный промпт
        """
        # Загружаем промпт из файла
        prompt_template = await workflow.execute_activity(
            "load_prompt",
            args=[CHARACTER_AVATAR_PROMPT],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Генерируем описание структуры ответа
        structure_description = model_to_template(CharacterAvatarPromptResponse)
        
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
            appearance_description=appearance_description,
            avatar_description=avatar_description,
            avatar_style=avatar_style,
            structure_description=structure_description
        )
        
        return prompt