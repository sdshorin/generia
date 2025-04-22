export interface Character {
  id: string;
  world_id: string;
  real_user_id?: string;
  is_ai: boolean;
  display_name: string;
  avatar_media_id?: string;
  meta?: string;
  created_at: string;
}

export interface CharacterListResponse {
  characters: Character[];
}