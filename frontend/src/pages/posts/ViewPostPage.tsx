import React, { useEffect, useState, useRef } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { formatDistanceToNow } from 'date-fns';
import { Layout } from '../../components/layout/Layout';
import { Button } from '../../components/ui/Button';
import { Loader } from '../../components/ui/Loader';
import { CommentCard } from '../../components/cards/CommentCard';
import { useWorld } from '../../hooks/useWorld';
import { useAuth } from '../../hooks/useAuth';
import { postsAPI, interactionsAPI } from '../../api/services';
import { Post, Comment } from '../../types';
import { formatNumber } from '../../utils/formatters';
import '../../styles/pages/feed.css';

// Helper functions and icons
const HeartIcon = ({ filled }: { filled: boolean }) => (
  <svg width="28" height="28" viewBox="0 0 24 24" fill={filled ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2">
    <path d="M20.84 4.61C20.3292 4.099 19.7228 3.69364 19.0554 3.41708C18.3879 3.14052 17.6725 2.99817 16.95 2.99817C16.2275 2.99817 15.5121 3.14052 14.8446 3.41708C14.1772 3.69364 13.5708 4.099 13.06 4.61L12 5.67L10.94 4.61C9.9083 3.5783 8.5091 2.9987 7.05 2.9987C5.5909 2.9987 4.1917 3.5783 3.16 4.61C2.1283 5.6417 1.5487 7.0409 1.5487 8.5C1.5487 9.9591 2.1283 11.3583 3.16 12.39L12 21.23L20.84 12.39C21.351 11.8792 21.7563 11.2728 22.0329 10.6053C22.3095 9.93789 22.4518 9.22248 22.4518 8.5C22.4518 7.77752 22.3095 7.06211 22.0329 6.39467C21.7563 5.72723 21.351 5.1208 20.84 4.61V4.61Z"></path>
  </svg>
);

const CommentIcon = () => (
  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M21 15C21 15.5304 20.7893 16.0391 20.4142 16.4142C20.0391 16.7893 19.5304 17 19 17H7L3 21V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H19C19.5304 3 20.0391 3.21071 20.4142 3.58579C20.7893 3.96086 21 4.46957 21 5V15Z"></path>
  </svg>
);

const SaveIcon = ({ filled }: { filled: boolean }) => (
  <svg width="28" height="28" viewBox="0 0 24 24" fill={filled ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2">
    <path d="M19 21L12 16L5 21V5C5 4.46957 5.21071 3.96086 5.58579 3.58579C5.96086 3.21071 6.46957 3 7 3H17C17.5304 3 18.0391 3.21071 18.4142 3.58579C18.7893 3.96086 19 4.46957 19 5V21Z"></path>
  </svg>
);

export const ViewPostPage: React.FC = () => {
  const { worldId, postId } = useParams<{ worldId: string; postId: string }>();
  const { currentWorld, loadCurrentWorld } = useWorld();
  const { user, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  
  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isCommentLoading, setIsCommentLoading] = useState(false);
  const [isLiking, setIsLiking] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLiked, setIsLiked] = useState(false);
  const [isSaved, setIsSaved] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  
  // Refs to prevent duplicate requests
  const isLoadingRef = useRef(false);
  const postIdRef = useRef<string | null>(null);
  const worldIdRef = useRef<string | null>(null);
  const commentTextareaRef = useRef<HTMLTextAreaElement>(null);
  
  // Load world and post data
  useEffect(() => {
    const fetchData = async () => {
      if (!worldId || !postId) return;
      
      // Skip if we're already loading this post or if it's the same post
      if (isLoadingRef.current || (postIdRef.current === postId && worldIdRef.current === worldId)) {
        return;
      }
      
      isLoadingRef.current = true;
      postIdRef.current = postId;
      worldIdRef.current = worldId;
      
      setIsLoading(true);
      setError(null);
      
      try {
        await loadCurrentWorld(worldId);
        
        const postData = await postsAPI.getPostById(worldId, postId);
        setPost(postData);
        setIsLiked(postData.user_liked || false);
        setLikesCount(postData.likes_count);
        
        const commentsData = await interactionsAPI.getPostComments(worldId, postId);
        setComments(commentsData.comments || []);
      } catch (err: any) {
        setError(err.message || 'Failed to load post');
        console.error('Error loading post:', err);
      } finally {
        setIsLoading(false);
        isLoadingRef.current = false;
      }
    };
    
    fetchData();
  }, [worldId, postId, loadCurrentWorld]);
  
  const handleLike = async () => {
    if (!worldId || !postId || isLiking) return;
    
    setIsLiking(true);
    
    try {
      if (isLiked) {
        await interactionsAPI.unlikePost(worldId, postId);
        setIsLiked(false);
        setLikesCount(prev => prev - 1);
      } else {
        await interactionsAPI.likePost(worldId, postId);
        setIsLiked(true);
        setLikesCount(prev => prev + 1);
      }
    } catch (err: any) {
      setError(err.message || 'Failed to like/unlike post');
    } finally {
      setIsLiking(false);
    }
  };

  const handleSave = () => {
    setIsSaved(!isSaved);
    // В реальном приложении здесь был бы API запрос
  };

  const focusCommentInput = () => {
    if (commentTextareaRef.current) {
      commentTextareaRef.current.focus();
    }
  };

  const handleGoToCharacter = (characterId: string) => {
    navigate(`/characters/${characterId}`);
  };
  
  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!worldId || !postId || !newComment.trim() || isCommentLoading) return;
    
    setIsCommentLoading(true);
    
    try {
      const comment = await interactionsAPI.addComment(worldId, postId, newComment);
      setComments(prev => [comment, ...prev]);
      setNewComment('');
      
      // Update comment count on post
      if (post) {
        setPost({
          ...post,
          comments_count: post.comments_count + 1
        });
      }
    } catch (err: any) {
      setError(err.message || 'Failed to add comment');
    } finally {
      setIsCommentLoading(false);
    }
  };
  
  // Format relative time (e.g., "2 hours ago")
  const formatRelativeTime = (dateString: string) => {
    return formatDistanceToNow(new Date(dateString), { addSuffix: true });
  };
  
  if (isLoading) {
    return (
      <Layout>
        <div style={{ display: 'flex', justifyContent: 'center', padding: 'var(--space-10)' }}>
          <Loader text="Loading post..." />
        </div>
      </Layout>
    );
  }
  
  if (!post || !currentWorld) {
    return (
      <Layout>
        <div className="container">
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <p style={{ color: 'var(--color-error)', marginBottom: '1rem' }}>
              {error || "Post not found or you don't have access."}
            </p>
            <Button onClick={() => navigate(`/worlds/${worldId}/feed`)}>
              Return to Feed
            </Button>
          </div>
        </div>
      </Layout>
    );
  }
  
  return (
    <Layout>
      {/* MAIN CONTENT */}
      <div className="feed-container">
        <div className="post-detail-container">
          
          {error && (
            <div style={{
              backgroundColor: 'rgba(239, 118, 122, 0.1)',
              color: 'var(--color-accent)',
              padding: 'var(--space-4)',
              borderRadius: 'var(--radius-md)',
              marginBottom: 'var(--space-6)'
            }}>
              {error}
            </div>
          )}
          
          {/* POST CARD */}
          <div className="post-detail-card">
            
            {/* Post Header */}
            <div className="post-detail-header">
              <div 
                className="post-detail-avatar" 
                style={{
                  backgroundImage: `url(${post.avatar_url || '/no-image.jpg'})`
                }}
                onClick={() => handleGoToCharacter(post.character_id)}
              />
              <div className="post-detail-author-info">
                <p 
                  className="post-detail-author-name" 
                  onClick={() => handleGoToCharacter(post.character_id)}
                >
                  {post.display_name}
                </p>
                <p className="post-detail-author-meta">
                  {post.is_ai ? 'AI Character' : 'User'} • {currentWorld.name}
                </p>
                <p className="post-detail-time">
                  {formatDistanceToNow(new Date(post.created_at), { addSuffix: true })}
                </p>
              </div>
              <button className="post-menu-btn">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12C11 12.5523 11.4477 13 12 13Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M12 6C12.5523 6 13 5.55228 13 5C13 4.44772 12.5523 4 12 4C11.4477 4 11 4.44772 11 5C11 5.55228 11.4477 6 12 6Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M12 20C12.5523 20 13 19.5523 13 19C13 18.4477 12.5523 18 12 18C11.4477 18 11 18.4477 11 19C11 19.5523 11.4477 20 12 20Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            </div>
            
            {/* Post Image */}
            {(post.media_url || post.image_url) && (
              <div 
                className="post-detail-image" 
                style={{
                  backgroundImage: `url(${post.media_url || post.image_url})`
                }}
              />
            )}
            
            {/* Post Actions and Caption */}
            <div className="post-detail-content">
              {/* Action Buttons */}
              <div className="post-detail-actions">
                <button 
                  className={`post-detail-action-btn like-btn ${isLiked ? 'liked' : ''}`} 
                  onClick={handleLike}
                  disabled={isLiking}
                  style={{ color: isLiked ? '#ef4444' : undefined }}
                >
                  <HeartIcon filled={isLiked} />
                </button>
                <button className="post-detail-action-btn" onClick={focusCommentInput}>
                  <CommentIcon />
                </button>
                <button 
                  className={`post-detail-action-btn post-detail-save-btn ${isSaved ? 'saved' : ''}`} 
                  onClick={handleSave}
                  style={{ color: isSaved ? 'currentColor' : undefined }}
                >
                  <SaveIcon filled={isSaved} />
                </button>
              </div>
              
              {/* Likes */}
              <p className="post-detail-likes">{formatNumber(likesCount)} likes</p>
              
              {/* Caption */}
              {post.caption && (
                <div className="post-detail-caption">
                  <p className="post-detail-caption-text">
                    <span className="post-detail-caption-author">{post.display_name}</span> {post.caption}
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* COMMENTS SECTION */}
          <div className="comments-section">
            <h3 className="comments-title">Comments</h3>
            
            {/* Comments List */}
            {comments.length > 0 ? (
              comments.map((comment) => (
                <CommentCard 
                  key={comment.id} 
                  comment={comment} 
                  currentWorldId={worldId || ''}
                />
              ))
            ) : (
              <div style={{ 
                textAlign: 'center', 
                padding: 'var(--space-6)', 
                color: 'var(--color-text-secondary)' 
              }}>
                No comments yet. Be the first to add one!
              </div>
            )}
            
            {/* Add Comment Form */}
            {isAuthenticated && (
              <div className="add-comment-form">
                <div className="comment-form-header">
                  <div 
                    className="user-avatar" 
                    style={{
                      backgroundImage: `url(${user?.avatar_url || '/no-image.jpg'})`
                    }}
                  />
                  <div className="comment-form-content">
                    <textarea 
                      ref={commentTextareaRef}
                      value={newComment}
                      onChange={(e) => setNewComment(e.target.value)}
                      placeholder="Add a comment..."
                      rows={1}
                      className="comment-textarea"
                      style={{
                        height: 'auto',
                        minHeight: '2.5rem'
                      }}
                      onInput={(e) => {
                        const target = e.target as HTMLTextAreaElement;
                        target.style.height = 'auto';
                        target.style.height = target.scrollHeight + 'px';
                        
                        // Show/hide actions based on content
                        const actions = target.parentElement?.querySelector('.comment-form-actions') as HTMLElement;
                        if (actions) {
                          actions.style.display = target.value.trim() ? 'flex' : 'none';
                        }
                      }}
                    />
                    <div 
                      className="comment-form-actions" 
                      style={{ 
                        display: newComment.trim() ? 'flex' : 'none',
                        justifyContent: 'flex-end',
                        gap: 'var(--space-2)',
                        marginTop: 'var(--space-3)'
                      }}
                    >
                      <button 
                        className="btn btn-secondary btn-sm" 
                        onClick={() => setNewComment('')}
                        type="button"
                      >
                        Cancel
                      </button>
                      <button 
                        className="btn btn-primary btn-sm" 
                        onClick={handleAddComment}
                        disabled={!newComment.trim() || isCommentLoading}
                        type="button"
                      >
                        {isCommentLoading ? 'Posting...' : 'Post'}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
          
        </div>
      </div>
    </Layout>
  );
};