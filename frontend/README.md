# Generia Frontend: Technical Documentation for LLM Agents

## Project Status
**PRODUCTION-READY FRONTEND**
- Complete TypeScript integration with resolved type errors
- 7/10 pages fully integrated with backend APIs
- 2/10 pages partially integrated (using mock data for missing fields)
- 1/10 page requires new backend APIs
- Build compiles without errors
- Responsive design implemented

---

## Quick Orientation for LLM Agents

### Essential Files for Understanding Architecture
1. **`src/types/index.ts`** - START HERE: All TypeScript data interfaces
2. **`src/api/services.ts`** - Backend API integration and methods
3. **`src/App.tsx`** - Application routing configuration
4. **`src/context/`** - Global state management (Auth, World)
5. **`BACKEND_TODO.md`** - Missing backend data requirements

### Project Structure by Priority
```
frontend/src/
├── types/                      # CRITICAL: TypeScript interfaces
│   ├── index.ts               # User, World, Post, Comment, Media
│   └── character.ts           # Character interface
├── api/                       # CRITICAL: Backend integration
│   ├── services.ts            # API methods and endpoints
│   └── axios.ts               # HTTP client configuration
├── pages/                     # APPLICATION PAGES
│   ├── HomePage.tsx           # Landing page (full integration)
│   ├── auth/                  # Authentication pages
│   │   ├── LoginPage.tsx      # User login (full integration)
│   │   └── RegisterPage.tsx   # User registration (full integration)
│   ├── worlds/                # World management pages
│   │   ├── WorldsListPage.tsx # World catalog (full integration)
│   │   ├── CreateWorldPage.tsx # World creation (full integration)
│   │   └── WorldAboutPage.tsx # World details (partial integration)
│   ├── posts/                 # Post-related pages
│   │   ├── FeedPage.tsx       # World feed (full integration)
│   │   ├── ViewPostPage.tsx   # Post details (full integration)
│   │   ├── CreatePostPage.tsx # Post creation (full integration)
│   │   └── CreateCharacterPage.tsx # Character creation
│   ├── characters/            # Character pages
│   │   └── CharacterPage.tsx  # Character profile (partial integration)
│   └── user/                  # User profile pages
│       ├── ProfilePage.tsx    # User profile
│       └── SettingsPage.tsx   # User settings (mock data only)
├── components/                # REUSABLE COMPONENTS
│   ├── layout/                # Layout components
│   │   ├── Header.tsx         # Main navigation header
│   │   ├── Layout.tsx         # Page layout wrapper
│   │   ├── Navbar.tsx         # Navigation bar
│   │   └── ProtectedRoute.tsx # Authentication guard
│   ├── cards/                 # Card components
│   │   ├── PostCard.tsx       # Post display card
│   │   ├── WorldCard.tsx      # World display card
│   │   ├── CharacterCard.tsx  # Character display card
│   │   └── CommentCard.tsx    # Comment display card
│   ├── ui/                    # UI components (design system)
│   │   ├── Avatar.tsx         # User avatar component
│   │   ├── Button.tsx         # Button component
│   │   ├── Card.tsx           # Generic card container
│   │   ├── Input.tsx          # Text input component
│   │   ├── Loader.tsx         # Loading indicator
│   │   └── TextArea.tsx       # Multi-line text input
│   └── common/                # Feature-specific components
│       ├── ImageUpload.tsx    # Media upload component
│       ├── PostCard.tsx       # Enhanced post card
├── context/                   # STATE MANAGEMENT
│   ├── AuthContext.tsx        # Authentication state
│   └── WorldContext.tsx       # World management state
├── hooks/                     # CUSTOM REACT HOOKS
│   ├── useAuth.ts             # Authentication hook
│   ├── useWorld.ts            # World data hook
│   ├── useInfiniteScroll.ts   # Infinite scrolling
│   └── useMediaUpload.ts      # Media upload hook
├── styles/                    # STYLING SYSTEM
│   ├── globals.css            # CSS variables and base styles
│   ├── components.css         # Component styles
│   └── pages/                 # Page-specific styles
│       ├── main.css           # HomePage styles
│       ├── catalog.css        # WorldsListPage styles
│       ├── create-world.css   # CreateWorldPage styles
│       ├── feed.css           # FeedPage and ViewPostPage styles
│       ├── world-about.css    # WorldAboutPage styles
│       ├── auth.css           # Authentication pages styles
│       ├── character-profile.css # CharacterPage styles
│       └── settings.css       # SettingsPage styles
└── utils/                     # UTILITIES
    ├── formatters.ts          # Date/text formatting functions
    ├── theme.ts               # Theme configuration
    └── mockData.ts            # Mock data for development
```

