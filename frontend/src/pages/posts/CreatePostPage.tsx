import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { Layout } from '../../components/layout/Layout';
import { Card } from '../../components/ui/Card';
import { TextArea } from '../../components/ui/TextArea';
import { Button } from '../../components/ui/Button';
import { Loader } from '../../components/ui/Loader';
import { ImageUpload } from '../../components/common/ImageUpload';
import { useWorld } from '../../hooks/useWorld';
import { postsAPI } from '../../api/services';

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

const PreviewContainer = styled.div`
  border-radius: var(--radius-md);
  overflow: hidden;
  margin-top: var(--space-4);
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

export const CreatePostPage: React.FC = () => {
  const { worldId } = useParams<{ worldId: string }>();
  const { currentWorld, loadCurrentWorld } = useWorld();
  const [caption, setCaption] = useState('');
  const [mediaId, setMediaId] = useState<string | null>(null);
  const [mediaUrl, setMediaUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();
  
  const worldIdRef = useRef<string | null>(null);
  
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
    
    if (!mediaId) {
      setError('Please upload an image');
      return;
    }
    
    if (!worldId) {
      setError('World ID is missing');
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      const post = await postsAPI.createPost(worldId, caption, mediaId);
      
      setIsSuccess(true);
      
      // Clear form after success
      setCaption('');
      setMediaId(null);
      setMediaUrl(null);
      
      // Redirect to the post after a short delay
      setTimeout(() => {
        navigate(`/worlds/${worldId}/posts/${post.id}`);
      }, 2000);
    } catch (err: any) {
      setError(err.message || 'Failed to create post');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  const handleCancel = () => {
    navigate(`/worlds/${worldId}/feed`);
  };
  
  const handleUploadComplete = (id: string, url: string) => {
    setMediaId(id);
    setMediaUrl(url);
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
          <Title>Create Post</Title>
          <Subtitle>in {currentWorld.name}</Subtitle>
        </PageHeader>
        
        <FormContainer padding="var(--space-6)" variant="elevated">
          {error && <ErrorMessage>{error}</ErrorMessage>}
          
          {isSuccess && (
            <SuccessMessage
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
            >
              Post created successfully! Redirecting to your post...
            </SuccessMessage>
          )}
          
          <Form onSubmit={handleSubmit}>
            <TextArea
              label="Caption"
              value={caption}
              onChange={(e) => setCaption(e.target.value)}
              placeholder="What's on your mind?"
              rows={3}
              maxRows={8}
            />
            
            <div>
              <label style={{ 
                fontSize: 'var(--font-sm)', 
                fontWeight: 500, 
                color: 'var(--color-text)',
                marginBottom: 'var(--space-2)',
                display: 'block'
              }}>
                Image *
              </label>
              <ImageUpload 
                worldId={worldId || ''}
                onUploadComplete={handleUploadComplete}
              />
            </div>
            
            {/* Preview is now handled inside the ImageUpload component */}
            
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
                disabled={isSubmitting || !mediaId}
                fullWidth
              >
                Post
              </Button>
            </ButtonsContainer>
          </Form>
        </FormContainer>
      </PageContainer>
    </Layout>
  );
};