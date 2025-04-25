import React, { createContext, useState, useEffect, ReactNode, useCallback, useContext, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { World } from '../types';
import { worldsAPI } from '../api/services';
import { AuthContext } from './AuthContext';

interface WorldContextType {
  worlds: World[];
  currentWorld: World | null;
  isLoading: boolean;
  error: string | null;
  loadWorlds: (limit?: number, cursor?: string) => Promise<void>;
  createWorld: (name: string, description: string, prompt: string) => Promise<World>;
  joinWorld: (worldId: string) => Promise<void>;
  setCurrentWorld: (world: World | null) => void;
  loadCurrentWorld: (worldId: string) => Promise<void>;
  clearError: () => void;
}

interface WorldProviderProps {
  children: ReactNode;
}

const initialState: WorldContextType = {
  worlds: [],
  currentWorld: null,
  isLoading: false,
  error: null,
  loadWorlds: async () => {},
  createWorld: async () => ({} as World),
  joinWorld: async () => {},
  setCurrentWorld: () => {},
  loadCurrentWorld: async () => {},
  clearError: () => {},
};

export const WorldContext = createContext<WorldContextType>(initialState);

export const WorldProvider: React.FC<WorldProviderProps> = ({ children }) => {
  const [worlds, setWorlds] = useState<World[]>([]);
  const [currentWorld, setCurrentWorld] = useState<World | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { isAuthenticated } = useContext(AuthContext);
  const navigate = useNavigate();
  const location = useLocation();

  // Clear error state
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Load worlds from API
  const loadWorlds = useCallback(async (limit = 10, cursor = '') => {
    if (!isAuthenticated) return;
    
    try {
      setIsLoading(true);
      setError(null);
      const data = await worldsAPI.getWorlds(limit, cursor);
      setWorlds(data.worlds || []);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load worlds');
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  // Create a new world
  const createWorld = async (name: string, description: string, prompt: string): Promise<World> => {
    setIsLoading(true);
    setError(null);
    
    try {
      const newWorld = await worldsAPI.createWorld(name, description, prompt);
      setWorlds(prevWorlds => [...prevWorlds, newWorld]);
      return newWorld;
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create world');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  // Join a world
  const joinWorld = async (worldId: string) => {
    setIsLoading(true);
    setError(null);
    
    try {
      await worldsAPI.joinWorld(worldId);
      // Update the world in the list to show it's joined
      setWorlds(prevWorlds => 
        prevWorlds.map(world => 
          world.id === worldId ? { ...world, is_joined: true } : world
        )
      );
      
      // If this is the current world, update it too
      if (currentWorld && currentWorld.id === worldId) {
        setCurrentWorld(prev => prev ? { ...prev, is_joined: true } : null);
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to join world');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  // Load a specific world by ID
  const loadCurrentWorld = async (worldId: string) => {
    if (!worldId || !isAuthenticated) return;
    
    try {
      setIsLoading(true);
      setError(null);
      const world = await worldsAPI.getWorldById(worldId);
      setCurrentWorld(world);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load world');
      navigate('/worlds');
    } finally {
      setIsLoading(false);
    }
  };

  // Try to extract worldId from URL path when it changes
  const previousPathRef = useRef(location.pathname);
  
  useEffect(() => {
    if (!isAuthenticated) return;
    
    const path = location.pathname;
    
    // Skip if the path hasn't changed to prevent unnecessary calls
    if (previousPathRef.current === path) return;
    previousPathRef.current = path;
    
    const match = path.match(/\/worlds\/([^/]+)/);
    
    if (match && match[1]) {
      const worldId = match[1];
      if (!currentWorld || currentWorld.id !== worldId) {
        loadCurrentWorld(worldId);
      }
    } else if (!path.includes('/worlds/')) {
      setCurrentWorld(null);
    }
  }, [location.pathname, isAuthenticated, currentWorld, loadCurrentWorld]);

  // Initial load of worlds when authenticated
  useEffect(() => {
    if (isAuthenticated) {
      loadWorlds();
    }
  }, [isAuthenticated, loadWorlds]);

  return (
    <WorldContext.Provider
      value={{
        worlds,
        currentWorld,
        isLoading,
        error,
        loadWorlds,
        createWorld,
        joinWorld,
        setCurrentWorld,
        loadCurrentWorld,
        clearError,
      }}
    >
      {children}
    </WorldContext.Provider>
  );
};