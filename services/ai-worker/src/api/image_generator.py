import asyncio
import time
import uuid
import random
import base64
import hashlib
import os
from typing import Dict, Any, Optional
from datetime import datetime

import aiohttp
from ..utils.logger import logger
from ..utils.circuit_breaker import circuit_breaker
from ..utils.retries import with_retries
from ..db.models import ApiRequestHistory
from ..config import MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET, MINIO_USE_SSL

class ImageGenerator:
    """
    Image generator.
    
    Currently uses a stub, but the structure is ready for integration
    with a real image generation API (e.g., Stable Diffusion API).
    """
    
    def __init__(self, db_manager=None, service_client=None):
        """
        Initializes image generator
        
        Args:
            db_manager: Optional database manager for request logging
            service_client: Client for interacting with services (needed for uploading images)
        """
        self.db_manager = db_manager
        self.service_client = service_client
        self.semaphore = asyncio.Semaphore(10)  # Limit the number of concurrent requests
        self.session = None
    
    async def _ensure_session(self):
        """Ensures that aiohttp session is created"""
        if self.session is None or self.session.closed:
            self.session = aiohttp.ClientSession()
    
    async def close(self):
        """Closes aiohttp session"""
        if self.session and not self.session.closed:
            await self.session.close()
            self.session = None
    
    @circuit_breaker(name="image_generator", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def generate_image(
        self,
        prompt: str,
        width: int = 512,
        height: int = 512,
        task_id: Optional[str] = None,
        world_id: Optional[str] = None,
        media_type: str = "image/png",
        filename: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Generates an image from a prompt and uploads it via media-service
        
        Args:
            prompt: Prompt for generation
            width: Image width
            height: Image height
            task_id: Task ID for logging
            world_id: World ID for logging
            media_type: MIME type of the image
            filename: Filename (if not specified, will be generated)
            
        Returns:
            Dictionary with information about the generated image
        """
        async with self.semaphore:
            start_time = time.time()
            request_id = str(uuid.uuid4())
            
            if filename is None:
                # Generate filename based on prompt hash and current time
                prompt_hash = hashlib.md5(prompt.encode()).hexdigest()[:10]
                filename = f"generated_{prompt_hash}_{int(time.time())}.png"
            
            try:
                # Prepare request
                request_data = {
                    "prompt": prompt,
                    "width": width,
                    "height": height,
                    "media_type": media_type,
                    "filename": filename
                }
                
                # Real implementation would use an external API for generation
                # Here we use a stub
                
                # Simulate delay
                await asyncio.sleep(random.uniform(1.0, 3.0))
                
                # Generate image stub (in real implementation this would be an API call)
                # For demonstration, we'll use a simple placeholder
                
                # Base64 string with a single-pixel PNG image (for stub)
                base64_image = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
                
                # In real implementation, this would upload the image via media-service
                image_url = None
                media_id = None
                
                if self.service_client:
                    # Upload image via media service
                    upload_result = await self.service_client.upload_media_base64(
                        base64_data=base64_image,
                        media_type=media_type,
                        filename=filename,
                        world_id=world_id,
                        task_id=task_id
                    )
                    
                    image_url = upload_result.get("url")
                    media_id = upload_result.get("id")
                else:
                    # Without service client, use a stub URL
                    image_id = uuid.uuid4().hex
                    image_url = f"https://placeholder.com/{width}x{height}?text={image_id}"
                    media_id = image_id
                
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Prepare response
                result = {
                    "image_url": image_url,
                    "media_id": media_id,
                    "width": width,
                    "height": height,
                    "prompt": prompt,
                    "filename": filename
                }
                
                # Log request if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="image",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="generate_image",
                        request_data=request_data,
                        response_data=result,
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.info(
                    f"Image generation completed in {duration_ms}ms. "
                    f"TaskID: {task_id or 'manual'}"
                )
                
                return result
                
            except Exception as e:
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Log error if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="image",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="generate_image",
                        request_data=request_data,
                        error=str(e),
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.error(f"Image generation error: {str(e)}")
                raise
    
    # Preparation for future implementation with real API
    # Method will be used to send requests to the API for image generation
    async def _send_generation_request(self, prompt: str, width: int, height: int) -> Dict[str, Any]:
        """
        Sends a request to generate an image to an external API
        
        Args:
            prompt: Prompt for generation
            width: Image width
            height: Image height
            
        Returns:
            API response as a dictionary
        """
        # This implementation will be filled in the future when integrating with a real API
        # Example request structure based on GENERATION_EXAMPLE.py:
        # 
        # data = {
        #     "request_id": hashlib.md5(str(int(time.time())).encode()).hexdigest(),
        #     "stages": [
        #         {
        #             "type": "INPUT_INITIALIZE",
        #             "inputInitialize": {
        #                 "seed": -1,
        #                 "count": 1
        #             }
        #         },
        #         {
        #             "type": "DIFFUSION",
        #             "diffusion": {
        #                 "width": width,
        #                 "height": height,
        #                 "prompts": [
        #                     {
        #                         "text": prompt
        #                     }
        #                 ],
        #                 "sampler": "DPM++ 2M Karras",
        #                 "sdVae": "Automatic",
        #                 "steps": 15,
        #                 "sd_model": "600423083519508503",
        #                 "clip_skip": 2,
        #                 "cfg_scale": 7
        #             }
        #         }
        #     ]
        # }
        # 
        # await self._ensure_session()
        # async with self.session.post(url, json=data, headers=headers) as response:
        #     return await response.json()
        
        # Stub
        return {
            "image_data": "base64_encoded_data_will_be_here",
            "seed": random.randint(1, 100000),
            "steps": 15
        }