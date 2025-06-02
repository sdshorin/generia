from datetime import timedelta
from typing import Dict, Any, List
import math

from temporalio.workflow import ParentClosePolicy
from temporalio import activity, workflow
from temporalio.common import RetryPolicy

from ..temporal.base_workflow import BaseWorkflow, WorkflowResult
from ..temporal.task_base import TaskInput, TaskRef
from ..constants import GenerationStage, GenerationStatus
from ..utils.model_to_template import model_to_template
from ..prompts import CHARACTER_BATCH_PROMPT, PREVIOUS_CHARACTERS_PROMPT, FIRST_BATCH_CHARACTERS_PROMPT
from ..schemas.character_batch import CharacterBatchResponse

# Максимальное количество персонажей, генерируемых за один раз
MAX_CHARACTERS_PER_BATCH = 10

# Максимальная глубина рекурсии для генерации персонажей
MAX_CHARACTER_RECURSION_DEPTH = 50


class GenerateCharacterBatchInput(TaskInput):
    """Входные данные для workflow генерации пакета персонажей"""
    world_id: str
    users_count: int
    posts_count: int
    remaining_posts_count: int = None
    total_users_count: int = None
    generated_characters_description: str = ""
    generated_count: int = 0
    count_run: int = 0
    recursion_depth: int = 0
    
    def model_post_init(self, __context):
        """Устанавливает значения по умолчанию (Pydantic v2)"""
        if self.remaining_posts_count is None:
            self.remaining_posts_count = self.posts_count
        if self.total_users_count is None:
            self.total_users_count = self.users_count


