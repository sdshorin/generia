import React, { useState, useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axiosInstance from '../api/axios';
import { World } from '../types';
import { AuthContext } from '../context/AuthContext';

const WorldsList: React.FC = () => {
  const [worlds, setWorlds] = useState<World[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const { isAuthenticated } = useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    fetchWorlds();
  }, [isAuthenticated, navigate]);

  const fetchWorlds = async () => {
    try {
      setLoading(true);
      const limit = 10;
      const offset = (page - 1) * limit;
      const response = await axiosInstance.get(`/worlds?limit=${limit}&offset=${offset}`);
      
      // Log response for debugging
      console.log("Worlds response:", response.data);
      
      if (!response.data.worlds || response.data.worlds.length === 0) {
        setHasMore(false);
      } else {
        setWorlds(prevWorlds => [...prevWorlds, ...response.data.worlds]);
        setPage(prevPage => prevPage + 1);
      }
    } catch (err) {
      setError('Failed to load worlds');
      console.error("Error fetching worlds:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleJoinWorld = async (worldId: string) => {
    try {
      await axiosInstance.post(`/worlds/${worldId}/join`);
      await handleSetActiveWorld(worldId);
    } catch (err) {
      console.error('Failed to join world', err);
      setError('Failed to join world');
    }
  };

  const handleSetActiveWorld = async (worldId: string) => {
    try {
      await axiosInstance.post('/worlds/set-active', { world_id: worldId });
      navigate('/feed');
    } catch (err) {
      console.error('Failed to set active world', err);
      setError('Failed to set active world');
    }
  };
  
  const handleCreateWorld = () => {
    navigate('/create-world');
  };

  return (
    <div className="worlds-container">
      <h2>Choose a World</h2>
      {error && <div className="error">{error}</div>}

      <button onClick={handleCreateWorld} className="create-world-button">
        Create a New World
      </button>

      <div className="worlds-list">
        {worlds.map(world => (
          <div key={world.id} className="world-card">
            <h3>{world.name}</h3>
            <p className="world-description">{world.description}</p>
            <div className="world-stats">
              <span>{world.users_count} users</span>
              <span>{world.posts_count} posts</span>
            </div>
            <div className="world-actions">
              {world.is_joined ? (
                <button 
                  onClick={() => handleSetActiveWorld(world.id)}
                  className={`world-button ${world.is_active ? 'active' : ''}`}
                >
                  {world.is_active ? 'Current World' : 'Enter World'}
                </button>
              ) : (
                <button 
                  onClick={() => handleJoinWorld(world.id)}
                  className="world-button"
                >
                  Join World
                </button>
              )}
            </div>
            {world.generation_status === 'pending' && (
              <div className="world-status pending">Waiting for generation...</div>
            )}
            {world.generation_status === 'in_progress' && (
              <div className="world-status in-progress">Generating world...</div>
            )}
          </div>
        ))}
      </div>

      {loading && <div className="loading">Loading...</div>}

      {!loading && hasMore && (
        <button onClick={fetchWorlds} className="load-more-button">
          Load More
        </button>
      )}

      {!loading && !hasMore && worlds.length > 0 && (
        <div className="no-more-worlds">No more worlds</div>
      )}

      {!loading && worlds.length === 0 && (
        <div className="no-worlds">
          <p>No worlds available. Create your first world!</p>
        </div>
      )}
    </div>
  );
};

export default WorldsList;