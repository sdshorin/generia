import React from 'react';
import { useNavigate } from 'react-router-dom';
import { World } from '../../types';
import '../../styles/components.css';

interface WorldCardProps {
  world: World;
  badge?: string; // Optional badge like "Popular", "New"
}

export const WorldCard: React.FC<WorldCardProps> = ({ world, badge }) => {
  const navigate = useNavigate();

  const handleWorldClick = () => {
    navigate(`/worlds/${world.id}/feed`);
  };

  const handleEnterClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    navigate(`/worlds/${world.id}/feed`);
  };

  const formatCount = (count: number): string => {
    if (count >= 1000) {
      return `${(count / 1000).toFixed(1)}K`;
    }
    return count.toString();
  };

  return (
    <div className="world-card" onClick={handleWorldClick}>
      {/* World Cover */}
      <div className="world-card-image-container">
        <div 
          className="world-card-image"
          style={{
            backgroundImage: `url(${world.image_url || '/no-image.jpg'})`
          }}
        />
        {/* World Icon */}
        <div 
          className="world-card-icon"
          style={{
            backgroundImage: `url(${world.icon_url || '/no-image.jpg'})`
          }}
        />
        {/* Badge (optional) */}
        {badge && (
          <div className="world-card-badge">{badge}</div>
        )}
      </div>
      
      {/* World Info */}
      <div className="world-card-body">
        <h3 className="world-card-title">{world.name}</h3>
        <p className="world-card-description">
          {world.description || world.prompt}
        </p>
        
        {/* Stats */}
        <div className="world-card-stats">
          <span>✨ {world.users_count || 0} Characters</span>
          <span>📸 {formatCount(world.posts_count || 0)} Posts</span>
          <span>❤️ {formatCount(10)}</span>
        </div>
        
        {/* Enter Button */}
        <button className="world-card-btn" onClick={handleEnterClick}>
          {world.is_joined ? 'Enter World' : 'Join World'}
        </button>
      </div>
    </div>
  );
};