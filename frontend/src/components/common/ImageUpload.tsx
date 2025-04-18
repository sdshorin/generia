import React, { useState, useCallback, useEffect } from 'react';
import styled from 'styled-components';
import { motion, HTMLMotionProps } from 'framer-motion';
import { useMediaUpload } from '../../hooks/useMediaUpload';
import { Button } from '../ui/Button';

interface ImageUploadProps {
  worldId: string;
  onUploadComplete: (mediaId: string, mediaUrl: string) => void;
  className?: string;
}

const UploadContainer = styled.div`
  width: 100%;
  margin-bottom: var(--space-4);
`;

const DropZone = styled.div<{ $isDragging: boolean; $hasPreview: boolean }>`
  width: 100%;
  min-height: 140px;
  border: 2px dashed ${props => props.$isDragging ? 'var(--color-primary)' : 'var(--color-border)'};
  background-color: ${props => props.$isDragging ? 'rgba(255, 199, 95, 0.08)' : 'var(--color-input-bg)'};
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  cursor: pointer;
  transition: all var(--duration-fast) ease;
  position: relative;
  overflow: hidden;
  
  &:hover {
    border-color: var(--color-primary);
    background-color: rgba(255, 199, 95, 0.08);
  }
  
  ${props => props.$hasPreview && `
    border-style: solid;
    background-color: var(--color-card);
  `}
`;

const UploadIcon = styled.div`
  font-size: 32px;
  color: var(--color-text-light);
  margin-bottom: var(--space-2);
`;

const UploadText = styled.div`
  font-size: var(--font-sm);
  color: var(--color-text-light);
  text-align: center;
`;

const FileInput = styled.input`
  display: none;
`;

// Define the type for the motion.img component
type PreviewImageProps = HTMLMotionProps<'img'>;

const PreviewImage = styled(motion.img)<PreviewImageProps>`
  max-width: 100%;
  max-height: 300px;
  object-fit: contain;
  border-radius: var(--radius-md);
`;

const ProgressBar = styled.div<{ $progress: number }>`
  width: 100%;
  height: 4px;
  background-color: var(--color-border);
  border-radius: var(--radius-full);
  margin-top: var(--space-4);
  overflow: hidden;
  
  &::after {
    content: '';
    display: block;
    height: 100%;
    width: ${props => props.$progress}%;
    background-color: var(--color-primary);
    transition: width 0.3s ease;
  }
`;

const ButtonContainer = styled.div`
  display: flex;
  gap: var(--space-2);
  margin-top: var(--space-4);
`;

const ErrorMessage = styled.div`
  color: var(--color-accent);
  font-size: var(--font-sm);
  margin-top: var(--space-2);
`;

export const ImageUpload: React.FC<ImageUploadProps> = ({
  worldId,
  onUploadComplete,
  className,
}) => {
  const [isDragging, setIsDragging] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploaded, setIsUploaded] = useState(false);
  
  const {
    isUploading,
    progress,
    error,
    uploadMedia,
    clearError
  } = useMediaUpload({ worldId });
  
  const handleDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);
  
  const handleDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);
  
  const handleDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  }, []);
  
  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFileSelect(e.target.files[0]);
    }
  };
  
  const handleFileSelect = (file: File) => {
    // Check if it's an image
    if (!file.type.match('image.*')) {
      alert('Please select an image file');
      return;
    }
    
    setSelectedFile(file);
    setIsUploaded(false);
    
    // Create preview URL
    const reader = new FileReader();
    reader.onload = (e) => {
      if (e.target?.result) {
        setPreviewUrl(e.target.result as string);
      }
    };
    reader.readAsDataURL(file);
    
    clearError();
  };
  
  const handleUpload = async () => {
    if (!selectedFile) return;
    
    const result = await uploadMedia(selectedFile);
    
    if (result && result.variants) {
      // Find the first available variant or original
      const mediaUrl = result.variants.original || Object.values(result.variants)[0];
      onUploadComplete(result.id, mediaUrl);
      setIsUploaded(true);
    }
  };
  
  const handleRemove = () => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setIsUploaded(false);
    onUploadComplete('', ''); // Clear any previously uploaded media
    
    // Refocus the file input to encourage selecting a new image
    setTimeout(() => {
      if (inputRef.current) {
        inputRef.current.click();
      }
    }, 100);
  };
  
  const inputRef = React.useRef<HTMLInputElement>(null);
  
  // Auto-upload when file is selected
  useEffect(() => {
    if (selectedFile && !isUploaded && !isUploading) {
      handleUpload();
    }
  }, [selectedFile, isUploaded, isUploading]);
  
  return (
    <UploadContainer className={className}>
      <DropZone
        $isDragging={isDragging}
        $hasPreview={!!previewUrl}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
      >
        {previewUrl ? (
          <>
            <PreviewImage 
              src={previewUrl}
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.3 }}
            />
            {isUploaded && (
              <div style={{ 
                position: 'absolute', 
                top: '10px', 
                right: '10px', 
                backgroundColor: 'var(--color-success)', 
                color: 'white',
                padding: '2px 8px',
                borderRadius: 'var(--radius-full)',
                fontSize: 'var(--font-xs)',
                opacity: 0.9
              }}>
                ✓ Ready to post
              </div>
            )}
          </>
        ) : (
          <>
            <UploadIcon>📸</UploadIcon>
            <UploadText>
              Drag & drop an image here, <br /> or click to select a file
            </UploadText>
          </>
        )}
        <FileInput
          type="file"
          accept="image/*"
          ref={inputRef}
          onChange={handleFileInput}
        />
      </DropZone>
      
      {isUploading && (
        <ProgressBar $progress={progress} />
      )}
      
      {error && (
        <ErrorMessage>{error}</ErrorMessage>
      )}
      
      {previewUrl && !isUploading && (
        <ButtonContainer>
          <Button 
            variant="ghost" 
            onClick={handleRemove}
            fullWidth
          >
            Remove Image
          </Button>
        </ButtonContainer>
      )}
    </UploadContainer>
  );
};