import React, { useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import { useNavigate } from 'react-router-dom';
import { Comment } from '../../types';
import { interactionsAPI } from '../../api/services';
import '../../styles/components.css';

interface CommentCardProps {
  comment: Comment;
  currentWorldId: string;
  isReply?: boolean;
  onReply?: (commentId: string, text: string) => void;
}

export const CommentCard: React.FC<CommentCardProps> = ({ 
  comment, 
  currentWorldId, 
  isReply = false,
  onReply 
}) => {
  const navigate = useNavigate();
  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [showReplyInput, setShowReplyInput] = useState(false);
  const [replyText, setReplyText] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleCharacterClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    navigate(`/characters/${comment.character_id}`);
  };

  const handleLike = async () => {
    try {
      // TODO: Implement comment like API
      setIsLiked(!isLiked);
      setLikesCount(prev => isLiked ? prev - 1 : prev + 1);
    } catch (error) {
      console.error('Failed to like comment:', error);
    }
  };

  const handleReplyToggle = () => {
    setShowReplyInput(!showReplyInput);
  };

  const handleReplySubmit = async () => {
    if (!replyText.trim() || isSubmitting) return;

    setIsSubmitting(true);
    try {
      if (onReply) {
        await onReply(comment.id, replyText.trim());
        setReplyText('');
        setShowReplyInput(false);
      }
    } catch (error) {
      console.error('Failed to post reply:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleReplySubmit();
    }
  };

  const formattedTime = formatDistanceToNow(new Date(comment.created_at), { addSuffix: true });
  const authorName = comment.display_name || 'Unknown';

  return (
    <div className={`comment ${isReply ? 'comment-reply' : ''}`}>
      <div className="comment-content">
        <div 
          className="comment-avatar"
          style={{
            backgroundImage: `url(${comment.avatar_url || '/no-image.jpg'})`
          }}
          onClick={handleCharacterClick}
        />
        <div className="comment-body">
          <div className="comment-header">
            <p className="comment-author" onClick={handleCharacterClick}>
              {authorName}
            </p>
            <span className="comment-time">{formattedTime}</span>
          </div>
          <p className="comment-text">{comment.text}</p>
          <div className="comment-actions">
            <button 
              className={`comment-action-btn ${isLiked ? 'liked' : ''}`} 
              onClick={handleLike}
            >
              {isLiked ? 'Liked' : 'Like'}
            </button>
            {!isReply && (
              <button className="comment-action-btn" onClick={handleReplyToggle}>
                Reply
              </button>
            )}
            {likesCount > 0 && (
              <span>{likesCount} likes</span>
            )}
          </div>
          
          {/* Reply input */}
          {showReplyInput && (
            <div className="comment-reply-input">
              <div className="comment-input-wrapper">
                <div 
                  className="comment-input-avatar"
                  style={{
                    backgroundImage: 'url(/no-image.jpg)'
                  }}
                />
                <div className="comment-input-form">
                  <div className="comment-input-row">
                    <input 
                      type="text" 
                      className="comment-input" 
                      placeholder="Write a reply..."
                      value={replyText}
                      onChange={(e) => setReplyText(e.target.value)}
                      onKeyPress={handleKeyPress}
                      autoFocus
                    />
                    <button 
                      className="comment-submit-btn" 
                      onClick={handleReplySubmit}
                      disabled={!replyText.trim() || isSubmitting}
                    >
                      {isSubmitting ? '...' : 'Post'}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};