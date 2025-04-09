import React, { useState, useContext, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axiosInstance from '../api/axios';
import { AuthContext } from '../context/AuthContext';
import axios from 'axios';

// Интерфейсы для типизации
interface UploadUrlResponse {
  media_id: string;
  upload_url: string;
  expires_at: number;
}

interface ConfirmUploadResponse {
  media_id: string;
  variants: Record<string, string>;
}

const CreatePost: React.FC = () => {
  const [caption, setCaption] = useState('');
  const [image, setImage] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [uploadMethod, setUploadMethod] = useState<'direct' | 'base64'>('direct');

  const { isAuthenticated } = useContext(AuthContext);
  const navigate = useNavigate();

  // Проверяем аутентификацию при монтировании
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  // Если пользователь не аутентифицирован, не рендерим компонент
  if (!isAuthenticated) {
    return null;
  }

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setImage(file);

      // Создаем URL для предпросмотра
      const fileReader = new FileReader();
      fileReader.onload = () => {
        setPreviewUrl(fileReader.result as string);
      };
      fileReader.readAsDataURL(file);
    }
  };

  // Функция для Direct Upload в S3/MinIO
  const handleDirectUpload = async () => {
    if (!image) {
      setError('Please select an image');
      return false;
    }

    try {
      // 1. Получаем предподписанный URL для загрузки
      const getUrlResponse = await axiosInstance.post<UploadUrlResponse>('/media/upload-url', {
        filename: image.name,
        content_type: image.type,
        size: image.size
      });

      const { media_id, upload_url } = getUrlResponse.data;

      // 2. Загружаем файл напрямую в S3/MinIO используя предподписанный URL
      await axios.put(upload_url, image, {
        headers: {
          'Content-Type': image.type
        },
        onUploadProgress: (progressEvent) => {
          const progress = Math.round((progressEvent.loaded * 100) / (progressEvent.total || image.size));
          setUploadProgress(progress);
        }
      });

      // 3. Подтверждаем загрузку на сервере
      const confirmResponse = await axiosInstance.post<ConfirmUploadResponse>('/media/confirm', {
        media_id
      });

      // 4. Создаем пост с ID загруженного медиа
      await axiosInstance.post('/posts', {
        caption,
        media_id
      });

      return true;
    } catch (err: any) {
      console.error('Direct upload error:', err);
      setError(err.response?.data || 'Failed to upload media');
      return false;
    }
  };

  // Функция для загрузки через Base64 (legacy)
  const handleBase64Upload = async () => {
    if (!image) {
      setError('Please select an image');
      return false;
    }

    try {
      // Подготавливаем изображение в Base64
      const reader = new FileReader();
      
      return new Promise<boolean>((resolve) => {
        reader.onload = async () => {
          try {
            const base64Image = reader.result as string;

            // Отправляем запрос
            await axiosInstance.post('/media', {
              media_data: base64Image,
              content_type: image.type,
              filename: image.name
            });

            resolve(true);
          } catch (err: any) {
            console.error('Base64 upload error:', err);
            setError(err.response?.data || 'Failed to upload media');
            resolve(false);
          }
        };
        
        reader.onerror = () => {
          setError('Failed to read file');
          resolve(false);
        };
        
        reader.readAsDataURL(image);
      });
    } catch (err: any) {
      console.error('Base64 processing error:', err);
      setError(err.response?.data || 'Failed to process image');
      return false;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);
    setUploadProgress(0);

    try {
      let success = false;

      // Используем выбранный метод загрузки
      if (uploadMethod === 'direct') {
        success = await handleDirectUpload();
      } else {
        success = await handleBase64Upload();
      }

      if (success) {
        navigate('/');
      } else {
        setIsLoading(false);
      }
    } catch (err: any) {
      console.error('Submit error:', err);
      setError(err.response?.data || 'Failed to create post');
      setIsLoading(false);
    }
  };

  return (
    <div className="create-post-container">
      <h2>Create Post</h2>
      {error && <div className="error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="caption">Caption</label>
          <textarea
            id="caption"
            value={caption}
            onChange={(e) => setCaption(e.target.value)}
            placeholder="Write a caption..."
            rows={4}
          />
        </div>
        <div className="form-group">
          <label htmlFor="image">Image</label>
          <input
            type="file"
            id="image"
            accept="image/*"
            onChange={handleImageChange}
            required
          />
        </div>
        {previewUrl && (
          <div className="image-preview">
            <img src={previewUrl} alt="Preview" />
          </div>
        )}
        
        <div className="form-group">
          <label>Upload Method</label>
          <div>
            <label>
              <input
                type="radio"
                name="uploadMethod"
                checked={uploadMethod === 'direct'}
                onChange={() => setUploadMethod('direct')}
              />
              Direct Upload (S3/MinIO)
            </label>
            <label>
              <input
                type="radio"
                name="uploadMethod"
                checked={uploadMethod === 'base64'}
                onChange={() => setUploadMethod('base64')}
              />
              Base64 Upload
            </label>
          </div>
        </div>
        
        {isLoading && uploadMethod === 'direct' && (
          <div className="progress-bar">
            <div className="progress" style={{ width: `${uploadProgress}%` }}></div>
            <span>{uploadProgress}%</span>
          </div>
        )}
        
        <button type="submit" disabled={isLoading || !image}>
          {isLoading ? 'Creating...' : 'Create Post'}
        </button>
      </form>
    </div>
  );
};

export default CreatePost;