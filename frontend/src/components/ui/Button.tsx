import React from 'react';
import styled, { css } from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';

type ButtonVariant = 'primary' | 'secondary' | 'accent' | 'ghost' | 'text';
type ButtonSize = 'small' | 'medium' | 'large';

// Combine HTML button props with motion props
type ButtonContainerProps = HTMLMotionProps<'button'> & {
  $variant: ButtonVariant;
  $size: ButtonSize;
  $fullWidth: boolean;
};

// Interface for the exported Button component
interface ButtonProps extends Omit<HTMLMotionProps<'button'>, '$variant' | '$size' | '$fullWidth'> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  fullWidth?: boolean;
  icon?: React.ReactNode;
  isLoading?: boolean;
}

const ButtonContainer = styled(motion.button)<ButtonContainerProps>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: none;
  border-radius: var(--radius-md);
  font-family: var(--font-sans);
  font-weight: 500;
  transition: all var(--duration-normal) ease;
  cursor: pointer;
  width: ${(props) => (props.$fullWidth ? '100%' : 'auto')};
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  /* Size variants */
  ${(props) => {
    switch (props.$size) {
      case 'small':
        return css`
          padding: 6px 12px;
          font-size: var(--font-sm);
          height: 32px;
        `;
      case 'large':
        return css`
          padding: 12px 24px;
          font-size: var(--font-lg);
          height: 48px;
        `;
      default: // medium
        return css`
          padding: 8px 16px;
          font-size: var(--font-md);
          height: 40px;
        `;
    }
  }}
  
  /* Color variants */
  ${(props) => {
    switch (props.$variant) {
      case 'secondary':
        return css`
          background-color: var(--color-secondary);
          color: var(--color-text);
          font-weight: 600;
          &:hover:not(:disabled) {
            background-color: #8da3fb;
            transform: translateY(-1px);
          }
          &:active:not(:disabled) {
            background-color: #7a90f7;
            transform: translateY(0);
          }
        `;
      case 'accent':
        return css`
          background-color: var(--color-accent);
          color: var(--color-text);
          font-weight: 600;
          &:hover:not(:disabled) {
            background-color: #e5636b;
            transform: translateY(-1px);
          }
          &:active:not(:disabled) {
            background-color: #d95258;
            transform: translateY(0);
          }
        `;
      case 'ghost':
        return css`
          background-color: transparent;
          box-shadow: inset 0 0 0 1px var(--color-border);
          color: var(--color-text);
          &:hover:not(:disabled) {
            background-color: rgba(0, 0, 0, 0.04);
          }
          &:active:not(:disabled) {
            background-color: rgba(0, 0, 0, 0.08);
          }
        `;
      case 'text':
        return css`
          background-color: transparent;
          color: var(--color-text);
          padding-left: 8px;
          padding-right: 8px;
          &:hover:not(:disabled) {
            color: var(--color-primary);
            background-color: transparent;
          }
        `;
      default: // primary
        return css`
          background-color: var(--color-primary);
          color: var(--color-text);
          font-weight: 600;
          &:hover:not(:disabled) {
            background-color: var(--color-primary-hover);
            transform: translateY(-1px);
            box-shadow: 0 4px 8px rgba(255, 199, 95, 0.3);
          }
          &:active:not(:disabled) {
            background-color: #efb03f;
            transform: translateY(0);
            box-shadow: none;
          }
        `;
    }
  }}
`;

const Spinner = styled.div`
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(47, 47, 47, 0.3);
  border-radius: 50%;
  border-top-color: var(--color-text);
  animation: spin 0.8s linear infinite;
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
`;

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  size = 'medium',
  fullWidth = false,
  icon,
  isLoading = false,
  disabled,
  ...props
}) => {
  return (
    <ButtonContainer
      $variant={variant}
      $size={size}
      $fullWidth={fullWidth}
      disabled={disabled || isLoading}
      whileTap={{ scale: disabled || isLoading ? 1 : 0.98 }}
      {...props}
    >
      {isLoading ? (
        <Spinner />
      ) : (
        <>
          {icon && <span className="button-icon">{icon}</span>}
          {children}
        </>
      )}
    </ButtonContainer>
  );
};