import asyncio
import os
import signal
import sys
import uuid
from datetime import datetime

from .config import validate_config, MAX_CONCURRENT_TASKS, MAX_CONCURRENT_LLM_REQUESTS, MAX_CONCURRENT_IMAGE_REQUESTS
from .constants import TaskType
from .db import MongoDBManager
from .kafka import KafkaConsumer, KafkaProducer
from .api import LLMClient, ImageGenerator, ServiceClient
from .core import TaskManager, JobFactory
from .utils import logger
from .utils.progress import ProgressManager
from .jobs import (
    InitWorldCreationJob,
    GenerateWorldDescriptionJob,
    GenerateWorldImageJob,
    GenerateCharacterBatchJob,
    GenerateCharacterJob,
    GenerateCharacterAvatarJob,
    GeneratePostBatchJob,
    GeneratePostJob,
    GeneratePostImageJob
)

# Global variables for application components
db_manager = None
kafka_consumer = None
kafka_producer = None
llm_client = None
image_generator = None
service_client = None
progress_manager = None
task_manager = None
job_factory = None

async def initialize_components():
    """Initializes all application components"""
    global db_manager, kafka_consumer, kafka_producer, llm_client, image_generator
    global service_client, progress_manager, task_manager, job_factory
    
    # Initialize DB manager
    db_manager = MongoDBManager()
    await db_manager.initialize()
    logger.info("MongoDB manager initialized")
    
    # Initialize Kafka
    kafka_producer = KafkaProducer()
    await kafka_producer.start()
    logger.info("Kafka producer started")
    
    # Initialize external API clients
    service_client = ServiceClient(db_manager=db_manager)
    logger.info("Service client initialized")
    
    # Initialize images and LLM (order is important)
    image_generator = ImageGenerator(db_manager=db_manager, service_client=service_client)
    logger.info("Image generator initialized")
    
    llm_client = LLMClient(db_manager=db_manager)
    logger.info("LLM client initialized")
    
    # Initialize progress manager
    progress_manager = ProgressManager(db_manager, None) #  kafka_producer)
    logger.info("Progress manager initialized")
    
    # Initialize job factory
    job_factory = JobFactory(
        db_manager=db_manager,
        llm_client=llm_client,
        image_generator=image_generator,
        service_client=service_client,
        progress_manager=progress_manager,
        kafka_producer=kafka_producer
    )
    
    # Register job classes
    job_factory.register_jobs({
        TaskType.INIT_WORLD_CREATION: InitWorldCreationJob,
        TaskType.GENERATE_WORLD_DESCRIPTION: GenerateWorldDescriptionJob,
        TaskType.GENERATE_WORLD_IMAGE: GenerateWorldImageJob,
        TaskType.GENERATE_CHARACTER_BATCH: GenerateCharacterBatchJob,
        TaskType.GENERATE_CHARACTER: GenerateCharacterJob,
        TaskType.GENERATE_CHARACTER_AVATAR: GenerateCharacterAvatarJob,
        TaskType.GENERATE_POST_BATCH: GeneratePostBatchJob,
        TaskType.GENERATE_POST: GeneratePostJob,
        TaskType.GENERATE_POST_IMAGE: GeneratePostImageJob
    })
    logger.info("Job factory initialized with registered job types")
    
    # Initialize task manager
    task_manager = TaskManager(
        db_manager=db_manager,
        job_factory=job_factory,
        progress_manager=progress_manager,
        kafka_producer=kafka_producer,
        max_tasks=MAX_CONCURRENT_TASKS
    )
    await task_manager.start()
    logger.info("Task manager started")
    
    # Initialize Kafka consumer
    kafka_consumer = KafkaConsumer(processor=process_kafka_message)
    await kafka_consumer.start()
    logger.info("Kafka consumer started")

async def process_kafka_message(message):
    """
    Processes a Kafka message and immediately starts task processing
    
    Args:
        message: Kafka message
    """
    try:
        event_type = message.get("event_type")
        
        if event_type == "task_created":
            # Process new task
            task_id = message.get("task_id")
            task_type = message.get("task_type")
            world_id = message.get("world_id")
            parameters = message.get("parameters", {})
            
            logger.info(f"Received task_created event for task {task_id} of type {task_type}")
            
            # Immediately start task processing
            success = await task_manager.process_task_by_id(task_id)
            if success:
                logger.info(f"Started processing task {task_id} from Kafka message")
            else:
                logger.warning(f"Could not start processing task {task_id} from Kafka message")
        
        elif event_type == "task_updated":
            # Process task update
            task_id = message.get("task_id")
            status = message.get("status")
            
            logger.info(f"Received task_updated event for task {task_id}, status: {status}")
            # Updates are already happening in the respective jobs
        
        else:
            logger.warning(f"Unknown event type: {event_type}")
    
    except Exception as e:
        logger.error(f"Error processing Kafka message: {str(e)}")

async def shutdown():
    """Closes all application components"""
    logger.info("Shutting down...")
    
    if task_manager:
        await task_manager.stop()
        logger.info("Task manager stopped")
    
    if kafka_consumer:
        await kafka_consumer.stop()
        logger.info("Kafka consumer stopped")
    
    if kafka_producer:
        await kafka_producer.stop()
        logger.info("Kafka producer stopped")
    
    if service_client:
        await service_client.close()
        logger.info("Service client closed")
    
    if image_generator:
        await image_generator.close()
        logger.info("Image generator closed")
    
    logger.info("Shutdown complete")

async def main():
    """Main application function"""
    # Check configuration
    config_status = validate_config()
    if not config_status["valid"]:
        logger.error(f"Invalid configuration: {config_status['issues']}")
        sys.exit(1)
    
    # Output configuration information
    logger.info(f"Starting AI Worker with configuration: {config_status['config']}")
    
    # Set signal handlers for graceful shutdown
    loop = asyncio.get_event_loop()
    
    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, lambda: asyncio.create_task(shutdown()))
    
    try:
        logger.info("Starting AI Worker service")
        worker_id = f"worker-{uuid.uuid4().hex[:8]}"
        logger.info(f"Worker ID: {worker_id}")
        
        # Initialize components
        await initialize_components()
        
        # Main application loop
        while True:
            await asyncio.sleep(3600)  # Just keep the application running
            
    except Exception as e:
        logger.error(f"Unexpected error: {str(e)}")
    finally:
        await shutdown()

if __name__ == "__main__":
    # Set policy for creating event loops
    if sys.platform == "linux" and "uvloop" in sys.modules:
        import uvloop
        uvloop.install()
        logger.info("Using uvloop event loop")
    
    # Run main function
    asyncio.run(main())