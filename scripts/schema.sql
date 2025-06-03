-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Worlds table (used by world-service)
CREATE TABLE IF NOT EXISTS worlds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    prompt TEXT NOT NULL,
    creator_id UUID,
    generation_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, in_progress, completed, failed
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, archived
    image_uuid UUID, -- UUID of the background image
    icon_uuid UUID, -- UUID of the world icon image
    params JSONB,
    users_count INTEGER DEFAULT 0, -- Actual number of AI users/characters
    posts_count INTEGER DEFAULT 0, -- Actual number of AI posts
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Users table (used by auth-service)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(30) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE, -- Allow NULL for AI users
    password_hash VARCHAR(255), -- Allow NULL for AI users
    -- todo - add credits
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


-- /api/v1/worlds/{world_id}/join - POST
-- /api/v1/worlds - POST
-- Пока что не используются - все миры видны всем пользователям (и так и хочется оставить. Хочется, чтобы часть миров были открытыми, а часть - приватными)
-- в будущем будет заменено таблицей world_memberships
-- User worlds table (many-to-many relationship for real users and worlds they can access)
CREATE TABLE IF NOT EXISTS user_worlds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, world_id)
);

-- Refresh tokens table (used by auth-service)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- World user characters table (used by character-service)
CREATE TABLE IF NOT EXISTS world_user_characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    real_user_id UUID REFERENCES users(id) ON DELETE SET NULL,    -- NULL => AI-NPC
    is_ai BOOLEAN GENERATED ALWAYS AS (real_user_id IS NULL) STORED,
    display_name TEXT NOT NULL,
    avatar_media_id UUID,
    meta JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Posts table (used by post-service)
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID REFERENCES world_user_characters(id) ON DELETE SET NULL,
    is_ai BOOLEAN NOT NULL DEFAULT FALSE,
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    caption TEXT,
    media_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


-- TODO: Разобарться как хранить медиа (все медиа - в s3)
-- Media table (used by media-service)
CREATE TABLE IF NOT EXISTS media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID, -- Nullable for world-level media
    world_id UUID NOT NULL REFERENCES worlds(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    bucket TEXT NOT NULL,
    object_name TEXT NOT NULL,
    media_type INTEGER NOT NULL DEFAULT 0, -- 0=unknown, 1=world_header, 2=world_icon, 3=character_avatar, 4=post_image
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Media variants table (used by media-service)
-- TODO: Разобарться как хранить медиа (все медиа - в s3)
CREATE TABLE IF NOT EXISTS media_variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    width INT NOT NULL,
    height INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);



-- Create indexes
CREATE INDEX IF NOT EXISTS idx_worlds_creator_id ON worlds(creator_id);
CREATE INDEX IF NOT EXISTS idx_worlds_status ON worlds(status);
CREATE INDEX IF NOT EXISTS idx_user_worlds_user_id ON user_worlds(user_id);
CREATE INDEX IF NOT EXISTS idx_user_worlds_world_id ON user_worlds(world_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- World user characters indexes
CREATE INDEX IF NOT EXISTS idx_world_user_characters_real_user_id ON world_user_characters(real_user_id);
CREATE INDEX IF NOT EXISTS idx_world_user_characters_world_id ON world_user_characters(world_id);
CREATE INDEX IF NOT EXISTS idx_world_user_characters_real_user_id_world_id ON world_user_characters(real_user_id, world_id);
CREATE INDEX IF NOT EXISTS idx_world_user_characters_is_ai ON world_user_characters(is_ai);

-- Posts indexes
CREATE INDEX IF NOT EXISTS idx_posts_character_id ON posts(character_id);
CREATE INDEX IF NOT EXISTS idx_posts_world_id ON posts(world_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
CREATE INDEX IF NOT EXISTS idx_posts_is_ai ON posts(is_ai);

-- Media indexes
CREATE INDEX IF NOT EXISTS idx_media_character_id ON media(character_id);
CREATE INDEX IF NOT EXISTS idx_media_world_id ON media(world_id);
CREATE INDEX IF NOT EXISTS idx_media_variants_media_id ON media_variants(media_id);


INSERT INTO users (id,username,email,password_hash,created_at,updated_at) VALUES
	 ('c35f05b3-16c6-4410-a18a-73aa5ed1a685'::uuid,'ser','serres123@yandex.ru','$2a$10$PAyEZQh7UrJ09B/FqQQDEO/4hHy5I9Mp99QUPmy/qhwl8i6CAZjwS','2025-06-01 15:31:46.961186+03','2025-06-01 15:31:46.961186+03')
    ON CONFLICT (id) DO NOTHING;
