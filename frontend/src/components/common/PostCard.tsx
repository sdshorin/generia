import React, { useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { Link } from 'react-router-dom';
import { Post } from '../../types';
import { interactionsAPI } from '../../api/services';
import '../../styles/components.css';

interface PostCardProps {
  post: Post;
  currentWorldId: string;
  onLike?: (postId: string, isLiked: boolean) => void;
}

export const PostCard: React.FC<PostCardProps> = ({
  post,
  currentWorldId,
  onLike,
}) => {
  const [isLiking, setIsLiking] = useState(false);
  const [isLiked, setIsLiked] = useState(post.user_liked || false);
  const [likesCount, setLikesCount] = useState(post.likes_count);

  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (isLiking) return;

    setIsLiking(true);
    try {
      if (isLiked) {
        await interactionsAPI.unlikePost(currentWorldId, post.id);
        setIsLiked(false);
        setLikesCount(prev => prev - 1);
      } else {
        await interactionsAPI.likePost(currentWorldId, post.id);
        setIsLiked(true);
        setLikesCount(prev => prev + 1);
      }

      if (onLike) {
        onLike(post.id, !isLiked);
      }
    } catch (error) {
      console.error('Failed to like/unlike post:', error);
    } finally {
      setIsLiking(false);
    }
  };

  const handlePostClick = () => {
    window.location.href = `/worlds/${currentWorldId}/posts/${post.id}`;
  };

  const handleCharacterClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    window.location.href = `/characters/${post.character_id}`;
  };

  const handleMenuClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    // TODO: Implement menu functionality
  };

  // Format timestamp to relative time (e.g., "2 hours ago")
  const formattedTime = formatDistanceToNow(new Date(post.created_at), { addSuffix: true });
  
  // Get author name and role
  const authorName = post.display_name || 'Unknown';
  const authorRole = post.is_ai ? 'AI Character' : 'User';

  return (
    <div className="post-card">
      {/* Post Header */}
      <div className="post-header">
        <div 
          className="post-avatar" 
          style={{
            backgroundImage: `url(${post.avatar_url || '/no-image.jpg'})`
          }}
          onClick={handleCharacterClick}
        />
        <div className="post-author-info">
          <p className="post-author-name" onClick={handleCharacterClick}>
            {authorName}
          </p>
          <p className="post-author-meta">
            {authorRole} • {formattedTime}
          </p>
        </div>
        <button className="post-menu-btn" onClick={handleMenuClick}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12C11 12.5523 11.4477 13 12 13Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M12 6C12.5523 6 13 5.55228 13 5C13 4.44772 12.5523 4 12 4C11.4477 4 11 4.44772 11 5C11 5.55228 11.4477 6 12 6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M12 20C12.5523 20 13 19.5523 13 19C13 18.4477 12.5523 18 12 18C11.4477 18 11 18.4477 11 19C11 19.5523 11.4477 20 12 20Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </div>
      
      {/* Post Image */}
      {(post.media_url || post.image_url) && (
        <div 
          className="post-image" 
          style={{
            backgroundImage: `url(${post.media_url || post.image_url})`
          }}
          onClick={handlePostClick}
        />
      )}
      
      {/* Post Content */}
      <div className="post-content">
        {/* Action Buttons */}
        <div className="post-actions">
          <button 
            className={`post-action-btn like-btn ${isLiked ? 'liked' : ''}`} 
            onClick={handleLike}
            disabled={isLiking}
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill={isLiked ? "currentColor" : "none"} xmlns="http://www.w3.org/2000/svg">
              <path d="M20.84 4.61C20.3292 4.099 19.7228 3.69364 19.0554 3.41708C18.3879 3.14052 17.6725 2.99817 16.95 2.99817C16.2275 2.99817 15.5121 3.14052 14.8446 3.41708C14.1772 3.69364 13.5708 4.099 13.06 4.61L12 5.67L10.94 4.61C9.9083 3.5783 8.5091 2.9987 7.05 2.9987C5.5909 2.9987 4.1917 3.5783 3.16 4.61C2.1283 5.6417 1.5487 7.0409 1.5487 8.5C1.5487 9.9591 2.1283 11.3583 3.16 12.39L12 21.23L20.84 12.39C21.351 11.8792 21.7563 11.2728 22.0329 10.6053C22.3095 9.93789 22.4518 9.22248 22.4518 8.5C22.4518 7.77752 22.3095 7.06211 22.0329 6.39467C21.7563 5.72723 21.351 5.1208 20.84 4.61V4.61Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
          <button className="post-action-btn" onClick={handlePostClick}>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M21 15C21 15.5304 20.7893 16.0391 20.4142 16.4142C20.0391 16.7893 19.5304 17 19 17H7L3 21V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H19C19.5304 3 20.0391 3.21071 20.4142 3.58579C20.7893 3.96086 21 4.46957 21 5V15Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
        
        {/* Likes */}
        <p className="post-likes">{likesCount} likes</p>
        
        {/* Caption */}
        {post.caption && (
          <div className="post-caption">
            <p className="post-caption-text">
              <span className="post-caption-author">{authorName.toLowerCase().replace(/\s+/g, '_')}</span> {post.caption.length > 250 ? post.caption.substring(0, 250) + '...' : post.caption}
            </p>
          </div>
        )}
        
        {/* View Comments */}
        <Link 
          to={`/worlds/${currentWorldId}/posts/${post.id}`} 
          className="post-comments-link"
        >
          View all {post.comments_count} comments
        </Link>
      </div>
    </div>
  );
};