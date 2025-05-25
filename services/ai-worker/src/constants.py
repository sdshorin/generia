# Task statuses
class TaskStatus:
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"

# Task types
class TaskType:
    INIT_WORLD_CREATION = "init_world_creation"
    GENERATE_WORLD_DESCRIPTION = "generate_world_description"
    GENERATE_WORLD_IMAGE = "generate_world_image"
    GENERATE_CHARACTER_BATCH = "generate_character_batch"
    GENERATE_CHARACTER = "generate_character"
    GENERATE_CHARACTER_AVATAR = "generate_character_avatar"
    GENERATE_POST_BATCH = "generate_post_batch"
    GENERATE_POST = "generate_post"
    GENERATE_POST_IMAGE = "generate_post_image"

# World generation statuses
class GenerationStatus:
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"

# World generation stages
class GenerationStage:
    INITIALIZING = "initializing"
    WORLD_DESCRIPTION = "world_description"
    WORLD_IMAGE = "world_image"
    CHARACTERS = "characters"
    POSTS = "posts"
    FINISHING = "finishing"

# MongoDB collections
class Collections:
    TASKS = "tasks"
    WORLD_GENERATION_STATUS = "world_generation_status"
    WORLD_PARAMETERS = "world_parameters"
    API_REQUESTS_HISTORY = "api_requests_history"

# Kafka event names
class KafkaEvents:
    TASK_CREATED = "task_created"
    TASK_UPDATED = "task_updated"
    TASK_COMPLETED = "task_completed"
    TASK_FAILED = "task_failed"
    PROGRESS_UPDATED = "progress_updated"

# Circuit Breaker states
class CircuitBreakerState:
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"

# Media types (matching proto enum values)
class MediaType:
    UNKNOWN = 0
    WORLD_HEADER = 1
    WORLD_ICON = 2
    CHARACTER_AVATAR = 3
    POST_IMAGE = 4

# Maximum number of task execution attempts
MAX_ATTEMPTS = {
    TaskType.INIT_WORLD_CREATION: 4,
    TaskType.GENERATE_WORLD_DESCRIPTION: 4,
    TaskType.GENERATE_WORLD_IMAGE: 4,
    TaskType.GENERATE_CHARACTER_BATCH: 4,
    TaskType.GENERATE_CHARACTER: 2,
    TaskType.GENERATE_CHARACTER_AVATAR: 2,
    TaskType.GENERATE_POST_BATCH: 2,
    TaskType.GENERATE_POST: 2,
    TaskType.GENERATE_POST_IMAGE: 2,
}

# Default values
DEFAULT_VALUES = {
    "users_count": 10,
    "posts_count": 50,
    "api_call_limits_LLM": 100,
    "api_call_limits_images": 50,
}