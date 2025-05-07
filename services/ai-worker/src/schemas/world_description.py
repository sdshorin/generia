from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional
from datetime import datetime



class Location(BaseModel):
    """Схема для описания примечательного места"""
    name: str = Field(..., description="Название места")
    description: str = Field(..., description="Подробное описание места")
    significance: str = Field(..., description="Значимость места для мира")
    visual_style: str = Field(..., description="Визуальный стиль места")

class CharacterType(BaseModel):
    """Схема для описания социальной группы"""
    type: str = Field(..., description="Тип персонажей")
    description: str = Field(..., description="Описание типа персонажей")
    role: str = Field(..., description="Роль в обществе")
    characteristics: List[str] = Field(..., description="Список характерных черт")

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
        description="Дополнительные уникальные детали мира, которые не вписываются в основную структуру (не ограниченное количество)"
    )


class UserPreferences(BaseModel):
    """Схема для описания предпочтений пользователя"""
    language: str = Field(..., description="Язык, который предпочитает пользователь (на каком языке должен быть сгенерирован мир - английский, русский, итальянский, французский, и т.д.)")
    other_preferences: List[str] = Field(
        default_factory=list,
        description="Дополнительные предпочтения пользователя, которые не вписываются в основную структуру (Взяты из промпта пользователя, если есть)"
    )

class WorldDescriptionResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации описания мира"""
    name: str = Field(..., description="Название мира (уникальное, запоминающееся, отражающее его суть)")
    user_preferences: UserPreferences = Field(..., description="Предпочтения пользователя")
    description_short: str = Field(..., description="Краткое описание мира (2-3 предложения)")
    description: str = Field(..., description="Подробное описание мира (1-5 абзацев)")
    theme: str = Field(..., description="Основная тема мира")
    technology_level: str = Field(..., description="Уровень технологического развития мира")
    social_structure: str = Field(..., description="Социальная структура общества")
    culture: str = Field(..., description="Культурные особенности мира")
    geography: str = Field(..., description="Географические особенности мира")
    visual_style: str = Field(..., description="Визуальный стиль мира (цветовая палитра, художественный стиль)")
    history: str = Field(..., description="История мира")
    # notable_locations: List[Location] = Field(..., description="Список примечательных мест. Содержит как конкретные места, так и примеры типичных мест/ареалов. 3-10 или более мест")
    # typical_characters: List[CharacterType] = Field(..., description="Список из типов персонажей. Это описание должно содержать описание социальных групп и ролей в мире, 3-10 или более типов")
    common_activities: List[str] = Field(..., description="Список распространенных занятий и активностей в этом мире, 5-20 или более активностей")
    typical_stories: List[str] = Field(..., description="Типичные истории и сюжеты, которые происходят в этом мире. 5-20 или более историй")
    additional_details: AdditionalDetails = Field(
        ..., description="Дополнительные детали и особенности мира, включающие: климат, ресурсы, конфликты, традиции, технологии, систему магии, временной период, язык и дополнительные детали"
    )


class WorldDescription(WorldDescriptionResponse):
    """
    Расширенная схема для описания мира.
    Наследуется от WorldDescriptionResponse и добавляет поля для MongoDB.
    """
    # Поля для MongoDB
    id: Optional[str] = Field(None, alias="_id", description="ID мира (используется в MongoDB)")
    created_at: Optional[datetime] = Field(None, description="Время создания записи")
    updated_at: Optional[datetime] = Field(None, description="Время последнего обновления")

    class Config:
        populate_by_name = True
        arbitrary_types_allowed = True

    @classmethod
    def from_mongo(cls, data: Dict[str, Any]) -> "WorldDescription":
        """
        Создает экземпляр WorldDescription из данных MongoDB

        Args:
            data: Данные из MongoDB

        Returns:
            Экземпляр WorldDescription
        """
        if not data:
            return None

        # Преобразуем вложенные объекты из словарей в объекты Pydantic
        # if "notable_locations" in data and isinstance(data["notable_locations"], list):
        #     data["notable_locations"] = [
        #         Location(**loc) if isinstance(loc, dict) else loc
        #         for loc in data["notable_locations"]
        #     ]

        # if "typical_characters" in data and isinstance(data["typical_characters"], list):
        #     data["typical_characters"] = [
        #         CharacterType(**char) if isinstance(char, dict) else char
        #         for char in data["typical_characters"]
        #     ]

        if "additional_details" in data and isinstance(data["additional_details"], dict):
            data["additional_details"] = AdditionalDetails(**data["additional_details"])

        return cls(**data)

    def to_mongo(self, world_id: str) -> Dict[str, Any]:
        """
        Преобразует экземпляр WorldDescription в словарь для сохранения в MongoDB

        Args:
            world_id: ID мира

        Returns:
            Словарь для сохранения в MongoDB
        """
        now = datetime.utcnow()

        # Преобразуем модель в словарь
        data = self.model_dump()

        # Добавляем/обновляем поля для MongoDB
        data["_id"] = world_id
        data["created_at"] = self.created_at or now
        data["updated_at"] = now

        # Преобразуем вложенные объекты в словари
        # data["notable_locations"] = [loc.model_dump() for loc in self.notable_locations]
        # data["typical_characters"] = [char.model_dump() for char in self.typical_characters]
        data["additional_details"] = self.additional_details.model_dump()

        return data