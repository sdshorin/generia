from typing import Dict, Type, Optional

from ..constants import TaskType
from ..db.models import Task
from .base_job import BaseJob
from ..utils.logger import logger

class JobFactory:
    """
    Фабрика для создания Job-объектов на основе типа задачи
    """
    
    def __init__(
        self,
        db_manager,
        llm_client=None,
        image_generator=None,
        service_client=None,
        progress_manager=None,
        kafka_producer=None
    ):
        """
        Инициализирует фабрику задач
        
        Args:
            db_manager: Менеджер базы данных
            llm_client: Клиент для LLM
            image_generator: Генератор изображений
            service_client: Клиент для взаимодействия с микросервисами
            progress_manager: Менеджер прогресса
            kafka_producer: Kafka продюсер
        """
        self.db_manager = db_manager
        self.llm_client = llm_client
        self.image_generator = image_generator
        self.service_client = service_client
        self.progress_manager = progress_manager
        self.kafka_producer = kafka_producer
        
        # Словарь для хранения классов job'ов по типу задачи
        # Будет заполнен при регистрации job'ов
        self.job_classes: Dict[str, Type[BaseJob]] = {}
    
    def register_job(self, task_type: str, job_class: Type[BaseJob]):
        """
        Регистрирует класс Job для определенного типа задачи
        
        Args:
            task_type: Тип задачи
            job_class: Класс для обработки задачи этого типа
        """
        self.job_classes[task_type] = job_class
        logger.debug(f"Registered job class for task type: {task_type}")
    
    def register_jobs(self, jobs_dict: Dict[str, Type[BaseJob]]):
        """
        Регистрирует несколько классов Job для разных типов задач
        
        Args:
            jobs_dict: Словарь с типами задач и соответствующими классами
        """
        for task_type, job_class in jobs_dict.items():
            self.register_job(task_type, job_class)
    
    def create_job(self, task: Task) -> Optional[BaseJob]:
        """
        Создает экземпляр Job для обработки задачи
        
        Args:
            task: Объект задачи
            
        Returns:
            Экземпляр BaseJob или None, если тип задачи неизвестен
        """
        job_class = self.job_classes.get(task.type)
        
        if not job_class:
            logger.error(f"No job class registered for task type: {task.type}")
            return None
        
        return job_class(
            task=task,
            db_manager=self.db_manager,
            llm_client=self.llm_client,
            image_generator=self.image_generator,
            service_client=self.service_client,
            progress_manager=self.progress_manager,
            kafka_producer=self.kafka_producer
        )