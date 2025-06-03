import asyncio
import signal
import sys
from datetime import timedelta
from typing import List
import logging


from temporalio.client import Client
from temporalio.worker import Worker, UnsandboxedWorkflowRunner
from temporalio.common import RetryPolicy

from .config import (
    validate_config, 
    MAX_CONCURRENT_TASKS, 
    MAX_CONCURRENT_LLM_REQUESTS, 
    MAX_CONCURRENT_IMAGE_REQUESTS,
    MAX_CONCURRENT_GRPC_CALLS,
    MAX_CONCURRENT_DB_OPERATIONS,
    MAX_WORKFLOW_TASKS_PER_WORKER,
    MAX_ACTIVITIES_PER_WORKER,
    TEMPORAL_HOST,
    TEMPORAL_NAMESPACE
)
from .temporal.shared_resources import SharedResourcesManager
from .temporal.activities import create_activity_functions
from .workflows import (
    InitWorldCreationWorkflow,
    GenerateWorldDescriptionWorkflow,
    GenerateWorldImageWorkflow,
    GenerateCharacterBatchWorkflow,
    GenerateCharacterWorkflow,
    GenerateCharacterAvatarWorkflow,
    GeneratePostBatchWorkflow,
    GeneratePostWorkflow,
    GeneratePostImageWorkflow
)

# Application state
resource_manager = None
temporal_client = None
workers: List[Worker] = []

async def initialize_components():
    """Initializes all application components with proper resource pooling"""
    global resource_manager, temporal_client


    # LOG_FMT = (
    #     "%(asctime)s | %(levelname)-7s | %(name)s | "
    #     "wf=%(temporal_workflow_id)s run=%(temporal_run_id)s "
    #     "act=%(temporal_activity_id)s | %(message)s"
    # )

    # Configure logging for both main application and Temporal workflows
    logging.basicConfig(
        level=logging.DEBUG,  # Enable DEBUG level to see all workflow.logger messages - or use INFO
        stream=sys.stdout,
        format='%(asctime)s | %(levelname)-7s | %(name)s | %(message)s'
    )
    
    # Ensure Temporal workflow logger is properly configured
    temporal_logger = logging.getLogger('temporalio.workflow')
    temporal_logger.setLevel(logging.DEBUG)

    # Initialize shared resources manager with proper connection pools
    resource_manager = SharedResourcesManager()
    await resource_manager.initialize()

    # Initialize Temporal client
    temporal_client = await Client.connect(TEMPORAL_HOST, namespace=TEMPORAL_NAMESPACE)
    # logger.info(f"Temporal client connected to {TEMPORAL_HOST} namespace {TEMPORAL_NAMESPACE}")

