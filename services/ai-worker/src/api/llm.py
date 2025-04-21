import asyncio
import time
import uuid
import json
import requests
from typing import Dict, Any, Optional, Type, List, TypeVar, Union
from datetime import datetime

from pydantic import BaseModel

from ..config import OPENROUTER_API_KEY
from ..utils.logger import logger
from ..utils.circuit_breaker import circuit_breaker
from ..utils.retries import with_retries
from ..db.models import ApiRequestHistory

T = TypeVar('T')

class LLMClient:
    """
    Client for working with LLM API via OpenRouter
    """
    
    def __init__(self, api_key: str = OPENROUTER_API_KEY, db_manager=None):
        """
        Initializes LLM client
        
        Args:
            api_key: API key for OpenRouter
            db_manager: Optional database manager for request logging
        """
        self.api_key = api_key
        self.db_manager = db_manager
        self.semaphore = asyncio.Semaphore(15)  # Limit the number of concurrent requests
    
    @circuit_breaker(name="llm_content", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def generate_content(
        self,
        prompt: str,
        model: str = "google/gemini-flash-1.5-8b",
        temperature: float = 0.7,
        max_output_tokens: int = 1024,
        task_id: Optional[str] = None,
        world_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Generates text content using LLM
        
        Args:
            prompt: Prompt for generation
            model: Model to use
            temperature: Temperature (creativity) of generation
            max_output_tokens: Maximum number of tokens in the response
            task_id: Task ID for logging
            world_id: World ID for logging
            
        Returns:
            Dictionary with API response
        """
        async with self.semaphore:
            start_time = time.time()
            request_id = str(uuid.uuid4())
            
            try:
                headers = {
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                }
                
                data = {
                    "model": model,
                    "messages": [
                        {"role": "user", "content": prompt},
                    ],
                    "temperature": temperature,
                    "max_tokens": max_output_tokens,
                }
                
                # For request logging
                request_data = {
                    "prompt": prompt,
                    "model": model,
                    "temperature": temperature,
                    "max_output_tokens": max_output_tokens,
                }
                
                # Execute request in a separate thread to avoid blocking
                loop = asyncio.get_event_loop()
                response = await loop.run_in_executor(
                    None,
                    lambda: requests.post(
                        "https://openrouter.ai/api/v1/chat/completions",
                        headers=headers,
                        json=data,
                    )
                )
                
                # Логируем статус ответа и заголовки
                logger.info(f"OpenRouter API response status: {response.status_code}")
                logger.info(f"OpenRouter API response headers: {dict(response.headers)}")
                
                if response.status_code != 200:
                    error_text = response.text
                    logger.error(f"OpenRouter API error response: {error_text}")
                    raise Exception(f"OpenRouter API error: {response.status_code} - {error_text}")
                
                response_data = response.json()
                
                # Логируем полный ответ от API
                logger.info(f"OpenRouter API full response: {json.dumps(response_data, indent=2)}")
                
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Проверяем наличие необходимых полей в ответе
                if "choices" not in response_data:
                    logger.error(f"Unexpected response format from OpenRouter API: {json.dumps(response_data, indent=2)}")
                    raise Exception(f"Unexpected response format from OpenRouter API: missing 'choices' field")
                
                if not response_data["choices"]:
                    logger.error(f"Empty choices array in OpenRouter API response: {json.dumps(response_data, indent=2)}")
                    raise Exception("Empty choices array in OpenRouter API response")
                
                if "message" not in response_data["choices"][0]:
                    logger.error(f"Unexpected response format from OpenRouter API: {json.dumps(response_data, indent=2)}")
                    raise Exception(f"Unexpected response format from OpenRouter API: missing 'message' field in first choice")
                
                if "content" not in response_data["choices"][0]["message"]:
                    logger.error(f"Unexpected response format from OpenRouter API: {json.dumps(response_data, indent=2)}")
                    raise Exception(f"Unexpected response format from OpenRouter API: missing 'content' field in message")
                
                # Get and parse JSON response
                content = response_data["choices"][0]["message"]["content"]
                
                # Логируем полученный контент
                logger.info(f"OpenRouter API content: {content}")
                
                # Prepare response for logging
                result = {
                    "text": content,
                    "model": model,
                    "finish_reason": response_data["choices"][0].get("finish_reason", "unknown"),
                }
                
                # Log request if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="llm",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="generate_content",
                        request_data=request_data,
                        response_data=result,
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.info(
                    f"LLM API call completed in {duration_ms}ms. "
                    f"Model: {model}, TaskID: {task_id or 'manual'}"
                )
                
                return result
                
            except Exception as e:
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Log error if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="llm",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="generate_content",
                        request_data=request_data,
                        error=str(e),
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.error(f"LLM API error: {str(e)}")
                raise
    
    @circuit_breaker(name="llm_structured", failure_threshold=3, recovery_timeout=60.0)
    @with_retries(max_retries=2)
    async def generate_structured_content(
        self,
        prompt: str,
        response_schema: Union[Type[BaseModel], List[Type[BaseModel]], Dict[str, Any]],
        # model: str = "openai/gpt-4o-mini",
        model: str = "google/gemini-flash-1.5-8b",
        temperature: float = 0.2,
        max_output_tokens: int = 2048,
        task_id: Optional[str] = None,
        world_id: Optional[str] = None
    ) -> Any:
        """
        Generates structured content using LLM
        
        Args:
            prompt: Prompt for generation
            response_schema: Pydantic schema or JSON schema dictionary for structured output
            model: Model to use
            temperature: Temperature (creativity) of generation
            max_output_tokens: Maximum number of tokens in the response
            task_id: Task ID for logging
            world_id: World ID for logging
            
        Returns:
            Object matching the response_schema
        """
        async with self.semaphore:
            start_time = time.time()
            request_id = str(uuid.uuid4())
            
            try:
                # Prepare JSON schema for request
                if isinstance(response_schema, dict):
                    # If JSON schema dictionary is passed
                    schema_dict = response_schema
                    schema_str = json.dumps(response_schema)
                else:
                    # If Pydantic schema is passed
                    schema_dict = response_schema.model_json_schema()
                    schema_str = str(response_schema)
                
                # Функция для замены всех $ref на соответствующие определения из $defs
                def replace_refs_with_defs(schema):
                    if isinstance(schema, dict):
                        if "$ref" in schema:
                            # Получаем имя модели из ссылки
                            ref_name = schema["$ref"].split("/")[-1]
                            # Находим определение модели в $defs
                            if "$defs" in schema_dict and ref_name in schema_dict["$defs"]:
                                # Заменяем $ref на содержимое определения
                                ref_content = schema_dict["$defs"][ref_name].copy()
                                schema.clear()
                                schema.update(ref_content)
                        for key, value in schema.items():
                            if isinstance(value, dict):
                                replace_refs_with_defs(value)
                            elif isinstance(value, list):
                                for item in value:
                                    if isinstance(item, dict):
                                        replace_refs_with_defs(item)
                    return schema
                
                # Заменяем все $ref на определения
                schema_dict = replace_refs_with_defs(schema_dict)
                
                # Упрощаем allOf конструкции
                def simplify_allof(schema):
                    if isinstance(schema, dict):
                        if "allOf" in schema and len(schema["allOf"]) == 1:
                            # Если allOf содержит только один элемент, заменяем на него
                            allof_content = schema["allOf"][0]
                            schema.clear()
                            schema.update(allof_content)
                        for key, value in schema.items():
                            if isinstance(value, dict):
                                simplify_allof(value)
                            elif isinstance(value, list):
                                for item in value:
                                    if isinstance(item, dict):
                                        simplify_allof(item)
                    return schema

                schema_dict = simplify_allof(schema_dict)
                
                # Удаляем $defs, так как все ссылки уже заменены
                if "$defs" in schema_dict:
                    del schema_dict["$defs"]
                
                # Ensure additionalProperties is set to false for strict validation
                if isinstance(schema_dict, dict):
                    # Add additionalProperties: false to the root schema
                    schema_dict["additionalProperties"] = False
                    
                    # Also add it to all object properties
                    def add_additional_properties_false(schema):
                        if isinstance(schema, dict):
                            if schema.get("type") == "object":
                                # Добавляем additionalProperties: false только если есть properties
                                if "properties" in schema and schema["properties"]:
                                    schema["additionalProperties"] = False
                            for key, value in schema.items():
                                if isinstance(value, dict):
                                    add_additional_properties_false(value)
                                elif isinstance(value, list):
                                    for item in value:
                                        if isinstance(item, dict):
                                            add_additional_properties_false(item)
                    
                    add_additional_properties_false(schema_dict)
                    
                    # Добавляем массив required, включающий все ключи из properties
                    if "properties" in schema_dict:
                        schema_dict["required"] = list(schema_dict["properties"].keys())
                        
                        # Рекурсивно добавляем required для вложенных объектов
                        def add_required_to_objects(schema):
                            if isinstance(schema, dict):
                                if schema.get("type") == "object" and "properties" in schema:
                                    schema["required"] = list(schema["properties"].keys())
                                for key, value in schema.items():
                                    if isinstance(value, dict):
                                        add_required_to_objects(value)
                                    elif isinstance(value, list):
                                        for item in value:
                                            if isinstance(item, dict):
                                                add_required_to_objects(item)
                        
                        add_required_to_objects(schema_dict)
                
                logger.info(f"schema_dict: {schema_dict}")
                headers = {
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json",
                }
                
                data = {
                    "model": model,
                    "messages": [
                        {"role": "user", "content": prompt},
                    ],
                    "temperature": temperature,
                    "max_tokens": max_output_tokens,
                    "response_format": {
                        "type": "json_schema",
                        "json_schema": {
                            "name": "response",
                            "strict": True,
                            "schema": schema_dict
                        },
                    },
                }
                
                # For request logging
                request_data = {
                    "prompt": prompt,
                    "model": model,
                    "temperature": temperature,
                    "max_output_tokens": max_output_tokens,
                    "response_schema": schema_str
                }
                
                # Execute request in a separate thread to avoid blocking
                loop = asyncio.get_event_loop()
                response = await loop.run_in_executor(
                    None,
                    lambda: requests.post(
                        "https://openrouter.ai/api/v1/chat/completions",
                        headers=headers,
                        json=data,
                    )
                )
                
                # Логируем статус ответа и заголовки
                logger.info(f"OpenRouter API response status: {response.status_code}")
                # logger.info(f"OpenRouter API response headers: {dict(response.headers)}")
                
                if response.status_code != 200:
                    error_text = response.text
                    # logger.error(f"OpenRouter API error response: {error_text}")
                    raise Exception(f"OpenRouter API error: {response.status_code} - {error_text}")
                
                response_data = response.json()
                
                # Логируем полный ответ от API
                logger.info(f"OpenRouter API full response: {json.dumps(response_data, indent=2)}")
                
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Проверяем наличие необходимых полей в ответе
                # if "choices" not in response_data:
                #     logger.error(f"Unexpected response format from OpenRouter API: {json.dumps(response_data, indent=2)}")
                #     raise Exception(f"Unexpected response format from OpenRouter API: missing 'choices' field")
                
                # if not response_data["choices"]:
                #     logger.error(f"Empty choices array in OpenRouter API response: {json.dumps(response_data, indent=2)}")
                #     raise Exception("Empty choices array in OpenRouter API response")
                
                # if "message" not in response_data["choices"][0]:
                #     logger.error(f"Unexpected response format from OpenRouter API: {json.dumps(response_data, indent=2)}")
                #     raise Exception(f"Unexpected response format from OpenRouter API: missing 'message' field in first choice")
                
                # if "content" not in response_data["choices"][0]["message"]:
                #     logger.error(f"Unexpected response format from OpenRouter API: {json.dumps(response_data, indent=2)}")
                #     raise Exception(f"Unexpected response format from OpenRouter API: missing 'content' field in message")
                
                # Get and parse JSON response
                content = response_data["choices"][0]["message"]["content"]
                
                # Логируем полученный контент
                # logger.info(f"OpenRouter API content: {content}")
                
                try:
                    structured_response = json.loads(content)
                    logger.info(f"Parsed structured response: {json.dumps(structured_response, indent=2)}")
                    
                    # Преобразуем словарь в объект Pydantic, если передан класс схемы
                    if not isinstance(response_schema, dict):
                        structured_response = response_schema.model_validate(structured_response)
                except json.JSONDecodeError as e:
                    logger.error(f"Failed to parse JSON content: {content}")
                    logger.error(f"JSON parse error: {str(e)}")
                    raise Exception(f"Failed to parse JSON content from OpenRouter API: {str(e)}")
                except Exception as e:
                    logger.error(f"Failed to validate response against schema: {str(e)}")
                    raise Exception(f"Failed to validate response against schema: {str(e)}")
                
                # Prepare response for logging
                result = {
                    "model": model,
                    "structured_data": content
                }
                
                # Log request if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="llm",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="generate_structured_content",
                        request_data=request_data,
                        response_data=result,
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.info(
                    f"LLM API structured call completed in {duration_ms}ms. "
                    f"Model: {model}, TaskID: {task_id or 'manual'}"
                )
                
                return structured_response
                
            except Exception as e:
                duration_ms = int((time.time() - start_time) * 1000)
                
                # Log error if db_manager is available
                if self.db_manager:
                    log_entry = ApiRequestHistory(
                        id=request_id,
                        api_type="llm",
                        task_id=task_id or "manual",
                        world_id=world_id or "unknown",
                        request_type="generate_structured_content",
                        request_data=request_data,
                        error=str(e),
                        duration_ms=duration_ms,
                        created_at=datetime.utcnow()
                    )
                    await self.db_manager.log_api_request(log_entry)
                
                logger.error(f"LLM API structured error: {str(e)}")
                raise