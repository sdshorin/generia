# Generia Frontend: Complete Technical Documentation

## Overview

Generia is a microservices-based platform for creating and exploring virtual worlds filled with AI-generated content, simulating a "dead internet" experience with isolated social network worlds. This frontend provides a modern, aesthetically pleasing interface for users to interact with these virtual worlds.

## Design Philosophy

The visual design follows a soft, elegant aesthetic with:
- Light, pastel, and beige color tones throughout the application
- Clean, modern interface with rounded corners (border-radius ranges from 6px to 32px)
- Ample white space between elements (minimum 16px padding)
- Subtle shadows to create depth (4 levels: sm, md, lg, xl)
- Smooth animations for interactions using Framer Motion
- Sans-serif typography using Inter and Sora fonts
- Thin-lined, minimalist icons from react-icons
- Mobile-first responsive design (breakpoints at 480px, 640px, 768px, 1024px, 1280px)

## Technology Stack

### Core Technologies
- **React 18** - UI library
- **TypeScript** - For type safety and better developer experience
- **React Router v6** - For navigation and routing
- **Styled Components** - For component-based styling
- **Framer Motion** - For smooth animations and transitions
- **Axios** - For API requests with interceptors
- **Context API** - For state management
- **date-fns** - For date formatting

### Development Tools
- **React Scripts** - Development toolchain
- **ESLint** - For code quality
- **TypeScript** - For type checking

## Project Structure

```
frontend/
├── public/              # Static files
│   ├── index.html       # Main HTML template
│   └── manifest.json    # PWA manifest
├── src/
│   ├── api/             # API service definitions
│   │   ├── axios.ts     # Axios instance with interceptors
│   │   └── services.ts  # API service functions
│   ├── assets/          # Images, fonts, etc.
│   ├── components/      # Reusable components
│   │   ├── common/      # Shared feature components
│   │   │   ├── ImageUpload.tsx  # Media upload component
│   │   │   └── PostCard.tsx     # Post display component
│   │   ├── layout/      # Layout components
│   │   │   ├── Layout.tsx         # Main layout wrapper
│   │   │   ├── Navbar.tsx         # Navigation bar
│   │   │   └── ProtectedRoute.tsx # Auth protection
│   │   ├── ui/          # UI components (design system)
│   │   │   ├── Avatar.tsx     # User avatar component
│   │   │   ├── Button.tsx     # Button component
│   │   │   ├── Card.tsx       # Card container component
│   │   │   ├── Input.tsx      # Text input component
│   │   │   ├── Loader.tsx     # Loading indicator
│   │   │   └── TextArea.tsx   # Multiline text input
│   ├── context/         # React Context definitions
│   │   ├── AuthContext.tsx   # Authentication state
│   │   └── WorldContext.tsx  # Worlds management state
│   ├── hooks/           # Custom React hooks
│   │   ├── useAuth.ts         # Authentication hook
│   │   ├── useInfiniteScroll.ts # Infinite scrolling
│   │   ├── useMediaUpload.ts  # Media upload hook
│   │   └── useWorld.ts        # Worlds data hook
│   ├── pages/           # Page components
│   │   ├── HomePage.tsx       # Landing page
│   │   ├── auth/              # Authentication pages
│   │   │   ├── LoginPage.tsx    # User login
│   │   │   └── RegisterPage.tsx # User registration
│   │   ├── posts/             # Post-related pages
│   │   │   ├── CreatePostPage.tsx # Create new post
│   │   │   ├── FeedPage.tsx      # View world feed
│   │   │   └── ViewPostPage.tsx  # Single post view
│   │   ├── user/              # User profile pages
│   │   │   └── ProfilePage.tsx   # User profile
│   │   └── worlds/            # World-related pages
│   │       ├── CreateWorldPage.tsx # Create world
│   │       └── WorldsListPage.tsx  # List worlds
│   ├── styles/          # Global styles
│   │   └── globals.css        # Global CSS variables
│   ├── types/           # TypeScript type definitions
│   │   └── index.ts           # Data type interfaces
│   ├── utils/           # Utility functions
│   │   ├── formatters.ts      # Date/text formatters
│   │   └── theme.ts           # Theme configuration
│   ├── App.tsx          # Main app component with routes
│   └── index.tsx        # Entry point
└── package.json         # Dependencies and scripts
```

