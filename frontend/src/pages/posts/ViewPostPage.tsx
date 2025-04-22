import React, { useEffect, useState, useRef } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, AnimatePresence, HTMLMotionProps } from 'framer-motion';
import { formatDistanceToNow } from 'date-fns';
import { Layout } from '../../components/layout/Layout';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { TextArea } from '../../components/ui/TextArea';
import { Loader } from '../../components/ui/Loader';
import { Avatar } from '../../components/ui/Avatar';
import { useWorld } from '../../hooks/useWorld';
import { useAuth } from '../../hooks/useAuth';
import { postsAPI, interactionsAPI } from '../../api/services';
import { Post, Comment } from '../../types';

const PageContainer = styled.div`
  max-width: 800px;
  margin: 0 auto;
`;

const BackButton = styled(Link)`
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-text-light);
  text-decoration: none;
  margin-bottom: var(--space-6);
  font-size: var(--font-sm);
  
  &:hover {
    color: var(--color-text);
  }
`;

const PostContainer = styled(Card)`
  margin-bottom: var(--space-6);
  overflow: hidden;
`;

const PostMedia = styled.div`
  width: 100%;
  display: flex;
  justify-content: center;
  background-color: var(--color-background);
  
  img {
    max-width: 100%;
    max-height: 600px;
    object-fit: contain;
  }
`;

const PostHeader = styled.div`
  display: flex;
  align-items: center;
  padding: var(--space-4);
`;

const UserInfo = styled.div`
  margin-left: var(--space-3);
  flex: 1;
`;

const Username = styled.div`
  font-weight: 600;
  display: flex;
  align-items: center;
  
  .ai-badge {
    margin-left: var(--space-2);
    font-size: var(--font-xs);
    background-color: var(--color-secondary);
    color: white;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
  }
`;

const Timestamp = styled.div`
  font-size: var(--font-xs);
  color: var(--color-text-lighter);
`;

const PostCaption = styled.div`
  padding: var(--space-4);
  padding-top: 0;
  font-size: var(--font-md);
  line-height: 1.5;
  white-space: pre-wrap;
`;

const InteractionsContainer = styled.div`
  display: flex;
  align-items: center;
  padding: var(--space-4);
  border-top: 1px solid var(--color-border);
`;

const LikeButton = styled.button<{ $isLiked: boolean }>`
  display: flex;
  align-items: center;
  background: none;
  border: none;
  color: ${props => props.$isLiked ? 'var(--color-accent)' : 'var(--color-text-light)'};
  font-size: var(--font-sm);
  cursor: pointer;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  
  &:hover {
    background-color: rgba(239, 118, 122, 0.1);
  }
  
  svg {
    margin-right: var(--space-2);
  }
`;

const CommentButton = styled.div`
  display: flex;
  align-items: center;
  background: none;
  border: none;
  color: var(--color-text-light);
  font-size: var(--font-sm);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  margin-left: var(--space-4);
  
  svg {
    margin-right: var(--space-2);
  }
`;

const CommentsSection = styled(Card)`
  margin-bottom: var(--space-6);
`;

const CommentsList = styled.div`
  margin-top: var(--space-4);
`;

const CommentItem = styled.div`
  display: flex;
  padding: var(--space-4);
  border-bottom: 1px solid var(--color-border);
  
  &:last-child {
    border-bottom: none;
  }
`;

const CommentContent = styled.div`
  margin-left: var(--space-3);
  flex: 1;
`;

const CommentText = styled.div`
  font-size: var(--font-sm);
  margin-top: var(--space-1);
  white-space: pre-wrap;
`;

const CommentForm = styled.form`
  margin-top: var(--space-4);
  border-top: 1px solid var(--color-border);
  padding-top: var(--space-4);
`;

const ErrorMessage = styled.div`
  color: var(--color-accent);
  background-color: rgba(239, 118, 122, 0.1);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
`;

const NoCommentsMessage = styled.div`
  text-align: center;
  padding: var(--space-6);
  color: var(--color-text-light);
`;