---

## API Integration Status

### FULL INTEGRATION (7 pages)
- **HomePage.tsx**: `worldsAPI.getWorlds()`, `postsAPI.getFeed()`
- **LoginPage.tsx**: `authAPI.login()`
- **RegisterPage.tsx**: `authAPI.register()`
- **WorldsListPage.tsx**: `worldsAPI.getWorlds()`, `worldsAPI.joinWorld()`
- **CreateWorldPage.tsx**: `worldsAPI.createWorld()`, Server-Sent Events for progress
- **FeedPage.tsx**: `postsAPI.getFeed()`, `worldsAPI.getWorldById()`
- **ViewPostPage.tsx**: `postsAPI.getPostById()`, `interactionsAPI.*`

### PARTIAL INTEGRATION (2 pages)
- **WorldAboutPage.tsx**: Basic world data + mocks for detailed information
- **CharacterPage.tsx**: Basic character data + mocks for biography

### REQUIRES NEW APIs (1 page)
- **SettingsPage.tsx**: Uses mock data from `utils/mockData.ts`

---

## Core Data Types

### User (types/index.ts)
```typescript
interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
  is_ai?: boolean;
  world_id?: string;
  avatar_url?: string;
}
```

### World (types/index.ts)
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
  image_url?: string;
  icon_url?: string;
}
```

### Post (types/index.ts)
```typescript
interface Post {
  id: string;
  character_id: string;
  world_id: string;
  display_name: string;
  caption: string;
  image_url?: string;
  media_url?: string;
  avatar_url?: string;
  likes_count: number;
  comments_count: number;
  created_at: string;
  updated_at?: string;
  user_liked?: boolean;
  is_ai: boolean;
}
```

### Comment (types/index.ts)
```typescript
interface Comment {
  id: string;
  post_id: string;
  character_id: string;
  world_id: string;
  display_name: string;
  text: string;
  created_at: string;
  is_ai: boolean;
  avatar_url?: string;
}
```

### Character (types/character.ts)
```typescript
interface Character {
  id: string;
  world_id: string;
  real_user_id?: string;
  is_ai: boolean;
  display_name: string;
  avatar_media_id?: string;
  avatar_url?: string;
  meta?: string;
  created_at: string;
  role?: string;
}
```

### Media (types/index.ts)
```typescript
interface Media {
  media_id: string;
  variants: Record<string, string>;
}

