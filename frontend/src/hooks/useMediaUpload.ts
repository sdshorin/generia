import { useState, useCallback, useRef } from 'react';
import { mediaAPI } from '../api/services';
import { Media } from '../types';

interface UseMediaUploadOptions {
  worldId: string;
  characterId: string;
}

export const useMediaUpload = ({ worldId, characterId }: UseMediaUploadOptions) => {
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [media, setMedia] = useState<Media | null>(null);

  const uploadInProgress = useRef(false);
  
  const uploadMedia = useCallback(
    async (file: File): Promise<Media | null> => {
      if (!file || uploadInProgress.current) return null;
      
      uploadInProgress.current = true;
      setIsUploading(true);
      setProgress(0);
      setError(null);
      setMedia(null);
      
      try {
        // Get pre-signed upload URL
        const { media_id, upload_url } = await mediaAPI.getUploadUrl(
          file.name,
          file.type,
          file.size,
          worldId,
          characterId
        );
        
        // Upload file directly to S3/MinIO
        await mediaAPI.uploadToUrl(upload_url, file);
        setProgress(80);
        
        // Confirm upload completion
        const mediaData = await mediaAPI.confirmUpload(media_id, characterId);
        setMedia(mediaData);
        setProgress(100);
        
        return mediaData;
      } catch (err: any) {
        setError(err.message || 'Failed to upload media');
        console.error('Media upload error:', err);
        return null;
      } finally {
        setIsUploading(false);
        uploadInProgress.current = false;
      }
    },
    [worldId, characterId]
  );

  // Upload media as base64 (alternative method)
  const uploadBase64 = useCallback(
    async (base64Data: string, filename: string, mimeType: string): Promise<Media | null> => {
      if (!base64Data || uploadInProgress.current) return null;
      
      uploadInProgress.current = true;
      setIsUploading(true);
      setProgress(0);
      setError(null);
      setMedia(null);
      
      try {
        // Strip data URL prefix if present
        const mediaData = base64Data.includes('base64,')
          ? base64Data.split('base64,')[1]
          : base64Data;
        
        // Upload media
        const response = await mediaAPI.uploadBase64(
          mediaData,
          mimeType,
          filename,
          worldId,
          characterId
        );
        
        setMedia(response);
        setProgress(100);
        return response;
      } catch (err: any) {
        setError(err.message || 'Failed to upload media');
        console.error('Media upload error:', err);
        return null;
      } finally {
        setIsUploading(false);
        uploadInProgress.current = false;
      }
    },
    [worldId, characterId]
  );

  return {
    isUploading,
    progress,
    error,
    media,
    uploadMedia,
    uploadBase64,
    clearError: () => setError(null),
    clearMedia: () => setMedia(null),
  };
};