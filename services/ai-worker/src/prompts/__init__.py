# Промпты для LLM
import os
import json

def load_prompt(filename: str) -> str:
    """Загружает промпт из файла"""
    prompts_dir = os.path.dirname(os.path.abspath(__file__))
    file_path = os.path.join(prompts_dir, filename)

    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            return f.read()
    except Exception as e:
        raise ValueError(f"Failed to load prompt {filename}: {str(e)}")

# Константы для промптов
WORLD_DESCRIPTION_PROMPT = 'world_description.txt'
WORLD_IMAGE_PROMPT = 'world_image.txt'
CHARACTER_BATCH_PROMPT = 'character_batch.txt'
CHARACTER_DETAIL_PROMPT = 'character_detail.txt'
CHARACTER_AVATAR_PROMPT = 'character_avatar.txt'
PREVIOUS_CHARACTERS_PROMPT = 'previous_characters.txt'
FIRST_BATCH_CHARACTERS_PROMPT = 'first_batch_characters.txt'
PREVIOUS_POSTS_PROMPT = 'previous_posts.txt'
FIRST_BATCH_POSTS_PROMPT = 'first_batch_posts.txt'
POST_BATCH_PROMPT = 'post_batch.txt'
POST_DETAIL_PROMPT = 'post_detail.txt'
POST_IMAGE_PROMPT = 'post_image.txt'