import React, { useState, useEffect, useContext } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import axiosInstance from '../api/axios';
import { Post, World } from '../types';
import { AuthContext } from '../context/AuthContext';

const Feed: React.FC = () => {
  const { worldId } = useParams<{ worldId: string }>();
  const [posts, setPosts] = useState<Post[]>([]);
  const [worldInfo, setWorldInfo] = useState<World | null>(null);
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
    
    if (!worldId) {
      navigate('/worlds');
      return;
    }
    
    // Get world info and initial posts
    fetchWorldInfo();
    setPosts([]);
    setPage(1);
    setHasMore(true);
    fetchPosts();
  }, [isAuthenticated, worldId, navigate]);

  const fetchWorldInfo = async () => {
    if (!worldId) return;
    
    try {
      const response = await axiosInstance.get(`/worlds/${worldId}`);
      if (response.data && response.data.id) {
        setWorldInfo(response.data);
      }
    } catch (err) {
      setError('Failed to load world information');
      console.error(err);
    }
  };

  const fetchPosts = async () => {
    if (!worldId) return;
    
    try {
      setLoading(true);
      const limit = 10;
      const offset = (page - 1) * limit;
      const response = await axiosInstance.get(`/worlds/${worldId}/feed?limit=${limit}&offset=${offset}`);
      
      if (response.data.posts.length === 0) {
        setHasMore(false);
      } else {
        setPosts(prevPosts => [...prevPosts, ...response.data.posts]);
        setPage(prevPage => prevPage + 1);
      }
    } catch (err) {
      setError('Failed to load posts');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleLike = async (postId: string, isLiked: boolean) => {
    if (!isAuthenticated || !worldId) {
      navigate('/login');
      return;
    }

    try {
      if (isLiked) {
        // Unlike
        await axiosInstance.delete(`/worlds/${worldId}/posts/${postId}/like`);
      } else {
        // Like
        await axiosInstance.post(`/worlds/${worldId}/posts/${postId}/like`);
      }

      // Update posts state
      setPosts(prevPosts =>
        prevPosts.map(post =>
          post.id === postId
            ? {
                ...post,
                likes_count: isLiked ? post.likes_count - 1 : post.likes_count + 1,
                user_liked: !isLiked
              }
            : post
        )
      );
    } catch (err) {
      console.error('Failed to like/unlike post', err);
    }
  };

  const handleChangeWorld = () => {
    navigate('/worlds');
  };

  const handleCreatePost = () => {
    navigate(`/worlds/${worldId}/create`);
  };

  if (!worldInfo) {
    return <div className="loading">Loading world...</div>;
  }

  return (
    <div className="feed-container">
      <div className="world-header">
        <h2>{worldInfo.name}</h2>
        <div className="world-actions">
          <button onClick={handleChangeWorld} className="change-world-button">
            Change World
          </button>
          <button onClick={handleCreatePost} className="create-post-button">
            Create Post
          </button>
        </div>
      </div>
      
      {worldInfo.description && (
        <p className="world-description">{worldInfo.description}</p>
      )}
      
      {error && <div className="error">{error}</div>}

      <div className="posts">
        {posts.map(post => (
          <div key={post.id} className="post-card">
            <div className="post-header">
              <span className="post-username">{post.username}</span>
              {post.is_ai && <span className="ai-badge">AI</span>}
            </div>
            <img src={post.media_url || post.image_url} alt={post.caption} className="post-image" />
            <div className="post-actions">
              <button
                className={`like-button ${post.user_liked ? 'liked' : ''}`}
                onClick={() => handleLike(post.id, post.user_liked || false)}
              >
                {post.user_liked ? '❤️' : '🤍'} {post.likes_count}
              </button>
            </div>
            <div className="post-caption">
              <span className="post-username">{post.username}</span> {post.caption}
            </div>
            <div className="post-comments">
              <a href={`/worlds/${worldId}/posts/${post.id}`}>View all {post.comments_count} comments</a>
            </div>
          </div>
        ))}
      </div>

      {loading && <div className="loading">Loading posts...</div>}

      {!loading && hasMore && (
        <button onClick={fetchPosts} className="load-more-button">
          Load More
        </button>
      )}

      {!loading && !hasMore && posts.length > 0 && (
        <div className="no-more-posts">No more posts</div>
      )}

      {!loading && posts.length === 0 && (
        <div className="no-posts">
          <p>No posts in this world yet. Be the first to post!</p>
          <button onClick={handleCreatePost} className="create-first-post-button">
            Create First Post
          </button>
        </div>
      )}
    </div>
  );
};

export default Feed;