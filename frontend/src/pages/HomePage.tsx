import React, { useEffect, useState, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { BiPlus } from 'react-icons/bi';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { Layout } from '../components/layout/Layout';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Avatar } from '../components/ui/Avatar';
import { Loader } from '../components/ui/Loader';
import { PostCard } from '../components/common/PostCard';
import { useAuth } from '../hooks/useAuth';
import { useWorld } from '../hooks/useWorld';
import { World, Post } from '../types';
import { worldsAPI, postsAPI } from '../api/services';

const HeroSection = styled.section`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  min-height: 50vh;
  padding: var(--space-16) var(--space-4);
  position: relative;
  overflow: hidden;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: radial-gradient(circle at center, rgba(213, 184, 152, 0.25) 0%, rgba(255, 255, 255, 0) 70%);
    z-index: -1;
  }
`;

const HeroTitle = styled(motion.h1)<HTMLMotionProps<'h1'>>`
  font-size: clamp(2.5rem, 5vw, 4.5rem);
  margin-bottom: var(--space-4);
  background: linear-gradient(135deg, #D5B898, #A78BFA);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1.2;
  font-weight: 700;
`;

const HeroSubtitle = styled(motion.h2)<HTMLMotionProps<'h2'>>`
  font-size: clamp(1.125rem, 2vw, 1.5rem);
  color: var(--color-text);
  font-weight: 400;
  margin-bottom: var(--space-8);
  max-width: 600px;
`;

const MainContent = styled.div`
  display: grid;
  grid-template-columns: minmax(0, 1fr) 350px;
  gap: var(--space-6);
  margin-top: var(--space-10);
  
  @media (max-width: 968px) {
    grid-template-columns: 1fr;
  }
`;

const WorldsPanel = styled.div`
  order: 2;
  
  @media (max-width: 968px) {
    order: 2;
  }
`;

const FeedPanel = styled.div`
  order: 1;
  
  @media (max-width: 968px) {
    order: 1;
  }
`;

const PanelTitle = styled.h3`
  font-size: var(--font-xl);
  margin-bottom: var(--space-4);
  color: var(--color-text);
  display: flex;
  justify-content: space-between;
  align-items: center;
`;

const CreatePostButton = styled(Link)`
  font-size: var(--font-sm);
  background-color: var(--color-primary);
  color: var(--color-text);
  font-weight: 600;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  transition: all 0.2s ease;
  
  &:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
    color: var(--color-text);
  }
`;

const WorldsList = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
`;

const WorldItem = styled(Card)<{ $isActive?: boolean }>`
  display: flex;
  align-items: center;
  padding: var(--space-4);
  transition: all 0.2s ease;
  cursor: pointer;
  border-left: ${props => props.$isActive ? '4px solid var(--color-primary)' : '4px solid transparent'};
  background-color: ${props => props.$isActive ? 'rgba(213, 184, 152, 0.2)' : 'var(--color-card)'};
  position: relative;
  
  &:hover {
    transform: translateY(-2px);
  }
  
  &:after {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 4px;
    background-color: ${props => props.$isActive ? 'var(--color-primary)' : 'transparent'};
  }
`;

const WorldInfo = styled.div`
  margin-left: var(--space-3);
  flex: 1;
`;

const WorldName = styled.h4`
  font-size: var(--font-md);
  margin-bottom: var(--space-1);
`;

const WorldDescription = styled.p`
  font-size: var(--font-sm);
  color: var(--color-text);
  margin-bottom: var(--space-2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
`;

const PortalCircle = styled.div<{ $index: number }>`
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, 
    ${props => {
      const colors = [
        'var(--color-primary), #FF9900',
        '#A78BFA, var(--color-secondary)',
        'var(--color-accent), #FB7185',
        '#6EE7B7, #34D399',
        '#60A5FA, #3B82F6'
      ];
      return colors[props.$index % colors.length];
    }}
  );
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 20px;
`;

const PostsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-6);
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
`;

const NoPostsMessage = styled.div`
  text-align: center;
  padding: var(--space-8) var(--space-4);
  color: var(--color-text);
  
  h4 {
    margin-bottom: var(--space-2);
  }
  
  p {
    margin-bottom: var(--space-6);
  }
`;

const heroVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.6,
      ease: "easeOut"
    }
  }
};

