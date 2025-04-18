import React, { useContext, useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import axiosInstance from '../api/axios';
import { World } from '../types';

const Navbar: React.FC = () => {
  const { isAuthenticated, user, logout } = useContext(AuthContext);
  const [activeWorld, setActiveWorld] = useState<World | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  // Извлекаем worldId из URL-пути, если он там есть
  const getWorldIdFromPath = (): string | null => {
    const match = location.pathname.match(/\/worlds\/([^/]+)/);
    return match ? match[1] : null;
  };

  useEffect(() => {
    if (isAuthenticated) {
      const worldId = getWorldIdFromPath();
      if (worldId) {
        fetchWorldInfo(worldId);
      } else {
        setActiveWorld(null);
      }
    }
  }, [isAuthenticated, location.pathname]);

  const fetchWorldInfo = async (worldId: string) => {
    try {
      const response = await axiosInstance.get(`/worlds/${worldId}`);
      if (response.data && response.data.id) {
        setActiveWorld(response.data);
        // Сохраняем ID активного мира в localStorage
        localStorage.setItem('activeWorldId', worldId);
      }
    } catch (err) {
      console.log('Failed to fetch world info:', err);
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
            {activeWorld && (
              <Link to={`/worlds/${activeWorld.id}/create`} className="navbar-item">Create Post</Link>
            )}
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