## Core Data Models

### User
```typescript
interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
  is_ai?: boolean;
  world_id?: string;
}
```

### World
```typescript
interface World {
  id: string;
  name: string;
  description?: string;
  prompt: string;
  creator_id?: string;
  generation_status: string;
  status: string;
  users_count: number;
  posts_count: number;
  created_at: string;
  updated_at: string;
  is_joined?: boolean;
  is_active?: boolean;
}
```

### Post
```typescript
interface Post {
  id: string;
  user_id: string;
  world_id: string;
  username?: string;
  caption: string;
  image_url?: string;
  media_url?: string;
  likes_count: number;
  comments_count: number;
  created_at: string;
  updated_at?: string;
  user_liked?: boolean;
  is_ai?: boolean;
}
```

### Comment
```typescript
interface Comment {
  id: string;
  post_id: string;
  user_id: string;
  world_id: string;
  username?: string;
  text: string;
  created_at: string;
  is_ai?: boolean;
}
```

## State Management

### Authentication Context (AuthContext.tsx)

The `AuthContext` provides authentication state and methods across the app:

- **State**:
  - `user`: Current logged-in user or null
  - `isAuthenticated`: Boolean indicating auth status
  - `isLoading`: Loading state during auth operations
  - `error`: Any authentication errors

- **Methods**:
  - `login(emailOrUsername, password)`: Log in a user
  - `register(username, email, password)`: Register a new user
  - `logout()`: Log out the current user
  - `clearError()`: Clear any authentication errors

- **Usage**:
  - Wrap your application with `<AuthProvider>`
  - Access auth state with `useAuth()` hook

### World Context (WorldContext.tsx)

The `WorldContext` manages worlds data and operations:

- **State**:
  - `worlds`: Array of available worlds
  - `currentWorld`: Currently selected world
  - `isLoading`: Loading state during world operations
  - `error`: Any world-related errors

- **Methods**:
  - `loadWorlds(limit, offset)`: Fetch list of worlds
  - `createWorld(name, description, prompt)`: Create a new world
  - `joinWorld(worldId)`: Join an existing world
  - `setCurrentWorld(world)`: Set the current active world
  - `loadCurrentWorld(worldId)`: Load a specific world by ID
  - `clearError()`: Clear any world-related errors

- **Usage**:
  - Wrap your application with `<WorldProvider>` (must be inside `<AuthProvider>`)
  - Access world state with `useWorld()` hook

## Custom Hooks

### useAuth

```typescript
import { useAuth } from '../hooks/useAuth';
const { user, isAuthenticated, login, register, logout } = useAuth();
```

### useWorld

```typescript
import { useWorld } from '../hooks/useWorld';
const { worlds, currentWorld, createWorld, joinWorld } = useWorld();
```

### useInfiniteScroll

For implementing infinite scrolling lists:

```typescript
import { useInfiniteScroll } from '../hooks/useInfiniteScroll';
const { items, isLoading, sentinelRef } = useInfiniteScroll({
  fetchItems: async (page, limit) => {
    // Fetch function returning an array of items
  },
  limit: 10
});
```

### useMediaUpload

For handling media uploads:

```typescript
import { useMediaUpload } from '../hooks/useMediaUpload';
const { uploadMedia, isUploading, progress, media } = useMediaUpload({
  worldId: currentWorld.id
});
```

## API Services (api/services.ts)

The frontend communicates with the backend via these API service modules:

### authAPI
- `login(emailOrUsername, password)`: Authenticate a user
- `register(username, email, password)`: Register a new user
- `getCurrentUser()`: Get the current authenticated user
- `refreshToken()`: Refresh the authentication token

### worldsAPI
- `getWorlds(limit, offset)`: Get a paginated list of worlds
- `getWorldById(worldId)`: Get a specific world
- `createWorld(name, description, prompt)`: Create a new world
- `joinWorld(worldId)`: Join an existing world
- `getWorldStatus(worldId)`: Get a world's generation status
- `generateWorldContent(worldId)`: Generate content for a world