@workflow.defn
class GenerateCharacterBatchWorkflow(BaseWorkflow):
    """
    Workflow для генерации пакета персонажей
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
            input = await self.load_task_data(task_ref, GenerateCharacterBatchInput)
            
            # logger.info(f"Starting character batch generation for world {input.world_id}")
            
            # Вычисляем максимально допустимую глубину рекурсии
            max_allowed_depth = min(
                math.ceil(input.total_users_count / 8) + 1, 
                MAX_CHARACTER_RECURSION_DEPTH
            )
            
            # Логируем входные параметры
            workflow.logger.debug(f"Character batch parameters: users_count={input.users_count}, "
                        f"total_users_count={input.total_users_count}, "
                        f"generated_count={input.generated_count}, "
                        f"recursion_depth={input.recursion_depth}, "
                        f"max_allowed_depth={max_allowed_depth}")
            
            # Проверяем глубину рекурсии
            if input.recursion_depth >= max_allowed_depth:
                workflow.logger.warning(f"Maximum recursion depth reached for world {input.world_id}")
                return WorkflowResult(
                    success=True,
                    data={
                        "characters_count": 0,
                        "total_characters_count": input.generated_count,
                        "remaining_characters": input.users_count,
                        "recursion_depth": input.recursion_depth,
                        "max_allowed_depth": max_allowed_depth,
                        "error": f"Maximum recursion depth reached"
                    }
                )
            
            # Ограничиваем размер текущего пакета
            current_batch_size = min(input.users_count, MAX_CHARACTERS_PER_BATCH)
            
            posts_count_for_batch = int(input.remaining_posts_count * (current_batch_size / input.users_count))
                        
            # Получаем параметры мира
            world_params = await workflow.execute_activity(
                "get_world_parameters",
                args=[input.world_id],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            workflow.logger.info(f"get_world_parameters: {world_params}")
            # Формируем промпт
            prompt = await self._build_character_batch_prompt(input, world_params, current_batch_size, posts_count_for_batch)
            
            # Увеличиваем счетчик LLM запросов
            await workflow.execute_activity(
                "increment_counter",
                args=[input.world_id, "api_calls_made_LLM", 1],
                task_queue="ai-worker-progress",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            
            # Генерируем пакет персонажей
            character_batch = await workflow.execute_activity(
                "generate_structured_content",
                args=[
                    prompt,
                    "CharacterBatchResponse",
                    input.world_id,
                    workflow.info().workflow_id,
                    0.9,  # temperature
                    8192  # max_output_tokens
                ],
                task_queue="ai-worker-llm",
                start_to_close_timeout=timedelta(minutes=10),
                retry_policy=RetryPolicy(
                    initial_interval=timedelta(seconds=3),
                    maximum_interval=timedelta(minutes=3),
                    maximum_attempts=3
                )
            )
            
            characters = character_batch.get("characters", [])
            
            workflow.logger.info(f"LLM generated {len(characters)} characters (requested {current_batch_size}) for world {input.world_id}")
            
            # Корректируем количество персонажей, если нужно
            characters = self._adjust_characters_count(characters, current_batch_size)
            
            # Корректируем количество постов для персонажей
            self._adjust_posts_count(characters, posts_count_for_batch)
            
            # Создаем описание сгенерированных персонажей
            character_descriptions = []
            for character in characters:
                desc = f"{character.get('concept_short', '')} Роль: {character.get('role_in_world', '')}. Черты: {', '.join(character.get('personality_traits', []))}."
                character_descriptions.append(desc)
            
            # Объединяем с предыдущим описанием
            generated_characters_description = input.generated_characters_description
            if character_descriptions:
                if generated_characters_description:
                    generated_characters_description += "\n\n"
                generated_characters_description += "\n".join(character_descriptions)
            
            # Используем описание из ответа LLM, если есть
            llm_description = character_batch.get("generated_characters_description", "")
            if llm_description:
                generated_characters_description = llm_description
            
            # Проверяем, что LLM вернула хотя бы одного персонажа после корректировки
            if len(characters) == 0:
                workflow.logger.warning(f"No characters available after adjustment for world {input.world_id}")
                return WorkflowResult(
                    success=True,
                    data={
                        "characters_count": 0,
                        "total_characters_count": input.generated_count,
                        "remaining_characters": input.users_count,
                        "error": "No characters generated",
                        "recursion_depth": input.recursion_depth,
                        "max_allowed_depth": max_allowed_depth
                    }
                )
            
            workflow.logger.info(f"Final characters count: {len(characters)} for world {input.world_id}")
            
            # Создаем и запускаем child workflows для каждого персонажа
            from .generate_character_workflow import GenerateCharacterInput
            
            for i, character in enumerate(characters):
                character_input = GenerateCharacterInput(
                    world_id=input.world_id,
                    character_data=character,
                    posts_per_character=character.get("posts_count", 5)
                )
                character_task_ref = await self.save_task_data(character_input, input.world_id)
                
                character_workflow_id = self.get_workflow_id(
                    input.world_id,
                    "generate-character",
                    character_task_ref.task_id
                )
                
                await workflow.start_child_workflow(
                    "GenerateCharacterWorkflow",
                    character_task_ref,
                    id=character_workflow_id,
                    task_queue="ai-worker-main",
                    parent_close_policy=ParentClosePolicy.ABANDON
                )
                
            
            # Обновляем счетчики (используем реальное количество созданных персонажей)
            actual_characters_generated = len(characters)
            new_generated_count = input.generated_count + actual_characters_generated
            remaining_users = input.users_count - actual_characters_generated
            
            
            # Если нужно сгенерировать еще персонажей, запускаем следующий пакет
            if remaining_users > 0:
                # Вычисляем количество постов, которые уже распределены среди персонажей
                posts_distributed_in_batch = sum(character.get("posts_count", 0) for character in characters)
                new_remaining_posts_count = input.remaining_posts_count - posts_distributed_in_batch
                
                workflow.logger.info(f"Posts distribution: batch={posts_distributed_in_batch}, "
                                   f"remaining_before={input.remaining_posts_count}, "
                                   f"remaining_after={new_remaining_posts_count}")
                
                next_batch_workflow_id = self.get_workflow_id(
                    input.world_id,
                    "generate-character-batch",
                    f"batch-{input.count_run + 1}"
                )
                
                next_batch_input = GenerateCharacterBatchInput(
                    world_id=input.world_id,
                    users_count=remaining_users,
                    posts_count=input.posts_count,
                    remaining_posts_count=new_remaining_posts_count,
                    total_users_count=input.total_users_count,
                    generated_characters_description=generated_characters_description,
                    generated_count=new_generated_count,
                    count_run=input.count_run + 1,
                    recursion_depth=input.recursion_depth + 1
                )
                next_batch_task_ref = await self.save_task_data(next_batch_input, input.world_id)
                
                next_batch_workflow_id = self.get_workflow_id(
                    input.world_id,
                    "generate-character-batch",
                    next_batch_task_ref.task_id
                )
                
                await workflow.start_child_workflow(
                    "GenerateCharacterBatchWorkflow",
                    next_batch_task_ref,
                    id=next_batch_workflow_id,
                    task_queue="ai-worker-main",
                    parent_close_policy=ParentClosePolicy.ABANDON
                )
                
                workflow.logger.info(f"Started next character batch workflow for {remaining_users} remaining characters")
            else:
                # Все персонажи сгенерированы, обновляем статус этапа
                await workflow.execute_activity(
                    "update_stage",
                    args=[input.world_id, GenerationStage.CHARACTERS, GenerationStatus.COMPLETED],
                    task_queue="ai-worker-progress",
                    start_to_close_timeout=timedelta(seconds=30),
                    retry_policy=RetryPolicy(maximum_attempts=3)
                )
            
            workflow.logger.info(f"Successfully completed character batch generation for world {input.world_id}")
            
            return WorkflowResult(
                success=True,
                data={
                    "characters_count": actual_characters_generated,
                    "total_characters_count": new_generated_count,
                    "remaining_characters": remaining_users,
                    "recursion_depth": input.recursion_depth,
                    "generated_characters_description": generated_characters_description
                }
            )
            
        except Exception as e:
            error_msg = f"Error generating character batch: {str(e)}"
            workflow.logger.error(f"Workflow failed for world {input.world_id}: {error_msg}")
            raise e
            # # Обновляем статус этапа на "Ошибка"
            # try:
            #     await workflow.execute_activity(
            #         "update_stage",
            #         args=[input.world_id, GenerationStage.CHARACTERS, GenerationStatus.FAILED],
            #         task_queue="ai-worker-progress",
            #         start_to_close_timeout=timedelta(seconds=30),
            #         retry_policy=RetryPolicy(maximum_attempts=3)
            #     )
            # except Exception as update_error:
            #     workflow.logger.error(f"Failed to update failure status: {str(update_error)}")
            
            # return WorkflowResult(success=False, error=error_msg)
    
    async def _build_character_batch_prompt(
        self, 
        input: GenerateCharacterBatchInput, 
        world_params: Dict[str, Any], 
        current_batch_size: int,
        posts_count_for_batch:int
    ) -> str:
        """
        Строит промпт для генерации пакета персонажей
        
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
            args=[CHARACTER_BATCH_PROMPT],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Формируем информацию о персонажах
        previous_characters_info = ""
        future_users_count = input.users_count - current_batch_size
        
        if input.generated_count > 0:
            # Есть уже сгенерированные персонажи
            previous_characters_template = await workflow.execute_activity(
                "load_prompt",
                args=[PREVIOUS_CHARACTERS_PROMPT],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            previous_characters_info = previous_characters_template.format(
                count_run=input.count_run,
                count=input.generated_count,
                total_users_count=input.total_users_count,
                current_batch_size=current_batch_size,
                future_users_count=future_users_count,
                description=input.generated_characters_description
            )
        elif input.users_count > current_batch_size:
            # Первая генерация, но будут еще
            first_batch_template = await workflow.execute_activity(
                "load_prompt",
                args=[FIRST_BATCH_CHARACTERS_PROMPT],
                task_queue="ai-worker-main",
                start_to_close_timeout=timedelta(seconds=30),
                retry_policy=RetryPolicy(maximum_attempts=3)
            )
            previous_characters_info = first_batch_template.format(
                total_users_count=input.total_users_count,
                current_batch_size=current_batch_size,
                future_users_count=future_users_count
            )
        
        # Формируем описание мира через activity
        workflow.logger.info(f"world_params: {world_params}")
        world_description = await workflow.execute_activity(
            "format_world_description",
            args=[world_params],
            task_queue="ai-worker-main",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        workflow.logger.info(f"world_description: {world_description}")

        structure_description = model_to_template(CharacterBatchResponse)
        # print(prompt_template)
        # Формируем итоговый промпт

        
        prompt = prompt_template.format(
            world_description=world_description,
            users_count=current_batch_size,
            posts_count=posts_count_for_batch,
            previous_characters_info=previous_characters_info,
            structure_description=structure_description
        )
        
        return prompt
    
    def _adjust_characters_count(self, characters: List[Dict[str, Any]], target_count: int) -> List[Dict[str, Any]]:
        """
        Корректирует количество персонажей до целевого значения
        
        Args:
            characters: Список персонажей от LLM
            target_count: Целевое количество персонажей
            
        Returns:
            Скорректированный список персонажей
        """
        current_count = len(characters)
        
        if current_count == target_count:
            workflow.logger.info(f"Characters count already matches target: {current_count}")
            return characters
        
        workflow.logger.info(f"Adjusting characters count from {current_count} to {target_count}")
        
        if current_count > target_count:
            # Обрезаем лишних персонажей
            characters = characters[:target_count]
            workflow.logger.info(f"Trimmed characters to {len(characters)}")
        elif current_count < target_count:
            # Дублируем существующих персонажей с небольшими изменениями
            additional_needed = target_count - current_count
            workflow.logger.info(f"Need to generate {additional_needed} additional characters")
            
            # Дублируем персонажей циклично
            for i in range(additional_needed):
                if not characters:  # Если вообще нет персонажей, создаем базового
                    additional_character = {
                        "concept": f"Generated character {i+1}",
                        "concept_short": f"Character {i+1}",
                        "id": f"generated_{i+1}",
                        "role_in_world": "citizen",
                        "posts_count": 1,
                        "personality_traits": ["friendly"],
                        "interests": ["general topics"]
                    }
                else:
                    # Берем персонажа циклично и немного модифицируем
                    base_character = characters[i % len(characters)].copy()
                    base_character["id"] = f"{base_character.get('id', 'char')}_{i+1}"
                    base_character["concept_short"] = f"{base_character.get('concept_short', 'Character')} (variant {i+1})"
                    additional_character = base_character
                
                characters.append(additional_character)
            
            workflow.logger.info(f"Added {additional_needed} characters, total: {len(characters)}")
        
        return characters
    
    def _adjust_posts_count(self, characters: List[Dict[str, Any]], target_posts_count: int) -> None:
        """
        Корректирует количество постов для персонажей так, чтобы сумма соответствовала целевому значению
        
        Args:
            characters: Список персонажей
            target_posts_count: Целевое общее количество постов
        """
        if not characters:
            return
        
        # Обеспечиваем минимум 1 пост на персонажа
        for character in characters:
            if character.get("posts_count", 0) < 1:
                character["posts_count"] = 1
        
        current_sum = sum(character.get("posts_count", 1) for character in characters)
        
        if current_sum == target_posts_count:
            workflow.logger.info(f"Posts count already matches target: {current_sum}")
            return
        
        workflow.logger.info(f"Adjusting posts count from {current_sum} to {target_posts_count}")
        
        if current_sum < target_posts_count:
            # Нужно добавить посты
            diff = target_posts_count - current_sum
            
            # Распределяем дополнительные посты равномерно
            posts_per_character = diff // len(characters)
            remainder = diff % len(characters)
            
            for i, character in enumerate(characters):
                character["posts_count"] = character.get("posts_count", 1) + posts_per_character
                if i < remainder:
                    character["posts_count"] += 1
                    
        elif current_sum > target_posts_count:
            # Нужно убрать посты
            diff = current_sum - target_posts_count
            
            # Сортируем персонажей по количеству постов (убираем у тех, у кого больше)
            sorted_characters = sorted(characters, key=lambda c: c.get("posts_count", 1), reverse=True)
            
            remaining_to_remove = diff
            
            for character in sorted_characters:
                if remaining_to_remove <= 0:
                    break
                
                current_posts = character.get("posts_count", 1)
                # Не убираем ниже 1 поста
                posts_to_remove = min(remaining_to_remove, current_posts - 1)
                
                if posts_to_remove > 0:
                    character["posts_count"] = current_posts - posts_to_remove
                    remaining_to_remove -= posts_to_remove
        
        # Проверяем результат
        new_sum = sum(character.get("posts_count", 1) for character in characters)
        workflow.logger.info(f"Adjusted posts count: {new_sum} (target: {target_posts_count})")
        
        # Логируем распределение постов по персонажам
        posts_distribution = [character.get("posts_count", 1) for character in characters]
        workflow.logger.debug(f"Posts per character: {posts_distribution}")
        
        # Проверяем, что у всех персонажей минимум 1 пост
        for character in characters:
            if character.get("posts_count", 0) < 1:
                workflow.logger.warning(f"Character has less than 1 post after adjustment, setting to 1")
                character["posts_count"] = 1
        
    