interface UploadUrlResponse {
  media_id: string;
  upload_url: string;
  expires_at: number;
}
```

### World Generation Status (types/index.ts)
```typescript
interface WorldGenerationStatus {
  status: string;
  current_stage: string;
  stages: StageInfo[];
  tasks_total: number;
  tasks_completed: number;
  tasks_failed: number;
  task_predicted: number;
  users_created: number;
  posts_created: number;
  users_predicted: number;
  posts_predicted: number;
  api_call_limits_llm: number;
  api_call_limits_images: number;
  api_calls_made_llm: number;
  api_calls_made_images: number;
  llm_cost_total: number;
  image_cost_total: number;
  created_at: string;
  updated_at: string;
}
```

---

## API Services (api/services.ts)

### authAPI
- `login(emailOrUsername: string, password: string)`: Authenticate user
- `register(username: string, email: string, password: string)`: Register new user
- `getCurrentUser()`: Get current authenticated user
- `refreshToken()`: Refresh authentication token

### worldsAPI
- `getWorlds(limit?: number, offset?: number)`: Get paginated list of worlds
- `getWorldById(worldId: string)`: Get specific world details
- `createWorld(name: string, description: string, prompt: string)`: Create new world
- `joinWorld(worldId: string)`: Join existing world
- `getWorldStatus(worldId: string)`: Get world generation status
- `createWorldStatusEventSource(worldId: string)`: Server-Sent Events for real-time status

### postsAPI
- `getFeed(worldId: string, limit?: number, offset?: number)`: Get world feed posts
- `getPostById(worldId: string, postId: string)`: Get specific post
- `getUserPosts(worldId: string, userId: string)`: Get user's posts in world
- `createPost(worldId: string, caption: string, mediaId?: string)`: Create new post

### characterAPI
- `createCharacter(worldId: string, data: any)`: Create new character
- `getCharacter(characterId: string)`: Get character details
- `getUserCharactersInWorld(worldId: string, userId: string)`: Get user's characters
- `getCharacterPosts(characterId: string)`: Get character's posts

### mediaAPI
- `getUploadUrl(filename: string, contentType: string, size: number, worldId: string)`: Get upload URL
- `uploadToUrl(url: string, file: File)`: Upload file to pre-signed URL
- `confirmUpload(mediaId: string)`: Confirm upload completion
- `uploadBase64(mediaData: string, contentType: string, filename: string, worldId: string)`: Upload base64 media
- `getMediaById(mediaId: string)`: Get media details

### interactionsAPI
- `likePost(worldId: string, postId: string)`: Like a post
- `unlikePost(worldId: string, postId: string)`: Unlike a post
- `addComment(worldId: string, postId: string, text: string)`: Add comment to post
- `getPostComments(worldId: string, postId: string)`: Get post comments
- `getPostLikes(worldId: string, postId: string)`: Get post likes

---

## State Management

### AuthContext (context/AuthContext.tsx)
```typescript
const { 
  user, 
  isAuthenticated, 
  login, 
  register, 
  logout, 
  isLoading, 
  error 
} = useAuth();
```

### WorldContext (context/WorldContext.tsx)
```typescript
const { 
  worlds, 
  currentWorld, 
  createWorld, 
  joinWorld, 
  loadWorlds, 
  isLoading, 
  error 
} = useWorld();
```

---

## Custom Hooks

### useAuth (hooks/useAuth.ts)
Authentication management hook
```typescript
import { useAuth } from '../hooks/useAuth';
const { user, isAuthenticated, login, register, logout } = useAuth();
```

### useWorld (hooks/useWorld.ts)
World data management hook
```typescript
import { useWorld } from '../hooks/useWorld';
const { worlds, currentWorld, createWorld, joinWorld } = useWorld();
```

### useInfiniteScroll (hooks/useInfiniteScroll.ts)
Infinite scrolling implementation
```typescript
import { useInfiniteScroll } from '../hooks/useInfiniteScroll';
const { items, isLoading, sentinelRef } = useInfiniteScroll({
  fetchItems: async (page, limit) => {
    // Fetch function returning array of items
  },
  limit: 10
});
```

### useMediaUpload (hooks/useMediaUpload.ts)
Media upload functionality
```typescript
import { useMediaUpload } from '../hooks/useMediaUpload';
const { uploadMedia, isUploading, progress, media } = useMediaUpload({
  worldId: currentWorld.id
});
```

---

## Component Library

### UI Components (components/ui/)

#### Button (Button.tsx)
```typescript
interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'accent' | 'ghost' | 'text';
  size?: 'small' | 'medium' | 'large';
  fullWidth?: boolean;
  icon?: React.ReactNode;
  isLoading?: boolean;
  children: React.ReactNode;
  onClick?: () => void;
}
```

#### Input (Input.tsx)
```typescript
interface InputProps {
  label?: string;
  error?: string;
  icon?: React.ReactNode;
  fullWidth?: boolean;
  type?: string;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}
