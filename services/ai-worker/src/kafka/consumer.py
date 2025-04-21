import asyncio
import json
from typing import Callable, Dict, Any, Optional
import uuid

from aiokafka import AIOKafkaConsumer
from aiokafka.errors import KafkaError

from ..config import KAFKA_BROKERS, KAFKA_TOPIC_TASKS, KAFKA_GROUP_ID
from ..utils.logger import logger

class KafkaConsumer:
    """
    Клиент для получения задач из Kafka
    """
    
    def __init__(
        self,
        bootstrap_servers: str = KAFKA_BROKERS,
        topic: str = KAFKA_TOPIC_TASKS,
        group_id: str = KAFKA_GROUP_ID,
        processor: Optional[Callable[[Dict[str, Any]], None]] = None
    ):
        """
        Инициализирует Kafka Consumer
        
        Args:
            bootstrap_servers: Список серверов Kafka
            topic: Имя темы для подписки
            group_id: ID группы для потребителя
            processor: Колбэк-функция для обработки сообщений
        """
        self.bootstrap_servers = bootstrap_servers
        self.topic = topic
        self.group_id = f"{group_id}-{uuid.uuid4().hex[:8]}"  # Уникальный group_id для каждого экземпляра
        self.processor = processor
        self.consumer = None
        self.running = False
        self._stop_event = asyncio.Event()
    
    async def start(self, processor: Optional[Callable[[Dict[str, Any]], None]] = None):
        """
        Запускает потребителя Kafka
        
        Args:
            processor: Колбэк-функция для обработки сообщений
        """
        if processor:
            self.processor = processor
        
        if not self.processor:
            raise ValueError("Processor function must be provided")
        
        try:
            self.consumer = AIOKafkaConsumer(
                self.topic,
                bootstrap_servers=self.bootstrap_servers,
                group_id=self.group_id,
                auto_offset_reset="earliest",
                enable_auto_commit=True,
                value_deserializer=lambda m: json.loads(m.decode("utf-8"))
            )
            
            await self.consumer.start()
            self.running = True
            self._stop_event.clear()
            
            logger.info(
                f"Kafka consumer started for topic {self.topic} "
                f"with group_id {self.group_id}"
            )
            
            # Запускаем обработчик сообщений в отдельной задаче
            asyncio.create_task(self._consume())
            
        except KafkaError as e:
            logger.error(f"Error starting Kafka consumer: {str(e)}")
            raise
    
    async def stop(self):
        """
        Останавливает потребителя Kafka
        """
        if self.running:
            self._stop_event.set()
            if self.consumer:
                await self.consumer.stop()
            self.running = False
            logger.info("Kafka consumer stopped")
    
    async def _consume(self):
        """
        Обрабатывает сообщения из Kafka
        """
        try:
            while self.running and not self._stop_event.is_set():
                try:
                    async for message in self.consumer:
                        if self._stop_event.is_set():
                            break
                        
                        try:
                            logger.debug(
                                f"Received message from topic {message.topic} "
                                f"partition {message.partition} offset {message.offset}"
                            )
                            
                            # Вызываем функцию-обработчик
                            await self.processor(message.value)
                            
                        except Exception as e:
                            logger.error(
                                f"Error processing message: {str(e)}. "
                                f"Message: {message.value}"
                            )
                
                except KafkaError as e:
                    logger.error(f"Kafka error during consumption: {str(e)}")
                    # Пауза перед повторной попыткой
                    await asyncio.sleep(5)
        
        except Exception as e:
            logger.error(f"Unexpected error in Kafka consumer: {str(e)}")
        
        finally:
            # Убеждаемся, что потребитель остановлен
            if self.consumer and self.running:
                await self.consumer.stop()
                self.running = False