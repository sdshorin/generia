import uuid
from typing import Dict, Any
from datetime import datetime

from ..core.base_job import BaseJob
from ..constants import TaskType, DEFAULT_VALUES, GenerationStage, GenerationStatus
from ..utils.logger import logger
from ..db.models import Task

class InitWorldCreationJob(BaseJob):
    """
    Задание для инициализации процесса генерации мира
    """
    
    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание по инициализации генерации мира
        
        Returns:
            Результат выполнения задания
        """
        # Получаем параметры из задачи
        world_id = self.task.world_id
        user_prompt = self.task.parameters.get("user_prompt", "")
        users_count = self.task.parameters.get("users_count", DEFAULT_VALUES["users_count"])
        posts_count = self.task.parameters.get("posts_count", DEFAULT_VALUES["posts_count"])
        
        # Проверяем валидность промпта
        if not user_prompt:
            raise ValueError("User prompt is required")
        
        # Устанавливаем лимиты на API-вызовы
        api_call_limits_llm = self.task.parameters.get(
            "api_call_limits_llm", DEFAULT_VALUES["api_call_limits_LLM"]
        )
        api_call_limits_images = self.task.parameters.get(
            "api_call_limits_images", DEFAULT_VALUES["api_call_limits_images"]
        )
        
        # Инициализируем запись о статусе генерации мира
        if self.progress_manager:
            await self.progress_manager.initialize_world_generation(
                world_id=world_id,
                users_count=users_count,
                posts_count=posts_count,
                user_prompt=user_prompt,
                api_call_limits_llm=api_call_limits_llm,
                api_call_limits_images=api_call_limits_images
            )
            
            # Обновляем статус этапа инициализации
            await self.progress_manager.update_stage(
                world_id=world_id,
                stage=GenerationStage.INITIALIZING,
                status=GenerationStatus.COMPLETED
            )
        
        # Создаем задачу для генерации описания мира
        next_task_id = str(uuid.uuid4())
        now = datetime.utcnow()
        
        next_task = Task(
            _id=next_task_id,
            type=TaskType.GENERATE_WORLD_DESCRIPTION,
            world_id=world_id,
            status="pending",
            worker_id=None,
            parameters={
                "user_prompt": user_prompt,
                "users_count": users_count,
                "posts_count": posts_count
            },
            created_at=now,
            updated_at=now,
            attempt_count=0
        )
        
        tasks_to_create = [{"task": next_task}]
        created_task_ids = await self.create_next_tasks(tasks_to_create)
        
        # Обновляем статус следующего этапа
        if self.progress_manager:
            await self.progress_manager.update_stage(
                world_id=world_id,
                stage=GenerationStage.WORLD_DESCRIPTION,
                status=GenerationStatus.IN_PROGRESS
            )
        
        return {
            "message": "World generation initialized successfully",
            "world_id": world_id,
            "next_tasks": created_task_ids
        }
    
    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания
        
        Args:
            result: Результат выполнения задания
        """
        logger.info(
            f"World generation initialized for world {self.task.world_id}. "
            f"Created next tasks: {result.get('next_tasks', [])}"
        )
    
    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания
        
        Args:
            error: Возникшая ошибка
        """
        logger.error(f"Failed to initialize world generation: {str(error)}")
        
        # Обновляем статус генерации мира
        if self.progress_manager:
            await self.progress_manager.update_stage(
                world_id=self.task.world_id,
                stage=GenerationStage.INITIALIZING,
                status=GenerationStatus.FAILED
            )
            
            # Устанавливаем общий статус генерации как неудачный
            await self.progress_manager.update_progress(
                world_id=self.task.world_id,
                updates={"status": GenerationStatus.FAILED}
            )