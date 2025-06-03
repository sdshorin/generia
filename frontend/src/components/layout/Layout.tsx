import React from 'react';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import Header from './Header';

interface LayoutProps {
  children: React.ReactNode;
  fullWidth?: boolean;
}

const MainContainer = styled.main<{ $fullWidth: boolean }>`
  flex: 1;
  width: 100%;
  // max-width: ${props => props.$fullWidth ? '100%' : '1200px'};
  max-width: '100%';
  margin: 0 auto;
  padding: ${props => props.$fullWidth ? '0' : 'var(--space-6) var(--space-4)'};
  
  @media (max-width: 768px) {
    padding: ${props => props.$fullWidth ? '0' : 'var(--space-4) var(--space-3)'};
  }
`;

const pageVariants = {
  initial: {
    opacity: 0,
  },
  in: {
    opacity: 1,
  },
  out: {
    opacity: 0,
  },
};

const pageTransition = {
  type: 'tween',
  ease: 'easeInOut',
  duration: 0.3,
};

export const Layout: React.FC<LayoutProps> = ({ children, fullWidth = false }) => {
  return (
    <>
      <Header />
      <MainContainer $fullWidth={fullWidth}>
        <motion.div
          initial="initial"
          animate="in"
          exit="out"
          variants={pageVariants}
          transition={pageTransition}
        >
          {children}
        </motion.div>
      </MainContainer>
    </>
  );
};