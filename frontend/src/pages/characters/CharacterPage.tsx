import React, { useEffect, useState } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { characterAPI } from '../../api/services';
import { Post } from '../../types';
import { Layout } from '../../components/layout/Layout';
import { Loader } from '../../components/ui/Loader';
import '../../styles/pages/character-profile.css';

// Mock data for character details (since not available in API)
const mockCharacterDetails = {
  bio: {
    description: "Lyra Starweaver is one of Eldoria's most innovative Arcane Engineers, specializing in the fusion of magical energies with mechanical systems. Born into a family of traditional enchanters, she broke convention by studying both mystical arts and engineering at the Academy of Convergent Sciences.",
    secondDescription: "Her groundbreaking research focuses on crystal resonance technology, which allows magical energies to power complex mechanical devices. She's particularly known for her work on the Cloud Bridge stabilization systems and the development of the first magically-enhanced construction automatons.",
    traits: [
      "Curious and Innovative",
      "Methodical Problem Solver", 
      "Collaborative Spirit",
      "Respectful of Tradition"
    ],
    specializations: [
      "Crystal Resonance Technology",
      "Magical-Mechanical Fusion",
      "Infrastructure Engineering", 
      "Automation Systems"
    ]
  },
  stats: {
    likes: 456,
    comments: 89,
    daysActive: 12
  },
  world: {
    id: 'eldoria',
    name: 'The Lost City of Eldoria',
    icon: '/no-image.jpg'
  }
};

export const CharacterPage: React.FC = () => {
  const { characterId } = useParams<{ characterId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
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

  const handleWorldClick = () => {
    if (character?.world_id) {
      navigate(`/worlds/${character.world_id}/about`);
    }
  };

  const handlePostClick = (post: Post) => {
    navigate(`/worlds/${post.world_id}/posts/${post.id}`);
  };

  if (loading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <Loader />
        </div>
      </Layout>
    );
  }

  if (error || !character) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="text-center">
            <h2 className="text-xl font-bold text-gray-800 mb-2">Character not found</h2>
            <p className="text-gray-600 mb-4">{error || 'The character you\'re looking for doesn\'t exist.'}</p>
            <button 
              onClick={() => navigate('/worlds')}
              className="btn btn-primary"
            >
              Browse Worlds
            </button>
          </div>
        </div>
      </Layout>
    );
  }

  const characterStats = [
    { label: 'Posts', value: posts.length },
    { label: 'Likes', value: mockCharacterDetails.stats.likes },
    { label: 'Comments', value: mockCharacterDetails.stats.comments },
    { label: 'Days Active', value: mockCharacterDetails.stats.daysActive }
  ];

  return (
    <Layout>
      <div className="min-h-screen flex flex-col bg-white">
        <main className="profile-page">
          <div className="profile-container">
            
            {/* CHARACTER HEADER */}
            <div className="character-header">
              <div className="character-info">
                
                {/* Character Avatar */}
                <div 
                  className="character-avatar-large" 
                  style={{ 
                    backgroundImage: `url('${character.avatar_url || '/no-image.jpg'}')` 
                  }}
                ></div>
                
                {/* Character Details */}
                <div className="character-details">
                  <h1 className="character-name">
                    {character.display_name}
                    {character.is_ai && <span className="ml-2 text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">AI</span>}
                  </h1>
                  <p className="character-role">{character.role || 'Arcane Engineer'}</p>
                  
                  {/* World Badge */}
                  <div className="world-badge">
                    <div 
                      className="world-icon" 
                      style={{ backgroundImage: `url('${mockCharacterDetails.world.icon}')` }}
                    ></div>
                    <div className="world-info">
                      <p>Lives in {mockCharacterDetails.world.name}</p>
                      <button 
                        onClick={handleWorldClick}
                        className="world-link"
                      >
                        Visit World →
                      </button>
                    </div>
                  </div>
                  
                  {/* Character Stats */}
                  <div className="character-stats">
                    {characterStats.map((stat, index) => (
                      <div key={index} className="stat-item">
                        <div className="stat-value">{stat.value}</div>
                        <div className="stat-label">{stat.label}</div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {/* CHARACTER BIO */}
            <div className="character-bio">
              <h2 className="bio-title">About {character.display_name}</h2>
              <div className="bio-content">
                <p className="bio-text">
                  {character.bio || mockCharacterDetails.bio.description}
                </p>
                <p className="bio-text secondary">
                  {mockCharacterDetails.bio.secondDescription}
                </p>
                
                {/* Character Traits */}
                <div className="traits-grid">
                  <div className="trait-section">
                    <h3>Personality Traits</h3>
                    <div className="trait-list">
                      {mockCharacterDetails.bio.traits.map((trait, index) => (
                        <div key={index} className="trait-item">
                          <span className="trait-dot primary"></span>
                          <span className="trait-text">{trait}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                  
                  <div className="trait-section">
                    <h3>Specializations</h3>
                    <div className="trait-list">
                      {mockCharacterDetails.bio.specializations.map((spec, index) => (
                        <div key={index} className="trait-item">
                          <span className="trait-dot success"></span>
                          <span className="trait-text">{spec}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* CHARACTER POSTS */}
            <div className="character-posts">
              <h2 className="posts-title">Recent Posts</h2>
              
              {/* Posts Grid */}
              <div className="posts-grid">
                {posts.length > 0 ? (
                  posts.map(post => (
                    <div 
                      key={post.id}
                      className="post-item"
                      onClick={() => handlePostClick(post)}
                    >
                      <div 
                        className="post-image" 
                        style={{ 
                          backgroundImage: `url('${post.image_url || post.media_url || '/no-image.jpg'}')` 
                        }}
                      ></div>
                      <div className="post-content">
                        <p className="post-text">{post.caption}</p>
                        <div className="post-meta">
                          <span>❤️ {post.likes_count}</span>
                          <span>💬 {post.comments_count}</span>
                          <span>{new Date(post.created_at).toLocaleDateString()}</span>
                        </div>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="col-span-full text-center py-8 text-gray-500">
                    No posts yet
                  </div>
                )}
              </div>
            </div>

          </div>
        </main>
      </div>
    </Layout>
  );
};