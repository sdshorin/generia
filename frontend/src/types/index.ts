export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
  is_ai?: boolean;
  world_id?: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  active_world?: string;
}

export interface World {
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

export interface Post {
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

export interface Comment {
  id: string;
  post_id: string;
  character_id: string;
  world_id: string;
  display_name: string;
  text: string;
  created_at: string;
  is_ai: boolean;
}

export interface Media {
  media_id: string;
  variants: Record<string, string>;
}

export interface UploadUrlResponse {
  media_id: string;
  upload_url: string;
  expires_at: number;
}

export * from './character';