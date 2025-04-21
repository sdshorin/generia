import asyncio
import random
from functools import wraps
from typing import Callable, TypeVar, Any, List, Optional, Type

from .logger import logger

T = TypeVar('T')

async def retry_async(
    func: Callable[..., Any],
    max_retries: int = 3,
    retry_exceptions: List[Type[Exception]] = None,
    initial_delay: float = 1.0,
    max_delay: float = 60.0,
    backoff_factor: float = 2.0,
    *args: Any,
    **kwargs: Any
) -> Any:
    """
    Executes an asynchronous function with retries on error
    
    Args:
        func: Function to execute
        max_retries: Maximum number of retry attempts
        retry_exceptions: List of exceptions that should trigger retries
        initial_delay: Initial delay before retry in seconds
        max_delay: Maximum delay between attempts in seconds
        backoff_factor: Factor by which to increase the delay
        *args: Function arguments
        **kwargs: Named function arguments
        
    Returns:
        Function execution result
        
    Raises:
        Exception: The last exception after exhausting all attempts
    """
    retry_exceptions = retry_exceptions or [Exception]
    delay = initial_delay
    
    for attempt in range(max_retries + 1):
        try:
            return await func(*args, **kwargs)
        except tuple(retry_exceptions) as e:
            if attempt >= max_retries:
                logger.error(f"All attempts exhausted ({max_retries + 1}). Last error: {str(e)}")
                raise e
            
            # Add random jitter (±10%)
            jitter = random.uniform(0.9, 1.1)
            sleep_time = min(delay * jitter, max_delay)
            
            logger.warning(
                f"Attempt {attempt + 1}/{max_retries + 1} failed: {str(e)}. "
                f"Retrying in {sleep_time:.2f} sec."
            )
            
            await asyncio.sleep(sleep_time)
            delay = min(delay * backoff_factor, max_delay)

def with_retries(
    max_retries: int = 3,
    retry_exceptions: Optional[List[Type[Exception]]] = None,
    initial_delay: float = 1.0,
    max_delay: float = 60.0,
    backoff_factor: float = 2.0
) -> Callable[[Callable[..., Any]], Callable[..., Any]]:
    """
    Decorator for executing an asynchronous function with retries on error
    
    Args:
        max_retries: Maximum number of retry attempts
        retry_exceptions: List of exceptions that should trigger retries
        initial_delay: Initial delay before retry in seconds
        max_delay: Maximum delay between attempts in seconds
        backoff_factor: Factor by which to increase the delay
        
    Returns:
        Decorated function
    """
    def decorator(func: Callable[..., Any]) -> Callable[..., Any]:
        @wraps(func)
        async def wrapper(*args: Any, **kwargs: Any) -> Any:
            return await retry_async(
                func,
                max_retries,
                retry_exceptions,
                initial_delay,
                max_delay,
                backoff_factor,
                *args,
                **kwargs
            )
        return wrapper
    return decorator