import React from 'react';
import styled, { css } from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';

type CardVariant = 'default' | 'elevated' | 'outline' | 'minimal';

// Props for the styled component
type CardContainerProps = HTMLMotionProps<'div'> & {
  $variant: CardVariant;
  $padding: string;
  $animateHover: boolean;
};

// Props for the exported Card component
interface CardProps {
  variant?: CardVariant;
  padding?: string;
  className?: string;
  children: React.ReactNode;
  onClick?: () => void;
  animateHover?: boolean;
}

const CardContainer = styled(motion.div)<CardContainerProps>`
  background-color: var(--color-card);
  border-radius: var(--radius-lg);
  overflow: hidden;
  padding: ${(props) => props.$padding};
  width: 100%;
  
  ${(props) => {
    switch (props.$variant) {
      case 'elevated':
        return css`
          box-shadow: var(--shadow-md);
        `;
      case 'outline':
        return css`
          border: 1px solid var(--color-border);
        `;
      case 'minimal':
        return css`
          background-color: transparent;
        `;
      default:
        return css`
          box-shadow: var(--shadow-sm);
        `;
    }
  }}
  
  ${(props) =>
    props.$animateHover &&
    css`
      cursor: pointer;
      transition: transform 0.2s ease, box-shadow 0.2s ease;
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: var(--shadow-lg);
      }
    `}
`;

export const Card: React.FC<CardProps> = ({
  variant = 'default',
  padding = 'var(--space-6)',
  className,
  children,
  onClick,
  animateHover = false,
}) => {
  return (
    <CardContainer
      $variant={variant}
      $padding={padding}
      $animateHover={animateHover}
      className={className}
      onClick={onClick}
      whileTap={onClick ? { scale: 0.98 } : undefined}
    >
      {children}
    </CardContainer>
  );
};