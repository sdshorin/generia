from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class PostDetailResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации детального поста"""
    content: str = Field(..., description="Полный текст поста")
    image_prompt: Optional[str] = Field(None, description="Промпт для генерации изображения к посту")
    image_style: Optional[str] = Field(None, description="Стиль изображения для поста")
    hashtags: List[str] = Field(..., description="Хэштеги для поста")
    mood: str = Field(..., description="Настроение поста")
    context: str = Field(..., description="Контекст написания поста (что происходило с персонажем)")
    mentions: List[str] = Field(default=[], description="Упоминания других персонажей (если есть)")
    location: Optional[str] = Field(None, description="Место, где был создан пост (если применимо)")
    time_of_day: Optional[str] = Field(None, description="Время суток создания поста (если применимо)")