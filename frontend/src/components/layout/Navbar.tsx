import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { motion, AnimatePresence, HTMLMotionProps } from 'framer-motion';
import { useAuth } from '../../hooks/useAuth';
import { useWorld } from '../../hooks/useWorld';
import { Avatar } from '../ui/Avatar';
import { Button } from '../ui/Button';

const NavbarContainer = styled.nav`
  position: sticky;
  top: 0;
  z-index: 100;
  width: 100%;
  background-color: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--color-border);
  padding: var(--space-3) 0;
`;

const NavbarInner = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--space-4);
`;

const Logo = styled(Link)`
  font-family: var(--font-sora);
  font-weight: 700;
  font-size: var(--font-xl);
  color: var(--color-text);
  text-decoration: none;
  transition: color 0.2s;
  display: flex;
  align-items: center;
  
  &:hover {
    color: var(--color-primary);
  }
  
  span {
    background: linear-gradient(135deg, var(--color-primary), var(--color-accent));
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
`;

const NavLinks = styled.div`
  display: flex;
  align-items: center;
  gap: var(--space-6);
  
  @media (max-width: 768px) {
    display: none;
  }
`;

const NavLink = styled(Link)<{ $isActive: boolean }>`
  color: ${props => props.$isActive ? 'var(--color-primary)' : 'var(--color-text)'};
  text-decoration: none;
  font-weight: ${props => props.$isActive ? '600' : '500'};
  position: relative;
  
  &::after {
    content: '';
    position: absolute;
    bottom: -6px;
    left: 0;
    width: 100%;
    height: 2px;
    background-color: var(--color-primary);
    transform: scaleX(${props => props.$isActive ? 1 : 0});
    transform-origin: center;
    transition: transform 0.3s ease;
  }
  
  &:hover::after {
    transform: scaleX(1);
  }
`;

const AuthSection = styled.div`
  display: flex;
  align-items: center;
  gap: var(--space-3);
`;

const UserDropdown = styled.div`
  position: relative;
`;

const DropdownTrigger = styled.button`
  display: flex;
  align-items: center;
  gap: var(--space-2);
  background: none;
  border: none;
  padding: var(--space-2);
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background-color 0.2s;
  
  &:hover {
    background-color: var(--color-input-bg);
  }
`;

const UserName = styled.span`
  font-weight: 500;
  
  @media (max-width: 768px) {
    display: none;
  }
`;

const DropdownMenu = styled(motion.div)<HTMLMotionProps<'div'>>`
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background-color: var(--color-card);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  min-width: 180px;
`;

const DropdownItem = styled.button`
  width: 100%;
  text-align: left;
  background: none;
  border: none;
  padding: var(--space-3) var(--space-4);
  cursor: pointer;
  color: var(--color-text);
  transition: background-color 0.2s;
  
  &:hover {
    background-color: var(--color-input-bg);
  }
  
  &:not(:last-child) {
    border-bottom: 1px solid var(--color-border);
  }
`;

const WorldIndicator = styled.div`
  display: flex;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  gap: var(--space-2);
  font-size: var(--font-sm);
  font-weight: 500;
  color: var(--color-text);
  border-radius: var(--radius-full);
  background-color: var(--color-input-bg);
  
  span {
    max-width: 150px;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
`;

const MobileMenuButton = styled.button`
  display: none;
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  
  @media (max-width: 768px) {
    display: block;
  }
`;

const MobileMenu = styled(motion.div)<HTMLMotionProps<'div'>>`
  position: fixed;
  top: 0;
  right: 0;
  width: 80%;
  max-width: 320px;
  height: 100vh;
  background-color: var(--color-card);
  box-shadow: var(--shadow-xl);
  padding: var(--space-6) var(--space-4);
  z-index: 200;
`;

const MobileMenuHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);
`;

const MobileMenuClose = styled.button`
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
`;

const MobileNavLinks = styled.div`
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
`;

const MobileNavLink = styled(Link)<{ $isActive: boolean }>`
  color: ${props => props.$isActive ? 'var(--color-primary)' : 'var(--color-text)'};
  text-decoration: none;
  font-weight: ${props => props.$isActive ? '600' : '500'};
  font-size: var(--font-lg);
  padding: var(--space-2) 0;
`;

const Overlay = styled(motion.div)<HTMLMotionProps<'div'>>`
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 150;
`;

export const Navbar: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const { currentWorld } = useWorld();
  const location = useLocation();
  const navigate = useNavigate();
  
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  
  const toggleDropdown = () => {
    setIsDropdownOpen(!isDropdownOpen);
  };
  
  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };
  
  const isLinkActive = (path: string) => {
    return location.pathname.startsWith(path);
  };
  
  const handleLogout = () => {
    logout();
    navigate('/login');
    setIsDropdownOpen(false);
  };
  
  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (isDropdownOpen) {
        setIsDropdownOpen(false);
      }
    };
    
    document.addEventListener('click', handleClickOutside);
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  }, [isDropdownOpen]);
  
  return (
    <NavbarContainer>
      <NavbarInner>
        <Logo to="/">
          <span>Generia</span>
        </Logo>
        
        {isAuthenticated && (
          <NavLinks>
            <NavLink to="/" $isActive={location.pathname === '/'}>
              Home
            </NavLink>
            <NavLink to="/worlds" $isActive={isLinkActive('/worlds')}>
              Worlds
            </NavLink>
          </NavLinks>
        )}
        
        {/* World indicator removed */}
        
        <AuthSection>
          {isAuthenticated ? (
            <>
              <UserDropdown onClick={(e) => e.stopPropagation()}>
                <DropdownTrigger onClick={toggleDropdown}>
                  <Avatar name={user?.username} size="sm" />
                  <UserName>{user?.username}</UserName>
                </DropdownTrigger>
                
                <AnimatePresence>
                  {isDropdownOpen && (
                    <DropdownMenu
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      transition={{ duration: 0.2 }}
                    >
                      <DropdownItem onClick={() => {
                        setIsDropdownOpen(false);
                        navigate('/profile');
                      }}>
                        Profile
                      </DropdownItem>
                      <DropdownItem onClick={() => {
                        setIsDropdownOpen(false);
                        navigate('/create-world');
                      }}>
                        Create World
                      </DropdownItem>
                      <DropdownItem onClick={handleLogout}>
                        Logout
                      </DropdownItem>
                    </DropdownMenu>
                  )}
                </AnimatePresence>
              </UserDropdown>
            </>
          ) : (
            <>
              <Link to="/login">
                <Button variant="ghost" size="small">
                  Login
                </Button>
              </Link>
              <Link to="/register">
                <Button variant="primary" size="small">
                  Register
                </Button>
              </Link>
            </>
          )}
          
          <MobileMenuButton onClick={toggleMobileMenu}>
            ☰
          </MobileMenuButton>
        </AuthSection>
      </NavbarInner>
      
      {/* Mobile Menu */}
      <AnimatePresence>
        {isMobileMenuOpen && (
          <>
            <Overlay
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={toggleMobileMenu}
            />
            
            <MobileMenu
              initial={{ x: '100%' }}
              animate={{ x: 0 }}
              exit={{ x: '100%' }}
              transition={{ type: 'tween' }}
            >
              <MobileMenuHeader>
                <Logo to="/" onClick={toggleMobileMenu}>
                  <span>Generia</span>
                </Logo>
                <MobileMenuClose onClick={toggleMobileMenu}>
                  ✕
                </MobileMenuClose>
              </MobileMenuHeader>
              
              {isAuthenticated && (
                <MobileNavLinks>
                  <MobileNavLink 
                    to="/" 
                    $isActive={location.pathname === '/'} 
                    onClick={toggleMobileMenu}
                  >
                    Home
                  </MobileNavLink>
                  <MobileNavLink 
                    to="/worlds" 
                    $isActive={isLinkActive('/worlds')} 
                    onClick={toggleMobileMenu}
                  >
                    Worlds
                  </MobileNavLink>
                  <MobileNavLink 
                    to="/create-world" 
                    $isActive={location.pathname === '/create-world'} 
                    onClick={toggleMobileMenu}
                  >
                    Create World
                  </MobileNavLink>
                  
                  <MobileNavLink 
                    to="/profile" 
                    $isActive={location.pathname === '/profile'} 
                    onClick={toggleMobileMenu}
                  >
                    Profile
                  </MobileNavLink>
                  
                  <Button 
                    variant="ghost" 
                    onClick={() => {
                      handleLogout();
                      toggleMobileMenu();
                    }}
                  >
                    Logout
                  </Button>
                </MobileNavLinks>
              )}
              
              {!isAuthenticated && (
                <div style={{ display: 'flex', gap: 'var(--space-4)', flexDirection: 'column' }}>
                  <Link to="/login" onClick={toggleMobileMenu}>
                    <Button variant="ghost" fullWidth>
                      Login
                    </Button>
                  </Link>
                  <Link to="/register" onClick={toggleMobileMenu}>
                    <Button variant="primary" fullWidth>
                      Register
                    </Button>
                  </Link>
                </div>
              )}
            </MobileMenu>
          </>
        )}
      </AnimatePresence>
    </NavbarContainer>
  );
};