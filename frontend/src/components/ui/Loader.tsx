import React from 'react';
import styled, { keyframes } from 'styled-components';

type LoaderSize = 'sm' | 'md' | 'lg';
type LoaderVariant = 'primary' | 'secondary' | 'light' | 'dark';

interface LoaderProps {
  size?: LoaderSize;
  variant?: LoaderVariant;
  text?: string;
  fullScreen?: boolean;
}

const spin = keyframes`
  to {
    transform: rotate(360deg);
  }
`;

const ripple = keyframes`
  0% {
    transform: scale(0);
    opacity: 1;
  }
  100% {
    transform: scale(1);
    opacity: 0;
  }
`;

const getSize = (size: LoaderSize) => {
  switch (size) {
    case 'sm': return '24px';
    case 'lg': return '48px';
    default: return '36px';
  }
};

const getColor = (variant: LoaderVariant) => {
  switch (variant) {
    case 'secondary': return 'var(--color-secondary)';
    case 'light': return 'rgba(255, 255, 255, 0.8)';
    case 'dark': return 'var(--color-text)';
    default: return 'var(--color-primary)';
  }
};

const Container = styled.div<{ $fullScreen: boolean }>`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  padding: var(--space-6);
  ${props => props.$fullScreen && `
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 255, 255, 0.9);
    z-index: 1000;
  `}
`;

const SpinnerContainer = styled.div<{ $size: string }>`
  position: relative;
  width: ${props => props.$size};
  height: ${props => props.$size};
`;

const Spinner = styled.div<{ $size: string; $color: string }>`
  width: ${props => props.$size};
  height: ${props => props.$size};
  border: 2px solid transparent;
  border-top-color: ${props => props.$color};
  border-radius: 50%;
  animation: ${spin} 0.8s linear infinite;
`;

const RippleContainer = styled.div<{ $size: string }>`
  position: relative;
  width: ${props => props.$size};
  height: ${props => props.$size};
`;

const RippleCircle = styled.div<{ $color: string; $delay: number }>`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background-color: ${props => props.$color};
  opacity: 0;
  animation: ${ripple} 1.4s cubic-bezier(0, 0.2, 0.8, 1) ${props => props.$delay}s infinite;
`;

const LoaderText = styled.div`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  margin-top: var(--space-2);
  text-align: center;
`;

export const Loader: React.FC<LoaderProps> = ({
  size = 'md',
  variant = 'primary',
  text,
  fullScreen = false,
}) => {
  const sizeValue = getSize(size);
  const colorValue = getColor(variant);

  return (
    <Container $fullScreen={fullScreen}>
      <RippleContainer $size={sizeValue}>
        <RippleCircle $color={colorValue} $delay={0} />
        <RippleCircle $color={colorValue} $delay={0.5} />
      </RippleContainer>
      {text && <LoaderText>{text}</LoaderText>}
    </Container>
  );
};