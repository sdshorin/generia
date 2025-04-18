import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { useAuth } from '../../hooks/useAuth';
import { Input } from '../../components/ui/Input';
import { Button } from '../../components/ui/Button';

const LoginContainer = styled.div`
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
    background: radial-gradient(circle at 10% 20%, rgba(255, 199, 95, 0.15) 0%, rgba(255, 255, 255, 0) 80%);
    z-index: -1;
  }
`;

const LoginCard = styled(motion.div)<HTMLMotionProps<'div'>>`
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

const LoginHeader = styled.div`
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

const LoginForm = styled.form`
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
`;

const InputLabel = styled.label`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  margin-bottom: var(--space-1);
`;

const ForgotPassword = styled(Link)`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  text-align: right;
  margin-top: var(--space-1);
  
  &:hover {
    color: var(--color-primary);
  }
`;

const FormError = styled.div`
  color: var(--color-accent);
  font-size: var(--font-sm);
  padding: var(--space-3);
  background-color: rgba(239, 118, 122, 0.1);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-4);
`;

const RegisterLink = styled.div`
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

export const LoginPage: React.FC = () => {
  const [emailOrUsername, setEmailOrUsername] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState('');
  const { login, error, clearError, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  
  // Get redirect path from location state or default to home
  const from = location.state?.from?.pathname || '/';
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    clearError();
    
    if (!emailOrUsername.trim()) {
      setFormError('Please enter your email or username');
      return;
    }
    
    if (!password) {
      setFormError('Please enter your password');
      return;
    }
    
    try {
      await login(emailOrUsername, password);
      navigate(from, { replace: true });
    } catch (err) {
      // Error is already handled in the AuthContext
      console.error('Login failed');
    }
  };
  
  return (
    <LoginContainer>
      <LoginCard
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <LoginHeader>
          <LogoText>Generia</LogoText>
          <SubTitle>Sign in to your account</SubTitle>
        </LoginHeader>
        
        {(error || formError) && (
          <FormError>
            {formError || error}
          </FormError>
        )}
        
        <LoginForm onSubmit={handleSubmit}>
          <div>
            <InputLabel htmlFor="emailOrUsername">Email or Username</InputLabel>
            <Input
              id="emailOrUsername"
              type="text"
              value={emailOrUsername}
              onChange={(e) => setEmailOrUsername(e.target.value)}
              placeholder="Enter your email or username"
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
              placeholder="Enter your password"
              fullWidth
            />
            <ForgotPassword to="/forgot-password">Forgot password?</ForgotPassword>
          </div>
          
          <Button 
            type="submit" 
            fullWidth 
            isLoading={isLoading}
            disabled={isLoading}
          >
            Sign In
          </Button>
        </LoginForm>
        
        <RegisterLink>
          Don't have an account? <Link to="/register">Sign up</Link>
        </RegisterLink>
      </LoginCard>
    </LoginContainer>
  );
};