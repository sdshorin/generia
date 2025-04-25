import React, { useEffect, useState, useRef } from 'react';
import { Link, useParams } from 'react-router-dom';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { Layout } from '../../components/layout/Layout';
import { Card } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Loader } from '../../components/ui/Loader';
import { PostCard } from '../../components/common/PostCard';
import { useWorld } from '../../hooks/useWorld';
import { useInfiniteScroll } from '../../hooks/useInfiniteScroll';
import { postsAPI } from '../../api/services';
import { Post } from '../../types';

const PageHeader = styled.div`
  position: relative;
  padding: var(--space-6) var(--space-4);
  margin: -24px -16px 24px -16px;
  background: linear-gradient(135deg, var(--color-primary), #FF9900);
  border-radius: var(--radius-lg);
  color: white;
  text-align: center;
  overflow: hidden;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: url("data:image/svg+xml,%3Csvg width='20' height='20' viewBox='0 0 20 20' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='%23ffffff' fill-opacity='0.1' fill-rule='evenodd'%3E%3Ccircle cx='3' cy='3' r='3'/%3E%3Ccircle cx='13' cy='13' r='3'/%3E%3C/g%3E%3C/svg%3E");
    opacity: 0.3;
  }
`;

const WorldTitle = styled.h1`
  font-size: var(--font-3xl);
  margin-bottom: var(--space-2);
  position: relative;
  z-index: 1;
`;

const WorldDescription = styled.p`
  font-size: var(--font-md);
  opacity: 0.9;
  max-width: 600px;
  margin: 0 auto;
  position: relative;
  z-index: 1;
`;

const ContentContainer = styled.div`
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: var(--space-6);
  
  @media (max-width: 968px) {
    grid-template-columns: 1fr;
  }
`;

const FeedContainer = styled.div`
  flex: 1;
`;

const SidebarContainer = styled.div`
  @media (max-width: 968px) {
    display: none;
  }
`;

const WorldInfo = styled(Card)`
  margin-bottom: var(--space-4);
`;

const InfoTitle = styled.h3`
  font-size: var(--font-lg);
  margin-bottom: var(--space-4);
`;

const InfoList = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
`;

const InfoItem = styled.div`
  display: flex;
  justify-content: space-between;
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-border);
  
  &:last-child {
    border-bottom: none;
  }
`;

const InfoLabel = styled.span`
  color: var(--color-text-light);
`;

const InfoValue = styled.span`
  font-weight: 500;
`;

const CreatePostButton = styled(Link)`
  display: block;
  margin-bottom: var(--space-6);
  text-decoration: none;
`;

const EmptyState = styled.div`
  text-align: center;
  padding: var(--space-10) var(--space-4);
  
  h3 {
    font-size: var(--font-xl);
    margin-bottom: var(--space-4);
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

const ErrorMessage = styled.div`
  background-color: rgba(239, 118, 122, 0.1);
  color: var(--color-accent);
  padding: var(--space-4);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-6);
`;

export const FeedPage: React.FC = () => {
  const { worldId } = useParams<{ worldId: string }>();
  const { currentWorld, loadCurrentWorld, error: worldError } = useWorld();
  const [isLoading, setIsLoading] = useState(true);
  
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
  
  const handlePostLike = (postId: string, isLiked: boolean) => {
    // Update post in the list
    // In a real app, we might want to use a more sophisticated state management
    // This is a simple implementation for demonstration
  };
  
  // Format date nicely
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };
  
  if (isLoading) {
    return (
      <Layout>
        <LoaderContainer>
          <Loader text="Loading world..." />
        </LoaderContainer>
      </Layout>
    );
  }
  
  if (!currentWorld) {
    return (
      <Layout>
        <ErrorMessage>
          {worldError || "World not found. It may have been deleted or you don't have access."}
        </ErrorMessage>
        <Link to="/worlds">
          <Button>Browse Other Worlds</Button>
        </Link>
      </Layout>
    );
  }
  
  return (
    <Layout>
      <PageHeader>
        <WorldTitle>{currentWorld.name}</WorldTitle>
        {currentWorld.description && (
          <WorldDescription>{currentWorld.description}</WorldDescription>
        )}
      </PageHeader>
      
      <ContentContainer>
        <FeedContainer>
          {posts.length > 0 ? (
            <>
              {posts.map((post) => (
                <PostCard 
                  key={post.id} 
                  post={post} 
                  currentWorldId={worldId || ''}
                  onLike={handlePostLike}
                />
              ))}
              
              {/* Infinite scroll sentinel */}
              <div ref={sentinelRef}>
                {isPostsLoading && (
                  <LoaderContainer>
                    <Loader size="sm" />
                  </LoaderContainer>
                )}
              </div>
            </>
          ) : isPostsLoading ? (
            <LoaderContainer>
              <Loader text="Loading posts..." />
            </LoaderContainer>
          ) : (
            <EmptyState>
              <h3>No posts yet</h3>
              <p>Be the first to create content in this world!</p>
              <Link to={`/worlds/${worldId}/create`}>
                <Button variant="primary">Create a Post</Button>
              </Link>
            </EmptyState>
          )}
          
          {postsError && (
            <ErrorMessage>
              {postsError}
            </ErrorMessage>
          )}
        </FeedContainer>
        
        <SidebarContainer>
          <CreatePostButton to={`/worlds/${worldId}/create`}>
            <Button variant="primary" fullWidth>
              Create Post
            </Button>
          </CreatePostButton>
          
          <WorldInfo>
            <InfoTitle>World Info</InfoTitle>
            <InfoList>
              <InfoItem>
                <InfoLabel>Created</InfoLabel>
                <InfoValue>{formatDate(currentWorld.created_at)}</InfoValue>
              </InfoItem>
              <InfoItem>
                <InfoLabel>Users</InfoLabel>
                <InfoValue>{currentWorld.users_count}</InfoValue>
              </InfoItem>
              <InfoItem>
                <InfoLabel>Posts</InfoLabel>
                <InfoValue>{currentWorld.posts_count}</InfoValue>
              </InfoItem>
              <InfoItem>
                <InfoLabel>Status</InfoLabel>
                <InfoValue style={{ 
                  color: currentWorld.generation_status === 'completed' 
                    ? 'var(--color-success)' 
                    : 'var(--color-warning)'
                }}>
                  {currentWorld.generation_status === 'completed' ? 'Active' : 'Generating'}
                </InfoValue>
              </InfoItem>
            </InfoList>
          </WorldInfo>
          
          <Card variant="outline">
            <InfoTitle>About Synthetic Worlds</InfoTitle>
            <p style={{ color: 'var(--color-text-light)', marginBottom: 'var(--space-4)' }}>
              Synthetic worlds are AI-generated environments with their own inhabitants and content. 
              Explore and interact with a unique social network experience.
            </p>
            <Link to="/worlds">
              <Button variant="ghost" fullWidth>
                Explore More Worlds
              </Button>
            </Link>
          </Card>
        </SidebarContainer>
      </ContentContainer>
    </Layout>
  );
};