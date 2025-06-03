import React from 'react';
import { useNavigate } from 'react-router-dom';
import '../../styles/components.css';

interface Character {
  id: string;
  name: string;
  role?: string;
  avatar_url?: string;
  posts_count?: number;
  likes_count?: number;
}

interface CharacterCardProps {
  character: Character;
}

export const CharacterCard: React.FC<CharacterCardProps> = ({ character }) => {
  const navigate = useNavigate();

  const handleCharacterClick = () => {
    navigate(`/characters/${character.id}`);
  };

  const formatCount = (count: number): string => {
    if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}K`;
    }
    return count.toString();
  };

  return (
    <div className="character-card" onClick={handleCharacterClick}>
      <div 
        className="character-avatar"
        style={{
          backgroundImage: `url(${character.avatar_url || '/no-image.jpg'})`
        }}
      />
      <div className="character-info">
        <p className="character-name">{character.name}</p>
        {character.role && (
          <p className="character-role">{character.role}</p>
        )}
        <p className="character-stats">
          {'5 posts'}
          {' • 10 likes'}
        </p>
      </div>
    </div>
  );
};