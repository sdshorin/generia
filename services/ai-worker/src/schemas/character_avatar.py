from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class CharacterAvatarPromptResponse(BaseModel):
    """Schema for structured response from LLM when generating character avatar prompt"""
    prompt: str = Field(..., description="The optimized prompt for generating the character avatar image")
