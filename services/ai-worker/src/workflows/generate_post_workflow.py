from datetime import timedelta
from typing import Dict, Any
from temporalio.workflow import ParentClosePolicy

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..utils.model_to_template import model_to_template
from ..prompts import POST_DETAIL_PROMPT
from ..schemas.post import PostDetailResponse


class GeneratePostInput(TaskInput):
    """Входные данные для workflow генерации отдельного поста"""
    world_id: str
    character_id: str
    post_data: Dict[str, Any]
    character_detail: Dict[str, Any]
    post_index: int = 0
    character_index: int = 0


@workflow.defn
class GeneratePostWorkflow(BaseWorkflow):
    """
    Workflow для генерации отдельного поста
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
            input = await self.load_task_data(task_ref, GeneratePostInput)
            
            character_name = input.character_detail.get("display_name", "Unknown")
            username = input.character_detail.get("username", "unknown")
            
            workflow.logger.info(f"Starting post generation for character {character_name} ({input.character_id})")
            
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Формируем промпт для генерации детального поста
            prompt = await self._build_post_detail_prompt(input, world_params)
            
            # Увеличиваем счетчик LLM запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем пост с помощью LLM
            post_detail = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "PostDetailResponse",
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
            
            workflow.logger.info(f"Generated post detail for character {character_name}")
            
            # Создаем и запускаем workflow для генерации изображения к посту
            from .generate_post_image_workflow import GeneratePostImageInput
            
            image_input = GeneratePostImageInput(
                world_id=input.world_id,
                character_id=input.character_id,
                post_detail=post_detail,
                character_detail=input.character_detail,
                post_index=input.post_index,
                character_index=input.character_index
            )
            image_task_ref = await self.save_task_data(image_input, input.world_id)
            
            post_image_workflow_id = self.get_workflow_id(
                input.world_id,
                "generate-post-image",
                image_task_ref.task_id
            )
            
            await workflow.start_child_workflow(
                "GeneratePostImageWorkflow",
                image_task_ref,
                id=post_image_workflow_id,
                task_queue="ai-worker-main",
                parent_close_policy=ParentClosePolicy.ABANDON
            )
            
            workflow.logger.info(f"Started post image generation for character {character_name}")
            
            return WorkflowResult(
                success=True,
                data={
                    "character_id": input.character_id,
                    "character_name": character_name,
                    "username": username,
                    "content": post_detail["content"],
                    "image_prompt": post_detail["image_prompt"],
                    "hashtags": post_detail["hashtags"],
                    "mood": post_detail["mood"],
                    "context": post_detail["context"],
                    "image_workflow_id": post_image_workflow_id,
                }
            )
            
            
        except Exception as e:
            error_msg = f"Error generating post: {str(e)}"
            workflow.logger.error(f"Workflow failed for character {input.character_id}: {error_msg}")
            raise
            # return WorkflowResult(success=False, error=error_msg)
    
    async def _build_post_detail_prompt(
        self, 
        input: GeneratePostInput, 
        world_params: Dict[str, Any]
    ) -> str:
        """
        Строит промпт для генерации детального поста
        
        Args:
            input: Входные данные
            world_params: Параметры мира
            
        Returns:
            Сформированный промпт
        """
        # Загружаем промпт из файла
        prompt_template = await workflow.execute_activity(
            "load_prompt",
            args=[POST_DETAIL_PROMPT],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Извлекаем данные поста из post_data
        post_topic = input.post_data.get("topic", "")
        post_brief = input.post_data.get("content_brief", "")
        emotional_tone = input.post_data.get("emotional_tone", "")
        post_type = input.post_data.get("post_type", "")
        relevance_to_character = input.post_data.get("relevance_to_character", "")
        
        # Извлекаем данные персонажа
        character_name = input.character_detail.get("display_name", "Unknown")
        character_description = input.character_detail.get("personality", "")
        speaking_style = input.character_detail.get("speaking_style", "")
        appearance = input.character_detail.get("appearance", "")
        secret = input.character_detail.get("secret", "")
        daily_routine = input.character_detail.get("daily_routine", "")
        avatar_description = input.character_detail.get("avatar_description", "")
        avatar_style = input.character_detail.get("avatar_style", "")
        
        # Генерируем описание структуры ответа
        structure_description = model_to_template(PostDetailResponse)
        
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
            character_description=character_description,
            speaking_style=speaking_style,
            appearance=appearance,
            secret=secret,
            daily_routine=daily_routine,
            avatar_description=avatar_description,
            avatar_style=avatar_style,
            post_topic=post_topic,
            post_brief=post_brief,
            emotional_tone=emotional_tone,
            post_type=post_type,
            relevance_to_character=relevance_to_character,
            structure_description=structure_description
        )
        
        return prompt