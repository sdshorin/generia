import React, { useEffect, useState, useRef } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { Layout } from '../../components/layout/Layout';
import { Button } from '../../components/ui/Button';
import { Loader } from '../../components/ui/Loader';
import { PostCard } from '../../components/common/PostCard';
import { useAuth } from '../../hooks/useAuth';
import { useWorld } from '../../hooks/useWorld';
import { useInfiniteScroll } from '../../hooks/useInfiniteScroll';
import { postsAPI } from '../../api/services';
import { Post, WorldGenerationStatus } from '../../types';
import { formatNumber } from '../../utils/formatters';
import '../../styles/pages/feed.css';


export const FeedPage: React.FC = () => {
  const { worldId } = useParams<{ worldId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { currentWorld, loadCurrentWorld, error: worldError } = useWorld();
  const [isLoading, setIsLoading] = useState(true);
  const [showGenerationProgress, setShowGenerationProgress] = useState(false);
  const [loadingMorePosts, setLoadingMorePosts] = useState(false);
  
  const {
    items: posts,
    isLoading: isPostsLoading,
    error: postsError,
    loadMore,
    reset,
    sentinelRef
  } = useInfiniteScroll<Post>({
    fetchItems: async (limit, cursor) => {
      if (!worldId) return { items: [], hasMore: false };
      const response = await postsAPI.getFeed(worldId, limit, cursor);
      return { 
        items: response.posts || [],
        nextCursor: response.next_cursor || '',
        hasMore: response.has_more || false
      };
    },
    limit: 10
  });
  
  const worldIdRef = useRef<string | null>(null);
  
  useEffect(() => {
    async function fetchInitialPosts() {
      try {
        if (!currentWorld) return;
        setIsLoading(true);
        reset(); // Clear existing posts when switching worlds
        
        // Show generation progress if world is still generating
        setShowGenerationProgress(
          currentWorld.generation_status !== 'completed' && 
          currentWorld.generation_status !== 'failed'
        );
      } catch (error) {
        console.error('Failed to fetch initial posts:', error);
      } finally {
        setIsLoading(false);
      }
    }

    if (worldId && worldIdRef.current !== worldId) {
      worldIdRef.current = worldId;
      loadCurrentWorld(worldId);
      setIsLoading(false);
    } else if (currentWorld && currentWorld.id === worldId) {
      fetchInitialPosts();
    }
  }, [worldId, currentWorld, reset, loadCurrentWorld]);
  
  const handleGenerationComplete = (status: WorldGenerationStatus) => {
    setShowGenerationProgress(false);
    // Reload the current world to get updated data
    if (worldId) {
      loadCurrentWorld(worldId);
    }
    // Reset posts to load new content
    reset();
  };
  
  const handlePostsUpdated = (postsCount: number) => {
    // Reset posts to load new content when posts are created
    reset();
  };
  
  const handlePostLike = (postId: string, isLiked: boolean) => {
    // Update post in the list
    // In a real app, we might want to use a more sophisticated state management
    // This is a simple implementation for demonstration
  };

  const handleLoadMorePosts = async () => {
    setLoadingMorePosts(true);
    try {
      await loadMore();
    } catch (error) {
      console.error('Failed to load more posts:', error);
    } finally {
      setLoadingMorePosts(false);
    }
  };
  
  const handleGoBack = () => {
    navigate(-1);
  };

  const handleWorldDetails = () => {
    navigate(`/worlds/${worldId}/about`);
  };
  
  if (isLoading) {
    return (
      <Layout>
        <div style={{ padding: 'var(--space-6) 0', display: 'flex', justifyContent: 'center' }}>
          <Loader text="Loading world..." />
        </div>
      </Layout>
    );
  }
  
  if (!currentWorld) {
    return (
      <Layout>
        <div style={{ 
          backgroundColor: 'rgba(239, 118, 122, 0.1)',
          color: 'var(--color-accent)',
          padding: 'var(--space-4)',
          borderRadius: 'var(--radius-md)',
          marginBottom: 'var(--space-6)'
        }}>
          {worldError || "World not found. It may have been deleted or you don't have access."}
        </div>
        <Link to="/worlds">
          <Button>Browse Other Worlds</Button>
        </Link>
      </Layout>
    );
  }
  
  return (
    <Layout>
      {/* WORLD HEADER */}
      <div className="world-header">
        <div className="world-header-image" style={{
          backgroundImage: `linear-gradient(rgba(0, 0, 0, 0.2) 0%, rgba(0, 0, 0, 0.4) 100%), url(${currentWorld.image_url || '/no-image.jpg'})`
        }}>
          <div className="world-header-overlay">
            <div className="world-header-content">
              <div className="world-header-info">
                <div className="world-icon" style={{
                  backgroundImage: `url(${currentWorld.icon_url || '/no-image.jpg'})`
                }}></div>
                <div className="world-details">
                  <h1 className="world-title">{currentWorld.name}</h1>
                  <div className="world-stats">
                    <span>✨ {formatNumber(currentWorld.users_count || 127)} Characters</span>
                    <span>📸 {formatNumber(currentWorld.posts_count || 0)} Posts</span>
                    <span>❤️ 2.3K</span>
                  </div>
                </div>
              </div>
              
              <button className="world-details-btn" onClick={handleWorldDetails}>
                <span>World Details</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* FEED CONTAINER */}
      <div className="feed-container">
        <div className="feed-content">
          
          {posts.length > 0 ? (
            <>
              {posts.map((post) => (
                <div key={post.id} style={{ marginBottom: 'var(--space-6)' }}>
                  <PostCard 
                    post={post} 
                    currentWorldId={worldId || ''}
                    onLike={handlePostLike}
                  />
                </div>
              ))}
              
              {/* Load More Button */}
              <div className="load-more-container">
                <button 
                  className="load-more-btn" 
                  onClick={handleLoadMorePosts}
                  disabled={loadingMorePosts || isPostsLoading}
                >
                  {loadingMorePosts ? 'Loading...' : 'Load More Posts'}
                </button>
              </div>
              
              {/* Infinite scroll sentinel */}
              <div ref={sentinelRef}>
                {isPostsLoading && (
                  <div style={{ padding: 'var(--space-6) 0', display: 'flex', justifyContent: 'center' }}>
                    <Loader size="sm" />
                  </div>
                )}
              </div>
            </>
          ) : isPostsLoading ? (
            <div style={{ padding: 'var(--space-6) 0', display: 'flex', justifyContent: 'center' }}>
              <Loader text="Loading posts..." />
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: 'var(--space-10) var(--space-4)' }}>
              <h3 style={{ fontSize: 'var(--font-xl)', marginBottom: 'var(--space-4)' }}>No posts yet</h3>
              <p style={{ color: 'var(--color-text-light)', marginBottom: 'var(--space-6)' }}>
                {showGenerationProgress ? 'Posts will appear here as they are generated!' : 'Be the first to create content in this world!'}
              </p>
              {!showGenerationProgress && (
                <Link to={`/worlds/${worldId}/create`}>
                  <Button variant="primary">Create a Post</Button>
                </Link>
              )}
            </div>
          )}
          
          {postsError && (
            <div style={{
              backgroundColor: 'rgba(239, 118, 122, 0.1)',
              color: 'var(--color-accent)',
              padding: 'var(--space-4)',
              borderRadius: 'var(--radius-md)',
              marginBottom: 'var(--space-6)'
            }}>
              {postsError}
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
};