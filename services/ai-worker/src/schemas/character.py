from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class CharacterRelationship(BaseModel):
    """Схема для описания отношений с другими персонажами"""
    username: str = Field(..., description="Имя пользователя персонажа, с которым есть отношения")
    relationship_type: str = Field(..., description="Тип отношений (друг, враг, родственник и т.д.)")
    description: str = Field(..., description="Описание отношений с персонажем")

class CharacterDetailResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации детального описания персонажа"""
    username: str = Field(..., description="Уникальное имя пользователя (никнейм)")
    display_name: str = Field(..., description="Отображаемое имя персонажа")
    bio: str = Field(..., description="Биография/описание профиля (до 200 символов)")
    background_story: str = Field(..., description="Подробная история персонажа")
    personality: str = Field(..., description="Подробное описание личности персонажа")
    appearance: str = Field(..., description="Детальное описание рассы и внешности персонажа")
    interests: List[str] = Field(..., description="Интересы и хобби персонажа")
    speaking_style: str = Field(..., description="Стиль речи персонажа")
    common_topics: List[str] = Field(..., description="Темы, на которые персонаж часто общается")
    avatar_description: str = Field(..., description="Детальное описание для генерации аватара персонажа")
    avatar_style: str = Field(..., description="Стиль изображения для аватара (photorealistic, stylized, anime style и т.д.)")
    relationships: List[CharacterRelationship] = Field(..., description="Отношения с другими персонажами (если есть)")
    secret: str = Field(..., description="Секрет или скрытая черта персонажа")
    daily_routine: str = Field(..., description="Повседневная жизнь персонажа")