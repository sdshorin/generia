import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { WorldProvider } from './context/WorldContext';
import { ProtectedRoute } from './components/layout/ProtectedRoute';

// Pages
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/auth/LoginPage';
import { RegisterPage } from './pages/auth/RegisterPage';
import { WorldsListPage } from './pages/worlds/WorldsListPage';
import { CreateWorldPage } from './pages/worlds/CreateWorldPage';
import { FeedPage } from './pages/posts/FeedPage';
import { CreatePostPage } from './pages/posts/CreatePostPage';
import { ViewPostPage } from './pages/posts/ViewPostPage';
import { CreateCharacterPage } from './pages/posts/CreateCharacterPage';
import { ProfilePage } from './pages/user/ProfilePage';

const App: React.FC = () => {
  return (
    <Router>
      <AuthProvider>
        <WorldProvider>
          <Routes>
            {/* Public routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            
            {/* Protected routes */}
            <Route path="/" element={
              <ProtectedRoute>
                <HomePage />
              </ProtectedRoute>
            } />
            
            {/* Worlds routes */}
            <Route path="/worlds" element={
              <ProtectedRoute>
                <WorldsListPage />
              </ProtectedRoute>
            } />
            <Route path="/create-world" element={
              <ProtectedRoute>
                <CreateWorldPage />
              </ProtectedRoute>
            } />
            
            {/* Posts routes */}
            <Route path="/worlds/:worldId/feed" element={
              <ProtectedRoute>
                <FeedPage />
              </ProtectedRoute>
            } />
            <Route path="/worlds/:worldId/create" element={
              <ProtectedRoute>
                <CreatePostPage />
              </ProtectedRoute>
            } />
            <Route path="/worlds/:worldId/posts/:postId" element={
              <ProtectedRoute>
                <ViewPostPage />
              </ProtectedRoute>
            } />
            
            {/* Character routes */}
            <Route path="/worlds/:worldId/characters/create" element={
              <ProtectedRoute>
                <CreateCharacterPage />
              </ProtectedRoute>
            } />
            
            {/* User profile */}
            <Route path="/profile" element={
              <ProtectedRoute>
                <ProfilePage />
              </ProtectedRoute>
            } />
            <Route path="/profile/:userId" element={
              <ProtectedRoute>
                <ProfilePage />
              </ProtectedRoute>
            } />
            
            {/* Fallback route */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </WorldProvider>
      </AuthProvider>
    </Router>
  );
};

export default App;