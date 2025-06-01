"""
Utility для логирования в Temporal workflows и activities.
Автоматически определяет контекст и использует подходящий логгер.
"""

def get_workflow_logger():
    """
    Возвращает подходящий логгер в зависимости от контекста выполнения.
    
    В workflows использует workflow.logger, в activities - обычный logger.
    """
    try:
        from temporalio import workflow
        # Проверяем, находимся ли мы в контексте workflow
        if hasattr(workflow, 'info') and workflow.info() is not None:
            return workflow.logger
    except Exception:
        pass
    
    # Если мы не в workflow контексте, используем обычный логгер
    from .logger import get_logger
    return get_logger()


def log_info(message: str):
    """Удобная функция для info логирования"""
    get_workflow_logger().info(message)


def log_error(message: str):
    """Удобная функция для error логирования"""
    get_workflow_logger().error(message)


def log_warning(message: str):
    """Удобная функция для warning логирования"""
    get_workflow_logger().warning(message)


def log_debug(message: str):
    """Удобная функция для debug логирования"""
    get_workflow_logger().debug(message)