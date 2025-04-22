import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { Layout } from '../../components/layout/Layout';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { Button } from '../../components/ui/Button';
import { Loader } from '../../components/ui/Loader';
import { useWorld } from '../../hooks/useWorld';
import { useAuth } from '../../hooks/useAuth';
import { characterAPI } from '../../api/services';

const PageContainer = styled.div`
  max-width: 640px;
  margin: 0 auto;
`;

const PageHeader = styled.div`
  text-align: center;
  margin-bottom: var(--space-6);
`;

const Title = styled.h1`
  font-size: var(--font-3xl);
  margin-bottom: var(--space-2);
`;

const Subtitle = styled.p`
  color: var(--color-text-light);
`;

const FormContainer = styled(Card)`
  margin-bottom: var(--space-8);
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
`;

const ButtonsContainer = styled.div`
  display: flex;
  gap: var(--space-4);
  
  @media (max-width: 640px) {
    flex-direction: column;
  }
`;

const ErrorMessage = styled.div`
  color: var(--color-accent);
  background-color: rgba(239, 118, 122, 0.1);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
`;

const SuccessMessage = styled(motion.div)<HTMLMotionProps<'div'>>`
  background-color: rgba(110, 231, 183, 0.1);
  color: var(--color-success);
  padding: var(--space-4);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
  text-align: center;
`;

const InfoMessage = styled.div`
  background-color: rgba(165, 180, 252, 0.1);
  color: var(--color-info);
  padding: var(--space-4);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
`;

export const CreateCharacterPage: React.FC = () => {
  const { worldId } = useParams<{ worldId: string }>();
  const { currentWorld, loadCurrentWorld } = useWorld();
  const { user } = useAuth();
  const [displayName, setDisplayName] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  
  const worldIdRef = useRef<string | null>(null);
  const returnToRef = useRef<string | null>(
    new URLSearchParams(window.location.search).get('returnTo')
  );
  
  useEffect(() => {
    if (worldId && worldIdRef.current !== worldId) {
      worldIdRef.current = worldId;
      loadCurrentWorld(worldId)
        .then(() => setIsLoading(false))
        .catch(() => {
          setError('Failed to load world');
          setIsLoading(false);
        });
    }
  }, [worldId, loadCurrentWorld]);
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    
    if (!displayName.trim()) {
      setError('Please enter a display name');
      return;
    }
    
    if (!worldId) {
      setError('World ID is missing');
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      const character = await characterAPI.createCharacter(worldId, displayName);
      
      setIsSuccess(true);
      
      // Redirect after success
      setTimeout(() => {
        if (returnToRef.current) {
          navigate(returnToRef.current);
        } else {
          navigate(`/worlds/${worldId}/feed`);
        }
      }, 2000);
    } catch (err: any) {
      setError(err.message || 'Failed to create character');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  const handleCancel = () => {
    navigate(`/worlds/${worldId}/feed`);
  };
  
  if (isLoading) {
    return (
      <Layout>
        <div style={{ display: 'flex', justifyContent: 'center', padding: 'var(--space-10)' }}>
          <Loader text="Loading..." />
        </div>
      </Layout>
    );
  }
  
  if (!currentWorld) {
    return (
      <Layout>
        <ErrorMessage>
          World not found or you don't have access.
        </ErrorMessage>
        <Button onClick={() => navigate('/worlds')}>
          View All Worlds
        </Button>
      </Layout>
    );
  }
  
  return (
    <Layout>
      <PageContainer>
        <PageHeader>
          <Title>Create Character</Title>
          <Subtitle>in {currentWorld.name}</Subtitle>
        </PageHeader>
        
        <FormContainer padding="var(--space-6)" variant="elevated">
          <InfoMessage>
            You need to create a character in this world before you can post content.
            This character will represent you within this virtual world.
          </InfoMessage>
          
          {error && <ErrorMessage>{error}</ErrorMessage>}
          
          {isSuccess && (
            <SuccessMessage
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
            >
              Character created successfully! Redirecting...
            </SuccessMessage>
          )}
          
          <Form onSubmit={handleSubmit}>
            <Input
              label="Display Name"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="How you want to be known in this world"
              required
            />
            
            <ButtonsContainer>
              <Button
                type="button"
                variant="ghost"
                onClick={handleCancel}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                isLoading={isSubmitting}
                disabled={isSubmitting || !displayName.trim()}
                fullWidth
              >
                Create Character
              </Button>
            </ButtonsContainer>
          </Form>
        </FormContainer>
      </PageContainer>
    </Layout>
  );
};