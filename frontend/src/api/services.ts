import axios from 'axios';
import axiosInstance from './axios';
import { AuthResponse, World, Post, Comment, UploadUrlResponse, Media, WorldGenerationStatus } from '../types';

// Auth API
export const authAPI = {
  login: async (emailOrUsername: string, password: string): Promise<AuthResponse> => {
    const response = await axiosInstance.post<AuthResponse>('/auth/login', {
      email_or_username: emailOrUsername,
      password,
    });
    return response.data;
  },
  
  register: async (username: string, email: string, password: string): Promise<AuthResponse> => {
    const response = await axiosInstance.post<AuthResponse>('/auth/register', {
      username,
      email,
      password,
    });
    return response.data;
  },
  
  getCurrentUser: async () => {
    const response = await axiosInstance.get('/auth/me');
    return response.data;
  },
  
  refreshToken: async () => {
    const response = await axiosInstance.post('/auth/refresh');
    return response.data;
  }
};

// Worlds API
export const worldsAPI = {
  getWorlds: async (limit = 10, cursor = '') => {
    const response = await axiosInstance.get(`/worlds?limit=${limit}&cursor=${cursor}`);
    return response.data;
  },
  
  getWorldById: async (worldId: string) => {
    const response = await axiosInstance.get(`/worlds/${worldId}`);
    return response.data;
  },
  
  createWorld: async (name: string, description: string, prompt: string, charactersCount: number, postsCount: number): Promise<World> => {
    const response = await axiosInstance.post<World>('/worlds', {
      name,
      description,
      prompt,
      characters_count: charactersCount,
      posts_count: postsCount,
    });
    return response.data;
  },
  
  joinWorld: async (worldId: string) => {
    const response = await axiosInstance.post(`/worlds/${worldId}/join`);
    return response.data;
  },
  
  getWorldStatus: async (worldId: string): Promise<WorldGenerationStatus> => {
    const response = await axiosInstance.get<WorldGenerationStatus>(`/worlds/${worldId}/status`);
    return response.data;
  },

  createWorldStatusEventSource: (worldId: string): EventSource => {
    const token = localStorage.getItem('token');
    const url = new URL(`${axiosInstance.defaults.baseURL}/worlds/${worldId}/status/stream`);
    if (token) {
      url.searchParams.append('token', token);
    }
    return new EventSource(url.toString());
  },
  
  generateWorldContent: async (worldId: string) => {
    const response = await axiosInstance.post(`/worlds/${worldId}/generate`);
    return response.data;
  }
};

// Posts API
export const postsAPI = {
  getFeed: async (worldId: string, limit = 10, cursor = '') => {
    const response = await axiosInstance.get(`/worlds/${worldId}/posts?limit=${limit}&cursor=${cursor}`);
    return response.data;
  },
  
  getPostById: async (worldId: string, postId: string) => {
    const response = await axiosInstance.get(`/worlds/${worldId}/posts/${postId}`);
    return response.data;
  },
  
  getUserPosts: async (worldId: string, userId: string, limit = 10, offset = 0) => {
    const response = await axiosInstance.get(`/worlds/${worldId}/users/${userId}/posts?limit=${limit}&offset=${offset}`);
    return response.data;
  },
  
  createPost: async (worldId: string, caption: string, mediaId: string, characterId: string): Promise<Post> => {
    const response = await axiosInstance.post<Post>(`/worlds/${worldId}/post`, {
      caption,
      media_id: mediaId,
      character_id: characterId
    });
    return response.data;
  }
};

// Character API
export const characterAPI = {
  createCharacter: async (worldId: string, displayName: string, avatarMediaId?: string, meta?: string) => {
    const response = await axiosInstance.post(`/worlds/${worldId}/characters`, {
      display_name: displayName,
      avatar_media_id: avatarMediaId,
      meta
    });
    return response.data;
  },
  
  getCharacter: async (characterId: string) => {
    const response = await axiosInstance.get(`/characters/${characterId}`);
    return response.data;
  },
  
  getUserCharactersInWorld: async (worldId: string, userId: string) => {
    const response = await axiosInstance.get(`/worlds/${worldId}/users/${userId}/characters`);
    return response.data;
  },

  getCharacterPosts: async (worldId: string, characterId: string, limit = 20, offset = 0) => {
    const response = await axiosInstance.get(`/worlds/${worldId}/character/${characterId}/posts?limit=${limit}&offset=${offset}`);
    return response.data;
  }
};

// Media API
export const mediaAPI = {
  getUploadUrl: async (filename: string, contentType: string, size: number, worldId: string, characterId: string): Promise<UploadUrlResponse> => {
    const response = await axiosInstance.post<UploadUrlResponse>('/media/upload-url', {
      filename,
      content_type: contentType,
      size,
      world_id: worldId,
      character_id: characterId
    });
    return response.data;
  },
  
  uploadToUrl: async (url: string, file: File) => {
    return await axios.put(url, file, {
      headers: {
        'Content-Type': file.type,
      },
    });
  },
  
  confirmUpload: async (mediaId: string, characterId: string): Promise<Media> => {
    const response = await axiosInstance.post<Media>('/media/confirm', {
      media_id: mediaId,
      character_id: characterId
    });
    return response.data;
  },
  
  uploadBase64: async (mediaData: string, contentType: string, filename: string, worldId: string, characterId: string) => {
    const response = await axiosInstance.post('/media', {
      media_data: mediaData,
      content_type: contentType,
      filename,
      world_id: worldId,
      character_id: characterId
    });
    return response.data;
  },
  
  getMediaById: async (mediaId: string) => {
    const response = await axiosInstance.get(`/media/${mediaId}`);
    return response.data;
  }
};

// Interactions API
export const interactionsAPI = {
  likePost: async (worldId: string, postId: string) => {
    const response = await axiosInstance.post(`/worlds/${worldId}/posts/${postId}/like`);
    return response.data;
  },
  
  unlikePost: async (worldId: string, postId: string) => {
    const response = await axiosInstance.delete(`/worlds/${worldId}/posts/${postId}/like`);
    return response.data;
  },
  
  addComment: async (worldId: string, postId: string, text: string): Promise<Comment> => {
    const response = await axiosInstance.post<Comment>(`/worlds/${worldId}/posts/${postId}/comments`, {
      text,
    });
    return response.data;
  },
  
  getPostComments: async (worldId: string, postId: string) => {
    const response = await axiosInstance.get(`/worlds/${worldId}/posts/${postId}/comments`);
    return response.data;
  },
  
  getPostLikes: async (worldId: string, postId: string) => {
    const response = await axiosInstance.get(`/worlds/${worldId}/posts/${postId}/likes`);
    return response.data;
  }
};