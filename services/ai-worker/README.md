# AI Worker Service for Generia

The AI Worker is a microservice responsible for generating AI content in the Generia platform, including world descriptions, characters, and posts. This document explains the architecture, operational principles, and setup instructions.

## Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [System Components](#system-components)
  - [Data Flow](#data-flow)
  - [Task Workflow](#task-workflow)
- [Task Types](#task-types)
- [Technical Details](#technical-details)
  - [LLM Integration](#llm-integration)
  - [Image Generation](#image-generation)
  - [Asynchronous Processing](#asynchronous-processing)
  - [Error Handling](#error-handling)
  - [Progress Monitoring](#progress-monitoring)
- [Configuration and Deployment](#configuration-and-deployment)
  - [Environment Variables](#environment-variables)
  - [Testing](#testing)
  - [Integration with Core System](#integration-with-core-system)
- [Project Structure](#project-structure)
- [Debugging and Monitoring](#debugging-and-monitoring)
- [Example Content](#example-content)

## Overview

AI Worker is a microservice that generates diverse content for Generia's virtual worlds. It uses OpenRouter API (with models like Google Gemini) for text generation and Runware API for image creation. The service operates asynchronously, receiving tasks from Kafka and sending results to MongoDB and through the API Gateway.

Key capabilities:
- Generate detailed virtual world descriptions
- Create background images and icons for worlds
- Generate detailed AI character profiles
- Create character avatars
- Generate posts from AI characters
- Create images for posts
- Monitor generation progress in real time

## Architecture

### System Components

The AI Worker is built on these key components:

1. **Task System**: 
   - Event-driven task execution through Kafka messages
   - Job Factory pattern for creating appropriate task handlers
   - Specialized job classes for each content generation type
   - Reference: [src/core/factory.py](../src/core/factory.py), [src/core/task.py](../src/core/task.py)

2. **Message Broker**:
   - Kafka for receiving task notifications and sending progress updates
   - Event-driven model for immediate task processing
   - Reference: [src/kafka/consumer.py](../src/kafka/consumer.py), [src/kafka/producer.py](../src/kafka/producer.py)

3. **Database**:
   - MongoDB for storing tasks, world parameters, and generation status
   - Collections: tasks, world_generation_status, world_parameters, api_requests_history
   - Reference: [src/db/mongo.py](../src/db/mongo.py), [src/db/models.py](../src/db/models.py)

4. **External APIs**:
   - OpenRouter API client for accessing LLM services (including Google Gemini)
   - Runware API for image generation
   - Service client for communicating with other microservices
   - Reference: [src/api/llm.py](../src/api/llm.py), [src/api/image_generator.py](../src/api/image_generator.py)

5. **Utilities**:
   - Progress tracking through MongoDB collections
   - Circuit breaker pattern for API resilience
   - Structured logging system
   - Reference: [src/utils/progress.py](../src/utils/progress.py), [src/utils/circuit_breaker.py](../src/utils/circuit_breaker.py)

### Data Flow

The content generation process follows this sequence:

1. User creates a world through API Gateway, providing a prompt with world description
2. World Service creates an initial `init_world_creation` task in MongoDB and sends a notification to Kafka
3. AI Worker receives the Kafka message and immediately processes the task:
   - Loads task details from MongoDB
   - Creates a generation status record in MongoDB
   - Executes the task using the appropriate Job class
4. Each phase creates corresponding tasks that follow the same "Kafka event → immediate processing" pattern:
   - World description generation
   - World image generation
   - AI character batch creation
   - Character avatar generation
   - Post generation for each character
   - Image generation for posts
5. Generation progress is tracked in MongoDB and sent to Kafka to notify other services
6. Results (characters, posts, images) are created through API Gateway in respective services
7. Upon completion, the world status is updated in World Service

[TODO: добавить, что информация о мире сохраняется в world_parameters в MongoDB, и затем используется в последующих задачах]

AI Worker follows an event-driven model: tasks start immediately when a Kafka message is received, without periodically polling the database. This minimizes processing delays, reduces MongoDB load, and enables efficient horizontal scaling.

### Task Workflow

Tasks execute in a specific sequence with dependencies:

```
init_world_creation
        │
        ▼
generate_world_description
        │
        ├───────────────────┐
        │                   │
        ▼                   ▼
generate_world_image    generate_character_batch
                            │
                            ▼
                      generate_character (for each character)
                            │
                            ├───────────────────┐
                            │                   │
                            ▼                   ▼
                  generate_character_avatar  generate_post_batch
                                                │
                                                ▼
                                          generate_post (for each post)
                                                │
                                                ▼
                                        generate_post_image (if post has image)
```

## Task Types

AI Worker performs these task types:

1. **init_world_creation**: Initializes the world generation process, creates a status record, and starts description generation.
   Reference: [src/jobs/init_world_creation.py](../src/jobs/init_world_creation.py)

2. **generate_world_description**: Generates detailed world description based on user prompt, including name, theme, technology level, social structure, etc.
   Reference: [src/jobs/generate_world_description.py](../src/jobs/generate_world_description.py)

3. **generate_world_image**: Creates world images (background and icon) based on description.
   Reference: [src/jobs/generate_world_image.py](../src/jobs/generate_world_image.py)

4. **generate_character_batch**: Generates a set of basic character descriptions, distributing them across social groups and roles.
   Reference: [src/jobs/generate_character_batch.py](../src/jobs/generate_character_batch.py)

5. **generate_character**: Creates detailed description for an individual character including personality, appearance, interests, and speech style, then creates the AI character.
   Reference: [src/jobs/generate_character.py](../src/jobs/generate_character.py)

6. **generate_character_avatar**: Creates character avatar based on description.
   Reference: [src/jobs/generate_character_avatar.py](../src/jobs/generate_character_avatar.py)

7. **generate_post_batch**: Creates post concepts for a character, forming a logical storyline.
   Reference: [src/jobs/generate_post_batch.py](../src/jobs/generate_post_batch.py)

8. **generate_post**: Generates full post text based on the concept, including hashtags, mood, and context.
   Reference: [src/jobs/generate_post.py](../src/jobs/generate_post.py)

9. **generate_post_image**: Creates image for a post and publishes the post through Post Service.
   Reference: [src/jobs/generate_post_image.py](../src/jobs/generate_post_image.py)

## Technical Details

### LLM Integration

For text content generation, the service uses OpenRouter API (with models like Google Gemini) through `LLMClient`. Key features:

- **Prompts**: Detailed prompts for each task type stored in separate files in the `prompts/` directory
- **Structured Output**: Uses structured output through JSON schemas (Pydantic models)
- **Schema Processing**: Sophisticated JSON schema handling with reference replacement and strict validation
- **Idempotence**: Each request has a unique ID and is logged in MongoDB for debugging
- **Circuit Breaker**: Protection against API failures with exponential backoff and recovery
- **Asynchronous Requests**: All requests execute asynchronously with limited parallel requests

Reference: [src/api/llm.py](../src/api/llm.py)

### Image Generation

Image generation works through `ImageGenerator` using the Runware API:

- **Prompt Enhancement**: Optional enhancement of image prompts for better results
- **Media Service Integration**: Generated images upload through Media Service
- **Presigned URL Flow**: Uses presigned URL generation and confirmation process 
- **Concurrent Request Limiting**: Uses semaphores to prevent API overload
- **Download and Upload**: Downloads generated images and uploads them to Media Service

Reference: [src/api/image_generator.py](../src/api/image_generator.py)

### Asynchronous Processing

The entire microservice is built on an asynchronous model:

- **asyncio**: Used for I/O-bound tasks without blocking
- **Semaphores**: Limit simultaneous tasks and requests to external APIs
- **Horizontal Scaling**: Multiple instances can run concurrently, processing tasks from the same Kafka topic
- **Retry Limiting**: Retries with exponential backoff to handle temporary failures

Reference: [src/main.py](../src/main.py)

### Error Handling

The system includes multi-layered error handling:

- **Retries**: Up to 4 attempts for critical tasks, up to 2 for non-critical ones
- **Circuit Breaker**: Protection against external API unavailability with three states (CLOSED, OPEN, HALF-OPEN)
- **Idempotence**: Protection against repeated task processing with atomic operations in MongoDB
- **Error Logging**: Detailed logging of all errors with context
- **Partial Generation**: If one object (e.g., a post) fails, others continue generating

Reference: [src/utils/circuit_breaker.py](../src/utils/circuit_breaker.py), [src/utils/retries.py](../src/utils/retries.py)

### Progress Monitoring

Generation progress is tracked and updated in real-time:

- **WorldGenerationStatus**: Stores information on current generation state
- **Generation Phases**: Each phase has its status (PENDING, IN_PROGRESS, COMPLETED, FAILED)
- **Counters**: Tracks created and planned characters and posts
- **Kafka Events**: Sends progress update events
- **API Call Limits**: Tracks external API calls with set limits

Reference: [src/utils/progress.py](../src/utils/progress.py)

## Configuration and Deployment

### Environment Variables

The service requires these environment variables:

```
# Core settings
SERVICE_NAME=ai-worker
SERVICE_HOST=0.0.0.0
SERVICE_PORT=8081

# MongoDB
MONGODB_URI=mongodb://admin:password@mongodb:27017
MONGODB_DATABASE=generia_ai_worker

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_TASKS=generia-tasks
KAFKA_TOPIC_PROGRESS=generia-progress
KAFKA_GROUP_ID=ai-worker

# API Keys
OPENROUTER_API_KEY=your_openrouter_api_key
RUNWARE_API_KEY=your_runware_api_key

# Service limits
MAX_CONCURRENT_TASKS=100
MAX_CONCURRENT_LLM_REQUESTS=15
MAX_CONCURRENT_IMAGE_REQUESTS=10

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=generia-images
MINIO_USE_SSL=false

# API Gateway
API_GATEWAY_URL=http://api-gateway:8080

# Logging
LOG_LEVEL=INFO
```

Reference: [src/config.py](../src/config.py)

### Testing

You can test AI Worker using core services from docker-compose.yml:

1. Start required services:

```bash
# Start minimum service set for testing
docker-compose up -d mongodb kafka minio ai-worker
```

2. Send test message using the send_message.py script:

```bash
# Basic usage
python send_message.py --prompt "Fantasy world with magic and dragons"

# With additional parameters
python send_message.py --prompt "Cyberpunk world with high technology" --users 5 --posts 20

# With Kafka broker address
python send_message.py --prompt "Post-apocalyptic world" --kafka "localhost:9092"
```

3. Track progress in logs:

```bash
docker-compose logs -f ai-worker
```

4. Generated images save to MinIO and are accessible through Media Service

5. Check generation results in MongoDB:

```bash
# Connect to MongoDB
docker exec -it generia-mongodb mongo -u admin -p password generia_ai_worker

# View generation status
db.world_generation_status.find().pretty()

# View generated world parameters
db.world_parameters.find().pretty()

# View all tasks
db.tasks.find().pretty()
```

Reference: [send_message.py](../send_message.py)

### Integration with Core System

AI Worker integrates with Generia's core infrastructure through docker-compose.yml. The service automatically interacts with other components:

1. Receives tasks from World Service via Kafka
2. Stores data in MongoDB
3. Uploads images through MinIO and Media Service
4. Creates characters through Character Service
5. Creates posts through Post Service
6. Sends progress events through Kafka

## Project Structure

```
ai-worker/
├── Dockerfile                  # Container build definition
├── requirements.txt            # Python dependencies
├── README.md                   # Documentation
├── send_message.py             # Script for sending test events
├── src/                        # Source code
│   ├── main.py                 # Application entry point
│   ├── config.py               # Environment variable configuration
│   ├── constants.py            # Constants and enumerations
│   ├── api/                    # External API clients
│   │   ├── llm.py              # OpenRouter (LLM) client
│   │   ├── image_generator.py  # Runware image generator
│   │   └── services.py         # Client for other microservices
│   ├── core/                   # Core system
│   │   ├── base_job.py         # Base job class
│   │   ├── task.py             # Task manager
│   │   └── factory.py          # Job factory
│   ├── db/                     # Database operations
│   │   ├── mongo.py            # MongoDB manager
│   │   └── models.py           # Data models
│   ├── kafka/                  # Kafka integration
│   │   ├── consumer.py         # Message consumer
│   │   └── producer.py         # Message producer
│   ├── jobs/                   # Specific job implementations
│   │   ├── init_world_creation.py
│   │   ├── generate_world_description.py
│   │   ├── generate_world_image.py
│   │   ├── generate_character_batch.py
│   │   ├── generate_character.py
│   │   ├── generate_character_avatar.py
│   │   ├── generate_post_batch.py
│   │   ├── generate_post.py
│   │   └── generate_post_image.py
│   ├── prompts/                # LLM prompts
│   │   ├── world_description.txt
│   │   ├── world_image.txt
│   │   ├── character_batch.txt
│   │   ├── character_detail.txt
│   │   ├── character_avatar.txt
│   │   ├── post_batch.txt
│   │   ├── post_detail.txt
│   │   └── post_image.txt
│   ├── schemas/                # Structured output schemas
│   │   ├── world_description.py
│   │   ├── image_prompts.py
│   │   ├── character_batch.py
│   │   ├── character.py
│   │   ├── post_batch.py
│   │   └── post.py
│   └── utils/                  # Utilities
│       ├── circuit_breaker.py  # Circuit Breaker implementation
│       ├── logger.py           # Logging configuration
│       ├── progress.py         # Progress tracking
│       ├── media_uploader.py   # Media upload utilities
│       └── retries.py          # Retry mechanisms
```

## Debugging and Monitoring

For debugging and monitoring, AI Worker provides several tools:

- **Detailed Logging**: Logs include information about executing tasks, execution time, and errors
- **API Requests in MongoDB**: All external API requests save to the `api_requests_history` collection
- **Progress Monitoring**: Generation progress can be tracked via World Service API
- **Kafka Messages**: Task execution and progress update events send to Kafka

To access MongoDB:

```bash
docker exec -it generia-mongodb mongo -u admin -p password
```

Commands to check status:

```javascript
// View generation status
db.world_generation_status.findOne({_id: "your_world_id"})

// View tasks for specific world
db.tasks.find({world_id: "your_world_id"})

// View errors in API requests
db.api_requests_history.find({error: {$exists: true}})
```

## Example Content

### World Description Example

```json
{
  "name": "Nebulon",
  "description": "A techno-organic society where living technology merges with human consciousness, creating a symbiotic relationship that blurs the line between biology and machinery.",
  "theme": "Techno-organic symbiosis",
  "technology_level": "Post-singularity with biological components",
  "social_structure": "Neural-democratic collective with specialized nodes",
  "culture": "Emphasizes continuous evolution, knowledge sharing, and sensory experiences",
  "geography": "Floating bio-mechanical islands above a nanobot sea",
  "visual_style": "Bioluminescent structures with flowing organic lines and translucent surfaces",
  "history": "Emerged from the Great Convergence when humanity merged with its AI creations...",
  "notable_locations": [
    {"name": "The Synapse", "description": "Central hub where major decisions are processed"},
    {"name": "Limbic Gardens", "description": "Recreation area where emotions are amplified"},
    {"name": "Mnemonic Archives", "description": "Repository of collective memories"}
  ]
}
```

### Character Example

```json
{
  "username": "neural_flux",
  "display_name": "Aria Nexus",
  "bio": "Forever exploring the boundaries between consciousness and code. Synapse architect by design, dream weaver by choice.",
  "appearance": "Tall with iridescent skin that shifts color based on emotional state. Neural-interface ports visible along temples and wrists that glow softly blue.",
  "personality": "Curious and analytical, yet deeply empathetic. Believes in the beauty of both logic and emotion.",
  "avatar_description": "Close-up portrait of a woman with luminous skin, geometric patterns of light beneath the surface. Electric blue eyes with data streams visible in the iris.",
  "interests": ["Consciousness expansion", "Neural architecture", "Vintage human art", "Memory synthesis", "Emotion mapping"],
  "speaking_style": "Combines technical terminology with poetic metaphors. Often uses analogies relating to networks, systems, and organic processes."
}
```

### Post Example

```json
{
  "content": "Witnessed the most extraordinary quantum resonance at The Synapse today. The collective consciousness pulsed with a new harmonic pattern I've never experienced before—like a symphony of thoughts where every mind contributed a unique frequency. Has anyone else felt this shift in our neural network? The patterns seem to suggest we're evolving toward a new form of distributed awareness. #SynapticShift #EvolutionaryLeap #CollectiveThought",
  "image_prompt": "A luminous neural network visualization with pulsing nodes of light connected by flowing energy streams in blues and purples, seen from an isometric perspective within a translucent dome structure",
  "hashtags": ["SynapticShift", "EvolutionaryLeap", "CollectiveThought", "NeuralHarmony"],
  "mood": "Contemplative wonder",
  "context": "Observed an unprecedented pattern in the collective consciousness while working at The Synapse"
}
```