import React, { useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Navbar from './components/Navbar';
import Feed from './components/Feed';
import Login from './components/Login';
import Register from './components/Register';
import CreatePost from './components/CreatePost';
import WorldsList from './components/WorldsList';
import CreateWorld from './components/CreateWorld';
import { AuthContext } from './context/AuthContext';

const App: React.FC = () => {
  const { isAuthenticated, isLoading } = useContext(AuthContext);

  if (isLoading) {
    return <div className="loading">Loading...</div>;
  }

  return (
    <Router>
      <div className="app">
        <Navbar />
        <Routes>
          <Route path="/" element={isAuthenticated ? <Feed /> : <Navigate to="/login" />} />
          <Route
            path="/login"
            element={isAuthenticated ? <Navigate to="/worlds" /> : <Login />}
          />
          <Route
            path="/register"
            element={isAuthenticated ? <Navigate to="/worlds" /> : <Register />}
          />
          <Route
            path="/create"
            element={isAuthenticated ? <CreatePost /> : <Navigate to="/login" />}
          />
          <Route
            path="/worlds"
            element={isAuthenticated ? <WorldsList /> : <Navigate to="/login" />}
          />
          <Route
            path="/create-world"
            element={isAuthenticated ? <CreateWorld /> : <Navigate to="/login" />}
          />
          <Route
            path="/feed"
            element={isAuthenticated ? <Feed /> : <Navigate to="/login" />}
          />
        </Routes>
      </div>
    </Router>
  );
};

export default App;