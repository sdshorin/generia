import aiohttp
import asyncio
import time
import uuid
import json
import base64
from typing import Dict, Any, Optional, List
from datetime import datetime

from ..config import API_GATEWAY_URL
from ..utils.logger import logger
from ..utils.circuit_breaker import circuit_breaker
from ..utils.retries import with_retries
from ..db.models import ApiRequestHistory

class ServiceClient:
    """
    Client for interacting with other microservices via API Gateway
    """
    
    def __init__(self, api_gateway_url: str = API_GATEWAY_URL, db_manager=None):
        """
        Initializes client for interacting with microservices
        
        Args:
            api_gateway_url: API Gateway URL
            db_manager: Optional database manager for request logging
        """
        self.api_gateway_url = api_gateway_url.rstrip('/')
        self.db_manager = db_manager
        self.session = None
    
    async def _ensure_session(self):
        """
        Ensures that aiohttp session is created
        """
        if self.session is None or self.session.closed:
            self.session = aiohttp.ClientSession()
    
    async def close(self):
        """
        Closes aiohttp session
        """
        if self.session and not self.session.closed:
            await self.session.close()
            self.session = None
    
    async def _request(
        self,
        method: str,
        endpoint: str,
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
        params: Optional[Dict[str, Any]] = None,
        task_id: Optional[str] = None,
        world_id: Optional[str] = None,
        service_name: str = "api-gateway"
    ) -> Dict[str, Any]:
        """
        Performs HTTP request to microservice
        
        Args:
            method: HTTP method (GET, POST, PUT, DELETE)
            endpoint: API endpoint path
            data: JSON data to send
            headers: HTTP headers
            params: Request parameters
            task_id: Task ID for logging
            world_id: World ID for logging
            service_name: Service name for circuit breaker
            
        Returns:
            API response as a dictionary
        """
        await self._ensure_session()
        
        url = f"{self.api_gateway_url}{endpoint}"
        request_id = str(uuid.uuid4())
        start_time = time.time()
        
        # Prepare request for logging
        request_data = {
            "method": method,
            "url": url,
            "data": data,
            "params": params,
        }
        
        try:
            headers = headers or {}
            headers.update({
                "Content-Type": "application/json",
                "Accept": "application/json",
            })
            
            async with self.session.request(
                method=method,
                url=url,
                json=data,
                headers=headers,
                params=params
            ) as response:
                try:
                    response_data = await response.json()
                except:
                    response_data = {"status": "error", "message": await response.text()}
                
                duration_ms = int((time.time() - start_time) * 1000)
                
                if response.status >= 400:
                    error_message = f"API error: {response.status} - {response_data.get('message', 'Unknown error')}"
                    logger.error(error_message)
                    
                    # Log error if db_manager is available
                    if self.db_manager:
                        log_entry = ApiRequestHistory(
                            id=request_id,
                            api_type="service",
                            task_id=task_id or "manual",
                            world_id=world_id or "unknown",
                            request_type=f"{service_name}_{method.lower()}_{endpoint}",
                            request_data=request_data,
                            response_data={"status": response.status, "error": response_data},
                            error=error_message,
                            duration_ms=duration_ms,
                            created_at=datetime.utcnow()
                        )
                        await self.db_manager.log_api_request(log_entry)
                    
                    raise Exception(error_message)
                
                # Log request if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="service",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type=f"{service_name}_{method.lower()}_{endpoint}",
                        request_data=request_data,
                        response_data=response_data,
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.debug(
                    f"API call {method} {endpoint} completed in {duration_ms}ms. "
                    f"Status: {response.status}"
                )
                
                return response_data
                
        except aiohttp.ClientError as e:
            duration_ms = int((time.time() - start_time) * 1000)
            error_message = f"API client error: {str(e)}"
            logger.error(error_message)
            
            # Log error if db_manager is available
            if self.db_manager:
                log_entry = ApiRequestHistory(
                    id=request_id,
                    api_type="service",
                    task_id=task_id or "manual",
                    world_id=world_id or "unknown",
                    request_type=f"{service_name}_{method.lower()}_{endpoint}",
                    request_data=request_data,
                    error=error_message,
                    duration_ms=duration_ms,
                    created_at=datetime.utcnow()
                )
                await self.db_manager.log_api_request(log_entry)
            
            raise
    
    # Methods for working with World Service
    
    @circuit_breaker(name="world_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def get_world_info(self, world_id: str, task_id: Optional[str] = None) -> Dict[str, Any]:
        """
        Gets world information
        
        Args:
            world_id: World ID
            task_id: Task ID for logging
            
        Returns:
            World information
        """
        return await self._request(
            method="GET",
            endpoint=f"/api/v1/worlds/{world_id}",
            task_id=task_id,
            world_id=world_id,
            service_name="world-service"
        )
    
    @circuit_breaker(name="world_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def update_world_status(
        self, world_id: str, status: str, task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Updates world generation status
        
        Args:
            world_id: World ID
            status: New status
            task_id: Task ID for logging
            
        Returns:
            Operation result
        """
        return await self._request(
            method="PUT",
            endpoint=f"/api/v1/worlds/{world_id}/status",
            data={"status": status},
            task_id=task_id,
            world_id=world_id,
            service_name="world-service"
        )
    
    @circuit_breaker(name="world_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def update_world_images(
        self, 
        world_id: str, 
        header_url: str, 
        icon_url: str, 
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Updates world images
        
        Args:
            world_id: World ID
            header_url: Background image URL
            icon_url: Icon URL
            task_id: Task ID for logging
            
        Returns:
            Operation result
        """
        return await self._request(
            method="PUT",
            endpoint=f"/api/v1/worlds/{world_id}/images",
            data={
                "header_image_url": header_url,
                "icon_url": icon_url
            },
            task_id=task_id,
            world_id=world_id,
            service_name="world-service"
        )
    
    # Methods for working with Auth Service
    
    @circuit_breaker(name="auth_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def create_ai_user(
        self,
        world_id: str,
        username: str,
        display_name: str,
        avatar_url: Optional[str] = None,
        bio: Optional[str] = None,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Creates AI user
        
        Args:
            world_id: World ID
            username: User name
            display_name: Display name
            avatar_url: Avatar URL
            bio: Profile description
            task_id: Task ID for logging
            
        Returns:
            Information about created user
        """
        data = {
            "world_id": world_id,
            "username": username,
            "display_name": display_name,
            "ai_generated": True
        }
        
        if avatar_url:
            data["avatar_url"] = avatar_url
        
        if bio:
            data["bio"] = bio
        
        return await self._request(
            method="POST",
            endpoint="/api/v1/auth/create-ai-user",
            data=data,
            task_id=task_id,
            world_id=world_id,
            service_name="auth-service"
        )
    
    # Methods for working with Media Service
    
    @circuit_breaker(name="media_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def upload_media_base64(
        self,
        base64_data: str,
        media_type: str,
        filename: str,
        world_id: Optional[str] = None,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Uploads media file in base64 format
        
        Args:
            base64_data: File data in base64 format
            media_type: File MIME type
            filename: File name
            world_id: World ID
            task_id: Task ID for logging
            
        Returns:
            Information about uploaded file
        """
        data = {
            "file": base64_data,
            "media_type": media_type,
            "filename": filename,
            "world_id": world_id
        }
        
        return await self._request(
            method="POST",
            endpoint="/api/v1/media",
            data=data,
            task_id=task_id,
            world_id=world_id,
            service_name="media-service"
        )
    
    @circuit_breaker(name="media_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def get_upload_url(
        self,
        media_type: str,
        filename: str,
        world_id: Optional[str] = None,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Gets URL for direct file upload
        
        Args:
            media_type: File MIME type
            filename: File name
            world_id: World ID
            task_id: Task ID for logging
            
        Returns:
            Information with URL for upload and media ID
        """
        data = {
            "media_type": media_type,
            "filename": filename,
            "world_id": world_id
        }
        
        return await self._request(
            method="POST",
            endpoint="/api/v1/media/upload-url",
            data=data,
            task_id=task_id,
            world_id=world_id,
            service_name="media-service"
        )
    
    # Methods for working with Post Service
    
    @circuit_breaker(name="post_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def create_post(
        self,
        world_id: str,
        user_id: str,
        content: str,
        media_ids: Optional[List[str]] = None,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Creates post
        
        Args:
            world_id: World ID
            user_id: User ID
            world_id: ID мира
            user_id: ID пользователя
            content: Текст поста
            media_ids: Список ID медиа-файлов
            task_id: ID задачи для логирования
            
        Returns:
            Информация о созданном посте
        """
        data = {
            "user_id": user_id,
            "content": content
        }
        
        if media_ids:
            data["media_ids"] = media_ids
        
        return await self._request(
            method="POST",
            endpoint=f"/api/v1/worlds/{world_id}/posts",
            data=data,
            task_id=task_id,
            world_id=world_id,
            service_name="post-service"
        )
    
    @circuit_breaker(name="post_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def create_comment(
        self,
        world_id: str,
        post_id: str,
        user_id: str,
        content: str,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Создает комментарий к посту
        
        Args:
            world_id: ID мира
            post_id: ID поста
            user_id: ID пользователя
            content: Текст комментария
            task_id: ID задачи для логирования
            
        Returns:
            Информация о созданном комментарии
        """
        data = {
            "user_id": user_id,
            "content": content
        }
        
        return await self._request(
            method="POST",
            endpoint=f"/api/v1/worlds/{world_id}/posts/{post_id}/comments",
            data=data,
            task_id=task_id,
            world_id=world_id,
            service_name="post-service"
        )
    
    # Methods for working with Interaction Service
    
    @circuit_breaker(name="interaction_service", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def like_post(
        self,
        world_id: str,
        post_id: str,
        user_id: str,
        task_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Ставит лайк на пост
        
        Args:
            world_id: ID мира
            post_id: ID поста
            user_id: ID пользователя
            task_id: ID задачи для логирования
            
        Returns:
            Результат операции
        """
        data = {
            "user_id": user_id
        }
        
        return await self._request(
            method="POST",
            endpoint=f"/api/v1/worlds/{world_id}/posts/{post_id}/like",
            data=data,
            task_id=task_id,
            world_id=world_id,
            service_name="interaction-service"
        )