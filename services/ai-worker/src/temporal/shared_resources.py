"""
Shared resources manager для Temporal workers
Управляет connection pools и clients согласно best practices
"""
import asyncio
import os
from typing import Dict, Any, Optional, Type
from temporalio import workflow
with workflow.unsafe.imports_passed_through():
    from pydantic import BaseModel, Field
import motor.motor_asyncio as motor
import grpc.aio
import httpx




class SharedResourcesManager:
    """
    Менеджер shared resources с proper connection pooling
    Создает ресурсы один раз при запуске worker и переиспользует их
    """
    
    def __init__(self):
        # Connection pools будут инициализированы в initialize()
        self.mongo_client = None
        self.grpc_channels = {}
        self.http_client = None
        
        # Clients будут созданы поверх connection pools
        self.db_manager = None
        self.llm_client = None
        self.image_generator = None
        self.service_client = None
        
        # Schema mapping для LLM
        self.schema_mapping = {}
        
        # Semaphores для ограничения параллельных запросов
        from ..config import (
            MAX_CONCURRENT_LLM_REQUESTS,
            MAX_CONCURRENT_IMAGE_REQUESTS, 
            MAX_CONCURRENT_GRPC_CALLS,
            MAX_CONCURRENT_DB_OPERATIONS
        )
        
        self.llm_semaphore = asyncio.Semaphore(MAX_CONCURRENT_LLM_REQUESTS)
        self.image_semaphore = asyncio.Semaphore(MAX_CONCURRENT_IMAGE_REQUESTS)
        self.grpc_semaphore = asyncio.Semaphore(MAX_CONCURRENT_GRPC_CALLS)
        self.db_semaphore = asyncio.Semaphore(MAX_CONCURRENT_DB_OPERATIONS)
        
        # logger.info("SharedResourcesManager created")
    
    async def initialize(self):
        """Инициализирует все connection pools и clients"""
        # Импортируем конфиги
        from ..config import (
            MONGODB_URI,
            MAX_CONCURRENT_DB_OPERATIONS,
            CONSUL_HOST,
            CONSUL_PORT
        )
        
        # 1. Инициализируем MongoDB connection pool
        self.mongo_client = motor.AsyncIOMotorClient(
            MONGODB_URI,
            maxPoolSize=MAX_CONCURRENT_DB_OPERATIONS * 2,  # Buffer для пиковых нагрузок
            minPoolSize=10,
            maxIdleTimeMS=30_000,  # 30 секунд idle timeout
            serverSelectionTimeoutMS=5000,
            connectTimeoutMS=10000,
            socketTimeoutMS=None,  # Нет timeout для длинных операций
        )
        # logger.info(f"MongoDB connection pool initialized with maxPoolSize={MAX_CONCURRENT_DB_OPERATIONS * 2}")
        
        # 2. Инициализируем HTTP client для внешних API
        self.http_client = httpx.AsyncClient(
            timeout=httpx.Timeout(30.0),  # 30 seconds timeout
            limits=httpx.Limits(max_connections=100, max_keepalive_connections=20),
            headers={"User-Agent": "Generia-AI-Worker/1.0"}
        )
        # logger.info("HTTP client initialized with connection pooling")
        
        # 3. Инициализируем gRPC channels (будут создаваться по мере необходимости)
        # Channels создаются lazy в service_client при обнаружении сервисов
        
        # 4. Создаем clients поверх connection pools
        await self._initialize_clients()
        
        # 5. Инициализируем schema mapping
        self._initialize_schema_mapping()
        
        # logger.info("SharedResourcesManager fully initialized")
    
    async def _initialize_clients(self):
        """Создает clients поверх connection pools"""
        from ..db import MongoDBManager
        from ..api import LLMClient, ImageGenerator, ServiceClient
        
        # DB manager с готовым connection pool
        self.db_manager = MongoDBManager(mongo_client=self.mongo_client)
        await self.db_manager.initialize()
        
        # Service client с Consul discovery
        self.service_client = ServiceClient(db_manager=self.db_manager)
        await self.service_client.initialize()
        
        # LLM client
        self.llm_client = LLMClient(
            db_manager=self.db_manager,
            http_client=self.http_client
        )
        
        # Image generator
        self.image_generator = ImageGenerator(
            db_manager=self.db_manager, 
            service_client=self.service_client,
            http_client=self.http_client
        )
        
        # logger.info("All clients initialized with shared connection pools")
    
    def _initialize_schema_mapping(self):
        """Инициализирует mapping схем для LLM"""
        try:
            from ..schemas.world_description import WorldDescriptionResponse
            from ..schemas.character_batch import CharacterBatchResponse
            from ..schemas.character import CharacterDetailResponse
            from ..schemas.post_batch import PostBatchResponse
            from ..schemas.post import PostDetailResponse
            from ..schemas.image_prompts import ImagePromptResponse
            from ..schemas.post_image import PostImagePromptResponse
            from ..schemas.character_avatar import CharacterAvatarPromptResponse

            CharacterAvatarPromptResponse
            
            self.schema_mapping = {
                'WorldDescriptionResponse': WorldDescriptionResponse,
                'CharacterBatchResponse': CharacterBatchResponse,
                'CharacterResponse': CharacterDetailResponse,
                'CharacterDetailResponse': CharacterDetailResponse,
                'PostBatchResponse': PostBatchResponse,
                'PostResponse': PostDetailResponse,
                'PostDetailResponse': PostDetailResponse,
                'ImagePromptResponse': ImagePromptResponse,
                'PostImagePromptResponse': PostImagePromptResponse,
                'CharacterAvatarPromptResponse': CharacterAvatarPromptResponse
            }
            # logger.info("Schema mapping initialized")
        except ImportError as e:
            logger.error(f"Failed to initialize schema mapping: {str(e)}")
    
    def get_schema_by_name(self, schema_name: str) -> Optional[Type[BaseModel]]:
        """Получает класс схемы по имени"""
        return self.schema_mapping.get(schema_name)
    
    async def create_grpc_channel(self, host: str, port: int) -> grpc.aio.Channel:
        """Создает или переиспользует gRPC channel"""
        channel_key = f"{host}:{port}"
        
        if channel_key not in self.grpc_channels:
            # Создаем новый channel с connection pooling
            self.grpc_channels[channel_key] = grpc.aio.insecure_channel(
                f"{host}:{port}",
                options=[
                    ("grpc.keepalive_time_ms", 60_000),  # Send keepalive every 60s
                    ("grpc.keepalive_timeout_ms", 5_000),  # Wait 5s for keepalive response
                    ("grpc.keepalive_permit_without_calls", True),
                    ("grpc.http2.max_pings_without_data", 0),
                    ("grpc.http2.min_time_between_pings_ms", 10_000),
                    ("grpc.http2.min_ping_interval_without_data_ms", 300_000),
                    ("grpc.max_connection_idle_ms", 60_000),
                    ("grpc.max_connection_age_ms", 300_000),
                ]
            )
            # logger.info(f"Created new gRPC channel for {channel_key}")
        
        return self.grpc_channels[channel_key]
    
    async def close(self):
        """Gracefully closes all connections"""
        # logger.info("Closing shared resources...")
        
        # Close HTTP client
        if self.http_client:
            await self.http_client.aclose()
            # logger.info("HTTP client closed")
        
        # Close gRPC channels
        for channel_key, channel in self.grpc_channels.items():
            try:
                await channel.close()
                # logger.info(f"gRPC channel {channel_key} closed")
            except Exception as e:
                logger.error(f"Error closing gRPC channel {channel_key}: {str(e)}")
        
        # Close service client
        if self.service_client:
            await self.service_client.close()
            # logger.info("Service client closed")
        
        # Close MongoDB client
        if self.mongo_client:
            self.mongo_client.close()
            # logger.info("MongoDB client closed")
        
        # logger.info("All shared resources closed")

