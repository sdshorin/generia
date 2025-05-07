import asyncio
import time
import uuid
import json
import base64
import grpc
from typing import Dict, Any, Optional, List, Tuple
from datetime import datetime
from google.protobuf.json_format import MessageToDict

from ..utils.logger import logger
from ..utils.circuit_breaker import circuit_breaker
from ..utils.retries import with_retries
from ..db.models import ApiRequestHistory
from ..utils.discovery import ConsulServiceDiscovery

# Import gRPC generated modules
try:
    # Try direct import first
    from ..grpc.character import character_pb2, character_pb2_grpc
    from ..grpc.media import media_pb2, media_pb2_grpc
    from ..grpc.post import post_pb2, post_pb2_grpc
    from ..grpc.world import world_pb2, world_pb2_grpc
except ImportError:
    # Fallback to absolute imports if relative imports fail
    import sys
    import os
    import importlib.util

    # Get the absolute path to the grpc directory
    grpc_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'grpc'))

    # Add to Python path if not already there
    if grpc_dir not in sys.path:
        sys.path.append(grpc_dir)

    # Import modules using importlib
    for service in ['character', 'media', 'post', 'world']:
        # Create the full path to the module
        pb2_path = os.path.join(grpc_dir, service, f"{service}_pb2.py")
        grpc_pb2_path = os.path.join(grpc_dir, service, f"{service}_pb2_grpc.py")

        # Check if files exist
        if not os.path.exists(pb2_path) or not os.path.exists(grpc_pb2_path):
            import logging
            logging.error(f"Missing gRPC files for {service} service")
            continue

        # Import the modules
        spec = importlib.util.spec_from_file_location(f"{service}_pb2", pb2_path)
        pb2_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(pb2_module)

        spec = importlib.util.spec_from_file_location(f"{service}_pb2_grpc", grpc_pb2_path)
        grpc_pb2_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(grpc_pb2_module)

        # Assign modules to global namespace
        globals()[f"{service}_pb2"] = pb2_module
        globals()[f"{service}_pb2_grpc"] = grpc_pb2_module

