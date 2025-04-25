import React, { useEffect, useState } from 'react';
import { useParams, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import { motion } from 'framer-motion';
import { characterAPI } from '../../api/services';
import { Post } from '../../types';
import { PostCard } from '../../components/common/PostCard';
import { Avatar } from '../../components/ui/Avatar';
import { Loader } from '../../components/ui/Loader';

const PageContainer = styled.div`
  max-width: 800px;
  margin: 0 auto;
  padding: var(--space-6);
`;

const CharacterHeader = styled(motion.div)`
  background-color: var(--color-card);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  margin-bottom: var(--space-6);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  gap: var(--space-6);
`;

const CharacterInfo = styled.div`
  flex: 1;
`;

const CharacterName = styled.h1`
  font-size: var(--font-2xl);
  font-weight: 600;
  margin-bottom: var(--space-2);
  display: flex;
  align-items: center;
  gap: var(--space-3);
`;

const CharacterBio = styled.p`
  color: var(--color-text-light);
  font-size: var(--font-md);
  line-height: 1.5;
`;

const PostsContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
`;

const NoPosts = styled.div`
  text-align: center;
  padding: var(--space-8);
  color: var(--color-text-light);
  font-size: var(--font-lg);
`;

export const CharacterPage: React.FC = () => {
  const { characterId } = useParams<{ characterId: string }>();
  const location = useLocation();
  const [character, setCharacter] = useState<any>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCharacterData = async () => {
      if (!characterId) return;

      try {
        setLoading(true);
        const characterData = await characterAPI.getCharacter(characterId);
        setCharacter(characterData);

        // Получаем worldId из state или из первого поста
        let worldId = location.state?.worldId;
        if (!worldId && characterData.world_id) {
          worldId = characterData.world_id;
        }

        if (worldId) {
          const postsData = await characterAPI.getCharacterPosts(worldId, characterId);
          setPosts(postsData.posts);
        }
      } catch (err) {
        setError('Failed to load character data');
        console.error('Error fetching character data:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchCharacterData();
  }, [characterId, location.state]);

  if (loading) {
    return (
      <PageContainer>
        <Loader />
      </PageContainer>
    );
  }

  if (error || !character) {
    return (
      <PageContainer>
        <div>Error: {error || 'Character not found'}</div>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <CharacterHeader
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <Avatar
          name={character.display_name}
          isAi={character.is_ai}
          size="xl"
        />
        <CharacterInfo>
          <CharacterName>
            {character.display_name}
            {character.is_ai && <span className="ai-badge">AI</span>}
          </CharacterName>
          <CharacterBio>
            {character.bio || 'No bio available yet'}
          </CharacterBio>
        </CharacterInfo>
      </CharacterHeader>

      <PostsContainer>
        {posts.length > 0 ? (
          posts.map(post => (
            <PostCard
              key={post.id}
              post={post}
              currentWorldId={post.world_id}
            />
          ))
        ) : (
          <NoPosts>No posts yet</NoPosts>
        )}
      </PostsContainer>
    </PageContainer>
  );
}; 