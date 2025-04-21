from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class ImagePromptResponse(BaseModel):
    """Схема для структурированного ответа от LLM для промптов изображений"""
    header_prompt: str = Field(..., description="Промпт для большого фонового изображения (хэдера) мира")
    icon_prompt: str = Field(..., description="Промпт для иконки мира")
    style_reference: str = Field(..., description="Описание стиля для согласованности изображений")
    visual_elements: List[str] = Field(..., description="Ключевые визуальные элементы для изображений")
    mood: str = Field(..., description="Настроение и атмосфера, которую должны передавать изображения")
    color_palette: List[str] = Field(..., description="Основные цвета для использования в изображениях")