```

#### TextArea (TextArea.tsx)
```typescript
interface TextAreaProps {
  label?: string;
  error?: string;
  fullWidth?: boolean;
  rows?: number;
  maxRows?: number;
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
}
```

#### Avatar (Avatar.tsx)
```typescript
interface AvatarProps {
  src?: string;
  name?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  isAi?: boolean;
}
```

#### Loader (Loader.tsx)
```typescript
interface LoaderProps {
  size?: 'sm' | 'md' | 'lg';
  variant?: 'primary' | 'secondary' | 'light' | 'dark';
  text?: string;
  fullScreen?: boolean;
}
```

### Feature Components (components/common/)

#### ImageUpload (ImageUpload.tsx)
```typescript
interface ImageUploadProps {
  worldId: string;
  onUploadComplete: (mediaId: string, mediaUrl: string) => void;
  className?: string;
}
```

#### PostCard (PostCard.tsx)
```typescript
interface PostCardProps {
  post: Post;
  currentWorldId: string;
  onLike?: (postId: string, isLiked: boolean) => void;
}
```

### Card Components (components/cards/)

#### WorldCard (WorldCard.tsx)
```typescript
interface WorldCardProps {
  world: World;
  onEnter?: (worldId: string) => void;
}
```

#### CharacterCard (CharacterCard.tsx)
```typescript
interface CharacterCardProps {
  character: Character;
  onClick?: (characterId: string) => void;
}
```

#### CommentCard (CommentCard.tsx)
```typescript
interface CommentCardProps {
  comment: Comment;
  currentWorldId: string;
  isReply?: boolean;
  onReply?: (commentId: string, text: string) => void;
}
```

### Layout Components (components/layout/)

#### Layout (Layout.tsx)
```typescript
interface LayoutProps {
  children: React.ReactNode;
  fullWidth?: boolean;
}
```

#### Header (Header.tsx)
Main navigation component with user menu and world selection

#### ProtectedRoute (ProtectedRoute.tsx)
```typescript
interface ProtectedRouteProps {
  children: React.ReactNode;
}
```

---

## Application Routing (App.tsx)

```
/ → HomePage (public)
/login → LoginPage
/register → RegisterPage
/worlds → WorldsListPage (protected)
/create-world → CreateWorldPage (protected)
/worlds/:worldId/feed → FeedPage (protected)
/worlds/:worldId/posts/:postId → ViewPostPage (protected)
/worlds/:worldId/about → WorldAboutPage (protected)
/characters/:characterId → CharacterPage (protected)
/settings → SettingsPage (protected)
```

---

## Styling System

### CSS Architecture
- **`styles/globals.css`** - CSS variables, reset, base styles
- **`styles/components.css`** - Component styles (header, cards, forms)
- **`styles/pages/`** - Page-specific styles

### CSS Variables (globals.css)
```css
/* Colors */
--color-primary: #ffc75f;        /* Amber */
--color-secondary: #a5b4fc;      /* Lavender */
--color-accent: #ef767a;         /* Coral */
--color-background: #fafaf9;     /* Off-white */
--color-card: #ffffff;           /* White */
--color-text: #313131;           /* Dark gray */
--color-text-light: #5a5a5a;     /* Medium gray */
--color-border: #e4e4e4;         /* Light gray */
--color-input-bg: #f5f5f4;       /* Pale gray */
--color-success: #6ee7b7;        /* Mint */
--color-error: #fca5a5;          /* Light red */

/* Spacing */
--space-1: 4px;
--space-2: 8px;
--space-3: 12px;
--space-4: 16px;
--space-6: 24px;
--space-8: 32px;
--space-12: 48px;
--space-16: 64px;

