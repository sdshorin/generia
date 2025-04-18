import React, { useState, useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axiosInstance from '../api/axios';
import { Post, World } from '../types';
import { AuthContext } from '../context/AuthContext';

const Feed: React.FC = () => {
  const [posts, setPosts] = useState<Post[]>([]);
  const [activeWorld, setActiveWorld] = useState<World | null>(null);
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
    
    // Get active world first
    fetchActiveWorld();
  }, [isAuthenticated, navigate]);

  // Fetch posts when active world changes
  useEffect(() => {
    if (activeWorld) {
      setPosts([]);
      setPage(1);
      setHasMore(true);
      fetchPosts();
    }
  }, [activeWorld]);

  const fetchActiveWorld = async () => {
    try {
      const response = await axiosInstance.get('/worlds/active');
      
      if (response.data && response.data.id) {
        setActiveWorld(response.data);
      } else {
        // No active world, redirect to worlds selection
        navigate('/worlds');
      }
    } catch (err: any) {
      // If no active world or error, redirect to worlds page
      if (err.response && err.response.status === 404) {
        navigate('/worlds');
      } else {
        setError('Failed to load active world');
        console.error(err);
      }
    }
  };

  const fetchPosts = async () => {
    if (!activeWorld) return;
    
    try {
      setLoading(true);
      const limit = 10;
      const offset = (page - 1) * limit;
      const response = await axiosInstance.get(`/feed?world_id=${activeWorld.id}&limit=${limit}&offset=${offset}`);
      
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
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    try {
      if (isLiked) {
        // Unlike
        await axiosInstance.delete(`/posts/${postId}/like`);
      } else {
        // Like
        await axiosInstance.post(`/posts/${postId}/like`);
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
    navigate('/create');
  };

  if (!activeWorld) {
    return <div className="loading">Loading world...</div>;
  }

  return (
    <div className="feed-container">
      <div className="world-header">
        <h2>{activeWorld.name}</h2>
        <div className="world-actions">
          <button onClick={handleChangeWorld} className="change-world-button">
            Change World
          </button>
          <button onClick={handleCreatePost} className="create-post-button">
            Create Post
          </button>
        </div>
      </div>
      
      {activeWorld.description && (
        <p className="world-description">{activeWorld.description}</p>
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
              <a href={`/posts/${post.id}`}>View all {post.comments_count} comments</a>
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