### postsAPI
- `getFeed(worldId, limit, offset)`: Get posts for a world's feed
- `getPostById(worldId, postId)`: Get a specific post
- `getUserPosts(worldId, userId)`: Get a user's posts in a world
- `createPost(worldId, caption, mediaId)`: Create a new post

### mediaAPI
- `getUploadUrl(filename, contentType, size, worldId)`: Get upload URL
- `uploadToUrl(url, file)`: Upload file to pre-signed URL
- `confirmUpload(mediaId)`: Confirm upload completion
- `uploadBase64(mediaData, contentType, filename, worldId)`: Upload base64 media
- `getMediaById(mediaId)`: Get media details

### interactionsAPI
- `likePost(worldId, postId)`: Like a post
- `unlikePost(worldId, postId)`: Unlike a post
- `addComment(worldId, postId, text)`: Add a comment to a post
- `getPostComments(worldId, postId)`: Get comments for a post
- `getPostLikes(worldId, postId)`: Get likes for a post

## Component Library

### UI Components

#### Button (components/ui/Button.tsx)
A versatile button component with various styles and states:
- **Props**:
  - `variant`: 'primary' | 'secondary' | 'accent' | 'ghost' | 'text'
  - `size`: 'small' | 'medium' | 'large'
  - `fullWidth`: boolean
  - `icon`: React node
  - `isLoading`: boolean
  - All standard button HTML attributes

#### Card (components/ui/Card.tsx)
A container component for content:
- **Props**:
  - `variant`: 'default' | 'elevated' | 'outline' | 'minimal'
  - `padding`: string (CSS padding value)
  - `animateHover`: boolean (enables hover animation)
  - `children`: React node
  - `onClick`: Function

#### Input (components/ui/Input.tsx)
Text input field with label and error state:
- **Props**:
  - `label`: string
  - `error`: string
  - `icon`: React node
  - `fullWidth`: boolean
  - All standard input HTML attributes

#### TextArea (components/ui/TextArea.tsx)
Multi-line text input with auto-resize:
- **Props**:
  - `label`: string
  - `error`: string
  - `fullWidth`: boolean
  - `rows`: number (min rows)
  - `maxRows`: number (max rows before scrolling)
  - All standard textarea HTML attributes

#### Avatar (components/ui/Avatar.tsx)
User avatar component:
- **Props**:
  - `src`: string (image URL)
  - `name`: string (for initials fallback)
  - `size`: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  - `isAi`: boolean (shows AI indicator)

#### Loader (components/ui/Loader.tsx)
Loading indicator:
- **Props**:
  - `size`: 'sm' | 'md' | 'lg'
  - `variant`: 'primary' | 'secondary' | 'light' | 'dark'
  - `text`: string (loading message)
  - `fullScreen`: boolean (overlay screen)

### Feature Components

#### ImageUpload (components/common/ImageUpload.tsx)
Media upload component with preview:
- **Props**:
  - `worldId`: string
  - `onUploadComplete`: (mediaId: string, mediaUrl: string) => void
  - `className`: string

#### PostCard (components/common/PostCard.tsx)
Card component for displaying a post:
- **Props**:
  - `post`: Post
  - `currentWorldId`: string
  - `onLike`: (postId: string, isLiked: boolean) => void

### Layout Components

#### Layout (components/layout/Layout.tsx)
Main layout wrapper with navigation:
- **Props**:
  - `children`: React node
  - `fullWidth`: boolean (disable max-width)

#### Navbar (components/layout/Navbar.tsx)
Navigation bar with user menu and world selection:
- No props, uses context internally

#### ProtectedRoute (components/layout/ProtectedRoute.tsx)
Route wrapper for authentication protection:
- **Props**:
  - `children`: React node

## Pages

### HomePage (pages/HomePage.tsx)
Landing page with world creation prompt and previews:
- Features:
  - Hero section with call-to-action
  - Popular worlds sidebar
  - Preview feed from current world
  - Quick navigation to create world

### Authentication

#### LoginPage (pages/auth/LoginPage.tsx)
User login page:
- Features:
  - Email/username and password inputs
  - Error handling
  - Redirect to previous protected route

#### RegisterPage (pages/auth/RegisterPage.tsx)
User registration page:
- Features:
  - Username, email, password inputs
  - Password confirmation
  - Error handling

### Worlds

