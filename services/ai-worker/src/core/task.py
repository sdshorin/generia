import asyncio
import uuid
import traceback
from typing import Dict, Any, List, Optional, Type
from datetime import datetime

from ..utils.logger import logger
from ..constants import TaskStatus, MAX_ATTEMPTS
from ..db.models import Task
from .base_job import BaseJob

class TaskManager:
    """
    Менеджер для обработки задач
    
    TaskManager отвечает за обработку задач, полученных через Kafka.
    Вместо периодического опроса базы данных, задачи запускаются
    непосредственно при получении сообщения из Kafka, что обеспечивает 
    низкую задержку и более эффективное использование ресурсов.
    """
    
    def __init__(
        self,
        db_manager,
        job_factory,
        progress_manager=None,
        kafka_producer=None,
        max_tasks=100
    ):
        """
        Инициализирует менеджер задач
        
        Args:
            db_manager: Менеджер базы данных
            job_factory: Фабрика для создания Job-объектов
            progress_manager: Менеджер прогресса
            kafka_producer: Kafka продюсер
            max_tasks: Максимальное количество задач, обрабатываемых одновременно
        """
        self.db_manager = db_manager
        self.job_factory = job_factory
        self.progress_manager = progress_manager
        self.kafka_producer = kafka_producer
        self.worker_id = f"worker-{uuid.uuid4().hex[:8]}"
        self.task_semaphore = asyncio.Semaphore(max_tasks)
        self.running = False
        self._stop_event = asyncio.Event()
        
        # Словарь для отслеживания обрабатываемых задач
        self.active_tasks = {}
    
    async def start(self):
        """
        Запускает менеджер задач
        """
        if self.running:
            return
        
        self.running = True
        self._stop_event.clear()
        logger.info(f"Task manager started with worker ID {self.worker_id}")
    
    async def stop(self):
        """
        Останавливает менеджер задач
        """
        if not self.running:
            return
        
        logger.info("Stopping task manager...")
        self._stop_event.set()
        self.running = False
        
        # Ждем завершения всех задач
        if self.active_tasks:
            logger.info(f"Waiting for {len(self.active_tasks)} active tasks to complete...")
            timeout = 60  # Максимальное время ожидания в секундах
            try:
                await asyncio.wait_for(self._wait_for_active_tasks(), timeout=timeout)
                logger.info("All active tasks completed")
            except asyncio.TimeoutError:
                logger.warning(f"Not all tasks completed within {timeout} seconds")
    
    async def _wait_for_active_tasks(self):
        """
        Ожидает завершения всех активных задач
        """
        while self.active_tasks:
            await asyncio.sleep(1)
    
    
    async def process_task_by_id(self, task_id: str) -> bool:
        """
        Загружает и обрабатывает задачу по её ID из Kafka
        
        Этот метод вызывается при получении сообщения из Kafka
        с event_type="task_created" для немедленного запуска обработки
        
        Args:
            task_id: ID задачи
            
        Returns:
            True если задача была обработана успешно, иначе False
        """
        # Проверяем, не превышен ли лимит параллельных задач
        if len(self.active_tasks) >= self.task_semaphore._value:
            logger.warning(f"Maximum concurrent tasks limit reached ({self.task_semaphore._value}), will process task {task_id} later")
            return False
        
        # Загружаем задачу из БД
        task = await self.db_manager.get_task(task_id)
        
        if not task:
            logger.warning(f"Task {task_id} not found in database")
            return False
            
        if task.status != TaskStatus.PENDING:
            logger.warning(f"Task {task_id} is not in pending status (current status: {task.status})")
            return False
            
        # Запускаем обработку задачи в отдельной задаче
        asyncio.create_task(self.process_task(task))
        return True
    
    async def process_task(self, task: Task):
        """
        Обрабатывает отдельную задачу
        
        Args:
            task: Объект задачи
        """
        # Получаем семафор для ограничения количества одновременных задач
        async with self.task_semaphore:
            try:
                # Пытаемся захватить задачу для обработки
                claimed = await self.db_manager.claim_task(task.id, self.worker_id)
                
                if not claimed:
                    logger.warning(f"Failed to claim task {task.id}")
                    return
                
                # Добавляем задачу в список активных
                self.active_tasks[task.id] = task
                
                # Проверяем количество попыток
                max_attempts = MAX_ATTEMPTS.get(task.type, 2)
                if task.attempt_count > max_attempts:
                    logger.warning(
                        f"Task {task.id} of type {task.type} has exceeded max attempts "
                        f"({task.attempt_count} > {max_attempts})"
                    )
                    
                    # Отмечаем задачу как неудачную
                    await self.db_manager.update_task_status(
                        task.id,
                        TaskStatus.FAILED,
                        error=f"Exceeded maximum number of attempts ({max_attempts})"
                    )
                    
                    # Обновляем счетчик неудачных задач
                    if self.progress_manager:
                        await self.progress_manager.increment_task_counter(
                            world_id=task.world_id,
                            field="tasks_failed"
                        )
                    
                    # Отправляем уведомление о неудаче
                    if self.kafka_producer:
                        await self.kafka_producer.send_task_update(
                            task_id=task.id,
                            status=TaskStatus.FAILED,
                            event_type="task_failed"
                        )
                    
                    return
                
                logger.info(
                    f"Processing task {task.id} of type {task.type} "
                    f"(attempt {task.attempt_count} of {max_attempts})"
                )
                
                # Создаем Job для выполнения задачи
                job = self.job_factory.create_job(task)
                
                if not job:
                    logger.error(f"Unknown task type: {task.type}")
                    await self.db_manager.update_task_status(
                        task.id,
                        TaskStatus.FAILED,
                        error=f"Unknown task type: {task.type}"
                    )
                    return
                
                # Выполняем задачу
                try:
                    result = await job.execute()
                    
                    # Обновляем статус задачи
                    await self.db_manager.update_task_status(
                        task.id,
                        TaskStatus.COMPLETED,
                        result=result
                    )
                    
                    # Обновляем счетчик выполненных задач
                    if self.progress_manager:
                        await self.progress_manager.increment_task_counter(
                            world_id=task.world_id,
                            field="tasks_completed"
                        )
                    
                    # Отправляем уведомление о завершении
                    if self.kafka_producer:
                        await self.kafka_producer.send_task_update(
                            task_id=task.id,
                            status=TaskStatus.COMPLETED,
                            result=result,
                            event_type="task_completed"
                        )
                    
                    # Вызываем обработчик успешного выполнения
                    await job.on_success(result)
                    
                    logger.info(f"Task {task.id} of type {task.type} completed successfully")
                    
                except Exception as e:
                    logger.error(
                        f"Error executing task {task.id}: {str(e)}\n"
                        f"Traceback:\n{traceback.format_exc()}"
                    )
                    
                    # Обновляем статус задачи
                    if task.attempt_count >= max_attempts:
                        status = TaskStatus.FAILED
                        event_type = "task_failed"
                        
                        # Обновляем счетчик неудачных задач
                        if self.progress_manager:
                            await self.progress_manager.increment_task_counter(
                                world_id=task.world_id,
                                field="tasks_failed"
                            )
                    else:
                        status = TaskStatus.PENDING
                        event_type = "task_retrying"
                    
                    await self.db_manager.update_task_status(
                        task.id,
                        status,
                        error=str(e)
                    )
                    
                    # Сбрасываем worker_id, чтобы задача могла быть взята другим воркером
                    await self.db_manager.update_task(
                        task.id,
                        {
                            "worker_id": None,
                            "updated_at": datetime.utcnow()
                        }
                    )
                    
                    # Отправляем уведомление о неудаче
                    if self.kafka_producer:
                        await self.kafka_producer.send_task_update(
                            task_id=task.id,
                            status=status,
                            event_type=event_type
                        )
                    
                    # Вызываем обработчик ошибки
                    try:
                        await job.on_failure(e)
                    except Exception as e2:
                        logger.error(f"Error in on_failure handler for task {task.id}: {str(e2)}")
            
            except Exception as e:
                logger.error(f"Unexpected error in process_task for task {task.id}: {str(e)}")
            
            finally:
                # Удаляем задачу из списка активных
                if task.id in self.active_tasks:
                    del self.active_tasks[task.id]
    
    async def create_task(
        self,
        task_type: str,
        world_id: str,
        parameters: Dict[str, Any]
    ) -> str:
        """
        Создает новую задачу
        
        Args:
            task_type: Тип задачи
            world_id: ID мира
            parameters: Параметры задачи
            
        Returns:
            ID созданной задачи
        """
        task_id = str(uuid.uuid4())
        now = datetime.utcnow()
        
        task = Task(
            _id=task_id,
            type=task_type,
            world_id=world_id,
            status=TaskStatus.PENDING,
            worker_id=None,
            parameters=parameters,
            created_at=now,
            updated_at=now,
            attempt_count=0
        )
        
        # Создаем задачу в БД
        await self.db_manager.create_task(task)
        
        # Обновляем счетчик задач
        if self.progress_manager:
            await self.progress_manager.increment_task_counter(
                world_id=world_id,
                field="tasks_total"
            )
        
        # Отправляем уведомление о создании задачи
        if self.kafka_producer:
            await self.kafka_producer.send_task(
                task_id=task_id,
                task_type=task_type,
                world_id=world_id,
                parameters=parameters
            )
        
        logger.info(f"Created task {task_id} of type {task_type} for world {world_id}")
        return task_id