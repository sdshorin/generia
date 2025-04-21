import asyncio
from typing import Dict, Any, List, Optional
from datetime import datetime

from motor.motor_asyncio import AsyncIOMotorClient
from pymongo.errors import DuplicateKeyError

from ..config import MONGODB_URI, MONGODB_DATABASE
from ..constants import Collections, GenerationStatus, GenerationStage
from ..utils.logger import logger
from .models import Task, WorldGenerationStatus, WorldParameters, ApiRequestHistory

class MongoDBManager:
    """
    Manager for working with MongoDB
    """
    
    def __init__(self, uri: str = MONGODB_URI, db_name: str = MONGODB_DATABASE):
        """
        Initializes connection to MongoDB
        
        Args:
            uri: URI for connecting to MongoDB
            db_name: Database name
        """
        self.client = AsyncIOMotorClient(uri)
        self.db = self.client[db_name]
        
        # Collections
        self.tasks_collection = self.db[Collections.TASKS]
        self.world_generation_status_collection = self.db[Collections.WORLD_GENERATION_STATUS]
        self.world_parameters_collection = self.db[Collections.WORLD_PARAMETERS]
        self.api_requests_history_collection = self.db[Collections.API_REQUESTS_HISTORY]
        
        # Lock for idempotent operations
        self._locks = {}
    
    async def initialize(self):
        """
        Initializes indexes and other database settings
        """
        # Index for tasks
        await self.tasks_collection.create_index("world_id")
        await self.tasks_collection.create_index("type")
        await self.tasks_collection.create_index("status")
        await self.tasks_collection.create_index([("world_id", 1), ("type", 1)])
        
        # Index for API request history
        await self.api_requests_history_collection.create_index("world_id")
        await self.api_requests_history_collection.create_index("task_id")
        await self.api_requests_history_collection.create_index("api_type")
        
        logger.info("MongoDB indexes initialized")
    
    def get_lock(self, key: str) -> asyncio.Lock:
        """
        Gets a lock for the specified key
        
        Args:
            key: Lock key
            
        Returns:
            Lock object
        """
        if key not in self._locks:
            self._locks[key] = asyncio.Lock()
        return self._locks[key]
    
    # Methods for working with tasks
    
    async def create_task(self, task: Task) -> str:
        """
        Creates a new task
        
        Args:
            task: Task object
            
        Returns:
            ID of the created task
            
        Raises:
            DuplicateKeyError: If a task with this ID already exists
        """
        task_dict = task.dict(by_alias=True)
        try:
            await self.tasks_collection.insert_one(task_dict)
            logger.info(f"Created task {task.id} of type {task.type}")
            return task.id
        except DuplicateKeyError:
            logger.warning(f"Task with ID {task.id} already exists")
            raise
    
    async def get_task(self, task_id: str) -> Optional[Task]:
        """
        Gets a task by ID
        
        Args:
            task_id: Task ID
            
        Returns:
            Task object or None if task not found
        """
        logger.info(f"Searching for task in database: task_id={task_id}, collection={self.tasks_collection.name}, database={self.tasks_collection.database.name}")
        
        # Log the query that will be executed
        query = {"_id": task_id}
        logger.info(f"Executing query: {query}")
        
        # Execute the query
        doc = await self.tasks_collection.find_one(query)
        
        if doc:
            logger.info(f"Task found: {doc}")
            return Task(**doc)
        
        logger.warning(f"Task not found: task_id={task_id}, collection={self.tasks_collection.name}, database={self.tasks_collection.database.name}")
        return None
    
    async def update_task(
        self, task_id: str, updates: Dict[str, Any]
    ) -> bool:
        """
        Updates a task
        
        Args:
            task_id: Task ID
            updates: Dictionary with field updates
            
        Returns:
            True if the task was updated, otherwise False
        """
        updates["updated_at"] = datetime.utcnow()
        
        result = await self.tasks_collection.update_one(
            {"_id": task_id},
            {"$set": updates}
        )
        
        success = result.matched_count > 0
        if success:
            logger.debug(f"Updated task {task_id} with fields {updates.keys()}")
        else:
            logger.warning(f"Failed to update task {task_id}, not found")
        
        return success
    
    async def update_task_status(
        self, task_id: str, status: str, result: Optional[Dict[str, Any]] = None, error: Optional[str] = None
    ) -> bool:
        """
        Updates task status
        
        Args:
            task_id: Task ID
            status: New status
            result: Task execution result
            error: Error description
            
        Returns:
            True if the task was updated, otherwise False
        """
        updates = {"status": status, "updated_at": datetime.utcnow()}
        
        if result is not None:
            updates["result"] = result
        
        if error is not None:
            updates["error"] = error
        
        result = await self.tasks_collection.update_one(
            {"_id": task_id},
            {"$set": updates}
        )
        
        success = result.matched_count > 0
        if success:
            logger.info(f"Updated task {task_id} status to {status}")
        else:
            logger.warning(f"Failed to update task {task_id} status, not found")
        
        return success
    
    async def claim_task(self, task_id: str, worker_id: str) -> bool:
        """
        Claims a task for processing by a specific worker
        
        Args:
            task_id: Task ID
            worker_id: Worker ID
            
        Returns:
            True if the task was claimed, otherwise False
        """
        # Use lock for atomic operation
        async with self.get_lock(f"task_{task_id}"):
            task = await self.get_task(task_id)
            
            if not task:
                logger.warning(f"Failed to claim task {task_id}, not found")
                return False
            
            if task.status != "pending":
                logger.warning(
                    f"Failed to claim task {task_id}, invalid status: {task.status}"
                )
                return False
            
            if task.worker_id:
                logger.warning(
                    f"Failed to claim task {task_id}, already being processed by worker {task.worker_id}"
                )
                return False
            
            result = await self.tasks_collection.update_one(
                {
                    "_id": task_id,
                    "status": "pending",
                    "worker_id": None
                },
                {
                    "$set": {
                        "status": "in_progress",
                        "worker_id": worker_id,
                        "updated_at": datetime.utcnow()
                    },
                    "$inc": {"attempt_count": 1}
                }
            )
            
            success = result.matched_count > 0
            if success:
                logger.info(f"Task {task_id} claimed by worker {worker_id}")
            else:
                logger.warning(f"Failed to claim task {task_id} (race condition)")
            
            return success
    
    
    async def find_tasks_by_world(self, world_id: str) -> List[Task]:
        """
        Finds all tasks for the specified world
        
        Args:
            world_id: World ID
            
        Returns:
            List of tasks
        """
        cursor = self.tasks_collection.find({"world_id": world_id})
        
        tasks = []
        async for doc in cursor:
            tasks.append(Task(**doc))
        
        return tasks
    
    # Methods for working with world parameters
    
    async def save_world_parameters(self, params: WorldParameters) -> str:
        """
        Saves world parameters
        
        Args:
            params: World parameters object
            
        Returns:
            ID of saved parameters
        """
        params_dict = params.dict(by_alias=True)
        
        # Check if parameters already exist for this world
        existing = await self.world_parameters_collection.find_one({"_id": params.id})
        
        if existing:
            # Update existing parameters
            await self.world_parameters_collection.update_one(
                {"_id": params.id},
                {"$set": {**params_dict, "updated_at": datetime.utcnow()}}
            )
            logger.info(f"Updated world parameters {params.id}")
        else:
            # Create new parameters
            await self.world_parameters_collection.insert_one(params_dict)
            logger.info(f"Created world parameters {params.id}")
        
        return params.id
    
    async def get_world_parameters(self, world_id: str) -> Optional[WorldParameters]:
        """
        Gets world parameters
        
        Args:
            world_id: World ID
            
        Returns:
            World parameters object or None if parameters not found
        """
        doc = await self.world_parameters_collection.find_one({"_id": world_id})
        if doc:
            return WorldParameters(**doc)
        return None
    
    # Methods for working with API request history
    
    async def log_api_request(self, request: ApiRequestHistory) -> str:
        """
        Logs API request to database
        
        Args:
            request: Request logging object
            
        Returns:
            Record ID
        """
        request_dict = request.dict(by_alias=True)
        await self.api_requests_history_collection.insert_one(request_dict)
        return request.id
        
    # Methods for working with world generation status
    
    async def initialize_world_generation_status(
        self,
        world_id: str,
        users_count: int,
        posts_count: int,
        user_prompt: str,
        api_call_limits_llm: int,
        api_call_limits_images: int
    ) -> WorldGenerationStatus:
        """
        Initializes world generation status record
        
        Args:
            world_id: World ID
            users_count: Number of users to generate
            posts_count: Number of posts to generate
            user_prompt: User prompt for world generation
            api_call_limits_llm: Limit on the number of LLM API calls
            api_call_limits_images: Limit on the number of image generation API calls
            
        Returns:
            Created WorldGenerationStatus object
            
        Raises:
            DuplicateKeyError: If a status with this ID already exists
        """
        now = datetime.utcnow()
        status = WorldGenerationStatus(
            _id=world_id,
            status=GenerationStatus.IN_PROGRESS,
            current_stage=GenerationStage.INITIALIZING,
            stages=[
                {"name": GenerationStage.INITIALIZING, "status": GenerationStatus.IN_PROGRESS},
                {"name": GenerationStage.WORLD_DESCRIPTION, "status": GenerationStatus.PENDING},
                {"name": GenerationStage.WORLD_IMAGE, "status": GenerationStatus.PENDING},
                {"name": GenerationStage.CHARACTERS, "status": GenerationStatus.PENDING},
                {"name": GenerationStage.POSTS, "status": GenerationStatus.PENDING},
                {"name": GenerationStage.FINISHING, "status": GenerationStatus.PENDING},
            ],
            tasks_total=0,
            tasks_completed=0,
            tasks_failed=0,
            task_predicted=0,
            users_created=0,
            posts_created=0,
            users_predicted=users_count,
            posts_predicted=posts_count,
            api_call_limits_LLM=api_call_limits_llm,
            api_call_limits_images=api_call_limits_images,
            api_calls_made_LLM=0,
            api_calls_made_images=0,
            parameters={
                "users_count": users_count,
                "posts_count": posts_count,
                "user_prompt": user_prompt,
            },
            created_at=now,
            updated_at=now,
        )
        
        try:
            await self.world_generation_status_collection.insert_one(status.dict(by_alias=True))
            logger.info(f"Initialized generation status for world {world_id}")
            return status
        except DuplicateKeyError:
            logger.warning(f"Generation status for world {world_id} already exists")
            raise
    
    async def get_world_generation_status(self, world_id: str) -> Optional[Dict[str, Any]]:
        """
        Gets world generation status
        
        Args:
            world_id: World ID
            
        Returns:
            World generation status document or None if not found
        """
        return await self.world_generation_status_collection.find_one({"_id": world_id})
    
    async def update_world_generation_stage(
        self, world_id: str, stage: str, status: str
    ) -> Optional[Dict[str, Any]]:
        """
        Updates generation stage status
        
        Args:
            world_id: World ID
            stage: Stage name
            status: New stage status
            
        Returns:
            Updated document or None if document not found
            
        Raises:
            ValueError: If world_id is not provided or stage/status is invalid
            RuntimeError: If generation status document is not found
        """
        if not world_id:
            raise ValueError("world_id is required")
            
        if not stage or not status:
            raise ValueError("stage and status are required")
            
        # Get current state
        current_status = await self.world_generation_status_collection.find_one({"_id": world_id})
        
        if not current_status:
            error_msg = f"Could not find generation status for world {world_id}"
            logger.error(error_msg)
            raise RuntimeError(error_msg)
        
        # Update stage status
        stages = current_status.get("stages", [])
        for s in stages:
            if s["name"] == stage:
                s["status"] = status
        
        # Set current stage if stage transitions to IN_PROGRESS
        if status == GenerationStatus.IN_PROGRESS:
            current_status["current_stage"] = stage
        
        # Check overall status
        failed_stages = [s for s in stages if s["status"] == GenerationStatus.FAILED]
        if failed_stages:
            overall_status = GenerationStatus.FAILED
        elif all(s["status"] == GenerationStatus.COMPLETED for s in stages):
            overall_status = GenerationStatus.COMPLETED
        else:
            overall_status = GenerationStatus.IN_PROGRESS
        
        current_status["status"] = overall_status
        current_status["updated_at"] = datetime.utcnow()
        
        try:
            # Save updated document
            await self.world_generation_status_collection.update_one(
                {"_id": world_id},
                {"$set": {
                    "stages": stages,
                    "current_stage": current_status["current_stage"],
                    "status": overall_status,
                    "updated_at": current_status["updated_at"]
                }}
            )
            
            logger.info(f"Updated stage {stage} status to {status} for world {world_id}")
            return current_status
        except Exception as e:
            error_msg = f"Failed to update stage status for world {world_id}: {str(e)}"
            logger.error(error_msg)
            raise RuntimeError(error_msg)
    
    async def increment_world_generation_counter(
        self, world_id: str, field: str, increment: int = 1
    ) -> Optional[Dict[str, Any]]:
        """
        Increments task counter
        
        Args:
            world_id: World ID
            field: Field name to increment (tasks_total, tasks_completed, tasks_failed, users_created, posts_created)
            increment: Value to increment by
            
        Returns:
            Updated document or None if document not found
            
        Raises:
            ValueError: If field is not allowed
        """
        allowed_fields = [
            "tasks_total", "tasks_completed", "tasks_failed", 
            "users_created", "posts_created", 
            "api_calls_made_LLM", "api_calls_made_images"
        ]
        
        if field not in allowed_fields:
            error_msg = f"Invalid field for increment: {field}"
            logger.error(error_msg)
            raise ValueError(error_msg)
        
        result = await self.world_generation_status_collection.update_one(
            {"_id": world_id},
            {
                "$inc": {field: increment},
                "$set": {"updated_at": datetime.utcnow()}
            }
        )
        
        if result.matched_count == 0:
            logger.error(f"Could not find generation status for world {world_id}")
            return None
        
        updated_status = await self.world_generation_status_collection.find_one({"_id": world_id})
        
        logger.debug(f"Incremented {field} counter by {increment} for world {world_id}")
        
        return updated_status
    
    async def update_world_generation_progress(
        self, world_id: str, updates: Dict[str, Any]
    ) -> Optional[Dict[str, Any]]:
        """
        Updates multiple fields in the generation status document
        
        Args:
            world_id: World ID
            updates: Dictionary with field updates
            
        Returns:
            Updated document or None if document not found
        """
        if not updates:
            return None
        
        # Add update time
        updates["updated_at"] = datetime.utcnow()
        
        result = await self.world_generation_status_collection.update_one(
            {"_id": world_id},
            {"$set": updates}
        )
        
        if result.matched_count == 0:
            logger.error(f"Could not find generation status for world {world_id}")
            return None
        
        updated_status = await self.world_generation_status_collection.find_one({"_id": world_id})
        
        logger.info(f"Updated progress for world {world_id}: {updates.keys()}")
        
        return updated_status