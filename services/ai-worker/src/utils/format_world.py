"""
Утилиты для форматирования описания мира и других объектов для промптов
"""
import re
from typing import Any, Dict, List, Type
from temporalio import workflow
with workflow.unsafe.imports_passed_through():
    from pydantic import BaseModel

from ..schemas.world_description import WorldDescription, WorldDescriptionResponse


def _clean_description(description: str) -> str:
    """
    Очищает описание поля от текста в скобках, полностью удаляя его

    Args:
        description: Исходное описание поля

    Returns:
        Очищенное описание
    """
    # Полностью удаляем текст в круглых скобках
    return re.sub(r'\([^)]*\)', '', description)


def format_model_to_text(model_instance: BaseModel, model_class: Type[BaseModel] = None) -> str:
    """
    Форматирует экземпляр модели для использования в промптах,
    используя описания полей из модели

    Args:
        model_instance: Экземпляр модели для форматирования
        model_class: Класс модели (если отличается от класса экземпляра)

    Returns:
        Отформатированный текст для вставки в промпт
    """
    # Если класс модели не указан, используем класс экземпляра
    if model_class is None:
        model_class = model_instance.__class__

    # Получаем описания полей из модели
    model_fields = model_class.model_fields

    # Форматируем поля
    formatted_fields = []

    # Обрабатываем все поля модели
    for field_name, field_info in model_fields.items():
        # Пропускаем служебные поля MongoDB
        if field_name in ['id', 'created_at', 'updated_at']:
            continue

        # Получаем значение поля
        field_value = getattr(model_instance, field_name, None)

        # Пропускаем пустые значения
        if field_value is None:
            continue

        # Получаем и очищаем описание поля
        field_desc = _clean_description(field_info.description)

        # Форматируем значение в зависимости от типа
        if isinstance(field_value, list):
            if field_value and isinstance(field_value[0], BaseModel):
                # Для списка моделей форматируем каждый элемент
                nested_items = []
                for item in field_value:
                    nested_text = format_model_to_text(item)
                    nested_items.append(f"- {nested_text}")

                formatted_fields.append(f"{field_desc}:\n" + "\n".join(nested_items))
            else:
                # Для простых списков объединяем элементы
                formatted_fields.append(f"{field_desc}: {', '.join(str(item) for item in field_value)}")
        elif isinstance(field_value, BaseModel):
            # Для вложенных моделей рекурсивно форматируем
            nested_text = format_model_to_text(field_value)
            formatted_fields.append(f"{field_desc}:\n{nested_text}")
        else:
            # Для простых типов просто выводим значение
            formatted_fields.append(f"{field_desc}: {field_value}")

    # Объединяем все в одну строку
    return '\n'.join(formatted_fields)


def format_world_description(world_params: WorldDescription) -> str:
    """
    Форматирует описание мира для использования в промптах,
    используя описания полей из модели

    Args:
        world_params: Параметры мира

    Returns:
        Отформатированное описание мира для вставки в промпт
    """
    return str(world_params) # temp
    return format_model_to_text(world_params, WorldDescriptionResponse)
