import React, { forwardRef, CSSProperties } from 'react';
import styled from 'styled-components';
import TextareaAutosize, { TextareaAutosizeProps } from 'react-textarea-autosize';

// Props for the component
interface TextAreaProps extends Omit<React.TextareaHTMLAttributes<HTMLTextAreaElement>, 'rows' | 'style'> {
  label?: string;
  error?: string;
  fullWidth?: boolean;
  rows?: number;
  maxRows?: number;
  style?: Omit<CSSProperties, 'resize'>; // Remove 'resize' from allowed style properties
}

const TextAreaContainer = styled.div<{ $fullWidth: boolean }>`
  display: flex;
  flex-direction: column;
  margin-bottom: var(--space-4);
  width: ${(props) => (props.$fullWidth ? '100%' : 'auto')};
`;

const TextAreaLabel = styled.label`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  margin-bottom: var(--space-2);
`;

const ErrorMessage = styled.div`
  font-size: var(--font-xs);
  color: var(--color-accent);
  margin-top: var(--space-1);
`;

// Create our own textarea to avoid styled-components issues
export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(
  ({ label, error, fullWidth = true, rows = 3, maxRows = 10, style, className, ...props }, ref) => {
    // Create a CSS class based on error state
    const hasError = !!error;
    const textareaClassName = `textarea ${hasError ? 'has-error' : ''} ${className || ''}`;
    
    return (
      <TextAreaContainer $fullWidth={fullWidth}>
        {label && <TextAreaLabel>{label}</TextAreaLabel>}
        <div className="textarea-container">
          <TextareaAutosize
            minRows={rows}
            maxRows={maxRows}
            className={textareaClassName.trim()}
            style={{
              width: '100%',
              padding: 'var(--space-3) var(--space-4)',
              fontSize: 'var(--font-md)',
              backgroundColor: 'var(--color-input-bg)',
              color: 'var(--color-text)',
              border: hasError ? '1px solid var(--color-error)' : '1px solid transparent',
              borderRadius: 'var(--radius-md)',
              fontFamily: 'var(--font-sans)',
              ...(style || {})
            } as TextareaAutosizeProps['style']}
            ref={ref}
            {...props}
          />
        </div>
        {error && <ErrorMessage>{error}</ErrorMessage>}
      </TextAreaContainer>
    );
  }
);

// Add display name for debugging
TextArea.displayName = 'TextArea';