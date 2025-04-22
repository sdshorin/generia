import os
import aiohttp
import asyncio
from typing import Optional, Dict, List
from ..utils.logger import logger
from ..utils.retries import with_retries

class ConsulServiceDiscovery:
    """
    Client for service discovery using Consul.
    Allows finding service addresses dynamically.
    """
    
    def __init__(self, consul_address: str = None):
        """
        Initialize Consul service discovery client
        
        Args:
            consul_address: Address of Consul server (defaults to CONSUL_ADDRESS env var)
        """
        self.consul_address = consul_address or os.getenv("CONSUL_ADDRESS", "consul:8500")
        self.session = None
        self.service_cache = {}
        self.cache_ttl = 30  # Cache service addresses for 30 seconds
        self.cache_timestamps = {}
    
    async def initialize(self):
        """Initialize HTTP session"""
        if self.session is None:
            self.session = aiohttp.ClientSession()
        return self
    
    async def close(self):
        """Close HTTP session"""
        if self.session:
            await self.session.close()
            self.session = None
    
    def _get_consul_url(self, path: str) -> str:
        """Generate Consul API URL for the given path"""
        return f"http://{self.consul_address}/v1{path}"
    
    async def _ensure_session(self):
        """Ensure HTTP session is initialized"""
        if self.session is None:
            await self.initialize()
    
    @with_retries(max_retries=3)
    async def resolve_service(self, service_name: str) -> str:
        """
        Resolve a service address by name
        
        Args:
            service_name: Name of the service to resolve
            
        Returns:
            The resolved address in format "host:port"
        """
        # Check cache first
        current_time = asyncio.get_event_loop().time()
        if service_name in self.service_cache and (
            current_time - self.cache_timestamps.get(service_name, 0) < self.cache_ttl
        ):
            logger.info(f"Using cached address for service {service_name}: {self.service_cache[service_name]}")
            return self.service_cache[service_name]
        
        await self._ensure_session()
        
        # Query Consul for healthy service instances
        url = self._get_consul_url(f"/health/service/{service_name}")
        params = {"passing": "true"}
        
        try:
            async with self.session.get(url, params=params) as response:
                if response.status != 200:
                    error_text = await response.text()
                    logger.error(f"Consul API error: {response.status} - {error_text}")
                    raise Exception(f"Failed to query Consul: {response.status}")
                
                services = await response.json()
                
                if not services:
                    # Try to get unhealthy services too
                    async with self.session.get(self._get_consul_url(f"/health/service/{service_name}")) as fallback_response:
                        all_services = await fallback_response.json()
                        
                        if not all_services:
                            logger.error(f"No instances of service {service_name} found (healthy or unhealthy)")
                            raise Exception(f"No instances of service {service_name} found")
                        
                        logger.warning(f"No healthy instances found for service {service_name}, using DNS resolution")
                        service_port = all_services[0]["Service"]["Port"]
                        address = f"{service_name}:{service_port}"
                        self._update_cache(service_name, address)
                        return address
                
                # Get the first healthy service
                service = services[0]["Service"]
                service_address = service["Address"]
                service_port = service["Port"]
                
                # If address is empty, fall back to service name (Docker DNS)
                if not service_address:
                    service_address = service_name
                    logger.warning(f"Empty service address detected for {service_name}, falling back to service name")
                
                address = f"{service_address}:{service_port}"
                logger.info(f"Resolved service {service_name} to {address}")
                
                self._update_cache(service_name, address)
                return address
        
        except aiohttp.ClientError as e:
            logger.error(f"HTTP error when contacting Consul: {str(e)}")
            # Fall back to DNS-based service discovery
            logger.warning(f"Falling back to DNS-based discovery for {service_name}")
            fallback_address = f"{service_name}:50051"  # Default gRPC port
            self._update_cache(service_name, fallback_address)
            return fallback_address
    
    def _update_cache(self, service_name: str, address: str):
        """Update the service cache"""
        self.service_cache[service_name] = address
        self.cache_timestamps[service_name] = asyncio.get_event_loop().time()