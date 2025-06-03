import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useWorld } from '../hooks/useWorld';
import { Layout } from '../components/layout/Layout';
import { PostCard } from '../components/common/PostCard';
import { Loader } from '../components/ui/Loader';
import { World, Post } from '../types';
import { worldsAPI, postsAPI } from '../api/services';
import '../styles/pages/main.css';

export const HomePage: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const { currentWorld, loadCurrentWorld } = useWorld();
  const [popularWorlds, setPopularWorlds] = useState<World[]>([]);
  const [selectedWorld, setSelectedWorld] = useState<World | null>(null);
  const [recentPosts, setRecentPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingPosts, setIsLoadingPosts] = useState(false);
  const navigate = useNavigate();

  // Default world data for demo
  const defaultWorlds = [
    {
      id: 'eldoria',
      name: '',
      description: '',
      prompt: '',
      creator_id: 'demo',
      generation_status: 'completed',
      status: 'active',
      users_count: 0,
      posts_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      image_url: '/no-image.jpg',
      icon_url: '/no-image.jpg'
    },
    
  ];

  useEffect(() => {
    const fetchPopularWorlds = async () => {
      try {
        setIsLoading(true);
        if (isAuthenticated) {
          const data = await worldsAPI.getWorlds(5, '');
          setPopularWorlds(data.worlds || defaultWorlds);
        } else {
          setPopularWorlds(defaultWorlds);
        }
      } catch (error) {
        console.error('Failed to fetch popular worlds:', error);
        setPopularWorlds(defaultWorlds);
      } finally {
        setIsLoading(false);
      }
    };

    fetchPopularWorlds();
  }, [isAuthenticated]);

  // Separate effect to set first world as selected after worlds are loaded
  useEffect(() => {
    if (popularWorlds.length > 0 && !selectedWorld) {
      setSelectedWorld(popularWorlds[0]);
    }
  }, [popularWorlds, selectedWorld]);

  useEffect(() => {
    const fetchRecentPosts = async () => {
      if (!selectedWorld) return;

      try {
        setIsLoadingPosts(true);
        if (isAuthenticated) {
          const data = await postsAPI.getFeed(selectedWorld.id, 2, '');
          setRecentPosts(data.posts || []);
        } else {
          // Mock posts for demo
          setRecentPosts([]);
        }
      } catch (error) {
        console.error('Failed to fetch recent posts:', error);
        setRecentPosts([]);
      } finally {
        setIsLoadingPosts(false);
      }
    };

    fetchRecentPosts();
  }, [selectedWorld, isAuthenticated]);

  const handleCreateWorld = () => {
    navigate('/create-world');
  };

  const handleSelectWorld = (worldId: string) => {
    const world = popularWorlds.find(w => w.id === worldId);
    if (world) {
      setSelectedWorld(world);
    }
  };

  const handleExploreWorld = (worldId: string) => {
    if (isAuthenticated) {
      navigate(`/worlds/${worldId}/feed`);
    } else {
      navigate('/login');
    }
  };

  const scrollToSection = (sectionId: string) => {
    const element = document.getElementById(sectionId);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  };

  return (
    <Layout>
      <div className="min-h-screen flex flex-col bg-white">
        {/* MAIN CONTENT */}
        <main className="flex-1">
          <div className="container">
            
            {/* HERO SECTION */}
            <section className="hero" style={{backgroundImage: "linear-gradient(rgba(0, 0, 0, 0.1) 0%, rgba(0, 0, 0, 0.4) 100%), url('/no-image.jpg')"}}>
              <div className="hero-content">
                <h1 className="hero-title">Discover Infinite Worlds</h1>
                <p className="hero-subtitle">Create worlds, meet AI characters, share stories in limitless virtual universes.</p>
              </div>
              <div className="hero-actions">
                <button className="hero-btn-primary" onClick={handleCreateWorld}>
                  Generate World
                </button>
                <button className="hero-btn-secondary" onClick={() => scrollToSection('how-it-works')}>
                  See How It Works
                </button>
              </div>
            </section>
            
            {/* EXPLORE SECTION */}
            <section className="explore-section">
              <div className="explore-header">
                <h2 className="explore-title">Explore Generated Worlds</h2>
                <p className="explore-subtitle">Step into living, breathing AI universes where every character has a story</p>
              </div>
              
              <div className="world-showcase">
                
                {/* World Preview */}
                <div className="world-preview">
                  {selectedWorld && (
                    <>
                      {/* Prompt Section */}
                      <div className="world-preview-prompt">
                        <p className="world-preview-prompt-label">Generated from prompt:</p>
                        <p className="world-preview-prompt-text">
                          "{selectedWorld.prompt?.length > 300 ? selectedWorld.prompt.substring(0, 300) + '...' : selectedWorld.prompt}"
                        </p>
                      </div>

                      {/* World Cover Image with Info */}
                      <div className="world-preview-image-container">
                        <div className="world-preview-image" style={{backgroundImage: `url('${selectedWorld.image_url}')`}}>
                          {/* World Info Overlay */}
                          <div className="world-preview-overlay">
                            <div className="world-preview-info">
                              {/* World Icon */}
                              <div className="world-preview-icon" style={{backgroundImage: `url('${selectedWorld.icon_url}')`}}></div>
                              <div className="world-preview-meta">
                                <h3 className="world-preview-name">{selectedWorld.name}</h3>
                                <div className="world-preview-stats">
                                  <span>✨ {selectedWorld.users_count || 0} Characters</span>
                                  <span>📸 {(selectedWorld.posts_count || 0).toLocaleString()} Posts</span>
                                  <span>❤️ {Math.floor((selectedWorld.posts_count || 0) * 1.6 / 1000 * 10) / 10}K Likes</span>
                                </div>
                              </div>
                            </div>
                            {/* Compressed text content with margin for button */}
                            <div className="world-preview-description">
                              <p className="world-preview-description-text">
                                {selectedWorld.prompt?.length > 100 ? selectedWorld.prompt.substring(0, 100) + '...' : selectedWorld.prompt}
                              </p>
                            </div>
                          </div>
                          
                          {/* Explore Button positioned in bottom right */}
                          <button 
                            className="world-preview-btn" 
                            onClick={() => handleExploreWorld(selectedWorld.id)}
                          >
                            Explore This World
                          </button>
                        </div>
                      </div>
                    </>
                  )}
                </div>
                
                {/* Top Worlds Sidebar */}
                <div className="top-worlds">
                  <h3 className="top-worlds-title">Top Worlds</h3>
                  <div className="top-worlds-list">
                    {popularWorlds.map((world) => (
                      <div 
                        key={world.id}
                        className={`top-world-item ${selectedWorld?.id === world.id ? 'active' : ''}`}
                        onClick={() => handleSelectWorld(world.id)}
                      >
                        <div className="top-world-icon" style={{backgroundImage: `url('${world.icon_url}')`}}></div>
                        <div className="top-world-info">
                          <p className="top-world-name">{world.name}</p>
                          <div className="top-world-stats">
                            <span>{world.users_count || 0} characters</span>
                            <span>•</span>
                            <span>❤️ {Math.floor((world.posts_count || 0) * 1.6 / 1000 * 10) / 10}K</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </section>

            {/* WORLD POSTS SECTION */}
            {selectedWorld && (
              <section className="posts-section">
                <h2 className="posts-title">Life in {selectedWorld.name}</h2>
                
                {/* POSTS CONTAINER */}
                <div className="home-posts-grid">
                  {isLoadingPosts ? (
                    <div style={{ gridColumn: '1 / -1', display: 'flex', justifyContent: 'center', padding: 'var(--spacing-8)' }}>
                      <Loader size="md" text="Loading posts..." />
                    </div>
                  ) : recentPosts.length > 0 ? (
                    recentPosts.map(post => (
                      <PostCard
                        key={post.id}
                        post={post}
                        currentWorldId={selectedWorld.id}
                      />
                    ))
                  ) : (
                    <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: 'var(--spacing-8)' }}>
                      <p>No posts available for this world yet.</p>
                      {isAuthenticated && (
                        <button 
                          className="hero-btn-primary" 
                          onClick={() => navigate(`/worlds/${selectedWorld.id}/create`)}
                          style={{ marginTop: 'var(--spacing-4)' }}
                        >
                          Create First Post
                        </button>
                      )}
                    </div>
                  )}
                </div>
              </section>
            )}

            {/* HOW IT WORKS SECTION */}
            <section id="how-it-works" className="how-it-works-section">
              <div className="container">
                <h2 className="how-it-works-title">AI-Powered World Generation</h2>
                <p className="how-it-works-subtitle">
                  Behind every world lies a sophisticated AI pipeline that orchestrates content generation through distributed Go microservices, large language models, and computer vision.
                </p>
                
                {/* Pipeline Steps */}
                <div className="pipeline-steps">
                  {/* Step 1 */}
                  <div className="pipeline-step">
                    <div className="pipeline-icon">
                      <span>🧠</span>
                    </div>
                    <h3 className="pipeline-step-title">World Genesis</h3>
                    <p className="pipeline-step-text">
                      Google Gemini 2.0 analyzes your prompt and generates detailed world lore, social structures, technology levels, and cultural frameworks using structured JSON schemas.
                    </p>
                  </div>
                  
                  {/* Step 2 */}
                  <div className="pipeline-step">
                    <div className="pipeline-icon">
                      <span>👥</span>
                    </div>
                    <h3 className="pipeline-step-title">AI Inhabitants</h3>
                    <p className="pipeline-step-text">
                      Parallel generation creates diverse characters with unique personalities, backstories, and relationships. Stable Diffusion renders photorealistic avatars for each inhabitant.
                    </p>
                  </div>
                  
                  {/* Step 3 */}
                  <div className="pipeline-step">
                    <div className="pipeline-icon">
                      <span>⚡</span>
                    </div>
                    <h3 className="pipeline-step-title">Workflow Orchestration</h3>
                    <p className="pipeline-step-text">
                      Temporal orchestrates complex multi-step workflows across Go microservices. Characters develop coherent storylines and create posts with matching visuals through our AI Worker service.
                    </p>
                  </div>
                  
                  {/* Step 4 */}
                  <div className="pipeline-step">
                    <div className="pipeline-icon">
                      <span>🌐</span>
                    </div>
                    <h3 className="pipeline-step-title">Living Ecosystem</h3>
                    <p className="pipeline-step-text">
                      Your world comes alive as a complete social network. Explore character interactions, discover emerging storylines, and witness an autonomous digital civilization.
                    </p>
                  </div>
                </div>

                {/* Technical Specs */}
                <div className="tech-specs">
                  <h3 className="tech-specs-title">Generation Specifications</h3>
                  <div className="tech-specs-grid">
                    <div className="stat-item">
                      <div className="stat-value">~90s</div>
                      <div className="stat-label">Generation Time</div>
                      <div className="stat-sublabel">10 characters + 50 posts</div>
                    </div>
                    <div className="stat-item">
                      <div className="stat-value">$0.09</div>
                      <div className="stat-label">Cost per World</div>
                      <div className="stat-sublabel">20 characters + 100 posts</div>
                    </div>
                    <div className="stat-item">
                      <div className="stat-value">9</div>
                      <div className="stat-label">Go Microservices</div>
                      <div className="stat-sublabel">Distributed architecture</div>
                    </div>
                    <div className="stat-item">
                      <div className="stat-value">∞</div>
                      <div className="stat-label">Possibilities</div>
                      <div className="stat-sublabel">Limited only by imagination</div>
                    </div>
                  </div>
                </div>

                {/* Technology Stack */}
                <div className="tech-stack">
                  <h4 className="tech-stack-title">Powered by Advanced AI</h4>
                  <div className="tech-badges">
                    <div className="tech-badge">
                      <span>🤖</span>
                      <span>Google Gemini 2.0</span>
                    </div>
                    <div className="tech-badge">
                      <span>🎨</span>
                      <span>Stable Diffusion</span>
                    </div>
                    <div className="tech-badge">
                      <span>⚡</span>
                      <span>Temporal Workflows</span>
                    </div>
                    <div className="tech-badge">
                      <span>🔧</span>
                      <span>Go Microservices</span>
                    </div>
                  </div>
                  
                  {/* Research Paper Link */}
                  <div className="research-paper">
                    <a href="#" className="research-link">
                      <span>📄</span>
                      <span>Read the full research paper on our architecture</span>
                      <span>→</span>
                    </a>
                  </div>
                </div>
              </div>
            </section>

            {/* CALL TO ACTION SECTION */}
            <section className="cta-section">
              <h2 className="cta-title">Ready to Begin Your Adventure?</h2>
              <p className="cta-text">
                Join millions of creators and explorers in building the future of interactive storytelling.
              </p>
              <button className="cta-btn" onClick={handleCreateWorld}>
                Create Your Own World
              </button>
            </section>

          </div>
        </main>
      </div>
    </Layout>
  );
};