class ServiceClient:
    """
    Client for interacting with other microservices via gRPC
    """

    def __init__(self, db_manager=None):
        """
        Initializes client for interacting with microservices via gRPC

        Args:
            db_manager: Optional database manager for request logging
        """
        self.db_manager = db_manager
        self.discovery_client = ConsulServiceDiscovery()

        # gRPC stubs
        self.character_stub = None
        self.media_stub = None
        self.post_stub = None
        self.world_stub = None

        # gRPC channels
        self.character_channel = None
        self.media_channel = None
        self.post_channel = None
        self.world_channel = None

    async def initialize(self):
        """Initialize service client"""
        await self.discovery_client.initialize()

        # Initialize gRPC stubs
        await self._init_character_stub()
        await self._init_media_stub()
        await self._init_post_stub()
        await self._init_world_stub()

        return self

    async def close(self):
        """Close all gRPC channels"""
        # Close all channels
        channels = [
            self.character_channel,
            self.media_channel,
            self.post_channel,
            self.world_channel
        ]

        for channel in channels:
            if channel:
                await channel.close()

        # Reset stubs and channels
        self.character_stub = None
        self.media_stub = None
        self.post_stub = None
        self.world_stub = None

        self.character_channel = None
        self.media_channel = None
        self.post_channel = None
        self.world_channel = None

        if self.discovery_client:
            await self.discovery_client.close()

    async def _log_grpc_request(
        self,
        service_name: str,
        method_name: str,
        request_data: Any,
        response_data: Any,
        duration_ms: int,
        error: Optional[str] = None,
        task_id: Optional[str] = None,
        world_id: Optional[str] = None
    ):
        """Log gRPC request to database"""
        if not self.db_manager:
            return

        request_id = str(uuid.uuid4())

        # Convert request and response to dict if needed
        if hasattr(request_data, "DESCRIPTOR"):
            request_dict = MessageToDict(request_data)
        else:
            request_dict = request_data

        if hasattr(response_data, "DESCRIPTOR"):
            response_dict = MessageToDict(response_data)
        else:
            response_dict = response_data

        log_entry = ApiRequestHistory(
            id=request_id,
            api_type="grpc",
            task_id=task_id or "manual",
            world_id=world_id or "unknown",
            request_type=f"{service_name}_{method_name}",
            request_data=request_dict,
            response_data=response_dict,
            error=error,
            duration_ms=duration_ms,
            created_at=datetime.utcnow()
        )

        await self.db_manager.log_api_request(log_entry)

    async def _init_character_stub(self):
        """Initialize character service gRPC stub"""
        try:
            # Get service address from Consul
            address = await self.discovery_client.resolve_service("character-service")
            logger.info(f"Using character service at: {address}")

            # Create channel and stub
            self.character_channel = grpc.aio.insecure_channel(address)
            self.character_stub = character_pb2_grpc.CharacterServiceStub(self.character_channel)

            return self.character_stub
        except Exception as e:
            logger.error(f"Failed to initialize character stub: {str(e)}")
            raise

    async def _init_media_stub(self):
        """Initialize media service gRPC stub"""
        try:
            # Get service address from Consul
            address = await self.discovery_client.resolve_service("media-service")
            logger.info(f"Using media service at: {address}")

            # Create channel and stub
            self.media_channel = grpc.aio.insecure_channel(address)
            self.media_stub = media_pb2_grpc.MediaServiceStub(self.media_channel)

            return self.media_stub
        except Exception as e:
            logger.error(f"Failed to initialize media stub: {str(e)}")
            raise

    async def _init_post_stub(self):
        """Initialize post service gRPC stub"""
        try:
            # Get service address from Consul
            address = await self.discovery_client.resolve_service("post-service")
            logger.info(f"Using post service at: {address}")

            # Create channel and stub
            self.post_channel = grpc.aio.insecure_channel(address)
            self.post_stub = post_pb2_grpc.PostServiceStub(self.post_channel)

            return self.post_stub
        except Exception as e:
            logger.error(f"Failed to initialize post stub: {str(e)}")
            raise

    async def _init_world_stub(self):
        """Initialize world service gRPC stub"""
        try:
            # Get service address from Consul
            address = await self.discovery_client.resolve_service("world-service")
            logger.info(f"Using world service at: {address}")

            # Create channel and stub
            self.world_channel = grpc.aio.insecure_channel(address)
            self.world_stub = world_pb2_grpc.WorldServiceStub(self.world_channel)

            return self.world_stub
        except Exception as e:
            logger.error(f"Failed to initialize world stub: {str(e)}")
            raise

    async def _ensure_character_stub(self):
        """Ensure character service stub is available"""
        if not self.character_stub:
            await self._init_character_stub()
        return self.character_stub

    async def _ensure_media_stub(self):
        """Ensure media service stub is available"""
        if not self.media_stub:
            await self._init_media_stub()
        return self.media_stub

    async def _ensure_post_stub(self):
        """Ensure post service stub is available"""
        if not self.post_stub:
            await self._init_post_stub()
        return self.post_stub

    async def _ensure_world_stub(self):
        """Ensure world service stub is available"""
        if not self.world_stub:
            await self._init_world_stub()
        return self.world_stub

    # Character Service methods

    @circuit_breaker(name="character_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def create_character(
        self,
        world_id: str,
        display_name: str,
        meta: Optional[Dict[str, Any]] = None,
        avatar_media_id: Optional[str] = None,
        task_id: Optional[str] = None
    ) -> Tuple[str, Dict[str, Any]]:
        """
        Creates a character in the specified world

        Args:
            world_id: World ID
            display_name: Character display name
            meta: Additional character metadata (will be serialized to JSON)
            avatar_media_id: Optional avatar media ID
            task_id: Task ID for logging

        Returns:
            Tuple of (character_id, character_data)
        """
        stub = await self._ensure_character_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = character_pb2.CreateCharacterRequest(
                world_id=world_id,
                display_name=display_name,
            )

            # Add optional fields
            if meta:
                request.meta = json.dumps(meta)

            if avatar_media_id:
                request.avatar_media_id = avatar_media_id

            # Call gRPC method
            response = await stub.CreateCharacter(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="character-service",
                method_name="CreateCharacter",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            # Convert to dictionary and return
            response_dict = MessageToDict(response)
            return response.id, response_dict

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for create_character in {world_id} with display name {display_name} and meta {meta} and avatar media id {avatar_media_id}"
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="character-service",
                method_name="CreateCharacter",
                request_data={
                    "world_id": world_id,
                    "display_name": display_name,
                    "meta": meta,
                    "avatar_media_id": avatar_media_id
                },
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            raise Exception(f"Failed to create character: {str(e)}")

    @circuit_breaker(name="character_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def get_character(
        self,
        character_id: str,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Gets a character by ID

        Args:
            character_id: Character ID
            task_id: Task ID for logging

        Returns:
            Character data
        """
        stub = await self._ensure_character_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = character_pb2.GetCharacterRequest(
                character_id=character_id
            )

            # Call gRPC method
            response = await stub.GetCharacter(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="character-service",
                method_name="GetCharacter",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id
            )

            # Convert to dictionary and return
            return MessageToDict(response)

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for get_character with character id {character_id}"
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="character-service",
                method_name="GetCharacter",
                request_data={"character_id": character_id},
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id
            )

            raise Exception(f"Failed to get character: {str(e)}")

    @circuit_breaker(name="character_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def update_character(
        self,
        character_id: str,
        display_name: Optional[str] = None,
        avatar_media_id: Optional[str] = None,
        meta: Optional[Dict[str, Any]] = None,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Updates a character

        Args:
            character_id: Character ID
            display_name: Optional new display name
            avatar_media_id: Optional new avatar media ID
            meta: Optional new metadata (will be serialized to JSON)
            task_id: Task ID for logging

        Returns:
            Updated character data
        """
        stub = await self._ensure_character_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = character_pb2.UpdateCharacterRequest(
                character_id=character_id
            )

            # Add optional fields
            if display_name is not None:
                request.display_name = display_name

            if avatar_media_id is not None:
                request.avatar_media_id = avatar_media_id

            if meta is not None:
                request.meta = json.dumps(meta)

            # Call gRPC method
            response = await stub.UpdateCharacter(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="character-service",
                method_name="UpdateCharacter",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id
            )

            # Convert to dictionary and return
            return MessageToDict(response)

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for update_character with character id {character_id}"
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="character-service",
                method_name="UpdateCharacter",
                request_data={
                    "character_id": character_id,
                    "display_name": display_name,
                    "avatar_media_id": avatar_media_id,
                    "meta": meta
                },
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id
            )

            raise Exception(f"Failed to update character: {str(e)}")

    # Media Service methods

    @circuit_breaker(name="media_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def get_presigned_upload_url(
        self,
        character_id: str,
        world_id: str,
        filename: str,
        content_type: str,
        size: int,
        task_id: Optional[str] = None
    ) -> Tuple[str, str, int]:
        """
        Gets a presigned URL for uploading media

        Args:
            character_id: Character ID
            world_id: World ID
            filename: File name
            content_type: File content type (MIME)
            size: File size in bytes
            task_id: Task ID for logging

        Returns:
            Tuple of (media_id, upload_url, expires_at)
        """
        stub = await self._ensure_media_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = media_pb2.GetPresignedUploadURLRequest(
                character_id=character_id,
                world_id=world_id,
                filename=filename,
                content_type=content_type,
                size=size
            )

            # Call gRPC method
            response = await stub.GetPresignedUploadURL(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="media-service",
                method_name="GetPresignedUploadURL",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            return response.media_id, response.upload_url, response.expires_at

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for "
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="media-service",
                method_name="GetPresignedUploadURL",
                request_data={
                    "character_id": character_id,
                    "world_id": world_id,
                    "filename": filename,
                    "content_type": content_type,
                    "size": size
                },
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            raise Exception(f"Failed to get presigned upload URL: {str(e)}")

    @circuit_breaker(name="media_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def confirm_upload(
        self,
        media_id: str,
        character_id: str,
        task_id: Optional[str] = None
    ) -> bool:
        """
        Confirms that a media file has been uploaded

        Args:
            media_id: Media ID
            character_id: Character ID
            task_id: Task ID for logging

        Returns:
            True if confirmation was successful
        """
        stub = await self._ensure_media_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = media_pb2.ConfirmUploadRequest(
                media_id=media_id,
                character_id=character_id
            )

            # Call gRPC method
            response = await stub.ConfirmUpload(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="media-service",
                method_name="ConfirmUpload",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id
            )

            # Возвращаем булево значение
            return bool(response.success)

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for confirm_upload with media id {media_id} and character id {character_id} "
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="media-service",
                method_name="ConfirmUpload",
                request_data={
                    "media_id": media_id,
                    "character_id": character_id
                },
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id
            )

            raise Exception(f"Failed to confirm upload: {str(e)}")

    # Post Service methods

    @circuit_breaker(name="post_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def create_ai_post(
        self,
        character_id: str,
        caption: str,
        media_id: str,
        world_id: str,
        tags: List[str] = None,
        task_id: Optional[str] = None
    ) -> Tuple[str, str]:
        """
        Creates an AI post

        Args:
            character_id: Character ID
            caption: Post caption
            media_id: Media ID
            world_id: World ID
            tags: Optional list of tags
            task_id: Task ID for logging

        Returns:
            Tuple of (post_id, created_at)
        """
        stub = await self._ensure_post_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = post_pb2.CreateAIPostRequest(
                character_id=character_id,
                caption=caption,
                media_id=media_id,
                world_id=world_id,
            )

            # Add tags if available
            if tags:
                request.tags.extend(tags)

            # Call gRPC method
            response = await stub.CreateAIPost(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="post-service",
                method_name="CreateAIPost",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            # Возвращаем словарь с результатом
            return {
                "post_id": response.post_id,
                "created_at": response.created_at
            }

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for create_ai_post with character id {character_id} with caption {caption} and media id {media_id} and world id {world_id} and tags {tags}"
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="post-service",
                method_name="CreateAIPost",
                request_data={
                    "character_id": character_id,
                    "caption": caption,
                    "media_id": media_id,
                    "world_id": world_id,
                    "tags": tags or []
                },
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            raise Exception(f"Failed to create AI post: {str(e)}")

    # World Service methods

    @circuit_breaker(name="world_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def get_world(
        self,
        world_id: str,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Gets world information

        Args:
            world_id: World ID
            task_id: Task ID for logging

        Returns:
            World information
        """
        stub = await self._ensure_world_stub()
        start_time = time.time()
        try:
            # Prepare request
            request = world_pb2.GetWorldRequest(
                id=world_id
            )

            # Call gRPC method
            response = await stub.GetWorld(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="world-service",
                method_name="GetWorld",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            # Convert to dictionary and return
            return MessageToDict(response)

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for get_world with world id {world_id}"
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="world-service",
                method_name="GetWorld",
                request_data={"id": world_id},
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            raise Exception(f"Failed to get world: {str(e)}")

    @circuit_breaker(name="world_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def update_world_images(
        self,
        world_id: str,
        header_image_id: str,
        icon_image_id: str,
        task_id: Optional[str] = None
    ) -> bool:
        """
        Updates world images

        Args:
            world_id: World ID
            header_image_id: Header image media ID
            icon_image_id: Icon image media ID
            task_id: Task ID for logging

        Returns:
            True if update was successful
        """
        stub = await self._ensure_world_stub()
        start_time = time.time()

        try:
            # Prepare request
            request = world_pb2.UpdateWorldImagesRequest(
                id=world_id,
                header_image_id=header_image_id,
                icon_image_id=icon_image_id
            )

            # Call gRPC method
            response = await stub.UpdateWorldImages(request)

            # Log request
            duration_ms = int((time.time() - start_time) * 1000)
            await self._log_grpc_request(
                service_name="world-service",
                method_name="UpdateWorldImages",
                request_data=request,
                response_data=response,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            return response.success

        except grpc.RpcError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"gRPC error: {str(e)} for update_world_images with world id {world_id} with header image id {header_image_id} and icon image id {icon_image_id}"
            logger.error(error_message)

            # Log error
            await self._log_grpc_request(
                service_name="world-service",
                method_name="UpdateWorldImages",
                request_data={
                    "id": world_id,
                    "header_image_id": header_image_id,
                    "icon_image_id": icon_image_id
                },
                response_data=None,
                error=error_message,
                duration_ms=duration_ms,
                task_id=task_id,
                world_id=world_id
            )

            raise Exception(f"Failed to update world images: {str(e)}")