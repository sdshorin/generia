from datetime import timedelta
from typing import Dict, Any, List
import math
from temporalio.workflow import ParentClosePolicy

from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..utils.model_to_template import model_to_template
from ..prompts import POST_BATCH_PROMPT, PREVIOUS_POSTS_PROMPT, FIRST_BATCH_POSTS_PROMPT
from ..schemas.post_batch import PostBatchResponse

# Максимальное количество постов, генерируемых за один раз
MAX_POSTS_PER_BATCH = 10

# Максимальная глубина рекурсии для генерации постов
MAX_POST_RECURSION_DEPTH = 30


class GeneratePostBatchInput(TaskInput):
    """Входные данные для workflow генерации пакета постов"""
    world_id: str
    character_id: str
    posts_count: int
    character_detail: Dict[str, Any]
    generated_posts_count: int = 0
    generated_posts_description: str = ""
    count_run: int = 0
    recursion_depth: int = 0
    total_posts_count: int = None
    
    def model_post_init(self, __context):
        """Устанавливает значения по умолчанию (Pydantic v2)"""
        if self.total_posts_count is None:
            self.total_posts_count = self.posts_count


@workflow.defn 
class GeneratePostBatchWorkflow(BaseWorkflow):
    """
    Workflow для генерации пакета постов для персонажа
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
            input = await self.load_task_data(task_ref, GeneratePostBatchInput)
            
            character_name = input.character_detail.get("display_name", "Unknown")
            username = input.character_detail.get("username", "unknown")
            
            # logger.info(f"Starting post batch generation for character {character_name} ({input.character_id})")
            
            # Вычисляем максимально допустимую глубину рекурсии
            max_allowed_depth = min(
                math.ceil(input.total_posts_count / 8) + 1, 
                MAX_POST_RECURSION_DEPTH
            )
            
            # Логируем входные параметры
            workflow.logger.debug(f"Post batch parameters: posts_count={input.posts_count}, "
                        f"total_posts_count={input.total_posts_count}, "
                        f"generated_posts_count={input.generated_posts_count}, "
                        f"recursion_depth={input.recursion_depth}")
            
            # Проверяем глубину рекурсии
            if input.recursion_depth >= max_allowed_depth:
                workflow.logger.warning(f"Maximum recursion depth reached for character {character_name}")
                return WorkflowResult(
                    success=True,
                    data={
                        "posts_count": 0,
                        "total_posts_count": input.generated_posts_count,
                        "remaining_posts": input.posts_count - input.generated_posts_count,
                        "recursion_depth": input.recursion_depth,
                        "max_allowed_depth": max_allowed_depth,
                        "character_id": input.character_id,
                        "character_name": character_name,
                        "username": username,
                        "error": "Maximum recursion depth reached"
                    }
                )
            elif input.total_posts_count == 0:
                workflow.logger.warning(f"No posts for character {character_name}")
                return WorkflowResult(
                    success=True,
                    data={
                        "posts_count": 0,
                        "total_posts_count": input.generated_posts_count,
                        "remaining_posts": input.posts_count - input.generated_posts_count,
                        "recursion_depth": input.recursion_depth,
                        "max_allowed_depth": max_allowed_depth,
                        "character_id": input.character_id,
                        "character_name": character_name,
                        "username": username,
                        "error": "No posts for character"
                    }
                )

            
            # Ограничиваем размер текущего пакета
            current_batch_size = min(
                input.posts_count - input.generated_posts_count, 
                MAX_POSTS_PER_BATCH
            )
            
            if current_batch_size <= 0:
                workflow.logger.info(f"All posts generated for character {character_name}")
                return WorkflowResult(
                    success=True,
                    data={
                        "posts_count": 0,
                        "total_posts_count": input.generated_posts_count,
                        "character_id": input.character_id,
                        "message": "All posts already generated"
                    }
                )
            
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Формируем промпт
            prompt = await self._build_post_batch_prompt(input, world_params, current_batch_size)
            
            # Увеличиваем счетчик LLM запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем пакет постов
            post_batch = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "PostBatchResponse",
                    input.world_id,
                    workflow.info().workflow_id,
                    0.9,  # temperature
                    6144  # max_output_tokens
                ],
                task_queue="ai-worker-llm",
                start_to_close_timeout=timedelta(minutes=8),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=3),
                    maximum_interval=timedelta(minutes=3),
                    maximum_attempts=3
                )
            )
            
            posts = post_batch.get("posts", [])
            
            workflow.logger.info(f"LLM generated {len(posts)} posts (requested {current_batch_size}) for character {character_name}")
            
            # Корректируем количество постов, если нужно
            posts = self._adjust_posts_count(posts, current_batch_size, character_name)
            
            # Создаем описание сгенерированных постов
            post_descriptions = []
            for post in posts:
                desc = f"Тема: {post.get('topic', '')}. Краткое содержание: {post.get('content_brief', '')}. Тон: {post.get('emotional_tone', '')}. Тип: {post.get('post_type', '')}."
                post_descriptions.append(desc)
            
            # Объединяем с предыдущим описанием
            generated_posts_description = input.generated_posts_description
            if post_descriptions:
                if generated_posts_description:
                    generated_posts_description += "\n\n"
                generated_posts_description += "\n".join(post_descriptions)
            
            # Проверяем, что LLM вернула хотя бы один пост после корректировки
            if len(posts) == 0:
                workflow.logger.warning(f"No posts available after adjustment for character {character_name}")
                return WorkflowResult(
                    success=True,
                    data={
                        "posts_count": 0,
                        "total_posts_count": input.generated_posts_count,
                        "remaining_posts": input.posts_count - input.generated_posts_count,
                        "character_id": input.character_id,
                        "character_name": character_name,
                        "username": username,
                        "error": "No posts generated",
                        "recursion_depth": input.recursion_depth,
                        "max_allowed_depth": max_allowed_depth
                    }
                )
            
            workflow.logger.info(f"Final posts count: {len(posts)} for character {character_name}")
            
            # Создаем и запускаем child workflows для каждого поста
            from .generate_post_workflow import GeneratePostInput
            
            for i, post in enumerate(posts):
                post_input = GeneratePostInput(
                    world_id=input.world_id,
                    character_id=input.character_id,
                    post_data=post,
                    character_detail=input.character_detail,
                    post_index=i + input.generated_posts_count,
                    character_index=input.character_detail.get("character_index", 0)
                )
                post_task_ref = await self.save_task_data(post_input, input.world_id)
                
                post_workflow_id = self.get_workflow_id(
                    input.world_id,
                    "generate-post",
                    post_task_ref.task_id
                )
                
                await workflow.start_child_workflow(
                    "GeneratePostWorkflow",
                    post_task_ref,
                    id=post_workflow_id,
                    task_queue="ai-worker-main",
                    parent_close_policy=ParentClosePolicy.ABANDON
                )

            
            # Обновляем счетчики (используем реальное количество созданных постов)
            actual_posts_generated = len(posts)
            new_generated_count = input.generated_posts_count + actual_posts_generated
            remaining_posts = input.posts_count - new_generated_count
            

            # Если нужно сгенерировать еще посты, запускаем следующий пакет
            if remaining_posts > 0:
                next_batch_input = GeneratePostBatchInput(
                    world_id=input.world_id,
                    character_id=input.character_id,
                    posts_count=input.posts_count,
                    character_detail=input.character_detail,
                    generated_posts_count=new_generated_count,
                    generated_posts_description=generated_posts_description,
                    count_run=input.count_run + 1,
                    recursion_depth=input.recursion_depth + 1,
                    total_posts_count=input.total_posts_count
                )
                next_batch_task_ref = await self.save_task_data(next_batch_input, input.world_id)
                
                next_batch_workflow_id = self.get_workflow_id(
                    input.world_id,
                    "generate-post-batch",
                    next_batch_task_ref.task_id
                )
                
                await workflow.start_child_workflow(
                    "GeneratePostBatchWorkflow",
                    next_batch_task_ref,
                    id=next_batch_workflow_id,
                    task_queue="ai-worker-main",
                    parent_close_policy=ParentClosePolicy.ABANDON
                )
                
                workflow.logger.info(f"Started next post batch workflow for {remaining_posts} remaining posts")
            
            workflow.logger.info(f"Successfully completed post batch generation for character {character_name}")
            
            return WorkflowResult(
                success=True,
                data={
                    "posts_count": actual_posts_generated,
                    "total_posts_count": new_generated_count,
                    "remaining_posts": remaining_posts,
                    "recursion_depth": input.recursion_depth,
                    "character_id": input.character_id,
                    "character_name": character_name,
                    "generated_posts_description": generated_posts_description
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating post batch: {str(e)}"
            workflow.logger.error(f"Workflow failed for character {input.character_id}: {error_msg}")
            raise 
            # return WorkflowResult(success=False, error=error_msg)
    
    async def _build_post_batch_prompt(
        self, 
        input: GeneratePostBatchInput, 
        world_params: Dict[str, Any], 
        current_batch_size: int
    ) -> str:
        """
        Строит промпт для генерации пакета постов
        
        Args:
            input: Входные данные
            world_params: Параметры мира
            current_batch_size: Размер текущего пакета
            
        Returns:
            Сформированный промпт
        """
        # Загружаем базовый промпт
        prompt_template = await workflow.execute_activity(
            "load_prompt", 
            args=[POST_BATCH_PROMPT],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Извлекаем данные персонажа
        username = input.character_detail.get("username", "unknown")
        display_name = input.character_detail.get("display_name", "Unknown")
        bio = input.character_detail.get("bio", "")
        personality = input.character_detail.get("personality", "")
        interests = input.character_detail.get("interests", [])
        speaking_style = input.character_detail.get("speaking_style", "")
        
        # Преобразуем списки в строки
        interests_str = ", ".join(interests) if isinstance(interests, list) else str(interests)
        
        # Формируем информацию о предыдущих постах
        previous_posts_info = ""
        future_posts_count = input.posts_count - input.generated_posts_count - current_batch_size
        
        if input.generated_posts_count > 0:
            # Есть уже сгенерированные посты
            previous_posts_template = await workflow.execute_activity(
                "load_prompt", 
                args=[PREVIOUS_POSTS_PROMPT],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            previous_posts_info = previous_posts_template.format(
                count_run=input.count_run,
                count=input.generated_posts_count,
                total_posts_count=input.total_posts_count,
                current_batch_size=current_batch_size,
                future_posts_count=future_posts_count,
                description=input.generated_posts_description
            )
        elif input.posts_count > current_batch_size:
            # Первая генерация, но будут еще
            first_batch_template = await workflow.execute_activity(
                "load_prompt", 
                args=[FIRST_BATCH_POSTS_PROMPT],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            previous_posts_info = first_batch_template.format(
                total_posts_count=input.total_posts_count,
                current_batch_size=current_batch_size,
                future_posts_count=future_posts_count
            )
        
        # Формируем описание мира
        world_description = await workflow.execute_activity(
            "format_world_description",
            args=[world_params],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        structure_description = model_to_template(PostBatchResponse)
        
        # Формируем итоговый промпт
        prompt = prompt_template.format(
            world_description=world_description,
            character_name=display_name,
            character_description=f"{bio} {personality}",
            interests=interests_str,
            speaking_style=speaking_style,
            common_topics=input.character_detail.get("common_topics", ""),
            appearance=input.character_detail.get("appearance", ""),
            secret=input.character_detail.get("secret", ""),
            daily_routine=input.character_detail.get("daily_routine", ""),
            avatar_description=input.character_detail.get("avatar_description", ""),
            avatar_style=input.character_detail.get("avatar_style", ""),
            posts_count=current_batch_size,
            previous_posts_info=previous_posts_info,
            structure_description=structure_description
        )
        
        return prompt
    
    def _adjust_posts_count(self, posts: List[Dict[str, Any]], target_count: int, character_name: str) -> List[Dict[str, Any]]:
        """
        Корректирует количество постов до целевого значения
        
        Args:
            posts: Список постов от LLM
            target_count: Целевое количество постов
            character_name: Имя персонажа для логирования
            
        Returns:
            Скорректированный список постов
        """
        current_count = len(posts)
        
        if current_count == target_count:
            workflow.logger.info(f"Posts count already matches target for {character_name}: {current_count}")
            return posts
        
        workflow.logger.info(f"Adjusting posts count for {character_name} from {current_count} to {target_count}")
        
        if current_count > target_count:
            # Обрезаем лишние посты
            posts = posts[:target_count]
            workflow.logger.info(f"Trimmed posts for {character_name} to {len(posts)}")
            
        elif current_count < target_count:
            # Дублируем существующие посты с небольшими изменениями
            additional_needed = target_count - current_count
            workflow.logger.info(f"Need to generate {additional_needed} additional posts for {character_name}")
            
            # Дублируем посты циклично
            for i in range(additional_needed):
                if not posts:  # Если вообще нет постов, создаем базовый
                    additional_post = {
                        "topic": f"General topic {i+1}",
                        "content_brief": f"A post about daily life or thoughts {i+1}",
                        "emotional_tone": "neutral",
                        "post_type": "personal",
                        "relevance_to_character": "Reflects the character's everyday experiences"
                    }
                else:
                    # Берем пост циклично и немного модифицируем
                    base_post = posts[i % len(posts)].copy()
                    # Немного изменяем содержание
                    base_post["topic"] = f"{base_post.get('topic', 'Topic')} (variant {i+1})"
                    base_post["content_brief"] = f"{base_post.get('content_brief', 'Content')} (variation {i+1})"
                    additional_post = base_post
                
                posts.append(additional_post)
            
            workflow.logger.info(f"Added {additional_needed} posts for {character_name}, total: {len(posts)}")
        
        return posts