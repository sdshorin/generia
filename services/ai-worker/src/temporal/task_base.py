"""
Базовые классы для работы с задачами в Temporal workflows
"""
from temporalio import workflow
with workflow.unsafe.imports_passed_through():
    from pydantic import BaseModel, Field
from typing import Optional, Type, TypeVar, Any, Dict

# Generic type для TaskInput классов
T = TypeVar('T', bound='TaskInput')


class TaskInput(BaseModel):
    """
    Базовый класс для всех Input данных workflow'ов
    Содержит полные данные задачи
    """
    
    class Config:
        arbitrary_types_allowed = True
        # Разрешаем дополнительные поля для совместимости
        extra = "allow"


class TaskRef(BaseModel):
    """
    Класс для передачи только task_id между workflow'ами
    """
    task_id: str
    
    class Config:
        arbitrary_types_allowed = True


def get_task_type_from_class(input_class: Type[TaskInput]) -> str:
    """
    Получает тип задачи из имени класса
    
    Args:
        input_class: Класс Input данных
        
    Returns:
        Строка с типом задачи для MongoDB
    """
    class_name = input_class.__name__
    
    # Убираем суффикс "Input" и конвертируем в snake_case
    if class_name.endswith('Input'):
        class_name = class_name[:-5]  # Убираем "Input"
    
    # Конвертируем CamelCase в snake_case
    import re
    task_type = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', class_name)
    task_type = re.sub('([a-z0-9])([A-Z])', r'\1_\2', task_type).lower()
    
    return task_type