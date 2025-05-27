import React, { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import { Card } from '../ui/Card';
import { WorldGenerationStatus, StageInfo } from '../../types';
import { worldsAPI } from '../../api/services';
import { useWorld } from '../../hooks/useWorld';

const ProgressContainer = styled(Card)`
  margin-bottom: var(--space-6);
  background: linear-gradient(135deg, var(--color-primary), #FF9900);
  color: white;
  position: relative;
  overflow: hidden;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.1);
    z-index: 1;
  }
`;

const ProgressContent = styled.div`
  position: relative;
  z-index: 2;
`;

const ProgressHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
`;

const ProgressTitle = styled.h3`
  font-size: var(--font-lg);
  margin: 0;
`;

const StatusBadge = styled.span<{ $status: string }>`
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  font-size: var(--font-sm);
  font-weight: 500;
  background: ${props => {
    switch (props.$status) {
      case 'completed': return 'rgba(34, 197, 94, 0.8)';
      case 'in_progress': return 'rgba(59, 130, 246, 0.8)';
      case 'failed': return 'rgba(239, 68, 68, 0.8)';
      default: return 'rgba(107, 114, 128, 0.8)';
    }
  }};
`;

const StagesContainer = styled.div`
  margin-bottom: var(--space-4);
`;

const StageItem = styled.div<{ $status: string; $isCurrent: boolean }>`
  display: flex;
  align-items: center;
  padding: var(--space-2) 0;
  position: relative;
  
  &:not(:last-child)::after {
    content: '';
    position: absolute;
    left: 10px;
    top: 100%;
    width: 2px;
    height: var(--space-2);
    background: ${props => 
      props.$status === 'completed' ? 'rgba(34, 197, 94, 0.5)' : 'rgba(255, 255, 255, 0.3)'
    };
  }
`;

const StageIcon = styled.div<{ $status: string; $isCurrent: boolean }>`
  width: 20px;
  height: 20px;
  border-radius: 50%;
  margin-right: var(--space-3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  background: ${props => {
    if (props.$status === 'completed') return 'rgba(34, 197, 94, 0.8)';
    if (props.$isCurrent) return 'rgba(59, 130, 246, 0.8)';
    return 'rgba(255, 255, 255, 0.3)';
  }};
  
  ${props => props.$isCurrent && `
    animation: pulse 2s infinite;
  `}
  
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
`;

const StageName = styled.span<{ $isCurrent: boolean }>`
  font-weight: ${props => props.$isCurrent ? '600' : '400'};
  text-transform: capitalize;
`;

const ProgressBarsContainer = styled.div`
  margin-bottom: var(--space-4);
`;

const ProgressBarWrapper = styled.div`
  margin-bottom: var(--space-3);
  
  &:last-child {
    margin-bottom: 0;
  }
`;

const ProgressBarLabel = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-1);
  font-size: var(--font-sm);
`;

const ProgressBar = styled.div`
  height: 8px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: var(--radius-full);
  overflow: hidden;
`;

const ProgressBarFill = styled(motion.div)<{ $color: string }>`
  height: 100%;
  background: ${props => props.$color};
  border-radius: var(--radius-full);
`;

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: var(--space-3);
  margin-bottom: var(--space-4);
`;

const StatItem = styled.div`
  text-align: center;
`;

const StatValue = styled.div`
  font-size: var(--font-lg);
  font-weight: 600;
  margin-bottom: var(--space-1);
`;

const StatLabel = styled.div`
  font-size: var(--font-sm);
  opacity: 0.8;
`;

interface WorldGenerationProgressProps {
  worldId: string;
  onGenerationComplete?: (status: WorldGenerationStatus) => void;
  onPostsUpdated?: (postsCount: number) => void;
}

export const WorldGenerationProgress: React.FC<WorldGenerationProgressProps> = ({
  worldId,
  onGenerationComplete,
  onPostsUpdated
}) => {
  const [status, setStatus] = useState<WorldGenerationStatus | null>(null);
  const [isVisible, setIsVisible] = useState(true);
  const eventSourceRef = useRef<EventSource | null>(null);
  const { loadCurrentWorld } = useWorld();
  const previousPostsCountRef = useRef<number>(0);

  useEffect(() => {
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
            
            // Check if posts count has increased
            if (data.posts_created > previousPostsCountRef.current) {
              previousPostsCountRef.current = data.posts_created;
              onPostsUpdated?.(data.posts_created);
            }
            
            setStatus(data);
            
            // Check if world_image stage completed to refresh world data
            if (data.current_stage === 'characters' && data.stages) {
              const worldImageStage = data.stages.find((stage: StageInfo) => stage.name === 'world_image');
              if (worldImageStage && worldImageStage.status === 'completed') {
                loadCurrentWorld(worldId);
              }
            }
            
            // Hide progress when generation is completed
            if (data.status === 'completed' || data.status === 'failed') {
              setTimeout(() => {
                setIsVisible(false);
                onGenerationComplete?.(data);
              }, 2000);
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
            
            // Check if posts count has increased
            if (statusData.posts_created > previousPostsCountRef.current) {
              previousPostsCountRef.current = statusData.posts_created;
              onPostsUpdated?.(statusData.posts_created);
            }
            
            setStatus(statusData);
            
            if (statusData.status === 'completed' || statusData.status === 'failed') {
              setTimeout(() => {
                setIsVisible(false);
                onGenerationComplete?.(statusData);
              }, 2000);
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

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, [worldId, onGenerationComplete, loadCurrentWorld]);

  if (!status || !isVisible) return null;

  const getStageDisplayName = (stageName: string) => {
    const names: Record<string, string> = {
      'initializing': 'Initializing',
      'world_description': 'Creating Description',
      'world_image': 'Generating Image',
      'characters': 'Creating Characters',
      'posts': 'Generating Posts',
      'finishing': 'Finishing Up'
    };
    return names[stageName] || stageName;
  };

  const usersProgress = status.users_predicted > 0 ? (status.users_created / status.users_predicted) * 100 : 0;
  const postsProgress = status.posts_predicted > 0 ? (status.posts_created / status.posts_predicted) * 100 : 0;
  const tasksProgress = status.tasks_total > 0 ? (status.tasks_completed / status.tasks_total) * 100 : 0;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -20 }}
        transition={{ duration: 0.3 }}
      >
        <ProgressContainer>
          <ProgressContent>
            <ProgressHeader>
              <ProgressTitle>Generating World</ProgressTitle>
              <StatusBadge $status={status.status}>
                {status.status === 'in_progress' ? 'In Progress' : status.status}
              </StatusBadge>
            </ProgressHeader>

            <StagesContainer>
              {status.stages.map((stage, index) => (
                <StageItem
                  key={stage.name}
                  $status={stage.status}
                  $isCurrent={stage.name === status.current_stage}
                >
                  <StageIcon
                    $status={stage.status}
                    $isCurrent={stage.name === status.current_stage}
                  >
                    {stage.status === 'completed' ? '✓' : index + 1}
                  </StageIcon>
                  <StageName $isCurrent={stage.name === status.current_stage}>
                    {getStageDisplayName(stage.name)}
                  </StageName>
                </StageItem>
              ))}
            </StagesContainer>

            {(status.users_predicted > 0 || status.posts_predicted > 0) && (
              <ProgressBarsContainer>
                {status.users_predicted > 0 && (
                  <ProgressBarWrapper>
                    <ProgressBarLabel>
                      <span>Characters</span>
                      <span>{status.users_created} / {status.users_predicted}</span>
                    </ProgressBarLabel>
                    <ProgressBar>
                      <ProgressBarFill
                        $color="rgba(34, 197, 94, 0.8)"
                        initial={{ width: 0 }}
                        animate={{ width: `${usersProgress}%` }}
                        transition={{ duration: 0.5 }}
                      />
                    </ProgressBar>
                  </ProgressBarWrapper>
                )}

                {status.posts_predicted > 0 && (
                  <ProgressBarWrapper>
                    <ProgressBarLabel>
                      <span>Posts</span>
                      <span>{status.posts_created} / {status.posts_predicted}</span>
                    </ProgressBarLabel>
                    <ProgressBar>
                      <ProgressBarFill
                        $color="rgba(59, 130, 246, 0.8)"
                        initial={{ width: 0 }}
                        animate={{ width: `${postsProgress}%` }}
                        transition={{ duration: 0.5 }}
                      />
                    </ProgressBar>
                  </ProgressBarWrapper>
                )}

                <ProgressBarWrapper>
                  <ProgressBarLabel>
                    <span>Overall Progress</span>
                    <span>{status.tasks_completed} / {status.tasks_total}</span>
                  </ProgressBarLabel>
                  <ProgressBar>
                    <ProgressBarFill
                      $color="rgba(168, 85, 247, 0.8)"
                      initial={{ width: 0 }}
                      animate={{ width: `${tasksProgress}%` }}
                      transition={{ duration: 0.5 }}
                    />
                  </ProgressBar>
                </ProgressBarWrapper>
              </ProgressBarsContainer>
            )}

            <StatsGrid>
              <StatItem>
                <StatValue>{status.api_calls_made_llm}</StatValue>
                <StatLabel>AI Calls</StatLabel>
              </StatItem>
              <StatItem>
                <StatValue>{status.api_calls_made_images}</StatValue>
                <StatLabel>Images</StatLabel>
              </StatItem>
              <StatItem>
                <StatValue>${(status.llm_cost_total + status.image_cost_total).toFixed(3)}</StatValue>
                <StatLabel>Cost</StatLabel>
              </StatItem>
              <StatItem>
                <StatValue>{status.tasks_failed}</StatValue>
                <StatLabel>Failed</StatLabel>
              </StatItem>
            </StatsGrid>
          </ProgressContent>
        </ProgressContainer>
      </motion.div>
    </AnimatePresence>
  );
};