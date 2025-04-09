import React, { useState, useEffect, useContext } from 'react';
import axiosInstance from '../api/axios';
import { Post } from '../types';
import { AuthContext } from '../context/AuthContext';

const Feed: React.FC = () => {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  const { isAuthenticated } = useContext(AuthContext);

  useEffect(() => {
    fetchPosts();
  }, []);

  const fetchPosts = async () => {
    try {
      setLoading(true);
      const limit = 10;
      const offset = (page - 1) * limit;
      const response = await axiosInstance.get(`/feed?limit=${limit}&offset=${offset}`);
      
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
      window.location.href = '/login';
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

      // Обновляем состояние постов
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

  return (
    <div className="feed-container">
      <h2>Feed</h2>
      {error && <div className="error">{error}</div>}

      <div className="posts">
        {posts.map(post => (
          <div key={post.id} className="post-card">
            <div className="post-header">
              <span className="post-username">{post.username}</span>
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

      {loading && <div className="loading">Loading...</div>}

      {!loading && hasMore && (
        <button onClick={fetchPosts} className="load-more-button">
          Load More
        </button>
      )}

      {!loading && !hasMore && posts.length > 0 && (
        <div className="no-more-posts">No more posts</div>
      )}

      {!loading && posts.length === 0 && (
        <div className="no-posts">No posts yet</div>
      )}
    </div>
  );
};

export default Feed;