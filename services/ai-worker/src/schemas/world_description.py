from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class Location(BaseModel):
    """Схема для описания примечательного места"""
    name: str = Field(..., description="Название места")
    description: str = Field(..., description="Описание места")
    significance: str = Field(..., description="Значимость места для мира")
    visual_style: str = Field(..., description="Визуальный стиль места")

class CharacterType(BaseModel):
    """Схема для описания типа персонажа"""
    type: str = Field(..., description="Тип персонажа")
    description: str = Field(..., description="Описание типа персонажа")
    role: str = Field(..., description="Роль в обществе")
    characteristics: List[str] = Field(..., description="Характерные черты")

class AdditionalDetails(BaseModel):
    """Схема для дополнительных деталей мира"""
    climate: str = Field(..., description="Климат и погодные условия")
    resources: str = Field(..., description="Основные ресурсы и их распределение")
    conflicts: str = Field(..., description="Основные конфликты и противоречия")
    traditions: str = Field(..., description="Важные традиции и обычаи")
    technology: str = Field(..., description="Особенности технологий и их использования")
    magic_system: str = Field(description="Система магии, если применимо")
    time_period: str = Field(..., description="Временной период мира")
    language: str = Field(..., description="Особенности языка и коммуникации")
    custom_details: List[str] = Field(
        default_factory=list,
        description="Дополнительные уникальные детали мира, которые не вписываются в основную структуру"
    )

class WorldDescriptionResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации описания мира"""
    name: str = Field(..., description="Название мира")
    description: str = Field(..., description="Краткое описание мира (2-3 предложения)")
    theme: str = Field(..., description="Основная тема мира")
    technology_level: str = Field(..., description="Уровень технологического развития мира")
    social_structure: str = Field(..., description="Социальная структура общества")
    culture: str = Field(..., description="Культурные особенности мира")
    geography: str = Field(..., description="Географические особенности мира")
    visual_style: str = Field(..., description="Визуальный стиль мира (цветовая палитра, художественный стиль)")
    history: str = Field(..., description="Краткая история мира")
    notable_locations: List[Location] = Field(..., description="Список примечательных мест с названиями и описаниями")
    typical_characters: List[CharacterType] = Field(..., description="Типы персонажей, которые населяют мир")
    common_activities: List[str] = Field(..., description="Распространенные занятия и активности в этом мире")
    additional_details: AdditionalDetails = Field(
        ..., description="Дополнительные детали и особенности мира"
    )