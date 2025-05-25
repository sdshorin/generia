import React, { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { Layout } from '../../components/layout/Layout';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { TextArea } from '../../components/ui/TextArea';
import { Button } from '../../components/ui/Button';
import { Loader } from '../../components/ui/Loader';
import { useWorld } from '../../hooks/useWorld';

const PageContainer = styled.div`
  max-width: 720px;
  margin: 0 auto;
`;

const PageHeader = styled.div`
  text-align: center;
  margin-bottom: var(--space-6);
`;

const Title = styled.h1`
  font-size: var(--font-3xl);
  margin-bottom: var(--space-2);
  background: linear-gradient(135deg, var(--color-primary), #FF9900);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
`;

const Subtitle = styled.p`
  color: var(--color-text-light);
  font-size: var(--font-md);
  max-width: 500px;
  margin: 0 auto;
`;

const FormContainer = styled(Card)`
  margin-bottom: var(--space-8);
`;

const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
`;

const FieldGroup = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
`;

const Label = styled.label`
  font-size: var(--font-sm);
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: var(--space-1);
`;

const HelperText = styled.p`
  font-size: var(--font-xs);
  color: var(--color-text-lighter);
  margin-top: var(--space-1);
`;

const ButtonsContainer = styled.div`
  display: flex;
  gap: var(--space-4);
  margin-top: var(--space-2);
  
  @media (max-width: 640px) {
    flex-direction: column;
  }
`;

const ExamplePrompts = styled.div`
  margin-top: var(--space-4);
`;

const ExamplePromptTitle = styled.h4`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  margin-bottom: var(--space-2);
`;

const ExamplePromptCard = styled(motion.div)<HTMLMotionProps<'div'>>`
  background-color: var(--color-input-bg);
  border-radius: var(--radius-md);
  padding: var(--space-3);
  font-size: var(--font-sm);
  color: var(--color-text);
  cursor: pointer;
  transition: background-color 0.2s;
  margin-bottom: var(--space-2);
  
  &:hover {
    background-color: var(--color-border);
  }
`;

const ErrorMessage = styled.div`
  color: var(--color-accent);
  font-size: var(--font-sm);
  padding: var(--space-3);
  background-color: rgba(239, 118, 122, 0.1);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
`;

const SliderContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
`;

const SliderWrapper = styled.div`
  position: relative;
  margin: var(--space-3) 0;
`;

const SliderInput = styled.input`
  width: 100%;
  height: 6px;
  background: var(--color-input-bg);
  border-radius: var(--radius-full);
  outline: none;
  -webkit-appearance: none;
  
  &::-webkit-slider-thumb {
    -webkit-appearance: none;
    width: 20px;
    height: 20px;
    background: var(--color-primary);
    border-radius: 50%;
    cursor: pointer;
    box-shadow: var(--shadow-sm);
    transition: all 0.2s;
    
    &:hover {
      transform: scale(1.1);
      background: var(--color-primary-hover);
    }
  }
  
  &::-moz-range-thumb {
    width: 20px;
    height: 20px;
    background: var(--color-primary);
    border-radius: 50%;
    cursor: pointer;
    border: none;
    box-shadow: var(--shadow-sm);
  }
`;

const SliderLabels = styled.div`
  display: flex;
  justify-content: space-between;
  font-size: var(--font-xs);
  color: var(--color-text-lighter);
  margin-top: var(--space-1);
`;

const SliderValue = styled.div`
  text-align: center;
  font-size: var(--font-sm);
  font-weight: 500;
  color: var(--color-primary);
  margin-bottom: var(--space-2);
`;

const SlidersRow = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-6);
  
  @media (max-width: 640px) {
    grid-template-columns: 1fr;
  }
`;

// Example prompts for world generation
const examplePrompts = [
  "A cyberpunk city where nature has reclaimed technology, with neon-lit trees and digital wildlife.",
  "A peaceful medieval village where everyone specializes in unique magical crafts.",
  "A retro 1980s mall filled with bizarre shops that sell impossible objects.",
  "A tropical island community where residents communicate through colorful sand art.",
  "A steampunk space colony orbiting Jupiter where Victorian fashion meets advanced astronomy."
];

export const CreateWorldPage: React.FC = () => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [prompt, setPrompt] = useState('');
  const [charactersCount, setCharactersCount] = useState(25);
  const [postsCount, setPostsCount] = useState(150);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { createWorld, error, isLoading, clearError } = useWorld();
  const navigate = useNavigate();
  
  const validateForm = (): boolean => {
    const formErrors: Record<string, string> = {};
    
    if (!name.trim()) {
      formErrors.name = 'World name is required';
    } else if (name.length < 3) {
      formErrors.name = 'World name must be at least 3 characters';
    }
    
    if (!prompt.trim()) {
      formErrors.prompt = 'Prompt is required to generate the world';
    } else if (prompt.length < 10) {
      formErrors.prompt = 'Please provide a more detailed prompt (at least 10 characters)';
    }
    
    setErrors(formErrors);
    return Object.keys(formErrors).length === 0;
  };
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();
    
    if (!validateForm()) {
      return;
    }
    
    try {
      const world = await createWorld(name, description, prompt, charactersCount, postsCount);
      navigate(`/worlds/${world.id}/feed`);
    } catch (err) {
      console.error('Failed to create world:', err);
    }
  };
  
  const handleUseExamplePrompt = useCallback((examplePrompt: string) => {
    setPrompt(examplePrompt);
  }, []);
  
  const handleRandomPrompt = useCallback(() => {
    const randomIndex = Math.floor(Math.random() * examplePrompts.length);
    setPrompt(examplePrompts[randomIndex]);
  }, []);
  
  return (
    <Layout>
      <PageContainer>
        <PageHeader>
          <Title>Generate a New World</Title>
          <Subtitle>
            Create a synthetic world with its own unique theme and AI-generated inhabitants.
          </Subtitle>
        </PageHeader>
        
        <FormContainer padding="var(--space-6)" variant="elevated">
          {error && <ErrorMessage>{error}</ErrorMessage>}
          
          <Form onSubmit={handleSubmit}>
            <FieldGroup>
              <Label htmlFor="name">World Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Enter a name for your world"
                error={errors.name}
              />
            </FieldGroup>
            
            <FieldGroup>
              <Label htmlFor="description">Description (Optional)</Label>
              <TextArea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Briefly describe what makes this world unique"
                rows={2}
              />
            </FieldGroup>
            
            <FieldGroup>
              <Label htmlFor="prompt">World Generation Prompt</Label>
              <TextArea
                id="prompt"
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder="Describe the theme, aesthetics, and characteristics of your world"
                rows={4}
                error={errors.prompt}
              />
              <HelperText>
                Be specific and detailed about the world you want to create. This will guide the AI in generating users and content.
              </HelperText>
            </FieldGroup>
            
            <FieldGroup>
              <Label>World Size Settings</Label>
              <SlidersRow>
                <SliderContainer>
                  <Label htmlFor="charactersCount">Number of Characters</Label>
                  <SliderValue>{charactersCount} characters</SliderValue>
                  <SliderWrapper>
                    <SliderInput
                      id="charactersCount"
                      type="range"
                      min="1"
                      max="40"
                      value={charactersCount}
                      onChange={(e) => setCharactersCount(parseInt(e.target.value))}
                    />
                    <SliderLabels>
                      <span>1</span>
                      <span>40</span>
                    </SliderLabels>
                  </SliderWrapper>
                </SliderContainer>
                
                <SliderContainer>
                  <Label htmlFor="postsCount">Number of Posts</Label>
                  <SliderValue>{postsCount} posts</SliderValue>
                  <SliderWrapper>
                    <SliderInput
                      id="postsCount"
                      type="range"
                      min="1"
                      max="250"
                      value={postsCount}
                      onChange={(e) => setPostsCount(parseInt(e.target.value))}
                    />
                    <SliderLabels>
                      <span>1</span>
                      <span>250</span>
                    </SliderLabels>
                  </SliderWrapper>
                </SliderContainer>
              </SlidersRow>
              <HelperText>
                Configure how many AI characters and posts will be generated for your world. More content creates a richer experience but takes longer to generate.
              </HelperText>
            </FieldGroup>
            
            <ExamplePrompts>
              <ExamplePromptTitle>Not sure what to write? Try one of these examples:</ExamplePromptTitle>
              {examplePrompts.slice(0, 3).map((examplePrompt, index) => (
                <ExamplePromptCard
                  key={index}
                  onClick={() => handleUseExamplePrompt(examplePrompt)}
                  whileHover={{ scale: 1.01 }}
                  whileTap={{ scale: 0.99 }}
                >
                  "{examplePrompt}"
                </ExamplePromptCard>
              ))}
            </ExamplePrompts>
            
            <ButtonsContainer>
              <Button
                type="button"
                variant="ghost"
                onClick={handleRandomPrompt}
              >
                Surprise Me
              </Button>
              <Button
                type="submit"
                isLoading={isLoading}
                disabled={isLoading}
                fullWidth
              >
                Generate World
              </Button>
            </ButtonsContainer>
          </Form>
        </FormContainer>
        
        {isLoading && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.3 }}
          >
            <Card variant="outline" padding="var(--space-6)">
              <div style={{ textAlign: 'center' }}>
                <Loader />
                <h3 style={{ marginTop: 'var(--space-4)' }}>Creating your world...</h3>
                <p style={{ color: 'var(--color-text-light)', margin: 'var(--space-3) 0' }}>
                  We're building your synthetic world and populating it with AI-generated users and content.
                </p>
                <p style={{ color: 'var(--color-text-light)' }}>
                  This may take a moment as we craft a unique experience based on your prompt.
                </p>
              </div>
            </Card>
          </motion.div>
        )}
      </PageContainer>
    </Layout>
  );
};