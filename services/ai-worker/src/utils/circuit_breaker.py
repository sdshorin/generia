import asyncio
import time
from functools import wraps
from typing import Callable, Any, TypeVar, Optional

from .logger import logger
from ..constants import CircuitBreakerState

T = TypeVar('T')

class CircuitBreaker:
    """
    Implementation of the Circuit Breaker pattern to protect against unavailable services
    """
    
    def __init__(
        self,
        name: str,
        failure_threshold: int = 5,
        recovery_timeout: float = 30.0,
        recovery_threshold: int = 2,
        timeout: float = 10.0
    ):
        """
        Initializes Circuit Breaker
        
        Args:
            name: Name to identify the circuit breaker
            failure_threshold: Number of errors after which the circuit breaker opens
            recovery_timeout: Time in seconds after which the circuit breaker transitions to half-open
            recovery_threshold: Number of successful requests required to return to closed state
            timeout: Timeout for function execution in seconds
        """
        self.name = name
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.recovery_threshold = recovery_threshold
        self.timeout = timeout
        
        self.state = CircuitBreakerState.CLOSED
        self.failure_count = 0
        self.success_count = 0
        self.last_failure_time = 0
        self._lock = asyncio.Lock()
    
    async def __call__(self, func: Callable[..., Any], *args: Any, **kwargs: Any) -> Any:
        """
        Executes a function protected by the circuit breaker.
        
        Args:
            func: Function to execute
            *args: Function arguments
            **kwargs: Named function arguments
            
        Returns:
            Function execution result
        
        Raises:
            Exception: If the circuit breaker is open or an error occurred during function execution
        """
        async with self._lock:
            if self.state == CircuitBreakerState.OPEN:
                if time.time() > self.last_failure_time + self.recovery_timeout:
                    # logger.info(f"Circuit Breaker '{self.name}' transitioning to HALF-OPEN state")
                    self.state = CircuitBreakerState.HALF_OPEN
                    self.success_count = 0
                else:
                    raise Exception(f"Circuit Breaker '{self.name}' is open")
        
        try:
            # Execute function with timeout
            result = await asyncio.wait_for(func(*args, **kwargs), timeout=self.timeout)
            
            async with self._lock:
                if self.state == CircuitBreakerState.HALF_OPEN:
                    self.success_count += 1
                    if self.success_count >= self.recovery_threshold:
                        # logger.info(f"Circuit Breaker '{self.name}' recovered, transitioning to CLOSED state")
                        self.state = CircuitBreakerState.CLOSED
                        self.failure_count = 0
                        self.success_count = 0
                elif self.state == CircuitBreakerState.CLOSED:
                    self.failure_count = 0
            
            return result
            
        except Exception as e:
            async with self._lock:
                self.last_failure_time = time.time()
                
                if self.state == CircuitBreakerState.CLOSED:
                    self.failure_count += 1
                    if self.failure_count >= self.failure_threshold:
                        # logger.warning(
                        #     f"Circuit Breaker '{self.name}' opened after {self.failure_count} errors"
                        # )
                        self.state = CircuitBreakerState.OPEN
                elif self.state == CircuitBreakerState.HALF_OPEN:
                    # logger.warning(f"Circuit Breaker '{self.name}' opened again after error in HALF-OPEN state")
                    self.state = CircuitBreakerState.OPEN
            
            # logger.error(f"Circuit Breaker '{self.name}' error: {str(e)}")
            raise e


def circuit_breaker(
    name: Optional[str] = None,
    failure_threshold: int = 5,
    recovery_timeout: float = 30.0,
    recovery_threshold: int = 2,
    timeout: float = 10.0
) -> Callable[[Callable[..., Any]], Callable[..., Any]]:
    """
    Decorator for applying the Circuit Breaker pattern to an asynchronous function
    
    Args:
        name: Name to identify the circuit breaker
        failure_threshold: Number of errors after which the circuit breaker opens
        recovery_timeout: Time in seconds after which the circuit breaker transitions to half-open
        recovery_threshold: Number of successful requests required to return to closed state
        timeout: Timeout for function execution in seconds
    
    Returns:
        Decorated function
    """
    breakers = {}
    
    def decorator(func: Callable[..., Any]) -> Callable[..., Any]:
        breaker_name = name or func.__name__
        
        if breaker_name not in breakers:
            breakers[breaker_name] = CircuitBreaker(
                breaker_name,
                failure_threshold,
                recovery_timeout,
                recovery_threshold,
                timeout
            )
        
        @wraps(func)
        async def wrapper(*args: Any, **kwargs: Any) -> Any:
            breaker = breakers[breaker_name]
            return await breaker(func, *args, **kwargs)
        
        return wrapper
    
    return decorator