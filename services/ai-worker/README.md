# AI Worker Service for Generia - Temporal with Task Storage

The AI Worker is a microservice responsible for generating AI content in the Generia platform using **Temporal Workflow Engine** with **MongoDB Task Storage**. This document provides comprehensive technical details for developers and AI agents working with the codebase.

## Contents

- [Architecture Overview](#architecture-overview)
- [Task Storage System](#task-storage-system)
- [Temporal Implementation](#temporal-implementation)
- [Resource Management](#resource-management)
- [Code Structure & File Guide](#code-structure--file-guide)
- [Key Implementation Patterns](#key-implementation-patterns)
- [Cost Tracking System](#cost-tracking-system)
- [Configuration & Environment](#configuration--environment)
- [Development & Testing](#development--testing)
- [Migration Notes](#migration-notes)

## Architecture Overview

### System Design Principles

AI Worker follows modern microservices patterns with **Temporal Workflow Engine** and **MongoDB Task Storage**:

1. **Event-Driven Workflows**: Uses Temporal workflows for orchestration
2. **Task Storage**: Task data stored in MongoDB, only task_id passed between workflows
3. **Resource Pooling**: Proper connection pooling for MongoDB, HTTP, and gRPC
4. **Dependency Injection**: Activities receive shared resources via closure injection
5. **Horizontal Scaling**: Multiple worker instances share task queues
6. **Cost Tracking**: Automatic tracking of LLM and image generation costs

### Core Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Temporal      │    │  AI Worker       │    │  External APIs  │
│   Workflows     │◄──►│  Activities      │◄──►│  (LLM/Image)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │
         │                       ▼
         │              ┌──────────────────┐
         │              │  Shared Resources │
         │              │  • MongoDB Pool   │
         │              │  • HTTP Client    │
         │              │  • gRPC Channels  │
         │              └──────────────────┘
         │                       │
         ▼                       ▼
┌─────────────────┐    ┌──────────────────┐
│   Task Storage  │    │   Task Activities│
│   • Task Data   │◄──►│  • create_task   │
│   • Parameters  │    │  • get_task      │
│   • Metadata    │    │  • update_task   │
└─────────────────┘    └──────────────────┘
```

## Task Storage System

### Overview

**Location**: `src/schemas/task_base.py`, `src/temporal/activities.py`

The AI Worker uses a MongoDB-based task storage system to overcome Temporal's parameter size limitations and improve scalability.

### Task Storage Architecture

**Problem Solved**: Temporal has limitations on workflow parameter sizes, causing failures with large task data.

**Solution**: Store task data in MongoDB, pass only `task_id` between workflows.

### Key Components

#### 1. Base Classes

**File**: `src/schemas/task_base.py`

```python
class TaskInput(BaseModel):
    """Base class for all workflow input data"""
    # Contains full task parameters
    
class TaskRef(BaseModel):
    """Reference object passed between workflows"""
    task_id: str
```

#### 2. BaseWorkflow Methods

**File**: `src/temporal/base_workflow.py`

```python
async def save_task_data(self, input_data: TaskInput, world_id: str) -> TaskRef:
    """Saves task data to MongoDB, returns TaskRef with task_id"""
    
async def load_task_data(self, task_ref: TaskRef, input_class: Type[T]) -> T:
    """Loads task data from MongoDB by task_id"""
```

#### 3. Task Storage Activities

**File**: `src/temporal/activities.py`

- `create_task`: Creates task in MongoDB
- `get_task`: Retrieves task data by ID
- `update_task`: Updates task in MongoDB

### Task Flow Pattern

```python
# Parent Workflow
input_data = GenerateWorldDescriptionInput(world_id="123", user_prompt="fantasy")
task_ref = await self.save_task_data(input_data, world_id)  # Returns TaskRef

await workflow.start_child_workflow("ChildWorkflow", task_ref)  # Pass only TaskRef

# Child Workflow  
@workflow.run
async def run(self, task_ref: TaskRef) -> WorkflowResult:
    input_data = await self.load_task_data(task_ref, GenerateWorldDescriptionInput)
    # Now work with full input_data
```

### Benefits

- **Scalability**: No Temporal parameter size limits
- **Reliability**: Task data persisted in MongoDB
- **Performance**: Reduced Temporal server load
- **Traceability**: All tasks stored with metadata
- **Flexibility**: Easy to extend task data without Temporal changes

## Temporal Implementation

### Workflow Types

**Location**: `src/workflows/`

AI Worker implements these Temporal workflows with task storage:

1. **InitWorldCreationWorkflow**: Orchestrates world generation, accepts full parameters
2. **GenerateWorldDescriptionWorkflow**: Creates world parameters, loads from task_id
3. **GenerateWorldImageWorkflow**: Generates world images, loads from task_id
4. **GenerateCharacterBatchWorkflow**: Creates character batches, loads from task_id
5. **GenerateCharacterWorkflow**: Individual character creation, loads from task_id
6. **GenerateCharacterAvatarWorkflow**: Generates avatars, loads from task_id
7. **GeneratePostBatchWorkflow**: Creates post storylines, loads from task_id
8. **GeneratePostWorkflow**: Individual post generation, loads from task_id
9. **GeneratePostImageWorkflow**: Generates post images, loads from task_id

### Workflow Input Pattern

**All workflows except InitWorldCreationWorkflow**:
```python
@workflow.run
async def run(self, task_ref: TaskRef) -> WorkflowResult:
    # Load full input data from MongoDB
    input = await self.load_task_data(task_ref, InputClass)
    
    # Create child workflow tasks
    child_input = ChildInputClass(...)
    child_task_ref = await self.save_task_data(child_input, world_id)
    
    # Start child with only task_id
    await workflow.start_child_workflow("ChildWorkflow", child_task_ref)
```

**InitWorldCreationWorkflow only**:
```python  
@workflow.run
async def run(self, input: InitWorldCreationInput) -> WorkflowResult:
    # This is the only workflow that receives full parameters
    # It creates the first task for the chain
    description_input = GenerateWorldDescriptionInput(...)
    task_ref = await self.save_task_data(description_input, input.world_id)
```

### Activity Implementation

**Location**: `src/temporal/activities.py`

Activities use **dependency injection pattern** via `create_activity_functions(resource_manager)`:

```python
# Activities are created with injected resources
activities = create_activity_functions(resource_manager)

# Each activity has access to shared connection pools
@activity.defn(name="generate_structured_content")
async def generate_structured_content(...):
    async with resource_manager.llm_semaphore:
        result = await resource_manager.llm_client.generate_structured_content(...)
```

### Worker Specialization

**Location**: `src/main.py:55-156`

Five specialized workers handle different task types:

1. **Main Worker** (`ai-worker-main`): General workflows + activities
2. **LLM Worker** (`ai-worker-llm`): LLM generation only (15 concurrent)
3. **Image Worker** (`ai-worker-images`): Image generation only (10 concurrent)
4. **Progress Worker** (`ai-worker-progress`): Database operations (20 concurrent)
5. **Service Worker** (`ai-worker-services`): gRPC calls only (100 concurrent)

## Resource Management

### SharedResourcesManager

**Location**: `src/temporal/shared_resources.py`

**Key Innovation**: Replaces singleton pattern with proper dependency injection.

```python
class SharedResourcesManager:
    async def initialize(self):
        # 1. MongoDB connection pool
        self.mongo_client = motor.AsyncIOMotorClient(
            maxPoolSize=MAX_CONCURRENT_DB_OPERATIONS * 2,  # 40 connections
            minPoolSize=10,
            maxIdleTimeMS=30_000
        )
        
        # 2. HTTP client with connection pooling
        self.http_client = httpx.AsyncClient(
            limits=httpx.Limits(max_connections=100, max_keepalive_connections=20)
        )
        
        # 3. gRPC channels with keepalive
        # Created lazily in create_grpc_channel()
```

### Connection Pool Configuration

**MongoDB Pool Settings**:
- `maxPoolSize`: 40 (20 × 2 buffer for peak loads)
- `minPoolSize`: 10 (always-ready connections)
- `maxIdleTimeMS`: 30s (prevents Kubernetes connection kills)

**HTTP Pool Settings**:
- `max_connections`: 100 (total HTTP connections)
- `max_keepalive_connections`: 20 (persistent connections)
- `timeout`: 30s (request timeout)

**gRPC Settings**:
- `keepalive_time_ms`: 60s (ping interval)
- `max_connection_idle_ms`: 60s (idle timeout)

### Semaphore Limits

**Location**: `src/config.py:35-39`

```python
MAX_CONCURRENT_LLM_REQUESTS = 50      # OpenRouter API limit
MAX_CONCURRENT_IMAGE_REQUESTS = 30    # Runware API limit  
MAX_CONCURRENT_GRPC_CALLS = 100       # Internal service calls
MAX_CONCURRENT_DB_OPERATIONS = 20     # MongoDB operations
```

## Code Structure & File Guide

### Core Application Files

| File | Purpose | Key Functions |
|------|---------|---------------|
| `src/main.py` | **Entry point**, worker initialization | `initialize_components()`, `create_workers()` |
| `src/config.py` | **Environment configuration** | `validate_config()`, all env vars |
| `src/temporal/shared_resources.py` | **Resource pooling manager** | `SharedResourcesManager` class |
| `src/temporal/activities.py` | **Temporal activity functions** | `create_activity_functions()` |
| `src/temporal/base_activity.py` | **Legacy base classes** | DEPRECATED |

### Database Layer

| File | Purpose | Key Methods |
|------|---------|-------------|
| `src/db/mongo.py` | **MongoDB operations** | `increment_world_generation_cost()` |
| `src/db/models.py` | **Data models** | `WorldGenerationStatus`, `Task`, `ApiRequestHistory` |

### External API Clients

| File | Purpose | Key Features |
|------|---------|-------------|
| `src/api/llm.py` | **LLM client (OpenRouter)** | Auto-cost tracking, structured output, circuit breaker |
| `src/api/image_generator.py` | **Image generation (Runware)** | Auto-cost tracking ($0.0006/image), media upload |
| `src/api/services.py` | **gRPC service client** | Character/Post/Media service integration |

### Workflow Definitions

| File | Purpose | Orchestrates |
|------|---------|-------------|
| `src/workflows/init_world_creation_workflow.py` | **Master orchestrator** | Entire world generation process |
| `src/workflows/generate_world_description_workflow.py` | **World creation** | LLM → structured world data → database |
| `src/workflows/generate_character_batch_workflow.py` | **Character creation** | Batch generation with relationships |
| `src/workflows/generate_post_batch_workflow.py` | **Content creation** | Post storylines and narratives |

### Schema Definitions

| File | Purpose | Used For |
|------|---------|----------|
| `src/schemas/world_description.py` | **World data structure** | LLM structured output, database storage |
| `src/schemas/character.py` | **Character data** | Character creation and storage |
| `src/schemas/post.py` | **Post data** | Post generation and publishing |
| `src/schemas/character_batch.py` | **Batch generation** | Multiple character creation |

### Utility Modules

| File | Purpose | Key Features |
|------|---------|-------------|
| `src/utils/circuit_breaker.py` | **API resilience** | Automatic failure detection and recovery |
| `src/utils/retries.py` | **Retry mechanisms** | Exponential backoff for transient failures |
| `src/utils/logger.py` | **Structured logging** | JSON logging with correlation IDs |
| `src/utils/media_uploader.py` | **Media handling** | Image download and upload to MinIO |

## Key Implementation Patterns

### 1. Task Storage Pattern (NEW)

**Problem Solved**: Temporal has parameter size limitations, large workflow data causes failures.

**Implementation**:
```python
# Parent workflow creates task
input_data = GenerateWorldDescriptionInput(world_id="123", ...)
task_ref = await self.save_task_data(input_data, world_id)

# Child workflow loads task data
input = await self.load_task_data(task_ref, GenerateWorldDescriptionInput)
```

**Task Type Resolution**: Automatic conversion of class names to task types:
```python
# GenerateWorldDescriptionInput -> "generate_world_description"
# GenerateCharacterInput -> "generate_character"
```

### 2. Dependency Injection for Activities

**Problem Solved**: Temporal activities need access to shared resources without global state.

**Implementation**:
```python
# src/temporal/activities.py
def create_activity_functions(resource_manager):
    @activity.defn(name="generate_structured_content")
    async def generate_structured_content(...):
        # resource_manager is captured in closure
        async with resource_manager.llm_semaphore:
            return await resource_manager.llm_client.generate_structured_content(...)
    
    # Task storage activities
    @activity.defn(name="create_task")
    async def create_task(task_type, world_id, parameters):
        task = Task(id=str(uuid.uuid4()), type=task_type, ...)
        return await resource_manager.db_manager.create_task(task)
    
    return {'generate_structured_content': generate_structured_content,
            'create_task': create_task, 'get_task': get_task, ...}
```

### 3. Resource Initialization Pattern

**Problem Solved**: Expensive resources (DB connections, HTTP clients) should be created once per worker.

**Implementation**:
```python
# src/main.py
async def initialize_components():
    resource_manager = SharedResourcesManager()
    await resource_manager.initialize()  # Creates all connection pools
    
async def create_workers():
    activities = create_activity_functions(resource_manager)  # Inject resources
    workers = [Worker(..., activities=[activities['name']], ...)]
```

### 4. Graceful Shutdown Pattern

**Implementation**:
```python
# src/main.py:158-180
async def shutdown():
    # 1. Stop workers first
    for worker in workers:
        worker.shutdown()
    
    # 2. Close resource pools
    await resource_manager.close()  # Closes MongoDB, HTTP, gRPC
    
    # 3. Close Temporal client
    await temporal_client.close()
```

### 4. Semaphore-Based Rate Limiting

**Problem Solved**: Prevent API overload while maximizing throughput.

**Implementation**:
```python
async def llm_activity(...):
    async with resource_manager.llm_semaphore:  # Max 50 concurrent
        result = await llm_client.generate(...)
```

## Cost Tracking System

### Automatic Cost Tracking

**LLM Costs** (`src/api/llm.py:151-161`):
```python
if "usage" in response_data and "cost" in response_data["usage"]:
    cost = float(response_data["usage"]["cost"])
    await self.db_manager.increment_world_generation_cost(
        world_id=world_id, cost_type="llm", cost=cost
    )
```

**Image Costs** (`src/api/image_generator.py:215-225`):
```python
# Fixed cost per image: $0.0006
await self.db_manager.increment_world_generation_cost(
    world_id=world_id, cost_type="image", cost=IMAGE_GENERATION_COST
)
```

### Cost Storage Schema

**MongoDB Document** (`world_generation_status` collection):
```javascript
{
  "_id": "world_123",
  "llm_cost_total": 0.08,      // Sum of all LLM API costs
  "image_cost_total": 0.0018,  // Sum of all image generation costs
  "api_calls_made_LLM": 5,    // Number of LLM requests
  "api_calls_made_images": 3,  // Number of images generated
  // ... other status fields
}
```


## Configuration & Environment

### Required Environment Variables

**Location**: `src/config.py`

**Critical Variables**:
```bash
# Temporal
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default

# MongoDB  
MONGODB_URI=mongodb://admin:password@mongodb:27017
MONGODB_DATABASE=generia_ai_worker

# APIs
OPENROUTER_API_KEY=your_key_here    # Required
RUNWARE_API_KEY=your_key_here       # Required

# Performance Tuning
MAX_CONCURRENT_LLM_REQUESTS=50
MAX_CONCURRENT_IMAGE_REQUESTS=30
MAX_CONCURRENT_DB_OPERATIONS=20
MAX_ACTIVITIES_PER_WORKER=200
```

### Dependencies

**Location**: `requirements.txt`

**Key Dependencies**:
- `temporalio==1.9.0` - Temporal Workflow Engine
- `motor==3.3.2` - Async MongoDB driver
- `httpx==0.25.2` - Async HTTP client with connection pooling
- `grpcio==1.71.0` - gRPC for service communication
- `pydantic==2.5.3` - Data validation and schemas

### Debugging Workflows

**Temporal UI**: http://localhost:8090 (when running locally)

**Log Levels**: Set `LOG_LEVEL=DEBUG` for detailed activity logs

**Key Log Patterns**:
- `Activity {type} - {action} for world {id}` - Activity execution
- `Updated {cost_type} cost for world {id}: ${amount}` - Cost tracking
- `SharedResourcesManager fully initialized` - Resource setup complete

## Migration Notes

### From Kafka to Temporal

**Previous Architecture** (Deprecated):
- Kafka consumers for task processing
- Singleton resource sharing via `shared_resources`
- Manual retry and error handling
- Progress manager as separate component

**Current Architecture** (Temporal):
- Temporal workflows for orchestration  
- Dependency injection for resource sharing
- Built-in retry policies and error handling
- Progress tracking integrated into activities

### Backward Compatibility

**Deprecated Components**:
- `src/temporal/base_activity.py` - Use activity functions instead
- Global `shared_resources` singleton - Use `SharedResourcesManager`
- Manual progress manager - Use integrated cost tracking

**Migration Path**:
1. New workflows use `create_activity_functions()` pattern
2. Old workflows can be gradually migrated
3. Resource pools are backward compatible with existing clients

### Performance Improvements

**Connection Pooling Benefits**:
- **Before**: New MongoDB connection per task (~100ms overhead)
- **After**: Reused connections from pool (~1ms overhead)

**Resource Utilization**:
- **Before**: Up to 500 MongoDB connections per worker
- **After**: Maximum 40 connections per worker with pooling

**Scaling Characteristics**:
- **Horizontal**: Add more worker pods, automatic load balancing
- **Vertical**: Increase semaphore limits and pool sizes
- **Cost**: Linear scaling with automatic cost tracking per world

---

## Quick Reference for AI Agents

**Looking for specific functionality?**

| Task | Primary Files | Key Functions |
|------|---------------|---------------|
| **Add new workflow** | `src/workflows/` | Extend TaskInput pattern, use task storage |
| **Add new task type** | `src/schemas/task_base.py` | Create TaskInput subclass |
| **Modify LLM generation** | `src/api/llm.py` | `generate_structured_content()` |
| **Change image generation** | `src/api/image_generator.py` | `generate_and_upload_image()` |
| **Update cost tracking** | `src/db/mongo.py` | `increment_world_generation_cost()` |
| **Add new activity** | `src/temporal/activities.py` | Add to `create_activity_functions()` |
| **Modify task storage** | `src/temporal/activities.py` | `create_task`, `get_task`, `update_task` |
| **Modify resource pools** | `src/temporal/shared_resources.py` | `SharedResourcesManager.initialize()` |
| **Change worker configuration** | `src/main.py` | `create_workers()` |
| **Update data schemas** | `src/schemas/` | Pydantic model definitions |

**Task Storage Patterns**:
- **New Workflow**: Inherit from `TaskInput`, accept `TaskRef` in run method
- **Parent Workflow**: Use `save_task_data()` before starting child workflows
- **Child Workflow**: Use `load_task_data()` to get full input data
- **Task ID Generation**: Automatic UUID generation in `create_task` activity

**Architecture Patterns**:
- **Resource Access**: Always use `resource_manager.{client_name}` 
- **Rate Limiting**: Wrap operations with `async with resource_manager.{type}_semaphore:`
- **Error Handling**: Activities auto-retry via Temporal, clients use circuit breakers
- **Cost Tracking**: Automatic in LLM/Image clients, manual via `increment_world_generation_cost()`
- **Database**: Use `resource_manager.db_manager` for all MongoDB operations
- **Task Storage**: Use BaseWorkflow methods for task data persistence

## Migration from Kafka to Temporal with Task Storage

### v2.0 Task Storage Implementation

**Date**: January 2025

**Breaking Changes**: All workflows now use task storage system.

**Migration Required For**:
- Any custom workflow implementations
- External workflow triggers (world-service integration)
- Monitoring and logging systems

**Key Changes**:

1. **Workflow Signatures**: All workflows except `InitWorldCreationWorkflow` now accept `TaskRef` instead of full parameters
2. **Input Classes**: All `@dataclass` input classes converted to Pydantic `TaskInput` models
3. **Child Workflow Calls**: Must use `save_task_data()` → `start_child_workflow(TaskRef)`
4. **Database Schema**: New `tasks` collection in MongoDB for task storage

**Benefits Delivered**:
- ✅ Eliminated Temporal parameter size limitations
- ✅ Improved workflow reliability and debugging
- ✅ Enhanced task traceability in MongoDB
- ✅ Reduced Temporal server memory usage
- ✅ Better horizontal scaling characteristics

**Integration Points**:
- `world-service` at `services/world-service/internal/service/world_service.go:570-690` triggers `InitWorldCreationWorkflow`
- All other workflows are triggered internally via task storage system

This documentation serves as the primary reference for understanding and modifying the AI Worker codebase.