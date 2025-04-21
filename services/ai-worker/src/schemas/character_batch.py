from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class CharacterBase(BaseModel):
    """Базовая информация о персонаже для создания пакета персонажей"""
    concept: str = Field(..., description="Концепция персонажа (2-4 предложения)")
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



class CharacterBatchResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации пакета персонажей"""
    characters: List[CharacterBase] = Field(..., description="Список базовых описаний персонажей")
    world_interpretation: str = Field(..., description="Общее понимание мира, отраженное в персонажах")
    character_connections: List[CharacterConnection] = Field(..., description="Связи между персонажами", default_factory=list)