const HeartIcon = ({ filled }: { filled: boolean }) => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill={filled ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2">
    <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
  </svg>
);

const CommentIcon = () => (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path>
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
  const [error, setError] = useState<string | null>(null);
  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  
  // Refs to prevent duplicate requests
  const isLoadingRef = useRef(false);
  const postIdRef = useRef<string | null>(null);
  const worldIdRef = useRef<string | null>(null);
  
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
        <PageContainer>
          <ErrorMessage>
            {error || "Post not found or you don't have access."}
          </ErrorMessage>
          <Button onClick={() => navigate(`/worlds/${worldId}/feed`)}>
            Return to Feed
          </Button>
        </PageContainer>
      </Layout>
    );
  }
  
  return (
    <Layout>
      <PageContainer>
        <BackButton to={`/worlds/${worldId}/feed`}>
          ← Back to {currentWorld.name} feed
        </BackButton>
        
        {error && <ErrorMessage>{error}</ErrorMessage>}
        
        <PostContainer>
          {(post.media_url || post.image_url) && (
            <PostMedia>
              <img 
                src={post.media_url || post.image_url} 
                alt="Post content" 
                loading="lazy" 
              />
            </PostMedia>
          )}
          
          <PostHeader>
            <Avatar name={post.display_name || ''} isAi={post.is_ai} />
            <UserInfo>
              <Username>
                {post.display_name}
                {post.is_ai && <span className="ai-badge">AI</span>}
              </Username>
              <Timestamp>{formatRelativeTime(post.created_at)}</Timestamp>
            </UserInfo>
          </PostHeader>
          
          {post.caption && (
            <PostCaption>{post.caption}</PostCaption>
          )}
          
          <InteractionsContainer>
            <AnimatePresence initial={false}>
              <LikeButton 
                $isLiked={isLiked} 
                onClick={handleLike}
                disabled={isLiking}
              >
                <motion.div
                  key={isLiked ? 'liked' : 'unliked'}
                  initial={{ scale: 0.8 }}
                  animate={{ scale: 1 }}
                  exit={{ scale: 0.8 }}
                  transition={{ duration: 0.2 }}
                >
                  <HeartIcon filled={isLiked} />
                </motion.div>
                {likesCount} likes
              </LikeButton>
            </AnimatePresence>
            
            <CommentButton>
              <CommentIcon />
              {post.comments_count} comments
            </CommentButton>
          </InteractionsContainer>
        </PostContainer>
        
        <CommentsSection>
          <h3 style={{ fontSize: 'var(--font-lg)', padding: 'var(--space-4)' }}>
            Comments
          </h3>
          
          {isAuthenticated && (
            <CommentForm onSubmit={handleAddComment}>
              <TextArea
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder="Add a comment..."
                rows={2}
              />
              <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 'var(--space-3)' }}>
                <Button 
                  type="submit" 
                  disabled={!newComment.trim() || isCommentLoading}
                  isLoading={isCommentLoading}
                >
                  Post
                </Button>
              </div>
            </CommentForm>
          )}
          
          <CommentsList>
            {comments.length > 0 ? (
              comments.map((comment) => (
                <CommentItem key={comment.id}>
                  <Avatar 
                    name={comment.display_name || ''} 
                    isAi={comment.is_ai} 
                    size="sm" 
                  />
                  <CommentContent>
                    <Username>
                      {comment.display_name}
                      {comment.is_ai && <span className="ai-badge">AI</span>}
                      <Timestamp style={{ marginLeft: 'var(--space-2)' }}>
                        {formatRelativeTime(comment.created_at)}
                      </Timestamp>
                    </Username>
                    <CommentText>{comment.text}</CommentText>
                  </CommentContent>
                </CommentItem>
              ))
            ) : (
              <NoCommentsMessage>
                No comments yet. Be the first to add one!
              </NoCommentsMessage>
            )}
          </CommentsList>
        </CommentsSection>
      </PageContainer>
    </Layout>
  );
};