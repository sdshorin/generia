import React, { forwardRef } from 'react';
import styled, { css } from 'styled-components';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  icon?: React.ReactNode;
  fullWidth?: boolean;
}

const InputContainer = styled.div<{ $fullWidth: boolean }>`
  display: flex;
  flex-direction: column;
  margin-bottom: var(--space-4);
  width: ${(props) => (props.$fullWidth ? '100%' : 'auto')};
`;

const InputLabel = styled.label`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  margin-bottom: var(--space-2);
`;

const InputWrapper = styled.div`
  position: relative;
  display: flex;
  align-items: center;
`;

const StyledInput = styled.input<{ $hasError: boolean; $hasIcon: boolean }>`
  width: 100%;
  padding: var(--space-3) var(--space-4);
  padding-left: ${(props) => (props.$hasIcon ? 'var(--space-10)' : 'var(--space-4)')};
  font-size: var(--font-md);
  background-color: var(--color-input-bg);
  color: var(--color-text);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  transition: all var(--duration-fast) ease;
  
  &:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(255, 199, 95, 0.2);
  }
  
  &::placeholder {
    color: var(--color-text-lighter);
  }
  
  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  ${(props) =>
    props.$hasError &&
    css`
      border-color: var(--color-error);
      
      &:focus {
        box-shadow: 0 0 0 2px rgba(252, 165, 165, 0.2);
      }
    `}
`;

const IconWrapper = styled.div`
  position: absolute;
  left: var(--space-3);
  color: var(--color-text-light);
  display: flex;
  align-items: center;
  justify-content: center;
`;

const ErrorMessage = styled.div`
  font-size: var(--font-xs);
  color: var(--color-accent);
  margin-top: var(--space-1);
`;

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, icon, fullWidth = true, ...props }, ref) => {
    return (
      <InputContainer $fullWidth={fullWidth}>
        {label && <InputLabel>{label}</InputLabel>}
        <InputWrapper>
          {icon && <IconWrapper>{icon}</IconWrapper>}
          <StyledInput
            ref={ref}
            $hasError={!!error}
            $hasIcon={!!icon}
            {...props}
          />
        </InputWrapper>
        {error && <ErrorMessage>{error}</ErrorMessage>}
      </InputContainer>
    );
  }
);