async def create_workers():
    """Creates and configures Temporal workers with injected resources"""
    global workers, resource_manager, temporal_client
    
    # Create activity functions with injected resources
    activities = create_activity_functions(resource_manager)
    
    # Default retry policy
    default_retry_policy = RetryPolicy(
        initial_interval=timedelta(seconds=1),
        backoff_coefficient=2.0,
        maximum_interval=timedelta(minutes=10),
        maximum_attempts=5,
    )
    
    # Main worker - workflows and general activities
    main_worker = Worker(
        temporal_client,
        task_queue="ai-worker-main",
        activities=[
            activities['load_prompt'],
            activities['generate_structured_content'],
            activities['enhance_prompt'],
            activities['save_world_parameters'],
            activities['get_world_parameters'],
            activities['generate_image'],
            activities['upload_image_to_media_service'],
            activities['create_character'],
            activities['create_post'],
            activities['update_character_avatar'],
            activities['create_task'],
            activities['get_task'],
            activities['format_world_description'],
        ],
        workflows=[
            InitWorldCreationWorkflow,
            GenerateWorldDescriptionWorkflow,
            GenerateWorldImageWorkflow,
            GenerateCharacterBatchWorkflow,
            GenerateCharacterWorkflow,
            GenerateCharacterAvatarWorkflow,
            GeneratePostBatchWorkflow,
            GeneratePostWorkflow,
            GeneratePostImageWorkflow
        ],
        max_concurrent_activities=MAX_ACTIVITIES_PER_WORKER,
        max_concurrent_workflow_tasks=MAX_WORKFLOW_TASKS_PER_WORKER,
        workflow_runner=UnsandboxedWorkflowRunner(), # AGENT_LIMITS
    )
    
    # Specialized LLM worker
    llm_worker = Worker(
        temporal_client,
        task_queue="ai-worker-llm",
        activities=[
            activities['generate_structured_content'],
            activities['enhance_prompt'],
        ],
        max_concurrent_activities=MAX_CONCURRENT_LLM_REQUESTS,
        workflow_runner=UnsandboxedWorkflowRunner(), # AGENT_LIMITS
    )
    
    # Specialized image worker
    image_worker = Worker(
        temporal_client,
        task_queue="ai-worker-images",
        activities=[
            activities['generate_image'],
            activities['upload_image_to_media_service'],
        ],
        max_concurrent_activities=MAX_CONCURRENT_IMAGE_REQUESTS,
        workflow_runner=UnsandboxedWorkflowRunner(), # AGENT_LIMITS
    )
    
    # Specialized progress worker
    progress_worker = Worker(
        temporal_client,
        task_queue="ai-worker-progress",
        activities=[
            activities['initialize_world_generation'],
            activities['update_stage'],
            activities['increment_counter'],
            activities['increment_cost'],
            activities['update_progress'],
            activities['create_task'],
            activities['get_task'],
        ],
        max_concurrent_activities=MAX_CONCURRENT_DB_OPERATIONS,
        workflow_runner=UnsandboxedWorkflowRunner(), # AGENT_LIMITS
    )
    
    # Service worker for gRPC calls
    service_worker = Worker(
        temporal_client,
        task_queue="ai-worker-services",
        activities=[
            activities['create_character'],
            activities['create_post'],
            activities['update_character_avatar'],
            activities['update_world_image'],
        ],
        max_concurrent_activities=MAX_CONCURRENT_GRPC_CALLS,
        workflow_runner=UnsandboxedWorkflowRunner(), # AGENT_LIMITS
    )
    
    workers = [
        main_worker,
        llm_worker, 
        image_worker, 
        progress_worker, 
        service_worker
    ]
    # logger.info(f"Created {len(workers)} specialized Temporal workers with resource injection")

async def shutdown():
    """Closes all application components gracefully"""
    # logger.info("Shutting down...")

    # Stop all workers
    for i, worker in enumerate(workers):
        try:
            worker.shutdown()
            # logger.info(f"Worker {i+1} stopped")
        except Exception as e:
            pass
            # logger.error(f"Error stopping worker {i+1}: {str(e)}")

    # Close shared resources manager and all connection pools
    if resource_manager:
        await resource_manager.close()
        # logger.info("Shared resources manager closed")

    # Close temporal client
    # if temporal_client:
    #     await temporal_client.close()
    #     # logger.info("Temporal client closed")

    # logger.info("Shutdown complete")

async def main():
    """Main application function"""
    # Check configuration
    config_status = validate_config()
    if not config_status["valid"]:
        # logger.error(f"Invalid configuration: {config_status['issues']}")
        sys.exit(1)

    # Output configuration information
    # logger.info(f"Starting AI Worker with Temporal configuration: {config_status['config']}")

    # Set signal handlers for graceful shutdown
    loop = asyncio.get_event_loop()
    shutdown_event = asyncio.Event()

    def signal_handler():
        # logger.info("Received shutdown signal")
        shutdown_event.set()

    for sig in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig, signal_handler)

    try:
        # logger.info("Starting AI Worker service with Temporal")
        
        # Initialize components
        await initialize_components()
        
        # Create workers
        await create_workers()
        
        # Start all workers using context managers for proper cleanup
        try:
            worker_tasks = []
            for i, worker in enumerate(workers):
                task = asyncio.create_task(worker.run(), name=f"worker-{i+1}")
                worker_tasks.append(task)
                # logger.info(f"Started Temporal worker {i+1} on task queue {worker.task_queue}")
            
            # logger.info(f"All {len(workers)} Temporal workers started and running")
            
            # Wait for shutdown signal
            await shutdown_event.wait()
            
        finally:
            # Cancel all worker tasks
            for task in worker_tasks:
                if not task.done():
                    task.cancel()
            
            # Wait for tasks to finish cancellation
            await asyncio.gather(*worker_tasks, return_exceptions=True)
        
    except Exception as e:
        # logger.error(f"Unexpected error: {str(e)}")
        raise
    finally:
        await shutdown()

if __name__ == "__main__":
    # Set policy for creating event loops
    if sys.platform == "linux":
        try:
            import uvloop
            uvloop.install()
            # logger.info("Using uvloop event loop")
        except ImportError:
            pass
            # logger.info("uvloop not available, using default event loop")

    # Run main function
    asyncio.run(main())