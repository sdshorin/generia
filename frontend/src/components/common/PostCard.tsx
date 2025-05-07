import React, { useState } from 'react';
import styled from 'styled-components';
import { formatDistanceToNow } from 'date-fns';
import { motion, AnimatePresence, HTMLMotionProps } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Post } from '../../types';
import { interactionsAPI } from '../../api/services';
import { Avatar } from '../ui/Avatar';

interface PostCardProps {
  post: Post;
  currentWorldId: string;
  onLike?: (postId: string, isLiked: boolean) => void;
}

// Properly type the motion.div component
const Card = styled(motion.div)<HTMLMotionProps<'div'>>`
  background-color: var(--color-card);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  margin-bottom: var(--space-6);
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;

  &:hover {
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }
`;

const CardHeader = styled.div`
  display: flex;
  align-items: center;
  padding: var(--space-4);
`;

const UserInfo = styled.div`
  margin-left: var(--space-3);
  flex: 1;
`;

const Username = styled(Link)`
  font-weight: 600;
  display: flex;
  align-items: center;
  text-decoration: none;
  color: var(--color-text);

  &:hover {
    color: var(--color-accent);
  }

  .ai-badge {
    margin-left: var(--space-2);
    font-size: var(--font-xs);
    background-color: var(--color-secondary);
    color: var(--color-text);
    font-weight: 600;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
  }
`;

const Timestamp = styled.div`
  font-size: var(--font-xs);
  color: var(--color-text);
  font-weight: 500;
`;

const PostContent = styled.div`
  padding: ${props => props.children ? 'var(--space-4)' : '0'};
  padding-top: 0;
  font-size: var(--font-md);
  line-height: 1.5;
  color: var(--color-text);
  white-space: pre-wrap;
  overflow: visible;
`;

const PostImage = styled.img`
  width: 100%;
  max-height: 500px;
  object-fit: contain;
  background-color: var(--color-background);
`;

const CardFooter = styled.div`
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

const CommentButton = styled(Link)`
  display: flex;
  align-items: center;
  background: none;
  border: none;
  color: var(--color-text-light);
  font-size: var(--font-sm);
  cursor: pointer;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  margin-left: var(--space-4);
  text-decoration: none;

  &:hover {
    background-color: rgba(0, 0, 0, 0.05);
    color: var(--color-text);
  }

  svg {
    margin-right: var(--space-2);
  }
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

export const PostCard: React.FC<PostCardProps> = ({
  post,
  currentWorldId,
  onLike,
}) => {
  const [isLiking, setIsLiking] = useState(false);
  const [isLiked, setIsLiked] = useState(post.user_liked || false);
  const [likesCount, setLikesCount] = useState(post.likes_count);

  const handleLike = async () => {
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

  // Format timestamp to relative time (e.g., "2 hours ago")
  const formattedTime = formatDistanceToNow(new Date(post.created_at), { addSuffix: true });
  console.log(post);
  return (
    <Card
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <CardHeader>
        <Avatar
          src={post.avatar_url}
          name={post.display_name || ''}
          isAi={post.is_ai}
          size="md"
        />
        <UserInfo>
          <Username
            to={`/characters/${post.character_id}`}
            state={{ worldId: currentWorldId }}
          >
            {post.display_name}
            {post.is_ai && <span className="ai-badge">AI</span>}
          </Username>
          <Timestamp>{formattedTime}</Timestamp>
        </UserInfo>
      </CardHeader>

      {post.caption && (
        <PostContent>{post.caption}</PostContent>
      )}

      {(post.media_url || post.image_url) && (
        <PostImage src={post.media_url || post.image_url} alt="Post image" loading="lazy" />
      )}

      <CardFooter>
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
            {likesCount}
          </LikeButton>
        </AnimatePresence>

        <CommentButton to={`/worlds/${currentWorldId}/posts/${post.id}`}>
          <CommentIcon />
          {post.comments_count}
        </CommentButton>
      </CardFooter>
    </Card>
  );
};