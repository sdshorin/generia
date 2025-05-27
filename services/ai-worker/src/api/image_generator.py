import asyncio
import time
import uuid
import random
import base64
import hashlib
import os
import io
from typing import Dict, Any, Optional, List, Tuple
from datetime import datetime

import aiohttp
from runware import Runware, IImageInference, IPromptEnhance
import requests
from PIL import Image

from ..config import RUNWARE_API_KEY
from ..utils.logger import logger
from ..utils.circuit_breaker import circuit_breaker
from ..utils.retries import with_retries
from ..db.models import ApiRequestHistory
from ..utils.media_uploader import download_and_upload_image

MODEL_ID = "runware:100@1"
# Cost per image generation in USD
IMAGE_GENERATION_COST = 0.0006

class ImageGenerator:
    """
    Image generator using Runware API.
    Handles image generation and upload to media-service.
    """

    def __init__(self, api_key: str = RUNWARE_API_KEY, db_manager=None, service_client=None, progress_manager=None):
        """
        Initializes the image generator

        Args:
            api_key: API key for Runware
            db_manager: Optional database manager for request logging
            service_client: Client for interacting with other services
            progress_manager: Optional progress manager for cost tracking
        """
        self.api_key = api_key
        self.db_manager = db_manager
        self.service_client = service_client
        self.progress_manager = progress_manager
        self.semaphore = asyncio.Semaphore(10)  # Limit concurrent requests
        self._runware = None

    async def _get_runware(self):
        """Get or create Runware client instance"""
        if self._runware is None:
            self._runware = Runware(api_key=self.api_key)
            await self._runware.connect()
        return self._runware

    @circuit_breaker(name="prompt_enhance", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def enhance_prompt(
        self,
        prompt: str,
        versions: int = 3,
        max_length: int = 100,
        task_id: Optional[str] = None,
        world_id: Optional[str] = None
    ) -> str:
        """
        Enhances a prompt for better image generation results

        Args:
            prompt: The original prompt to enhance
            versions: Number of enhanced versions to generate
            max_length: Maximum length of enhanced prompt
            task_id: Task ID for logging
            world_id: World ID for logging

        Returns:
            The best enhanced prompt
        """
        async with self.semaphore:
            start_time = time.time()
            request_id = str(uuid.uuid4())

            try:
                runware = await self._get_runware()

                prompt_enhancer = IPromptEnhance(
                    prompt=prompt,
                    promptVersions=versions,
                    promptMaxLength=max_length,
                )

                enhanced_prompts = await runware.promptEnhance(promptEnhancer=prompt_enhancer)

                # Pick the first enhanced prompt (best match)
                best_prompt = enhanced_prompts[0].text if enhanced_prompts else prompt

                duration_ms = int((time.time() - start_time) * 1000)

                # Log request if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="image_generation",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="enhance_prompt",
                        request_data={"prompt": prompt},
                        response_data={"enhanced_prompt": best_prompt},
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)

                logger.info(f"Enhanced prompt in {duration_ms}ms. TaskID: {task_id or 'manual'}")
                return best_prompt

            except Exception as e:
                duration_ms = int((time.time() - start_time) * 1000)

                # Log error if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="image_generation",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="enhance_prompt",
                        request_data={"prompt": prompt},
                        error=str(e),
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)

                logger.error(f"Prompt enhancement error: {str(e)}")
                # Return original prompt if enhancement fails
                return prompt

    @circuit_breaker(name="image_generator", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def generate_image(
        self,
        prompt: str,
        world_id: str,
        media_type_enum: int,
        character_id: Optional[str] = None,
        width: int = 512,
        height: int = 512,
        task_id: Optional[str] = None,
        media_type: str = "image/png",
        filename: Optional[str] = None,
        enhance_prompt: bool = False,
        model: str = MODEL_ID
        # model: str = "civitai:101055@128078"  # Default model
    ) -> Dict[str, Any]:
        """
        Generates an image from a prompt and uploads it via media-service

        Args:
            prompt: Prompt for generation
            world_id: World ID for logging and storage path
            media_type_enum: Media type enum value (from MediaType constants)
            character_id: Character ID for media association (optional for world-level media)
            width: Image width
            height: Image height
            task_id: Task ID for logging
            media_type: MIME type of the image
            filename: Filename (if not specified, will be generated)
            enhance_prompt: Whether to enhance the prompt automatically
            model: The model ID to use for generation

        Returns:
            Dictionary with information about the generated image
        """
        async with self.semaphore:
            start_time = time.time()
            request_id = str(uuid.uuid4())

            if filename is None:
                filename = f"generia_{uuid.uuid4()}.png"

            try:
                # Enhance prompt if requested
                if enhance_prompt:
                    prompt = await self.enhance_prompt(prompt, task_id=task_id, world_id=world_id)

                runware = await self._get_runware()

                # Prepare image generation request
                request_image = IImageInference(
                    positivePrompt=prompt,
                    model=model,
                    numberResults=1,  # We only need one image
                    negativePrompt="blurry, deformed, disfigured, bad anatomy, ugly, text, watermark",
                    height=height,
                    width=width,
                )

                # For request logging
                request_data = {
                    "prompt": prompt,
                    "width": width,
                    "height": height,
                    "model": model
                }

                # Generate the image
                images = await runware.imageInference(requestImage=request_image)

                if not images:
                    raise Exception("No images were generated")

                # Update cost in progress manager if available
                if self.progress_manager and world_id:
                    await self.progress_manager.increment_cost(
                        world_id=world_id,
                        cost_type="image",
                        cost=IMAGE_GENERATION_COST
                    )

                # Get the image URL
                image_url = images[0].imageURL
                logger.info(f"Generated image at URL: {image_url}")

                # Upload image to media service
                if not self.service_client:
                    logger.warning("Service client not available, cannot upload image")
                    return {
                        "image_url": image_url,
                        "width": width,
                        "height": height
                    }

                logger.info("Getting presigned URL for upload")
                # Get presigned URL for upload
                media_id, upload_url, expires_at = await self.service_client.get_presigned_upload_url(
                    world_id=world_id,
                    character_id=character_id or "",  # Empty string if None
                    filename=filename,
                    content_type=media_type,
                    size=0,  # Мы не знаем размер заранее, поэтому передаем 0
                    media_type_enum=media_type_enum,
                    task_id=task_id
                )

                if not media_id or not upload_url:
                    raise Exception("Failed to get presigned upload URL")

                logger.info(f"Got presigned URL with media_id: {media_id}")

                # Скачиваем и загружаем изображение
                image_data, upload_result = await download_and_upload_image(
                    download_url=image_url,
                    upload_url=upload_url,
                    content_type=media_type,
                    timeout=60
                )

                if not upload_result["success"]:
                    raise Exception(f"Failed to upload image: {upload_result.get('error', 'Unknown error')}")

                if image_data is None:
                    raise Exception("Image download failed")

                logger.info(f"Successfully uploaded image, confirming upload with media_id: {media_id}")

                # Confirm upload
                success = await self.service_client.confirm_upload(
                    media_id=media_id,
                    task_id=task_id
                )

                if not success:
                    raise Exception("Failed to confirm upload")

                # Так как нам не возвращаются варианты из метода confirm_upload,
                # мы будем использовать оригинальный URL для отображения
                public_url = image_url
                variants = []

                duration_ms = int((time.time() - start_time) * 1000)

                # Prepare result
                result = {
                    "media_id": media_id,
                    "image_url": public_url or image_url,
                    "width": width,
                    "height": height,
                    "variants": variants,
                    "cost": IMAGE_GENERATION_COST
                }

                # Log request if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="image_generation",
                        task_id=task_id or "manual",
                        world_id=world_id,
                        request_type="generate_image",
                        request_data=request_data,
                        response_data=result,
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)

                logger.info(
                    f"Image generation completed in {duration_ms}ms. "
                    f"Media ID: {media_id}, TaskID: {task_id or 'manual'}"
                )

                return result

            except Exception as e:
                duration_ms = int((time.time() - start_time) * 1000)

                # Log error if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="image_generation",
                        task_id=task_id or "manual",
                        world_id=world_id,
                        request_type="generate_image",
                        request_data=request_data,
                        error=str(e),
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)

                logger.error(f"Image generation error: {str(e)}")
                raise