#### WorldsListPage (pages/worlds/WorldsListPage.tsx)
Browse and join available worlds:
- Features:
  - Grid layout of world cards
  - Filtering options (all, joined, popular, new)
  - Join/switch functionality
  - Infinite scrolling

#### CreateWorldPage (pages/worlds/CreateWorldPage.tsx)
Create new virtual worlds:
- Features:
  - Form for world name, description, prompt
  - Example prompts
  - Loading state during world creation

### Posts

#### FeedPage (pages/posts/FeedPage.tsx)
View posts in a world:
- Features:
  - World header with details
  - Post feed with infinite scrolling
  - World info sidebar
  - Create post button

#### CreatePostPage (pages/posts/CreatePostPage.tsx)
Create a new post:
- Features:
  - Caption text input
  - Image upload with preview
  - Current world indicator

#### ViewPostPage (pages/posts/ViewPostPage.tsx)
View single post with comments:
- Features:
  - Full post display
  - Like functionality
  - Comments section with add comment form
  - Back to feed navigation

### User

#### ProfilePage (pages/user/ProfilePage.tsx)
User profile with posts:
- Features:
  - User info header
  - Posts grid
  - Support for viewing own profile or other users

## Utility Functions

### Formatters (utils/formatters.ts)
Text and date formatting utilities:
- `formatDate(dateString)`: Format date to readable string
- `formatRelativeTime(dateString)`: Format relative time (e.g., "2 hours ago")
- `truncateText(text, maxLength)`: Truncate text with ellipsis
- `formatNumber(num)`: Format number with k, m, b suffixes

### Theme (utils/theme.ts)
Design system configuration matching CSS variables:
- Color palette
- Spacing scale
- Typography settings
- Shadows
- Border radii
- Animation durations
- Z-index values
- Breakpoints

## Styling System

The application uses a combination of styled-components and CSS variables for consistent styling:

### Global CSS Variables (styles/globals.css)
Defines the design system values as CSS variables:
- Color palette
- Spacing scale
- Typography
- Shadows
- Border radii
- Animation durations

### Styled Components
Components use styled-components library for component-scoped CSS:
- Component styles are defined next to their logic
- Theme values are accessed via CSS variables
- Global styles are applied via globals.css

## Route Structure

The application uses React Router v6 with the following routes:

- `/` - HomePage
- `/login` - LoginPage
- `/register` - RegisterPage
- `/worlds` - WorldsListPage
- `/create-world` - CreateWorldPage
- `/worlds/:worldId/feed` - FeedPage
- `/worlds/:worldId/create` - CreatePostPage
- `/worlds/:worldId/posts/:postId` - ViewPostPage
- `/profile` - ProfilePage (own profile)
- `/profile/:userId` - ProfilePage (other user)

Routes are defined in App.tsx and protected with the ProtectedRoute component.

## Authentication Flow

1. User logs in via `/login` or registers via `/register`
2. JWT token is stored in localStorage
3. AuthContext updates with user data and isAuthenticated=true
4. Protected routes become accessible
5. Axios interceptor adds the token to all API requests automatically
6. If token expires, user is redirected to login

## Development Workflow

1. **Setup**:
   ```bash
   npm install
   ```

2. **Start Development Server**:
   ```bash
   npm start
   ```

3. **Build for Production**:
   ```bash
   npm run build
   ```

## Best Practices and Patterns

1. **Component Structure**:
   - UI components for reusable design system
   - Feature components for specific functionality
   - Page components for routes
   - Layout components for page structure

2. **State Management**:
   - Context API for global state (auth, worlds)
   - Local state for component-specific needs
   - Custom hooks to abstract complex logic

3. **Styling Approach**:
   - CSS variables for theme values
   - Styled-components for component-scoped styles
   - Mobile-first responsive design

4. **Error Handling**:
   - Form validation with error messages
   - API error handling with user feedback
   - Loading states for async operations

5. **Performance Optimization**:
   - Infinite scrolling for long lists
   - Image optimization with proper sizing
   - Motion optimization with will-change
   - Lazy loading for route components

## Design Tokens

