import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Layout } from '../../components/layout/Layout';
import { Loader } from '../../components/ui/Loader';
import { useAuth } from '../../hooks/useAuth';
import { useWorld } from '../../hooks/useWorld';
import { worldsAPI } from '../../api/services';
import { World } from '../../types';
import '../../styles/pages/catalog.css';

export const WorldsListPage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const { joinWorld } = useWorld();
  const [worlds, setWorlds] = useState<World[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isJoining, setIsJoining] = useState<Record<string, boolean>>({});
  const navigate = useNavigate();

  // Default world data for demo
  const defaultWorlds: World[] = [
    
  ];

  useEffect(() => {
    const fetchWorlds = async () => {
      try {
        setIsLoading(true);
        setError(null);
        
        if (isAuthenticated) {
          const data = await worldsAPI.getWorlds(10, '');
          setWorlds(data.worlds || defaultWorlds);
        } else {
          setWorlds(defaultWorlds);
        }
      } catch (error) {
        console.error('Failed to fetch worlds:', error);
        setError('Failed to load worlds');
        setWorlds(defaultWorlds);
      } finally {
        setIsLoading(false);
      }
    };

    fetchWorlds();
  }, [isAuthenticated]);

  const handleJoinWorld = async (worldId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    setIsJoining(prev => ({ ...prev, [worldId]: true }));

    try {
      await joinWorld(worldId);
      navigate(`/worlds/${worldId}/feed`);
    } catch (error) {
      console.error('Failed to join world:', error);
    } finally {
      setIsJoining(prev => ({ ...prev, [worldId]: false }));
    }
  };

  const handleGoToWorld = (worldId: string) => {
    if (isAuthenticated) {
      navigate(`/worlds/${worldId}/feed`);
    } else {
      navigate('/login');
    }
  };

  const loadMoreWorlds = () => {
    console.log('Loading more worlds...');
    // Placeholder for load more functionality
  };

  return (
    <Layout>
      <div className="min-h-screen flex flex-col bg-white">
        {/* MAIN CONTENT */}
        <main className="flex-1">
          <div className="container">
            
            {/* PAGE HEADER */}
            <div className="catalog-header">
              <h1 className="catalog-title">Explore Worlds</h1>
              <p className="catalog-subtitle">Discover infinite AI-generated universes, each with unique stories, characters, and endless possibilities.</p>
            </div>

            {/* ERROR MESSAGE */}
            {error && (
              <div style={{ 
                backgroundColor: 'rgba(239, 118, 122, 0.1)', 
                color: 'var(--color-error)', 
                padding: 'var(--spacing-4)', 
                borderRadius: 'var(--radius-md)', 
                marginBottom: 'var(--spacing-6)' 
              }}>
                {error}
              </div>
            )}

            {/* WORLDS GRID */}
            <div className="catalog-content">
              {isLoading ? (
                <div className="catalog-loading">
                  <div className="catalog-loading-spinner"></div>
                </div>
              ) : worlds.length > 0 ? (
                <>
                  <div className="worlds-grid">
                    {worlds.map((world) => (
                      <div 
                        key={world.id} 
                        className="world-card" 
                        onClick={() => handleGoToWorld(world.id)}
                      >
                        {/* World Cover */}
                        <div className="world-card-image-container">
                          <div 
                            className="world-card-image"
                            style={{ backgroundImage: `url('${world.image_url}')` }}
                          ></div>
                          {/* World Icon */}
                          <div 
                            className="world-card-icon"
                            style={{ backgroundImage: `url('${world.icon_url}')` }}
                          ></div>
                        </div>
                        
                        {/* World Info */}
                        <div className="world-card-body">
                          <h3 className="world-card-title">{world.name}</h3>
                          <p className="world-card-description">{world.description}</p>
                          
                          {/* Stats */}
                          <div className="world-card-stats">
                            <span>✨ {world.users_count || 0} Characters</span>
                            <span>📸 {(world.posts_count || 0).toLocaleString()} Posts</span>
                            <span>❤️ {Math.floor((world.posts_count || 0) * 1.6 / 1000 * 10) / 10}K</span>
                          </div>
                          
                          {/* Enter Button */}
                          <button 
                            className={`world-card-btn ${isJoining[world.id] ? 'loading' : ''}`}
                            onClick={(e) => handleJoinWorld(world.id, e)}
                            disabled={isJoining[world.id]}
                          >
                            {isJoining[world.id] ? 'Joining...' : 'Enter World'}
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Load More Button */}
                  <div className="catalog-load-more">
                    <button className="load-more-btn" onClick={loadMoreWorlds}>
                      Load More Worlds
                    </button>
                  </div>
                </>
              ) : (
                <div className="catalog-empty">
                  <div className="catalog-empty-icon">🌍</div>
                  <h2 className="catalog-empty-title">No worlds found</h2>
                  <p className="catalog-empty-text">
                    Be the first to create a synthetic world and start your adventure!
                  </p>
                  <button 
                    className="catalog-empty-btn"
                    onClick={() => navigate('/create-world')}
                  >
                    Create New World
                  </button>
                </div>
              )}
            </div>
          </div>
        </main>
      </div>
    </Layout>
  );
};