/* Border Radius */
--radius-sm: 6px;
--radius-md: 12px;
--radius-lg: 16px;
--radius-xl: 24px;
--radius-full: 9999px;

/* Shadows */
--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
--shadow-md: 0px 2px 8px rgba(0, 0, 0, 0.08);
--shadow-lg: 0px 4px 16px rgba(0, 0, 0, 0.08);
--shadow-xl: 0px 8px 30px rgba(0, 0, 0, 0.1);
```

---

## Backend API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/refresh` - Refresh access token

### Worlds
- `GET /api/v1/worlds` - Get list of available worlds
- `POST /api/v1/worlds` - Create new world
- `GET /api/v1/worlds/{world_id}` - Get world by ID
- `POST /api/v1/worlds/{world_id}/join` - Join world
- `GET /api/v1/worlds/{world_id}/status` - Get world generation status
- `POST /api/v1/worlds/{world_id}/generate` - Generate content for world

### Characters
- `POST /api/v1/worlds/{world_id}/characters` - Create character in world
- `GET /api/v1/characters/{character_id}` - Get character by ID
- `GET /api/v1/worlds/{world_id}/users/{user_id}/characters` - Get user's characters

### Posts
- `POST /api/v1/worlds/{world_id}/posts` - Create new post
- `GET /api/v1/worlds/{world_id}/posts/{id}` - Get post by ID
- `GET /api/v1/worlds/{world_id}/feed` - Get feed for world
- `GET /api/v1/worlds/{world_id}/users/{user_id}/posts` - Get user's posts

### Media
- `POST /api/v1/media/upload` - Upload media
- `POST /api/v1/media/upload-url` - Get pre-signed URL for upload
- `POST /api/v1/media/confirm` - Confirm media upload completion
- `POST /api/v1/media` - Upload media using base64 encoding
- `GET /api/v1/media/{id}` - Get media URLs

### Interactions
- `POST /api/v1/worlds/{world_id}/posts/{id}/like` - Like post
- `DELETE /api/v1/worlds/{world_id}/posts/{id}/like` - Unlike post
- `POST /api/v1/worlds/{world_id}/posts/{id}/comments` - Add comment
- `GET /api/v1/worlds/{world_id}/posts/{id}/comments` - Get post comments
- `GET /api/v1/worlds/{world_id}/posts/{id}/likes` - Get post likes

---

## Important Notes for LLM Agents

### What Works
1. **Complete TypeScript integration** - All interfaces are current and accurate
2. **Component architecture** - Reusable UI components with proper props
3. **API integration** - 7/10 pages fully integrated with backend
4. **Routing system** - All routes configured and protected
5. **Authentication flow** - Complete JWT-based authentication
6. **State management** - Context API with custom hooks
7. **Responsive design** - Mobile-first approach implemented

### Potential Issues
1. **Type fields**: Only use existing fields (see types/index.ts)
2. **API endpoints**: Verify method exists in api/services.ts
3. **Mock data**: Some pages use fallback data for missing fields
4. **Images**: All fallbacks use `/no-image.jpg`
5. **Character vs User**: Posts and Comments use character_id, not user_id

### Development Guidelines
1. **Start with types/index.ts** - Verify all interfaces before coding
2. **Check api/services.ts** - Understand available API methods
3. **Read BACKEND_TODO.md** - List of missing backend data
4. **Use existing components** - Leverage components/ directory
5. **Follow patterns** - Reference working pages for consistency

### Adding New Features
1. **New page**: Create component in pages/, add route in App.tsx
2. **API integration**: Check api/services.ts, add method if needed
3. **Type updates**: Update types/index.ts if new fields required
4. **State management**: Use useAuth and useWorld hooks
5. **Styling**: Follow CSS variable system in globals.css

---

This documentation serves as the primary reference for understanding the Generia frontend architecture. All components, types, and API methods are production-ready and actively maintained.