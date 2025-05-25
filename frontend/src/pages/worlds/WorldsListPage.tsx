import React, { useEffect, useState, useRef } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, AnimatePresence, HTMLMotionProps } from 'framer-motion';
import { Layout } from '../../components/layout/Layout';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Loader } from '../../components/ui/Loader';
import { useWorld } from '../../hooks/useWorld';
import { useInfiniteScroll } from '../../hooks/useInfiniteScroll';
import { worldsAPI } from '../../api/services';

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);

  @media (max-width: 640px) {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-4);
  }
`;

const Title = styled.h1`
  font-size: var(--font-3xl);
  color: var(--color-text);
`;

const FiltersBar = styled.div`
  display: flex;
  gap: var(--space-3);
  margin-bottom: var(--space-6);
  flex-wrap: wrap;
`;

const FilterButton = styled.button<{ $isActive: boolean }>`
  padding: var(--space-2) var(--space-4);
  background-color: ${props => props.$isActive ? 'var(--color-primary)' : 'var(--color-input-bg)'};
  color: ${props => props.$isActive ? 'white' : 'var(--color-text)'};
  border: none;
  border-radius: var(--radius-full);
  font-size: var(--font-sm);
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background-color: ${props => props.$isActive ? 'var(--color-primary-hover)' : 'var(--color-border)'};
  }
`;

const WorldsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-6);
`;

const WorldCard = styled(motion(Card))`
  overflow: hidden;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  display: flex;
  flex-direction: column;

  &:hover {
    transform: translateY(-4px);
    box-shadow: var(--shadow-lg);
  }
`;

const WorldImage = styled.div<{ $index: number; $backgroundImage?: string }>`
  height: 140px;
  background: ${props => props.$backgroundImage
    ? `url(${props.$backgroundImage}) center/cover no-repeat`
    : `linear-gradient(135deg,
        ${(() => {
          const colors = [
            'var(--color-primary), #FF9900',
            '#A78BFA, var(--color-secondary)',
            'var(--color-accent), #FB7185',
            '#6EE7B7, #34D399',
            '#60A5FA, #3B82F6'
          ];
          return colors[props.$index % colors.length];
        })()}
      )`
  };
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 48px;
  font-weight: bold;
  position: relative;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.2);
    opacity: ${props => props.$backgroundImage ? 0.4 : 0};
  }
`;

const WorldContent = styled.div`
  padding: var(--space-4);
  flex: 1;
  display: flex;
  flex-direction: column;
`;

const WorldName = styled.h3`
  font-size: var(--font-lg);
  margin-bottom: var(--space-2);
`;

const WorldDescription = styled.p`
  font-size: var(--font-sm);
  color: var(--color-text);
  margin-bottom: var(--space-3);
  line-height: 1.5;
  flex: 1;

  /* Limit to 3 lines with ellipsis */
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
`;

const WorldStats = styled.div`
  display: flex;
  gap: var(--space-4);
  margin-bottom: var(--space-3);
  font-size: var(--font-sm);
`;

const StatItem = styled.div`
  color: var(--color-text);

  span {
    font-weight: 600;
    color: var(--color-text);
    margin-right: 4px;
  }
`;

