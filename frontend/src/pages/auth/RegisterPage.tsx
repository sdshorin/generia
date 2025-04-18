import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { useAuth } from '../../hooks/useAuth';
import { Input } from '../../components/ui/Input';
import { Button } from '../../components/ui/Button';

const RegisterContainer = styled.div`
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  position: relative;
  overflow: hidden;
  
  &::before {
    content: '';
    position: absolute;
    width: 100%;
    height: 100%;
    background: radial-gradient(circle at 90% 20%, rgba(239, 118, 122, 0.15) 0%, rgba(255, 255, 255, 0) 80%);
    z-index: -1;
  }
`;

const RegisterCard = styled(motion.div)<HTMLMotionProps<'div'>>`
  background-color: var(--color-card);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 420px;
  padding: var(--space-8);
  
  @media (max-width: 480px) {
    padding: var(--space-6);
  }
`;

const RegisterHeader = styled.div`
  text-align: center;
  margin-bottom: var(--space-6);
`;

const LogoText = styled.h1`
  font-family: var(--font-sora);
  font-size: var(--font-3xl);
  margin-bottom: var(--space-2);
  background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
`;

const SubTitle = styled.p`
  color: var(--color-text-light);
  font-size: var(--font-md);
`;

const RegisterForm = styled.form`
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
`;

const InputLabel = styled.label`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  margin-bottom: var(--space-1);
`;

const FormError = styled.div`
  color: var(--color-accent);
  font-size: var(--font-sm);
  padding: var(--space-3);
  background-color: rgba(239, 118, 122, 0.1);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
`;

const LoginLink = styled.div`
  text-align: center;
  margin-top: var(--space-6);
  font-size: var(--font-sm);
  color: var(--color-text-light);
  
  a {
    color: var(--color-primary);
    font-weight: 500;
    
    &:hover {
      text-decoration: underline;
    }
  }
`;

export const RegisterPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [formError, setFormError] = useState('');
  const { register, error, clearError, isLoading } = useAuth();
  const navigate = useNavigate();
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    clearError();
    
    // Basic validation
    if (!username.trim() || !email.trim() || !password || !confirmPassword) {
      setFormError('All fields are required');
      return;
    }
    
    if (username.trim().length < 3) {
      setFormError('Username must be at least 3 characters long');
      return;
    }
    
    if (!email.includes('@') || !email.includes('.')) {
      setFormError('Please enter a valid email address');
      return;
    }
    
    if (password.length < 6) {
      setFormError('Password must be at least 6 characters long');
      return;
    }
    
    if (password !== confirmPassword) {
      setFormError('Passwords do not match');
      return;
    }
    
    try {
      await register(username, email, password);
      navigate('/');
    } catch (err) {
      // Error is already handled in the AuthContext
      console.error('Registration failed');
    }
  };
  
  return (
    <RegisterContainer>
      <RegisterCard
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <RegisterHeader>
          <LogoText>Generia</LogoText>
          <SubTitle>Create your account</SubTitle>
        </RegisterHeader>
        
        {(error || formError) && (
          <FormError>
            {formError || error}
          </FormError>
        )}
        
        <RegisterForm onSubmit={handleSubmit}>
          <div>
            <InputLabel htmlFor="username">Username</InputLabel>
            <Input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Choose a username"
              fullWidth
            />
          </div>
          
          <div>
            <InputLabel htmlFor="email">Email</InputLabel>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Enter your email"
              fullWidth
            />
          </div>
          
          <div>
            <InputLabel htmlFor="password">Password</InputLabel>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Create a password"
              fullWidth
            />
          </div>
          
          <div>
            <InputLabel htmlFor="confirmPassword">Confirm Password</InputLabel>
            <Input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm your password"
              fullWidth
            />
          </div>
          
          <Button 
            type="submit" 
            fullWidth 
            isLoading={isLoading}
            disabled={isLoading}
          >
            Create Account
          </Button>
        </RegisterForm>
        
        <LoginLink>
          Already have an account? <Link to="/login">Sign in</Link>
        </LoginLink>
      </RegisterCard>
    </RegisterContainer>
  );
};