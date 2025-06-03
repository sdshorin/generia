import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import '../../styles/pages/auth.css';

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
    <div className="auth-container">
      <div className="auth-card">
        
        {/* Auth Header */}
        <div className="auth-header">
          <div className="auth-logo">
            <div className="auth-logo-icon"></div>
            <span className="auth-logo-text">Generia</span>
          </div>
          <h1 className="auth-title">Create your account</h1>
          <p className="auth-subtitle">Join Generia and start exploring infinite virtual worlds powered by AI</p>
        </div>

        {/* Register Form */}
        <form className="auth-form" onSubmit={handleSubmit}>
          
          {/* General Error Message */}
          {(error || formError) && (
            <div className="auth-error-message show" style={{textAlign: 'center', marginBottom: 'var(--spacing-4)'}}>
              {formError || error}
            </div>
          )}
          
          {/* Username Field */}
          <div className="auth-form-group">
            <label className="auth-form-label" htmlFor="username">Username</label>
            <input 
              type="text" 
              id="username" 
              name="username" 
              className="auth-form-input" 
              placeholder="Choose a unique username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              minLength={3}
              maxLength={20}
            />
          </div>

          {/* Email Field */}
          <div className="auth-form-group">
            <label className="auth-form-label" htmlFor="email">Email Address</label>
            <input 
              type="email" 
              id="email" 
              name="email" 
              className="auth-form-input" 
              placeholder="Enter your email address"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
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
              placeholder="Create a strong password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
            />
          </div>

          {/* Confirm Password Field */}
          <div className="auth-form-group">
            <label className="auth-form-label" htmlFor="confirm-password">Confirm Password</label>
            <input 
              type="password" 
              id="confirm-password" 
              name="confirm-password" 
              className="auth-form-input" 
              placeholder="Confirm your password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
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
              {isLoading ? '' : 'Create Account'}
            </button>
          </div>

        </form>

        {/* Auth Toggle */}
        <div className="auth-toggle">
          <span className="auth-toggle-text">Already have an account?</span>
          <Link to="/login" className="auth-toggle-link">Sign in</Link>
        </div>

      </div>
    </div>
  );
};