const EmptyState = styled.div`
  text-align: center;
  padding: var(--space-16) var(--space-4);

  h3 {
    font-size: var(--font-xl);
    margin-bottom: var(--space-4);
  }

  p {
    color: var(--color-text);
    margin-bottom: var(--space-6);
    max-width: 500px;
    margin-left: auto;
    margin-right: auto;
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

type FilterType = 'all' | 'joined' | 'created' | 'popular' | 'new';

export const WorldsListPage: React.FC = () => {
  const { worlds, loadWorlds, joinWorld, currentWorld, loadCurrentWorld, isLoading: isWorldLoading, error } = useWorld();
  const [filter, setFilter] = useState<FilterType>('all');
  const [filteredWorlds, setFilteredWorlds] = useState(worlds);
  const [isJoining, setIsJoining] = useState<Record<string, boolean>>({});
  const navigate = useNavigate();

  const {
    items: infiniteWorlds,
    isLoading,
    error: scrollError,
    loadMore,
    reset,
    sentinelRef
  } = useInfiniteScroll({
    fetchItems: async (limit, cursor) => {
      const response = await worldsAPI.getWorlds(limit, cursor);
      return {
        items: response.worlds || [],
        nextCursor: response.next_cursor || '',
        hasMore: response.has_more || false
      };
    },
    limit: 12
  });

  // Use a ref to track if we've already applied the filter for this world state
  const worldsFilteredRef = useRef(false);

  useEffect(() => {
    // Skip filter application if the world state hasn't changed
    // This prevents continuous re-renders
    if (worldsFilteredRef.current && filter === 'all') return;
    worldsFilteredRef.current = true;

    applyFilter(filter);
  }, [worlds, filter]);

  const applyFilter = (filterType: FilterType) => {
    let result = [...worlds];

    switch (filterType) {
      case 'joined':
        result = result.filter(world => world.is_joined);
        break;
      case 'created':
        // This would require having creator_id information
        // For now, we're just showing all worlds
        break;
      case 'popular':
        result = result.sort((a, b) => b.users_count - a.users_count);
        break;
      case 'new':
        result = result.sort((a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
        break;
      default:
        // 'all' filter, no change needed
        break;
    }

    setFilteredWorlds(result);
  };

  const handleFilterChange = (newFilter: FilterType) => {
    setFilter(newFilter);
    reset();
  };

  const handleJoinWorld = async (worldId: string) => {
    setIsJoining(prev => ({ ...prev, [worldId]: true }));

    try {
      await joinWorld(worldId);
    } catch (error) {
      console.error('Failed to join world:', error);
    } finally {
      setIsJoining(prev => ({ ...prev, [worldId]: false }));
    }
  };

  const handleSwitchWorld = (worldId: string) => {
    loadCurrentWorld(worldId);
  };

  const worldCardVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: (i: number) => ({
      opacity: 1,
      y: 0,
      transition: {
        delay: i * 0.05,
        duration: 0.3,
        ease: 'easeOut'
      }
    })
  };

  return (
    <Layout>
      <PageHeader>
        <Title>Explore Worlds</Title>
        <Link to="/create-world">
          <Button variant="primary">Create World</Button>
        </Link>
      </PageHeader>

      <FiltersBar>
        <FilterButton
          $isActive={filter === 'all'}
          onClick={() => handleFilterChange('all')}
        >
          All Worlds
        </FilterButton>
        <FilterButton
          $isActive={filter === 'joined'}
          onClick={() => handleFilterChange('joined')}
        >
          Joined
        </FilterButton>
        <FilterButton
          $isActive={filter === 'popular'}
          onClick={() => handleFilterChange('popular')}
        >
          Popular
        </FilterButton>
        <FilterButton
          $isActive={filter === 'new'}
          onClick={() => handleFilterChange('new')}
        >
          Newest
        </FilterButton>
      </FiltersBar>

      {error && (
        <ErrorMessage>
          {error}
        </ErrorMessage>
      )}

      <AnimatePresence>
        {filteredWorlds.length > 0 ? (
          <WorldsGrid>
            {filteredWorlds.map((world, index) => (
              <WorldCard
                key={world.id}
                variants={worldCardVariants}
                initial="hidden"
                animate="visible"
                custom={index}
                variant="elevated"
              >
                <WorldImage
                  $index={index}
                  $backgroundImage={world.image_url}
                >
                  {!world.image_url && world.name.charAt(0)}
                </WorldImage>
                <WorldContent>
                  <WorldName>{world.name}</WorldName>
                  <WorldDescription>
                    {world.description || 'No description'}
                  </WorldDescription>
                  <WorldStats>
                    <StatItem>
                      <span>{world.users_count}</span> users
                    </StatItem>
                    <StatItem>
                      <span>{world.posts_count}</span> posts
                    </StatItem>
                  </WorldStats>
                  <Button
                    variant={world.is_joined ? 'ghost' : 'primary'}
                    fullWidth
                    isLoading={isJoining[world.id]}
                    onClick={() => {
                      if (world.is_joined) {
                        navigate(`/worlds/${world.id}/feed`);
                      } else {
                        handleJoinWorld(world.id);
                      }
                    }}
                  >
                    {world.is_joined ? 'Open World' : 'Join'}
                  </Button>
                </WorldContent>
              </WorldCard>
            ))}
          </WorldsGrid>
        ) : isLoading || isWorldLoading ? (
          <LoaderContainer>
            <Loader text="Loading worlds..." />
          </LoaderContainer>
        ) : (
          <EmptyState>
            <h3>No worlds found</h3>
            <p>
              {filter !== 'all'
                ? "Try changing your filter or create your own world!"
                : "Be the first to create a synthetic world!"}
            </p>
            <Link to="/create-world">
              <Button variant="primary" size="large">
                Create New World
              </Button>
            </Link>
          </EmptyState>
        )}
      </AnimatePresence>

      {/* Infinite scroll sentinel */}
      {filteredWorlds.length > 0 && (
        <div ref={sentinelRef}>
          {isLoading && (
            <LoaderContainer>
              <Loader size="sm" />
            </LoaderContainer>
          )}
        </div>
      )}
    </Layout>
  );
};