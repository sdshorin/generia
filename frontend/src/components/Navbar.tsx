import React, { useContext, useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import axiosInstance from '../api/axios';
import { World } from '../types';

const Navbar: React.FC = () => {
  const { isAuthenticated, user, logout } = useContext(AuthContext);
  const [activeWorld, setActiveWorld] = useState<World | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      fetchActiveWorld();
    }
  }, [isAuthenticated]);

  const fetchActiveWorld = async () => {
    try {
      const response = await axiosInstance.get('/worlds/active');
      if (response.data && response.data.id) {
        setActiveWorld(response.data);
      }
    } catch (err) {
      // If no active world, this is not an error
      console.log('No active world found:', err);
      // Сбрасываем активный мир если есть ошибка
      setActiveWorld(null);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const handleWorldsClick = () => {
    navigate('/worlds');
  };

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <Link to="/">Generia</Link>
      </div>
      <div className="navbar-menu">
        {isAuthenticated ? (
          <>
            {activeWorld && (
              <div className="active-world" onClick={handleWorldsClick}>
                <span className="world-indicator">World:</span>
                <span className="world-name">{activeWorld.name}</span>
              </div>
            )}
            <Link to="/worlds" className="navbar-item">Worlds</Link>
            <Link to="/create" className="navbar-item">Create Post</Link>
            <span className="navbar-item">Welcome, {user?.username}</span>
            <button onClick={handleLogout} className="navbar-item logout-button">Logout</button>
          </>
        ) : (
          <>
            <Link to="/login" className="navbar-item">Login</Link>
            <Link to="/register" className="navbar-item">Register</Link>
          </>
        )}
      </div>
    </nav>
  );
};

export default Navbar;