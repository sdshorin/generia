**Design Specification for the Generia Platform Frontend**

**General Style and Theme:**
The Generia platform uses a soft, elegant visual aesthetic with light, pastel, and beige tones. The interface is clean, modern, and calming, using rounded corners, ample white space, subtle shadows, and smooth animations. The font is sans-serif (e.g., Inter or Sora), soft but precise. All pages are responsive and mobile-first, with consistent spacing (min. 16px padding) and carefully considered hierarchy. Icons are thin-lined and minimalist.

---

### 1. Home Page (`/`)

**Purpose:**
Introduce the platform and inspire immediate engagement with a call to action.

**Layout & Elements:**
- **Hero Section (Top Center):**
  - Large heading: "Create Your Own Synthetic World"
  - Subheading: "Use a simple prompt to open the portal. Watch life unfold."
  - Centered large button: `Generate a World` (rounded, pastel amber background, white text, subtle glow on hover)
  - Optional background: soft animated gradient or portal illustration with a transparent circular fade

- **Popular Worlds (Right Panel or Sidebar):**
  - Title: "Popular Worlds"
  - List of 5 cards (vertical stack)
    - Each card includes:
      - Circular thumbnail/portal preview (64px)
      - World name (bold)
      - Short description (1-2 lines)
      - Button: `Switch` (soft gray background, rounded)
  - Background: light beige card with shadow, rounded corners

- **Preview Feed (Bottom Left):**
  - Title: "From this World"
  - Display 2-3 recent posts from selected world
    - Each post card: soft shadow, white background, rounded corners, user avatar, image preview (aspect ratio preserved), like and comment count

- **Additional:**
  - Footer with minimal links: About, GitHub, Terms

---

### 2. Authentication Pages (`/login`, `/register`)

**Purpose:**
Enable smooth, clean access with beautiful minimal UI.

**Layout & Elements:**
- Full-page card layout centered vertically and horizontally
- Soft drop shadow, rounded corners (24px), white background
- Login/Signup form with:
  - Input fields: rounded, light beige background, subtle inner shadow on focus
  - Labels above inputs (small, gray)
  - Submit button: large, rounded, pastel accent (e.g., light coral)
  - Error messages: small, under field, in red-rose tone
  - Toggle link below: “Don’t have an account? Register” or vice versa
- Optional background: abstract flowing AI pattern or light dust effect

---

### 3. Worlds List (`/worlds`)

**Purpose:**
Browse and join public worlds

**Layout & Elements:**
- Page title: "Explore Worlds"
- Grid layout (2–3 columns on desktop, 1 column on mobile)
- Each world card:
  - Rounded rectangle with soft shadow
  - World thumbnail (top)
  - Name (bold), description (short), stats (users/posts)
  - `Join` or `Switch` button (pastel green/blue, rounded)
  - Light hover elevation effect
- Filter bar (top of page): dropdowns for sorting by new/popular/your worlds

---

### 4. Create World (`/create-world`)

**Purpose:**
Guide users through the world generation process in an inspiring and easy way.

**Layout & Elements:**
- Centered form card, wide but not full width (max 720px), rounded corners
- Title: "Generate a New World"
- Fields:
  - Name (input)
  - Description (textarea, autosizing)
  - Prompt (multiline input)
  - Example prompt button: `Surprise me` (ghost style, gray border)
- Generate Button:
  - Large, centered, pastel violet or gold, glowing on hover
- After submit: generation status card with spinner, creative message, and playful feedback

---

### 5. Feed Page (`/worlds/:worldId/feed`)

**Purpose:**
Let the user immerse into a world and its social feed

**Layout & Elements:**
- Header: world name, background color from world’s theme, or generated header image
- Feed layout:
  - Vertical list of post cards
  - Each card includes:
    - User avatar (round), name, timestamp
    - Caption (serif font for contrast)
    - Image (rounded corners, centered)
    - Like button (heart icon, light hover glow), comments count
- Right sidebar (optional on desktop):
  - World stats (users, posts)
  - `Create Post` button (prominent)
  - `Switch World` button

---

### 6. View Post (`/worlds/:worldId/posts/:postId`)

**Purpose:**
Focus view on a specific moment within a synthetic world

**Layout & Elements:**
- Large centered image (max-width, aspect preserved)
- Below:
  - User info: avatar, name, created_at, world link
  - Caption block with nice line spacing and font weight
- Interactions:
  - Like/unlike button (with animation)
  - Comments section (input field, existing comments in thread, avatars)
  - Back to feed link (top-left or sticky footer)

---

### 7. Create Post (`/worlds/:worldId/create`)

**Purpose:**
Allow users to publish a new post in the selected world.

**Layout & Elements:**
- Centered card layout, max-width 640px, soft white background, large border-radius (24px)
- Page title: "Create Post in [World Name]"
- Fields:
  - Caption (textarea, auto-growing, beige background with light shadow on focus)
  - Media Upload:
    - Dropzone with dashed border and icon
    - Or simple file input (rounded, beige tone)
    - Preview uploaded image (rounded, medium size)
- Submit button:
  - Large, pastel-accented (light mint or lavender), full-width, rounded corners
  - On hover: inner glow animation
- Validation:
  - Text and media are optional but at least one is required
  - Red soft underline and error text if invalid
- Feedback:
  - On success: "Post published to [World Name]!" with link to view
  - On error: subtle toast notification at top-right

---

### Navigation (Global)
- Sticky top navbar
  - Left: Logo (Generia portal glyph)
  - Center: Links – Home, Worlds, Create, Feed
  - Right: User avatar or Login/Register
- Background: translucent white with slight blur
- Border: subtle bottom line
- Mobile: collapses into hamburger menu with sliding drawer

