import React, { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { mockCredits } from '../../utils/mockData';
import '../../styles/components.css';

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const { user } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  const isActiveLink = (path: string) => {
    return location.pathname === path;
  };

  const getLinkClassName = (path: string) => {
    return `nav-link ${isActiveLink(path) ? 'active' : ''}`;
  };

  const getMobileLinkClassName = (path: string) => {
    return `mobile-nav-link ${isActiveLink(path) ? 'active' : ''}`;
  };

  const handleAvatarClick = () => {
    navigate('/settings');
  };

  return (
    <>
      <header className="header">
        <div className="header-left">
          {/* Logo */}
          <Link to="/" className="logo">
            <div className="logo-icon"></div>
            <span className="logo-text">Generia</span>
          </Link>
          
          {/* Navigation */}
          <nav className="nav">
            <Link to="/worlds" className={getLinkClassName('/worlds')}>
              Explore Worlds
            </Link>
            <a 
              href="https://your-research-paper-link.com" 
              className="nav-link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Research Paper
            </a>
            <Link to="/create-world" className="btn btn-primary btn-sm">
              Create World
            </Link>
          </nav>
          
          {/* Mobile menu button */}
          <button className="mobile-menu-btn" onClick={toggleMobileMenu}>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M3 12H21M3 6H21M3 18H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
        
        {/* User section */}
        <div className="header-right">
          {/* Credits display */}
          <div className="credits-display">
            <span>💎</span>
            <span className="credits-full">{mockCredits.balance.toLocaleString()} Credits</span>
            <span className="credits-short">{mockCredits.balance >= 1000 ? `${(mockCredits.balance / 1000).toFixed(1)}K` : mockCredits.balance}</span>
          </div>
          
          {/* User avatar */}
          <div 
            className="user-avatar" 
            style={{
              backgroundImage: user?.avatar_url 
                ? `url(${user.avatar_url})` 
                : 'url(/no-image.jpg)',
              cursor: 'pointer'
            }}
            onClick={handleAvatarClick}
            title="Settings"
          >
          </div>
        </div>
      </header>

      {/* Mobile Menu */}
      <div className={`mobile-menu ${isMobileMenuOpen ? 'show' : ''}`}>
        <nav className="mobile-menu-nav">
          <Link 
            to="/worlds" 
            className={getMobileLinkClassName('/worlds')}
            onClick={() => setIsMobileMenuOpen(false)}
          >
            Explore Worlds
          </Link>
          <a 
            href="https://your-research-paper-link.com" 
            className="mobile-nav-link"
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => setIsMobileMenuOpen(false)}
          >
            Research Paper
          </a>
          <Link 
            to="/create-world" 
            className="btn btn-primary"
            onClick={() => setIsMobileMenuOpen(false)}
          >
            Create World
          </Link>
        </nav>
      </div>
    </>
  );
};

export default Header;