export const HomePage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const { worlds, currentWorld, loadWorlds, loadCurrentWorld } = useWorld();
  const [popularWorlds, setPopularWorlds] = useState<World[]>([]);
  const [recentPosts, setRecentPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingPosts, setIsLoadingPosts] = useState(false);
  const navigate = useNavigate();
  
  useEffect(() => {
    const fetchPopularWorlds = async () => {
      if (!isAuthenticated) return;
      
      try {
        setIsLoading(true);
        const data = await worldsAPI.getWorlds(5, '');
        setPopularWorlds(data.worlds || []);
        
        // If we have worlds but no current world, load the first one
        if (data.worlds?.length > 0 && !currentWorld) {
          loadCurrentWorld(data.worlds[0].id);
        }
      } catch (error) {
        console.error('Failed to fetch popular worlds:', error);
      } finally {
        setIsLoading(false);
      }
    };
    
    fetchPopularWorlds();
  }, [isAuthenticated, loadCurrentWorld]);
  
  const currentWorldRef = useRef<string | null>(null);
  
  useEffect(() => {
    const fetchRecentPosts = async () => {
      if (!isAuthenticated || !currentWorld) return;
      
      // Skip if we've already loaded posts for this world
      if (currentWorldRef.current === currentWorld.id) return;
      currentWorldRef.current = currentWorld.id;
      
      try {
        setIsLoadingPosts(true);
        setRecentPosts([]); // Очистить прежние посты во время загрузки
        const data = await postsAPI.getFeed(currentWorld.id, 6, '');
        setRecentPosts(data.posts || []);
      } catch (error) {
        console.error('Failed to fetch recent posts:', error);
      } finally {
        setIsLoadingPosts(false);
      }
    };
    
    if (currentWorld) {
      fetchRecentPosts();
    }
  }, [isAuthenticated, currentWorld]); // Зависит от всего объекта currentWorld, чтобы реагировать на его изменения
  
  // Navigate directly to world feed instead of switching
  
  const handleCreateWorld = () => {
    navigate('/create-world');
  };
  
  return (
    <Layout>
      <HeroSection>
        <HeroTitle
          variants={heroVariants}
          initial="hidden"
          animate="visible"
        >
          Create Your Own Synthetic World
        </HeroTitle>
        <HeroSubtitle
          variants={heroVariants}
          initial="hidden"
          animate="visible"
          transition={{ delay: 0.2 }}
        >
          Use a simple prompt to open the portal. Watch life unfold.
        </HeroSubtitle>
        <motion.div
          variants={heroVariants}
          initial="hidden"
          animate="visible"
          transition={{ delay: 0.4 }}
        >
          <Button 
            onClick={handleCreateWorld} 
            size="large"
          >
            Generate a World
          </Button>
        </motion.div>
      </HeroSection>
      
      {isAuthenticated && (
        <MainContent>
          <WorldsPanel>
            <PanelTitle>Popular Worlds</PanelTitle>
            <WorldsList>
              {popularWorlds.map((world, index) => (
                <WorldItem 
                  key={world.id} 
                  variant="elevated" 
                  animateHover
                  $isActive={currentWorld?.id === world.id}
                  onClick={() => {
                    loadCurrentWorld(world.id);
                    currentWorldRef.current = null; // Сбросить для загрузки постов нового мира
                  }}
                >
                  <PortalCircle $index={index}>
                    {world.name.charAt(0)}
                  </PortalCircle>
                  <WorldInfo>
                    <WorldName>{world.name}</WorldName>
                    <WorldDescription>{world.description || 'No description'}</WorldDescription>
                  </WorldInfo>
                  <Button 
                    size="small" 
                    variant="primary"
                    onClick={(e) => {
                      e.stopPropagation(); // Предотвратить запуск клика родителя
                      navigate(`/worlds/${world.id}/feed`);
                    }}
                  >
                    Open World
                  </Button>
                </WorldItem>
              ))}
              <Link to="/worlds">
                <Button variant="ghost" fullWidth>
                  View All Worlds
                </Button>
              </Link>
            </WorldsList>
          </WorldsPanel>
          
          <FeedPanel>
            <PanelTitle>
              {currentWorld ? `Posts from ${currentWorld.name}` : 'Select a World'}
              {currentWorld && (
                <CreatePostButton to={`/worlds/${currentWorld.id}/create`}>
                  <BiPlus style={{ marginRight: '4px' }} /> Create Post
                </CreatePostButton>
              )}
            </PanelTitle>
            {currentWorld ? (
              <>
                {isLoadingPosts ? (
                  <div style={{ display: 'flex', justifyContent: 'center', padding: 'var(--space-8)' }}>
                    <Loader size="md" text="Loading posts..." />
                  </div>
                ) : recentPosts.length > 0 ? (
                  <>
                    <PostsGrid>
                      {recentPosts.map(post => (
                        <PostCard 
                          key={post.id} 
                          post={post} 
                          currentWorldId={currentWorld.id} 
                        />
                      ))}
                    </PostsGrid>
                    <Link to={`/worlds/${currentWorld.id}/feed`}>
                      <Button variant="ghost" fullWidth>
                        View All Posts
                      </Button>
                    </Link>
                  </>
                ) : (
                  <NoPostsMessage>
                    <h4>No posts yet</h4>
                    <p>Be the first to create content in this world!</p>
                    <Link to={`/worlds/${currentWorld.id}/create`}>
                      <Button>Create a Post</Button>
                    </Link>
                  </NoPostsMessage>
                )}
              </>
            ) : (
              <Card variant="elevated" padding="var(--space-6)">
                <NoPostsMessage>
                  <h4>No world selected</h4>
                  <p>Select a world from the right panel to see posts</p>
                  <Link to="/worlds">
                    <Button>Browse Worlds</Button>
                  </Link>
                </NoPostsMessage>
              </Card>
            )}
          </FeedPanel>
        </MainContent>
      )}
    </Layout>
  );
};