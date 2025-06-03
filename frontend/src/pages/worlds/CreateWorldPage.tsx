import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Layout } from '../../components/layout/Layout';
import { useWorld } from '../../hooks/useWorld';
import { WorldGenerationStatus, StageInfo } from '../../types';
import { worldsAPI } from '../../api/services';
import '../../styles/pages/create-world.css';

export const CreateWorldPage: React.FC = () => {
  const [currentState, setCurrentState] = useState<'form' | 'progress'>('form');
  const [prompt, setPrompt] = useState('');
  const [charactersCount, setCharactersCount] = useState(15);
  const [postsCount, setPostsCount] = useState(30);
  const [characterCost, setCharacterCost] = useState(45);
  const [postsCost, setPostsCost] = useState(30);
  const [totalCost, setTotalCost] = useState(75);
  const [createdWorldId, setCreatedWorldId] = useState<string | null>(null);
  
  // Progress state from real API
  const [status, setStatus] = useState<WorldGenerationStatus | null>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  
  const { createWorld, error, isLoading } = useWorld();
  const navigate = useNavigate();
  
  // Example prompts
  const prompts = {
    underwater: "A vast underwater civilization where merfolk, sea creatures, and coral cities thrive in the deep ocean. Ancient magic flows through the currents, and bioluminescent technology lights the abyssal depths.",
    space: "A massive space station at the edge of known space, where diverse alien species trade, explore, and live together. Advanced technology meets ancient wisdom as cultures blend in this cosmic melting pot.",
    medieval: "A sprawling medieval kingdom where knights, wizards, and mythical creatures coexist. Ancient castles dot the landscape, magic flows through enchanted forests, and political intrigue shapes the realm."
  };
  
  // Map real API stages to template stages
  const getStageNumber = (stageName: string): number => {
    const stageMap: Record<string, number> = {
      'initializing': 1,
      'world_description': 1,
      'world_image': 2,
      'characters': 3,
      'posts': 4,
      'finishing': 5
    };
    return stageMap[stageName] || 1;
  };

  // Check if generation is complete based on actual counts
  const isGenerationComplete = (statusData?: any): boolean => {
    const currentStatus = statusData || status;
    if (!currentStatus) return false;
    
    const usersComplete = currentStatus.users_predicted > 0 && currentStatus.users_created >= currentStatus.users_predicted;
    const postsComplete = currentStatus.posts_predicted > 0 && currentStatus.posts_created >= currentStatus.posts_predicted;
    
    return usersComplete && postsComplete;
  };
  
  // Update costs when character count changes
  const handleCharacterSliderChange = useCallback((value: number) => {
    setCharactersCount(value);
    const cost = value * 3; // 3 credits per character
    setCharacterCost(cost);
  }, []);
  
  // Update costs when posts count changes
  const handlePostsSliderChange = useCallback((value: number) => {
    setPostsCount(value);
    const cost = value * 1; // 1 credit per post
    setPostsCost(cost);
  }, []);
  
  // Update total cost
  useEffect(() => {
    setTotalCost(characterCost + postsCost);
  }, [characterCost, postsCost]);
  
  // Fill example prompt
  const fillPrompt = useCallback((type: keyof typeof prompts) => {
    setPrompt(prompts[type]);
  }, [prompts]);
  
  // Start generation process
  const startGeneration = useCallback(async () => {
    const trimmedPrompt = prompt.trim();
    if (!trimmedPrompt) {
      alert('Please describe your world first!');
      return;
    }
    
    try {
      // Use real API to create world
      const world = await createWorld(
        'Generated World', // You can add a name field later if needed
        '', // Description
        trimmedPrompt,
        charactersCount,
        postsCount
      );
      
      // Switch to progress view and set the created world ID
      setCreatedWorldId(world.id);
      setCurrentState('progress');
      
      // Start listening to progress updates
      startProgressTracking(world.id);
      
    } catch (err) {
      console.error('Failed to create world:', err);
    }
  }, [prompt, charactersCount, postsCount, createWorld]);
  
  // Start tracking progress with real API
  const startProgressTracking = useCallback((worldId: string) => {
    let eventSource: EventSource | null = null;

    const connectSSE = () => {
      try {
        eventSource = worldsAPI.createWorldStatusEventSource(worldId);
        eventSourceRef.current = eventSource;

        eventSource.onopen = () => {
          console.log('SSE connection opened');
        };

        eventSource.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'ping') return;
            
            setStatus(data);
            
            // Check if generation is actually complete based on counts
            if (data.status === 'completed' || isGenerationComplete(data)) {
              setTimeout(() => {
                navigate(`/worlds/${worldId}/feed`);
              }, 2000);
            } else if (data.status === 'failed') {
              // Go back to form on failure
              setCurrentState('form');
              setCreatedWorldId(null);
            }
          } catch (error) {
            console.error('Failed to parse SSE data:', error);
          }
        };

        eventSource.onerror = (error) => {
          console.error('SSE error:', error);
          eventSource?.close();
          
          // Retry connection after 5 seconds
          setTimeout(connectSSE, 5000);
        };
      } catch (error) {
        console.error('Failed to create EventSource:', error);
        
        // Fallback to polling
        const pollStatus = async () => {
          try {
            const statusData = await worldsAPI.getWorldStatus(worldId);
            setStatus(statusData);
            
            if (statusData.status === 'completed' || isGenerationComplete(statusData)) {
              setTimeout(() => {
                navigate(`/worlds/${worldId}/feed`);
              }, 2000);
            } else if (statusData.status === 'failed') {
              setCurrentState('form');
              setCreatedWorldId(null);
            } else {
              setTimeout(pollStatus, 1000);
            }
          } catch (error) {
            console.error('Failed to poll status:', error);
            setTimeout(pollStatus, 5000);
          }
        };
        
        pollStatus();
      }
    };

    connectSSE();
  }, [navigate]);
  
  // Cancel generation
  const cancelGeneration = useCallback(() => {
    if (window.confirm('Are you sure you want to cancel world generation?')) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      setCurrentState('form');
      setCreatedWorldId(null);
      setStatus(null);
      // Note: The actual API call to cancel generation would go here
    }
  }, []);
  
  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, []);
  
  // Update slider visual progress
  const updateSliderProgress = useCallback((slider: HTMLInputElement) => {
    const value = ((Number(slider.value) - Number(slider.min)) / (Number(slider.max) - Number(slider.min))) * 100;
    slider.style.background = `linear-gradient(to right, var(--color-primary) 0%, var(--color-primary) ${value}%, var(--color-border) ${value}%, var(--color-border) 100%)`;
  }, []);
  
  // Initialize sliders on mount
  useEffect(() => {
    const characterSlider = document.getElementById('character-slider') as HTMLInputElement;
    const postsSlider = document.getElementById('posts-slider') as HTMLInputElement;
    
    if (characterSlider) updateSliderProgress(characterSlider);
    if (postsSlider) updateSliderProgress(postsSlider);
  }, [updateSliderProgress]);
  
  // Get current step based on real API status
  const getCurrentStep = (): number => {
    if (!status) return 0;
    return getStageNumber(status.current_stage);
  };
  
  return (
    <Layout>
      <div className="min-h-screen flex flex-col bg-white">
        {/* MAIN CONTENT */}
        <main className="flex-1">
          <div className="container">
          
          {/* CREATE FORM (Initial State) */}
          {currentState === 'form' && (
            <div id="create-form" className="create-form-state">
              {/* Page Header */}
              <div className="page-header">
                <h1 className="page-title">Create Your World</h1>
                <p className="page-subtitle">Describe your vision and watch our AI bring it to life with characters, stories, and endless possibilities.</p>
              </div>

              {/* Error Display */}
              {error && (
                <div style={{ 
                  color: 'var(--color-accent)', 
                  background: 'rgba(239, 118, 122, 0.1)', 
                  padding: 'var(--spacing-4)', 
                  borderRadius: 'var(--radius-lg)', 
                  marginBottom: 'var(--spacing-6)' 
                }}>
                  {error}
                </div>
              )}

              {/* Main Form Card */}
              <div className="form-card">
                
                {/* Prompt Input */}
                <div className="form-group">
                  <label className="form-label">Describe Your World</label>
                  <textarea
                    id="world-prompt"
                    placeholder="A mystical ancient city floating among the clouds, where magic and technology intertwine..."
                    rows={6}
                    className="form-textarea"
                    value={prompt}
                    onChange={(e) => setPrompt(e.target.value)}
                  />
                  <p className="form-help-text">Be specific about the setting, culture, technology level, and atmosphere you envision.</p>
                </div>

                {/* Character Count Slider */}
                <div className="form-group">
                  <div className="slider-header">
                    <label className="form-label">Number of Characters</label>
                    <span id="character-count" className="slider-value">{charactersCount}</span>
                  </div>
                  <div className="slider-container">
                    <input
                      type="range"
                      id="character-slider"
                      min="5"
                      max="50"
                      value={charactersCount}
                      className="slider"
                      onChange={(e) => {
                        const value = parseInt(e.target.value);
                        handleCharacterSliderChange(value);
                        updateSliderProgress(e.target);
                      }}
                    />
                  </div>
                  <div className="slider-labels">
                    <span>5 characters</span>
                    <span>50 characters</span>
                  </div>
                </div>

                {/* Posts Count Slider */}
                <div className="form-group">
                  <div className="slider-header">
                    <label className="form-label">Number of Posts</label>
                    <span id="posts-count" className="slider-value">{postsCount}</span>
                  </div>
                  <div className="slider-container">
                    <input
                      type="range"
                      id="posts-slider"
                      min="10"
                      max="100"
                      value={postsCount}
                      className="slider"
                      onChange={(e) => {
                        const value = parseInt(e.target.value);
                        handlePostsSliderChange(value);
                        updateSliderProgress(e.target);
                      }}
                    />
                  </div>
                  <div className="slider-labels">
                    <span>10 posts</span>
                    <span>100 posts</span>
                  </div>
                </div>

                {/* Cost Display */}
                <div className="cost-display">
                  <div className="cost-header">
                    <span className="cost-label">Generation Cost</span>
                    <span id="total-cost" className="cost-value">{totalCost} 💎</span>
                  </div>
                  <div className="cost-breakdown">
                    <span>Characters: <span id="character-cost">{characterCost}</span> • Posts: <span id="posts-cost">{postsCost}</span> • ~90 seconds</span>
                  </div>
                </div>

                {/* Generate Button */}
                <button 
                  id="generate-btn" 
                  className="btn btn-primary" 
                  onClick={startGeneration}
                  disabled={isLoading}
                  style={{ 
                    width: '100%', 
                    height: '4rem', 
                    fontSize: 'var(--text-lg)', 
                    fontWeight: 'var(--font-bold)',
                    opacity: isLoading ? 0.6 : 1,
                    cursor: isLoading ? 'not-allowed' : 'pointer'
                  }}
                >
                  {isLoading ? 'Creating World...' : 'Generate World ✨'}
                </button>
              </div>

              {/* Examples Section */}
              <div className="example-prompts">
                <p className="example-prompts-text">Need inspiration? Try these prompts:</p>
                <div className="example-prompts-container">
                  <button onClick={() => fillPrompt('underwater')} className="example-prompt-btn">🌊 Underwater civilization</button>
                  <button onClick={() => fillPrompt('space')} className="example-prompt-btn">🚀 Space station colony</button>
                  <button onClick={() => fillPrompt('medieval')} className="example-prompt-btn">🏰 Medieval fantasy kingdom</button>
                </div>
              </div>
            </div>
          )}

          {/* GENERATION PROGRESS */}
          {currentState === 'progress' && createdWorldId && (
            <div id="generation-progress" className="generation-progress-state">
              {/* Progress Header */}
              <div className="progress-header">
                <h1 className="progress-title">Creating Your World</h1>
                <p className="progress-subtitle">AI is generating your unique world. This may take a few minutes...</p>
              </div>

              {/* Progress Steps - Using template design with real data */}
              <div className="progress-steps">
                <div className={`progress-step ${getCurrentStep() >= 1 ? 'active' : ''} ${getCurrentStep() > 1 ? 'completed' : ''}`} id="step-1">
                  <div className="progress-step-icon">1</div>
                  <div className="progress-step-content">
                    <h3>Generating World</h3>
                    <p>Creating world description and cover image</p>
                  </div>
                </div>
                
                <div className={`progress-step ${getCurrentStep() >= 2 ? 'active' : ''} ${getCurrentStep() > 2 ? 'completed' : ''}`} id="step-2">
                  <div className="progress-step-icon">2</div>
                  <div className="progress-step-content">
                    <h3>Creating World Image</h3>
                    <p>Generating beautiful world cover</p>
                  </div>
                </div>
                
                <div className={`progress-step ${getCurrentStep() >= 3 ? 'active' : ''} ${getCurrentStep() > 3 ? 'completed' : ''}`} id="step-3">
                  <div className="progress-step-icon">3</div>
                  <div className="progress-step-content">
                    <h3>Creating Characters</h3>
                    <p>Designing unique AI personalities</p>
                  </div>
                </div>
                
                <div className={`progress-step ${getCurrentStep() >= 4 ? 'active' : ''} ${getCurrentStep() > 4 ? 'completed' : ''}`} id="step-4">
                  <div className="progress-step-icon">4</div>
                  <div className="progress-step-content">
                    <h3>Generating Posts</h3>
                    <p>Creating stories and interactions</p>
                  </div>
                </div>
                
                <div className={`progress-step ${getCurrentStep() >= 5 ? 'active' : ''} ${isGenerationComplete() ? 'completed' : ''}`} id="step-5">
                  <div className="progress-step-icon">5</div>
                  <div className="progress-step-content">
                    <h3>Finalizing World</h3>
                    <p>Preparing your world for exploration</p>
                  </div>
                </div>
              </div>

              {/* Progress Bars - Using real data from API */}
              {status && (
                <div className="progress-bars">
                  <div className="progress-bar-item">
                    <div className="progress-bar-header">
                      <span>Characters Created</span>
                      <span id="characters-progress-text">{status.users_created} / {status.users_predicted || charactersCount}</span>
                    </div>
                    <div className="progress-bar">
                      <div 
                        id="characters-progress-fill" 
                        className="progress-bar-fill"
                        style={{ 
                          width: status.users_predicted > 0 
                            ? `${(status.users_created / status.users_predicted) * 100}%` 
                            : '0%'
                        }}
                      />
                    </div>
                  </div>
                  
                  <div className="progress-bar-item">
                    <div className="progress-bar-header">
                      <span>Posts Generated</span>
                      <span id="posts-progress-text">{status.posts_created} / {status.posts_predicted || postsCount}</span>
                    </div>
                    <div className="progress-bar">
                      <div 
                        id="posts-progress-fill" 
                        className="progress-bar-fill"
                        style={{ 
                          width: status.posts_predicted > 0 
                            ? `${(status.posts_created / status.posts_predicted) * 100}%` 
                            : '0%'
                        }}
                      />
                    </div>
                  </div>
                </div>
              )}

            </div>
          )}

          </div>
        </main>
      </div>
    </Layout>
  );
};