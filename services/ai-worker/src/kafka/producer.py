import json
from typing import Dict, Any, Optional
from datetime import datetime

from aiokafka import AIOKafkaProducer
from aiokafka.errors import KafkaError

from ..config import KAFKA_BROKERS, KAFKA_TOPIC_TASKS, KAFKA_TOPIC_PROGRESS
from ..constants import KafkaEvents
from ..utils.logger import logger

class KafkaProducer:
    """
    Клиент для отправки задач и обновлений прогресса в Kafka
    """
    
    def __init__(
        self,
        bootstrap_servers: str = KAFKA_BROKERS,
        tasks_topic: str = KAFKA_TOPIC_TASKS,
        progress_topic: str = KAFKA_TOPIC_PROGRESS
    ):
        """
        Инициализирует Kafka Producer
        
        Args:
            bootstrap_servers: Список серверов Kafka
            tasks_topic: Имя темы для задач
            progress_topic: Имя темы для обновлений прогресса
        """
        self.bootstrap_servers = bootstrap_servers
        self.tasks_topic = tasks_topic
        self.progress_topic = progress_topic
        self.producer = None
    
    async def start(self):
        """
        Запускает продюсера Kafka
        """
        try:
            self.producer = AIOKafkaProducer(
                bootstrap_servers=self.bootstrap_servers,
                value_serializer=lambda m: json.dumps(m).encode("utf-8")
            )
            
            await self.producer.start()
            logger.info("Kafka producer started")
            
        except KafkaError as e:
            logger.error(f"Error starting Kafka producer: {str(e)}")
            raise
    
    async def stop(self):
        """
        Останавливает продюсера Kafka
        """
        if self.producer:
            await self.producer.stop()
            logger.info("Kafka producer stopped")
    
    async def send_task(
        self, task_id: str, task_type: str, world_id: str, parameters: Dict[str, Any], event_type: str = KafkaEvents.TASK_CREATED
    ) -> bool:
        """
        Отправляет задачу в Kafka
        
        Args:
            task_id: ID задачи
            task_type: Тип задачи
            world_id: ID мира
            parameters: Параметры задачи
            event_type: Тип события
            
        Returns:
            True, если задача успешно отправлена, иначе False
        """
        if not self.producer:
            logger.error("Kafka producer not started")
            return False
        
        message = {
            "event_type": event_type,
            "task_id": task_id,
            "task_type": task_type,
            "world_id": world_id,
            # "parameters": parameters
        }
        
        try:
            await self.producer.send_and_wait(self.tasks_topic, message)
            logger.info(f"Sent task {task_id} of type {task_type} to Kafka")
            return True
            
        except KafkaError as e:
            logger.error(f"Error sending task to Kafka: {str(e)}")
            return False
    
    async def send_task_update(
        self, task_id: str, status: str, result: Optional[Dict[str, Any]] = None, event_type: str = KafkaEvents.TASK_UPDATED
    ) -> bool:
        """
        Отправляет обновление статуса задачи в Kafka
        
        Args:
            task_id: ID задачи
            status: Новый статус
            result: Результат выполнения задачи
            event_type: Тип события
            
        Returns:
            True, если обновление успешно отправлено, иначе False
        """
        if not self.producer:
            logger.error("Kafka producer not started")
            return False
        
        message = {
            "event_type": event_type,
            "task_id": task_id,
            "status": status
        }
        
        if result:
            message["result"] = result
        
        try:
            await self.producer.send_and_wait(self.tasks_topic, message)
            logger.info(f"Sent task update for {task_id}, status: {status}")
            return True
            
        except KafkaError as e:
            logger.error(f"Error sending task update to Kafka: {str(e)}")
            return False
    
    async def send_progress_update(self, world_id: str, progress_data: Dict[str, Any]) -> bool:
        """
        Отправляет обновление прогресса генерации в Kafka
        
        Args:
            world_id: ID мира
            progress_data: Данные о прогрессе
            
        Returns:
            True, если обновление успешно отправлено, иначе False
        """
        if not self.producer:
            logger.error("Kafka producer not started")
            return False
        
        message = {
            "event_type": KafkaEvents.PROGRESS_UPDATED,
            "world_id": world_id,
            "progress": progress_data
        }
        
        try:
            await self.producer.send_and_wait(self.progress_topic, message)
            logger.debug(f"Sent progress update for world {world_id}")
            return True
            
        except KafkaError as e:
            logger.error(f"Error sending progress update to Kafka: {str(e)}")
            return False