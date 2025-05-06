# Схемы для структурированного вывода LLM
from pydantic import BaseModel, Field, model_validator
from typing import Dict, Any, List, Optional
import json
import os
import importlib.resources as pkg_resources

def load_schema_file(file_name: str) -> Dict[str, Any]:
    """Загружает схему из JSON файла"""
    try:
        schemas_dir = os.path.dirname(os.path.abspath(__file__))
        file_path = os.path.join(schemas_dir, file_name)
        with open(file_path, 'r') as f:
            return json.load(f)
    except Exception as e:
        raise ValueError(f"Failed to load schema {file_name}: {str(e)}")

# Импорт схем
from .world_description import WorldDescription, WorldDescriptionResponse
from .image_prompts import ImagePromptResponse
from .character_batch import CharacterBatchResponse
from .character import CharacterDetailResponse
from .post_batch import PostBatchResponse
from .post import PostDetailResponse