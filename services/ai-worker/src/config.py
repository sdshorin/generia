import os
from typing import Dict, Any
from dotenv import load_dotenv
from os.path import join, dirname


# Load environment variables from .env file
load_dotenv(join(dirname(dirname(__file__)), '.env'))




# Main service settings
SERVICE_NAME = os.getenv("SERVICE_NAME", "ai-worker")
SERVICE_HOST = os.getenv("SERVICE_HOST", "0.0.0.0")
SERVICE_PORT = int(os.getenv("SERVICE_PORT", "8081"))

# MongoDB
MONGODB_URI = os.getenv("MONGODB_URI", "mongodb://admin:password@mongodb:27017")
MONGODB_DATABASE = os.getenv("MONGODB_DATABASE", "generia_ai_worker")

# Kafka
KAFKA_BROKERS = os.getenv("KAFKA_BROKERS", "localhost:9092")
KAFKA_TOPIC_TASKS = os.getenv("KAFKA_TOPIC_TASKS", "generia-tasks")
KAFKA_TOPIC_PROGRESS = os.getenv("KAFKA_TOPIC_PROGRESS", "generia-progress")
KAFKA_GROUP_ID = os.getenv("KAFKA_GROUP_ID", "ai-worker")

# API Keys
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY", "")
RUNWARE_API_KEY = os.getenv("RUNWARE_API_KEY", "")


# LLM Configuration
DEFAULT_LLM_MODEL = os.getenv("DEFAULT_LLM_MODEL", "openai/gpt-3.5-turbo")

# Service limits
MAX_CONCURRENT_TASKS = int(os.getenv("MAX_CONCURRENT_TASKS", "100"))
MAX_CONCURRENT_LLM_REQUESTS = int(os.getenv("MAX_CONCURRENT_LLM_REQUESTS", "15"))
MAX_CONCURRENT_IMAGE_REQUESTS = int(os.getenv("MAX_CONCURRENT_IMAGE_REQUESTS", "10"))

# MinIO
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "generia-images")
MINIO_USE_SSL = os.getenv("MINIO_USE_SSL", "false").lower() == "true"

# API Gateway
API_GATEWAY_URL = os.getenv("API_GATEWAY_URL", "http://api-gateway:8080")

# Logger
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")

# Check required environment variables
def validate_config() -> Dict[str, Any]:
    """Checks for the presence of all required environment variables and returns the configuration status."""
    issues = []
    
    if not OPENROUTER_API_KEY:
        issues.append("OPENROUTER_API_KEY not set")
        
    if not RUNWARE_API_KEY:
        issues.append("RUNWARE_API_KEY not set")
    
    return {
        "valid": len(issues) == 0,
        "issues": issues,
        "config": {
            "service": {
                "name": SERVICE_NAME,
                "host": SERVICE_HOST,
                "port": SERVICE_PORT,
            },
            "mongodb": {
                "uri": MONGODB_URI,
                "database": MONGODB_DATABASE,
            },
            "kafka": {
                "brokers": KAFKA_BROKERS,
                "topic_tasks": KAFKA_TOPIC_TASKS,
                "topic_progress": KAFKA_TOPIC_PROGRESS,
                "group_id": KAFKA_GROUP_ID,
            },
            "llm": {
                "model": DEFAULT_LLM_MODEL
            },
            "limits": {
                "max_concurrent_tasks": MAX_CONCURRENT_TASKS,
                "max_concurrent_llm_requests": MAX_CONCURRENT_LLM_REQUESTS,
                "max_concurrent_image_requests": MAX_CONCURRENT_IMAGE_REQUESTS,
            },
            "minio": {
                "endpoint": MINIO_ENDPOINT,
                "bucket": MINIO_BUCKET,
                "use_ssl": MINIO_USE_SSL,
            },
            "api_gateway": {
                "url": API_GATEWAY_URL,
            },
            "log_level": LOG_LEVEL,
        }
    }