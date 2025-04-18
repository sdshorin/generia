import React, { useState, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import axiosInstance from '../api/axios';
import { AuthContext } from '../context/AuthContext';

const CreateWorld: React.FC = () => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [prompt, setPrompt] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { isAuthenticated } = useContext(AuthContext);

  // Redirect if not authenticated
  React.useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!name || !prompt) {
      setError('Name and prompt are required');
      return;
    }

    try {
      setLoading(true);
      setError('');

      const response = await axiosInstance.post('/worlds', {
        name,
        description,
        prompt,
      });

      if (response.data && response.data.id) {
        navigate('/feed'); // Navigate to feed which will show the new world
      } else {
        setError('Failed to create world');
      }
    } catch (err) {
      console.error('Error creating world:', err);
      setError('Failed to create world. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="create-world-container">
      <h2>Create a New World</h2>
      
      <form onSubmit={handleSubmit} className="create-world-form">
        <div className="form-group">
          <label htmlFor="name">World Name</label>
          <input
            type="text"
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter a name for your world"
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="description">Description (optional)</label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe your world"
            rows={3}
          />
        </div>

        <div className="form-group">
          <label htmlFor="prompt">World Prompt</label>
          <textarea
            id="prompt"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Describe the theme, setting, and details of your world. This will be used to generate AI content."
            rows={6}
            required
          />
          <small className="form-helper">
            Be specific and detailed. Example: "A cyberpunk city in the year 2100 where corporations rule and technology has merged with humanity. The streets are neon-lit, rainy, and filled with augmented humans."
          </small>
        </div>

        {error && <div className="error">{error}</div>}

        <button 
          type="submit" 
          className="create-button"
          disabled={loading}
        >
          {loading ? 'Creating...' : 'Create World'}
        </button>
        
        <button 
          type="button" 
          className="cancel-button"
          onClick={() => navigate('/worlds')}
          disabled={loading}
        >
          Cancel
        </button>
      </form>
    </div>
  );
};

export default CreateWorld;