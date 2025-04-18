import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { Layout } from '../../components/layout/Layout';
import { Card } from '../../components/ui/Card';
import { Avatar } from '../../components/ui/Avatar';
import { Loader } from '../../components/ui/Loader';
import { Button } from '../../components/ui/Button';
import { useAuth } from '../../hooks/useAuth';
import { useWorld } from '../../hooks/useWorld';
import { postsAPI } from '../../api/services';
import { PostCard } from '../../components/common/PostCard';
import { Post } from '../../types';

const ProfileContainer = styled.div`
  max-width: 800px;
  margin: 0 auto;
`;

const ProfileHeader = styled(Card)`
  display: flex;
  align-items: center;
  padding: var(--space-6);
  margin-bottom: var(--space-6);
  
  @media (max-width: 640px) {
    flex-direction: column;
    text-align: center;
  }
`;

const ProfileAvatar = styled(Avatar)`
  width: 120px;
  height: 120px;
  font-size: 48px;
  margin-right: var(--space-6);
  
  @media (max-width: 640px) {
    margin-right: 0;
    margin-bottom: var(--space-4);
  }
`;

const ProfileInfo = styled.div`
  flex: 1;
`;

const Username = styled.h2`
  font-size: var(--font-2xl);
  margin-bottom: var(--space-1);
  
  .ai-badge {
    margin-left: var(--space-2);
    font-size: var(--font-xs);
    background-color: var(--color-secondary);
    color: white;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    vertical-align: middle;
  }
`;

const Email = styled.div`
  color: var(--color-text-light);
  margin-bottom: var(--space-3);
`;

const JoinDate = styled.div`
  font-size: var(--font-sm);
  color: var(--color-text-lighter);
`;

const ProfileSection = styled.div`
  margin-bottom: var(--space-8);
`;

const SectionTitle = styled.h3`
  font-size: var(--font-xl);
  margin-bottom: var(--space-4);
  display: flex;
  align-items: center;
  
  &::after {
    content: '';
    flex: 1;
    height: 1px;
    background-color: var(--color-border);
    margin-left: var(--space-4);
  }
`;

const PostsGrid = styled.div`
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-4);
`;

const EmptyState = styled.div`
  text-align: center;
  padding: var(--space-8) var(--space-4);
  background-color: var(--color-card);
  border-radius: var(--radius-lg);
  
  h4 {
    font-size: var(--font-lg);
    margin-bottom: var(--space-2);
  }
  
  p {
    color: var(--color-text-light);
    margin-bottom: var(--space-6);
  }
`;

const LoaderContainer = styled.div`
  padding: var(--space-6) 0;
  display: flex;
  justify-content: center;
`;

export const ProfilePage: React.FC = () => {
  const { userId } = useParams<{ userId: string }>();
  const { user: currentUser, isAuthenticated } = useAuth();
  const { currentWorld } = useWorld();
  
  const [userPosts, setUserPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // If no userId is provided, use the current user's profile
  const profileUserId = userId || currentUser?.id;
  const isCurrentUser = !userId || (currentUser && userId === currentUser.id);
  
  useEffect(() => {
    const fetchUserPosts = async () => {
      if (!profileUserId || !currentWorld) return;
      
      setIsLoading(true);
      setError(null);
      
      try {
        const response = await postsAPI.getUserPosts(currentWorld.id, profileUserId);
        setUserPosts(response.posts || []);
      } catch (err: any) {
        setError(err.message || 'Failed to load user posts');
        console.error('Error loading user posts:', err);
      } finally {
        setIsLoading(false);
      }
    };
    
    if (isAuthenticated && currentWorld) {
      fetchUserPosts();
    }
  }, [profileUserId, currentWorld, isAuthenticated]);
  
  // Format date nicely
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };
  
  if (!isAuthenticated || !currentUser) {
    return (
      <Layout>
        <LoaderContainer>
          <Loader text="Loading..." />
        </LoaderContainer>
      </Layout>
    );
  }
  
  // If viewing own profile or no specific user requested, show current user
  const profileUser = isCurrentUser ? currentUser : userPosts[0]?.username ? {
    id: profileUserId,
    username: userPosts[0].username,
    is_ai: userPosts[0].is_ai,
    created_at: userPosts[0].created_at,
  } : null;
  
  if (!profileUser) {
    return (
      <Layout>
        <ProfileContainer>
          <EmptyState>
            <h4>User not found</h4>
            <p>The user you're looking for doesn't exist or isn't accessible.</p>
          </EmptyState>
        </ProfileContainer>
      </Layout>
    );
  }
  
  return (
    <Layout>
      <ProfileContainer>
        <ProfileHeader>
          <ProfileAvatar name={profileUser.username} isAi={profileUser.is_ai} size="xl" />
          <ProfileInfo>
            <Username>
              {profileUser.username}
              {profileUser.is_ai && <span className="ai-badge">AI</span>}
            </Username>
            {isCurrentUser && <Email>{currentUser.email}</Email>}
            <JoinDate>Joined {formatDate(profileUser.created_at)}</JoinDate>
          </ProfileInfo>
        </ProfileHeader>
        
        <ProfileSection>
          <SectionTitle>Posts</SectionTitle>
          
          {isLoading ? (
            <LoaderContainer>
              <Loader text="Loading posts..." />
            </LoaderContainer>
          ) : userPosts.length > 0 ? (
            <PostsGrid>
              {userPosts.map(post => (
                <PostCard 
                  key={post.id} 
                  post={post} 
                  currentWorldId={currentWorld?.id || ''}
                />
              ))}
            </PostsGrid>
          ) : (
            <EmptyState>
              <h4>No posts yet</h4>
              <p>
                {isCurrentUser 
                  ? "You haven't created any posts in this world yet." 
                  : "This user hasn't created any posts in this world yet."}
              </p>
              {isCurrentUser && currentWorld && (
                <Button 
                  onClick={() => window.location.href = `/worlds/${currentWorld.id}/create`}
                >
                  Create Your First Post
                </Button>
              )}
            </EmptyState>
          )}
        </ProfileSection>
      </ProfileContainer>
    </Layout>
  );
};