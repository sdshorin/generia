from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any, Dict, Optional, Type, TypeVar
from datetime import timedelta
from temporalio import workflow
from temporalio.common import RetryPolicy

from ..db.models import Task
from ..schemas.world_description import WorldDescription
from ..temporal.task_base import TaskInput, TaskRef, get_task_type_from_class


# Generic type для TaskInput классов
T = TypeVar('T', bound=TaskInput)


@dataclass
class WorkflowResult:
    """Результат выполнения workflow"""
    success: bool
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


class BaseWorkflow(ABC):
    """
    Базовый класс для всех Temporal Workflows
    """
    
    @workflow.run
    @abstractmethod
    async def run(self, *args, **kwargs) -> WorkflowResult:
        """
        Основной метод выполнения workflow
        """
        pass
    
    @staticmethod
    def get_workflow_id(world_id: str, workflow_type: str, task_id: Optional[str] = None) -> str:
        """
        Генерирует ID для workflow
        
        Args:
            world_id: ID мира
            workflow_type: Тип workflow
            task_id: ID задачи (опционально)
            
        Returns:
            Уникальный ID для workflow
        """
        if task_id:
            return f"{workflow_type}-{task_id}"
        return f"{workflow_type}-{world_id}"
    
    @staticmethod
    def get_task_queue(workflow_type: str) -> str:
        """
        Возвращает имя task queue для workflow
        
        Args:
            workflow_type: Тип workflow
            
        Returns:
            Имя task queue
        """
        return f"ai-worker-{workflow_type}"
    
    async def save_task_data(self, input_data: TaskInput, world_id: str) -> TaskRef:
        """
        Сохраняет данные задачи в MongoDB и возвращает TaskRef с task_id
        
        Args:
            input_data: Полные данные задачи
            world_id: ID мира
            
        Returns:
            TaskRef объект с task_id
        """
        # Получаем тип задачи из класса
        task_type = get_task_type_from_class(type(input_data))
        
        # Сериализуем данные в dict
        parameters = input_data.model_dump()
        
        # Создаем задачу в MongoDB
        task_id = await workflow.execute_activity(
            "create_task",
            args=[task_type, world_id, parameters],
            task_queue="ai-worker-progress",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        return TaskRef(task_id=task_id)
    
    async def load_task_data(self, task_ref: TaskRef, input_class: Type[T]) -> T:
        """
        Загружает данные задачи из MongoDB и конвертирует в указанный класс
        
        Args:
            task_ref: TaskRef объект с task_id
            input_class: Класс, в который нужно десериализовать данные
            
        Returns:
            Объект указанного класса с загруженными данными
        """
        # Загружаем задачу из MongoDB
        task_data = await workflow.execute_activity(
            "get_task",
            args=[task_ref.task_id],
            task_queue="ai-worker-progress",
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )
        
        # Извлекаем параметры и создаем объект нужного класса
        parameters = task_data.get('parameters', {})
        return input_class(**parameters)