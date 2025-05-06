from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional

from ..db.models import Task
from ..schemas.world_description import WorldDescription
from ..utils.logger import logger

class BaseJob(ABC):
    """
    Базовый класс для всех заданий
    """

    def __init__(
        self,
        task: Task,
        db_manager,
        llm_client=None,
        image_generator=None,
        service_client=None,
        progress_manager=None,
        kafka_producer=None
    ):
        """
        Инициализирует базовое задание

        Args:
            task: Объект задачи
            db_manager: Менеджер базы данных
            llm_client: Клиент для LLM
            image_generator: Генератор изображений
            service_client: Клиент для взаимодействия с микросервисами
            progress_manager: Менеджер прогресса
            kafka_producer: Kafka продюсер
        """
        self.task = task
        self.db_manager = db_manager
        self.llm_client = llm_client
        self.image_generator = image_generator
        self.service_client = service_client
        self.progress_manager = progress_manager
        self.kafka_producer = kafka_producer

    @abstractmethod
    async def execute(self) -> Dict[str, Any]:
        """
        Выполняет задание

        Returns:
            Результат выполнения задания
        """
        pass

    @abstractmethod
    async def on_success(self, result: Dict[str, Any]) -> None:
        """
        Выполняется при успешном завершении задания

        Args:
            result: Результат выполнения задания
        """
        pass

    @abstractmethod
    async def on_failure(self, error: Exception) -> None:
        """
        Выполняется при ошибке во время выполнения задания

        Args:
            error: Возникшая ошибка
        """
        pass

    async def get_world_parameters(self, world_id: str) -> Optional[WorldDescription]:
        """
        Получает параметры мира

        Args:
            world_id: ID мира

        Returns:
            Объект параметров мира или None, если параметры не найдены
        """
        return await self.db_manager.get_world_parameters(world_id)

    async def create_next_tasks(self, tasks: List[Dict[str, Any]]) -> List[str]:
        """
        Создает следующие задачи и отправляет их в Kafka

        Args:
            tasks: Список словарей с описанием задач

        Returns:
            Список ID созданных задач
        """
        if not tasks:
            return []

        created_task_ids = []

        for task_data in tasks:
            try:
                # Создание задачи в БД
                task_id = await self.db_manager.create_task(task_data["task"])
                created_task_ids.append(task_id)

                # Отправка в Kafka
                if self.kafka_producer:
                    await self.kafka_producer.send_task(
                        task_id=task_id,
                        task_type=task_data["task"].type,
                        world_id=task_data["task"].world_id,
                        parameters=task_data["task"].parameters
                    )

                # Обновление счетчика задач
                if self.progress_manager:
                    await self.progress_manager.increment_task_counter(
                        world_id=task_data["task"].world_id,
                        field="tasks_total"
                    )

            except Exception as e:
                logger.error(f"Ошибка при создании следующей задачи: {str(e)}")

        return created_task_ids