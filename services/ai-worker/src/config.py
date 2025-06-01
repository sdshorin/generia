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

# Temporal
TEMPORAL_HOST = os.getenv("TEMPORAL_HOST", "localhost:7233")
TEMPORAL_NAMESPACE = os.getenv("TEMPORAL_NAMESPACE", "default")

# API Keys
OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY", "")
RUNWARE_API_KEY = os.getenv("RUNWARE_API_KEY", "")


# LLM Configuration
DEFAULT_LLM_MODEL = os.getenv("DEFAULT_LLM_MODEL", "openai/gpt-3.5-turbo")

# Service limits - optimized for high throughput
MAX_CONCURRENT_TASKS = int(os.getenv("MAX_CONCURRENT_TASKS", "500"))
MAX_CONCURRENT_LLM_REQUESTS = int(os.getenv("MAX_CONCURRENT_LLM_REQUESTS", "50"))
MAX_CONCURRENT_IMAGE_REQUESTS = int(os.getenv("MAX_CONCURRENT_IMAGE_REQUESTS", "30"))
MAX_CONCURRENT_GRPC_CALLS = int(os.getenv("MAX_CONCURRENT_GRPC_CALLS", "100"))
MAX_CONCURRENT_DB_OPERATIONS = int(os.getenv("MAX_CONCURRENT_DB_OPERATIONS", "20"))

# Worker specific limits
MAX_WORKFLOW_TASKS_PER_WORKER = int(os.getenv("MAX_WORKFLOW_TASKS_PER_WORKER", "100"))
MAX_ACTIVITIES_PER_WORKER = int(os.getenv("MAX_ACTIVITIES_PER_WORKER", "200"))

# MinIO
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "minio:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "generia-images")
MINIO_USE_SSL = os.getenv("MINIO_USE_SSL", "false").lower() == "true"

# API Gateway
API_GATEWAY_URL = os.getenv("API_GATEWAY_URL", "http://api-gateway:8080")

# Consul
CONSUL_HOST = os.getenv("CONSUL_HOST", "consul")
CONSUL_PORT = int(os.getenv("CONSUL_PORT", "8500"))

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
            "temporal": {
                "host": TEMPORAL_HOST,
                "namespace": TEMPORAL_NAMESPACE,
            },
            "llm": {
                "model": DEFAULT_LLM_MODEL
            },
            "limits": {
                "max_concurrent_tasks": MAX_CONCURRENT_TASKS,
                "max_concurrent_llm_requests": MAX_CONCURRENT_LLM_REQUESTS,
                "max_concurrent_image_requests": MAX_CONCURRENT_IMAGE_REQUESTS,
                "max_concurrent_grpc_calls": MAX_CONCURRENT_GRPC_CALLS,
                "max_concurrent_db_operations": MAX_CONCURRENT_DB_OPERATIONS,
                "max_workflow_tasks_per_worker": MAX_WORKFLOW_TASKS_PER_WORKER,
                "max_activities_per_worker": MAX_ACTIVITIES_PER_WORKER,
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