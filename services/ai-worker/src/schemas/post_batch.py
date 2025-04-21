from pydantic import BaseModel, Field
from typing import Dict, Any, List, Optional

class PostConcept(BaseModel):
    """Концепция поста для пакетной генерации"""
    topic: str = Field(..., description="Тема поста")
    content_brief: str = Field(..., description="Краткое описание содержания (2-3 предложения)")
    has_image: bool = Field(..., description="Нужно ли генерировать изображение для этого поста")
    emotional_tone: str = Field(..., description="Эмоциональный тон поста")
    post_type: str = Field(..., description="Тип поста (личное, новость, вопрос и т.д.)")
    relevance_to_character: str = Field(..., description="Как пост отражает личность персонажа")

class PostBatchResponse(BaseModel):
    """Схема для структурированного ответа от LLM при генерации пакета постов для персонажа"""
    posts: List[PostConcept] = Field(..., description="Список концепций постов")
    narrative_arc: str = Field(..., description="Общая сюжетная арка всех постов")
    character_development: str = Field(..., description="Как посты отражают развитие персонажа")
    posting_schedule: List[str] = Field(..., description="Примерный график публикации постов")
    recurring_themes: List[str] = Field(..., description="Повторяющиеся темы в постах персонажа")