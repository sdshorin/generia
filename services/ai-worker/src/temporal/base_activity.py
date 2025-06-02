"""
Base activity classes для Temporal
DEPRECATED: Use activity functions with dependency injection instead
"""
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Any, Dict, Optional
from temporalio import activity




@dataclass
class ActivityResult:
    """Результат выполнения activity"""
    success: bool
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


class BaseActivity(ABC):
    """
    Базовый класс для всех Temporal Activities
    DEPRECATED: Use activity functions with dependency injection instead
    """
    
    def __init__(
        self,
        db_manager=None,
        llm_client=None,
        image_generator=None,
        service_client=None
    ):
        """
        Инициализирует базовую activity
        
        Args:
            db_manager: Менеджер базы данных
            llm_client: Клиент для LLM
            image_generator: Генератор изображений
            service_client: Клиент для взаимодействия с микросервисами
        """
        logger.warning("BaseActivity is deprecated, use activity functions with dependency injection instead")
        self.db_manager = db_manager
        self.llm_client = llm_client
        self.image_generator = image_generator
        self.service_client = service_client
    
    @abstractmethod
    async def execute(self, *args, **kwargs) -> ActivityResult:
        """
        Выполняет activity
        
        Returns:
            Результат выполнения activity
        """
        pass
    
    def get_activity_info(self) -> str:
        """
        Возвращает информацию об activity для логирования
        
        Returns:
            Строка с информацией об activity
        """
        info = activity.info()
        return f"Activity: {info.activity_type}, Attempt: {info.attempt}, WorkflowID: {info.workflow_id}"