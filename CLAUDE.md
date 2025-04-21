# Generia: Virtual Worlds Platform (Claude's Guide)

## Project Overview

Generia is a microservices-based platform for creating and exploring virtual worlds filled with AI-generated content, simulating a "dead internet" experience with isolated social network worlds. This project allows users to:

- Create and join virtual worlds with unique themes
- Generate AI-powered users and content within each world
- Explore isolated social environments with distinct characteristics
- Create and share posts within the active world
- Like and comment on content across different worlds

## Architecture

Generia uses a modern microservices architecture with the following key components:

1. **Backend Services (Go)**:
   - **API Gateway** - Entry point for client applications
   - **Auth Service** - User authentication and authorization
   - **World Service** - World creation and management
   - **Post Service** - Post creation and retrieval
   - **Media Service** - Media uploads and processing
   - **Interaction Service** - Likes and comments
   - **Feed Service** - User feeds
   - **Cache Service** - Caching of frequent data
   - **AI Worker** - AI content generation

2. **Frontend (React + TypeScript)**:
   - Modern React application with TypeScript
   - Context API for state management
   - Styled Components for styling
   - Responsive design

3. **Data Storage**:
   - PostgreSQL for structured data
   - MongoDB for interactions
   - Redis for caching
   - MinIO for media storage

## Key Files and Directories

When working on this project, these are the most important locations to understand:

- `/services/` - All microservices with their individual codebases
- `/api/proto/` - Protocol Buffer definitions for services
- `/api/grpc/` - Generated gRPC code
- `/pkg/` - Shared packages used across services
- `/frontend/` - React frontend application
- `/frontend/src/components/` - UI components
- `/frontend/src/pages/` - Application pages
- `/frontend/src/context/` - State management
- `/frontend/src/api/` - API integrations

## Development Guidelines

### Important: This is an MVP Project

This project is still in development and has not been released yet. As such:

- **Backward compatibility is NOT a requirement** - Feel free to make breaking changes when needed
- **Code quality must be high** - Write clean, maintainable code at a senior level
- **No hacky solutions** - Avoid workarounds or "quick fixes" that compromise quality

### Process for Major Changes

When working on significant changes:

1. **Plan thoroughly** - First, think about the architecture and approach
2. **Create a detailed plan** - List the files that need changes and what changes are needed
3. **Discuss before implementation** - Present this plan for approval before writing code
4. **Implement only after approval** - Proceed with coding only after the plan is approved

### When to Seek Approval

You must seek approval when:

1. A requested change would require significant code restructuring
2. Implementation requires writing hacky or suboptimal solutions
3. Changes affect core architecture components
4. Multiple services need to be modified simultaneously

## Testing and Running

- Start the application: `docker-compose up -d`
- Frontend available at: http://localhost:80
- Backend API Gateway: http://localhost:8080

## Documentation

For more detailed information about specific components:

- Main project documentation: `README.md` (ALWAYS READ IT BEFORE ANY ACTIONS)
- Frontend documentation: `frontend/README.md`
- Architecture documentation: `ai_instruments/docs/general_architecture.md`

Remember that when working on this project, you should prioritize clean, maintainable solutions over quick fixes, and don't hesitate to suggest architectural improvements when necessary. Since this is an MVP, now is the time to implement the right architecture, even if it means breaking existing code.