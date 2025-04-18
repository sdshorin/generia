import React from 'react';
import styled from 'styled-components';

type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

interface AvatarProps {
  src?: string;
  name?: string;
  size?: AvatarSize;
  isAi?: boolean;
  className?: string;
}

const AvatarContainer = styled.div<{ $size: AvatarSize; $isAi: boolean }>`
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  overflow: hidden;
  background-color: var(--color-primary);
  color: white;
  font-weight: 600;
  
  /* Sizing */
  ${({ $size }) => {
    switch ($size) {
      case 'xs':
        return `
          width: 24px;
          height: 24px;
          font-size: 12px;
        `;
      case 'sm':
        return `
          width: 32px;
          height: 32px;
          font-size: 14px;
        `;
      case 'lg':
        return `
          width: 48px;
          height: 48px;
          font-size: 20px;
        `;
      case 'xl':
        return `
          width: 64px;
          height: 64px;
          font-size: 24px;
        `;
      default: // 'md'
        return `
          width: 40px;
          height: 40px;
          font-size: 16px;
        `;
    }
  }}
  
  /* AI indicator */
  ${({ $isAi }) =>
    $isAi &&
    `
    &::after {
      content: "";
      position: absolute;
      bottom: 0;
      right: 0;
      width: 30%;
      height: 30%;
      background-color: var(--color-secondary);
      border-radius: 50%;
      border: 2px solid white;
    }
  `}
`;

const AvatarImage = styled.img`
  width: 100%;
  height: 100%;
  object-fit: cover;
`;

export const Avatar: React.FC<AvatarProps> = ({
  src,
  name = '',
  size = 'md',
  isAi = false,
  className,
}) => {
  // Extract initials from name (up to 2 characters)
  const getInitials = (name: string) => {
    if (!name) return '';
    const parts = name.split(' ');
    if (parts.length >= 2) {
      return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
    }
    return name.substring(0, 2).toUpperCase();
  };

  return (
    <AvatarContainer $size={size} $isAi={isAi} className={className}>
      {src ? (
        <AvatarImage src={src} alt={name || 'User avatar'} />
      ) : (
        getInitials(name)
      )}
    </AvatarContainer>
  );
};