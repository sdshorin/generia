from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class PostImagePromptResponse(BaseModel):
    """Schema for structured response from LLM when generating post image prompt"""
    prompt: str = Field(..., description="The optimized prompt for generating the post image")
