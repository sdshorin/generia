import asyncio
from typing import Dict, Any, Optional
from datetime import datetime

from .logger import logger
from ..constants import GenerationStatus, GenerationStage
from ..db.models import WorldGenerationStatus

class ProgressManager:
    """
    Manages world generation state and updates progress
    """
    
    def __init__(self, db_manager, kafka_producer=None):
        """
        Initializes progress manager
        
        Args:
            db_manager: Database manager for updating status
            kafka_producer: Optional Kafka producer for sending progress update events
        """
        self.db_manager = db_manager
        self.kafka_producer = kafka_producer
        self._lock = asyncio.Lock()
    
    async def initialize_world_generation(
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
        """
        status = await self.db_manager.initialize_world_generation_status(
            world_id=world_id,
            users_count=users_count,
            posts_count=posts_count,
            user_prompt=user_prompt,
            api_call_limits_llm=api_call_limits_llm,
            api_call_limits_images=api_call_limits_images
        )
        
        # if self.kafka_producer:
        #     await self.kafka_producer.send_progress_update(world_id, status.dict())
        
        return status
    
    async def update_stage(
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
            
        async with self._lock:
            try:
                current_status = await self.db_manager.update_world_generation_stage(
                    world_id=world_id,
                    stage=stage,
                    status=status
                )
                
                if self.kafka_producer:
                    await self.kafka_producer.send_progress_update(world_id, current_status)
                
                return current_status
            except Exception as e:
                logger.error(f"Error updating stage status: {str(e)}")
                raise
    
    async def increment_task_counter(
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
        """
        async with self._lock:
            try:
                updated_status = await self.db_manager.increment_world_generation_counter(
                    world_id=world_id,
                    field=field,
                    increment=increment
                )
                
                # if updated_status and self.kafka_producer and field in ["tasks_completed", "users_created", "posts_created"]:
                #     await self.kafka_producer.send_progress_update(world_id, updated_status)
                
                return updated_status
            except Exception as e:
                logger.error(f"Error incrementing task counter: {str(e)}")
                return None
            
    async def update_progress(
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
        
        async with self._lock:
            try:
                updated_status = await self.db_manager.update_world_generation_progress(
                    world_id=world_id,
                    updates=updates
                )
                
                # if updated_status and self.kafka_producer:
                #     await self.kafka_producer.send_progress_update(world_id, updated_status)
                
                return updated_status
            except Exception as e:
                logger.error(f"Error updating progress: {str(e)}")
                return None