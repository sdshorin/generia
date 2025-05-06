from typing import Dict, Any, List, Optional
from datetime import datetime
from pydantic import BaseModel, Field

class Task(BaseModel):
    """Model for storing task information"""
    id: str = Field(..., alias="_id")
    type: str
    world_id: str
    status: str
    worker_id: Optional[str] = None
    parameters: Dict[str, Any]
    result: Optional[Dict[str, Any]] = None
    created_at: datetime
    updated_at: datetime
    attempt_count: int = 0
    error: Optional[str] = None

    class Config:
        populate_by_name = True
        arbitrary_types_allowed = True


class StageInfo(BaseModel):
    """Information about generation stage status"""
    name: str
    status: str

class WorldGenerationStatus(BaseModel):
    """Model for storing world generation progress information"""
    id: str = Field(..., alias="_id")
    status: str
    current_stage: str
    stages: List[StageInfo]
    tasks_total: int
    tasks_completed: int
    tasks_failed: int
    task_predicted: int
    users_created: int
    posts_created: int
    users_predicted: int
    posts_predicted: int
    api_call_limits_LLM: int
    api_call_limits_images: int
    api_calls_made_LLM: int
    api_calls_made_images: int
    parameters: Dict[str, Any]
    created_at: datetime
    updated_at: datetime

    class Config:
        populate_by_name = True
        arbitrary_types_allowed = True


class ApiRequestHistory(BaseModel):
    """Model for storing API request history"""
    id: str = Field(..., alias="_id")
    api_type: str  # "llm" or "image"
    task_id: str
    world_id: str
    request_type: str  # Request type (e.g., API method name)
    request_data: Dict[str, Any]  # Request data
    response_data: Optional[Dict[str, Any]] = None  # Response data
    error: Optional[str] = None  # Error, if any
    duration_ms: int  # Request duration in milliseconds
    created_at: datetime

    class Config:
        populate_by_name = True
        arbitrary_types_allowed = True