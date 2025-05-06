from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class CharacterBase(BaseModel):
    """Базовая информация о персонаже для создания пакета персонажей"""
    concept: str = Field(..., description="Концепция персонажа (2-4 предложения)")
    concept_short: str = Field(..., description="Краткое описание концепции персонажа (1 предложение)")
    id: str = Field(..., description="Идентификатор персонажа")
    role_in_world: str = Field(..., description="Роль персонажа в мире")
    posts_count: int = Field(..., description="Количество постов, которые создаст персонаж")
    personality_traits: List[str] = Field(..., description="Ключевые черты личности")
    interests: List[str] = Field(..., description="Интересы персонажа")

class CharacterConnection(BaseModel):
    """Схема для описания связи между персонажами"""
    character1_name: str = Field(..., description="Имя первого персонажа")
    character2_name: str = Field(..., description="Имя второго персонажа")
    connection_type: str = Field(..., description="Тип связи (семейная, дружеская, профессиональная и т.д.)")
    description: str = Field(..., description="Описание связи между персонажами")

class StoryAndEvent(BaseModel):
    """Схема для описания общих событий персонажей"""
    characters_ids: List[str] = Field(..., description="Идентификаторы персонажей, которые участвуют в истории")
    title: str = Field(..., description="Название события")
    story: str = Field(..., description="Описание совместных историй персонажей")
    location: str = Field(..., description="Место события")
    genre: str = Field(..., description="Жанр события")


class CharacterBatchResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации пакета персонажей"""
    characters: List[CharacterBase] = Field(..., description="Список базовых описаний персонажей")
    world_interpretation: str = Field(..., description="Общее понимание мира, отраженное в персонажах")
    character_connections: List[CharacterConnection] = Field(..., description="Связи между персонажами", default_factory=list)
    common_stories_and_events: List[StoryAndEvent] = Field(..., description="Общие истории и события, которые происходят с персонадами")
    generated_characters_description: str = Field("", description="Краткое описание сгенерированных персонажей (1-3 абзаца, про персонажей в общем, без деталей)")