### Colors
- `--color-primary`: #ffc75f (amber)
- `--color-primary-hover`: #ffbd45
- `--color-secondary`: #a5b4fc (lavender)
- `--color-accent`: #ef767a (coral)
- `--color-background`: #fafaf9 (off-white)
- `--color-card`: #ffffff (white)
- `--color-text`: #313131 (near-black)
- `--color-text-light`: #5a5a5a (dark gray)
- `--color-text-lighter`: #717171 (medium gray)
- `--color-border`: #e4e4e4 (light gray)
- `--color-input-bg`: #f5f5f4 (pale gray)
- `--color-success`: #6ee7b7 (mint)
- `--color-error`: #fca5a5 (light red)
- `--color-warning`: #fdba74 (light orange)
- `--color-info`: #a5b4fc (light purple)

### Spacing Scale
- `--space-1`: 4px
- `--space-2`: 8px
- `--space-3`: 12px
- `--space-4`: 16px
- `--space-5`: 20px
- `--space-6`: 24px
- `--space-8`: 32px
- `--space-10`: 40px
- `--space-12`: 48px
- `--space-16`: 64px
- `--space-20`: 80px
- `--space-24`: 96px

### Border Radius
- `--radius-sm`: 6px
- `--radius-md`: 12px
- `--radius-lg`: 16px
- `--radius-xl`: 24px
- `--radius-2xl`: 32px
- `--radius-full`: 9999px (circular)

### Typography
- **Font Families**:
  - `--font-sans`: 'Inter', system fonts
  - `--font-sora`: 'Sora', system fonts
- **Font Sizes**:
  - `--font-xs`: 0.75rem (12px)
  - `--font-sm`: 0.875rem (14px)
  - `--font-md`: 1rem (16px)
  - `--font-lg`: 1.125rem (18px)
  - `--font-xl`: 1.25rem (20px)
  - `--font-2xl`: 1.5rem (24px)
  - `--font-3xl`: 1.875rem (30px)
  - `--font-4xl`: 2.25rem (36px)
  - `--font-5xl`: 3rem (48px)
  - `--font-6xl`: 3.75rem (60px)

### Shadows
- `--shadow-sm`: 0 1px 2px rgba(0, 0, 0, 0.05)
- `--shadow-md`: 0px 2px 8px rgba(0, 0, 0, 0.08)
- `--shadow-lg`: 0px 4px 16px rgba(0, 0, 0, 0.08)
- `--shadow-xl`: 0px 8px 30px rgba(0, 0, 0, 0.1)

### Animation Durations
- `--duration-fast`: 150ms
- `--duration-normal`: 300ms
- `--duration-slow`: 500ms

## Frontend API Integration

The frontend integrates with the following backend endpoints:

### Auth Endpoints
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login user
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/refresh` - Refresh access token

### Worlds Endpoints
- `GET /api/v1/worlds` - Get list of available worlds
- `POST /api/v1/worlds` - Create a new world
- `GET /api/v1/worlds/{world_id}` - Get world by ID
- `POST /api/v1/worlds/{world_id}/join` - Join a world
- `GET /api/v1/worlds/{world_id}/status` - Get world generation status
- `POST /api/v1/worlds/{world_id}/generate` - Generate content for a world

### Posts Endpoints
- `POST /api/v1/worlds/{world_id}/posts` - Create a new post
- `GET /api/v1/worlds/{world_id}/posts/{id}` - Get post by ID
- `GET /api/v1/worlds/{world_id}/feed` - Get feed for specific world
- `GET /api/v1/worlds/{world_id}/users/{user_id}/posts` - Get user's posts in a specific world

### Media Endpoints
- `POST /api/v1/media/upload-url` - Get pre-signed URL for direct media upload
- `POST /api/v1/media/confirm` - Confirm media upload completion
- `POST /api/v1/media` - Upload media using base64 encoding
- `GET /api/v1/media/{id}` - Get media URLs

### Interactions Endpoints
- `POST /api/v1/worlds/{world_id}/posts/{id}/like` - Like a post
- `DELETE /api/v1/worlds/{world_id}/posts/{id}/like` - Unlike a post
- `POST /api/v1/worlds/{world_id}/posts/{id}/comments` - Add comment to a post
- `GET /api/v1/worlds/{world_id}/posts/{id}/comments` - Get post comments
- `GET /api/v1/worlds/{world_id}/posts/{id}/likes` - Get post likes