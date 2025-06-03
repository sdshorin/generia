import React, { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import '../../styles/pages/auth.css';

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
    <div className="auth-container">
      <div className="auth-card">
        
        {/* Auth Header */}
        <div className="auth-header">
          <div className="auth-logo">
            <div className="auth-logo-icon"></div>
            <span className="auth-logo-text">Generia</span>
          </div>
          <h1 className="auth-title">Welcome back</h1>
          <p className="auth-subtitle">Sign in to your account to continue exploring virtual worlds</p>
        </div>

        {/* Login Form */}
        <form className="auth-form" onSubmit={handleSubmit}>
          
          {/* General Error Message */}
          {(error || formError) && (
            <div className="auth-error-message show" style={{textAlign: 'center', marginBottom: 'var(--spacing-4)'}}>
              {formError || error}
            </div>
          )}
          
          {/* Email Field */}
          <div className="auth-form-group">
            <label className="auth-form-label" htmlFor="email">Email or Username</label>
            <input 
              type="text" 
              id="email" 
              name="email" 
              className="auth-form-input" 
              placeholder="Enter your email or username"
              value={emailOrUsername}
              onChange={(e) => setEmailOrUsername(e.target.value)}
              required
            />
          </div>

          {/* Password Field */}
          <div className="auth-form-group">
            <label className="auth-form-label" htmlFor="password">Password</label>
            <input 
              type="password" 
              id="password" 
              name="password" 
              className="auth-form-input" 
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>

          {/* Submit Button */}
          <div className="auth-actions">
            <button 
              type="submit" 
              className={`auth-submit-btn ${isLoading ? 'loading' : ''}`}
              disabled={isLoading}
            >
              {isLoading ? '' : 'Sign In'}
            </button>
          </div>

        </form>

        {/* Auth Toggle */}
        <div className="auth-toggle">
          <span className="auth-toggle-text">Don't have an account?</span>
          <Link to="/register" className="auth-toggle-link">Sign up for free</Link>
        </div>

      </div>
    </div>
  );
};