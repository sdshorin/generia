import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Layout } from "../../components/layout/Layout";
import { useWorld } from "../../hooks/useWorld";
import { World } from "../../types";
import "../../styles/pages/world-about.css";

// Mock data for world details (since not available in API)
const mockWorldDetails = {
  description: "",
  subDescription: "",
  history: [
    {
      title: "",
      content: "",
    },
    {
      title: "",
      content: "",
    },
    {
      title: "",
      content: "",
    },
  ],
  characteristics: {
    technology: [
      { name: "", value: 0, level: "" },
      // { name: "Energy Systems", value: 90, level: "Hybrid" }
    ],
    magic: [
      { name: "", value: 0, level: "" },
      // { name: "Elemental Control", value: 80, level: "Expert" }
    ],
    social: [
      "",
      // "⚖️ Guild-based Economy",
      // "🤝 Collaborative Governance",
      // "📚 Knowledge Sharing Culture"
    ],
    geography: [
      "",
      // "🌉 Cloud Bridge Networks",
      // "🏰 Multi-Level Architecture",
      // "🌪️ Weather Control Systems"
    ],
  },
  featuredCharacters: [
    // {
    //   id: 'character-1',
    //   name: 'Lyra Starweaver',
    //   role: 'Arcane Engineer',
    //   avatar: '/no-image.jpg',
    //   posts: 23,
    //   likes: 456
    // },
    // {
    //   id: 'character-2',
    //   name: 'Zephyr Cloudwright',
    //   role: 'Sky Merchant',
    //   avatar: '/no-image.jpg',
    //   posts: 18,
    //   likes: 287
    // },
    // {
    //   id: 'character-3',
    //   name: 'Mistral Aethermage',
    //   role: 'Wind Sorceress',
    //   avatar: '/no-image.jpg',
    //   posts: 31,
    //   likes: 612
    // },
    // {
    //   id: 'character-4',
    //   name: 'Thorne Mechanist',
    //   role: 'Gear Engineer',
    //   avatar: '/no-image.jpg',
    //   posts: 15,
    //   likes: 203
    // },
    // {
    //   id: 'character-5',
    //   name: 'Sage Elderwood',
    //   role: 'Ancient Scholar',
    //   avatar: '/no-image.jpg',
    //   posts: 27,
    //   likes: 523
    // }
  ],
};

export const WorldAboutPage: React.FC = () => {
  const { worldId } = useParams<{ worldId: string }>();
  const navigate = useNavigate();
  const { currentWorld, loadCurrentWorld } = useWorld();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchWorld = async () => {
      if (worldId && (!currentWorld || currentWorld.id !== worldId)) {
        try {
          await loadCurrentWorld(worldId);
        } catch (error) {
          console.error("Error loading world:", error);
        }
      }
      setLoading(false);
    };

    fetchWorld();
  }, [worldId, currentWorld, loadCurrentWorld]);

  const handleEnterWorld = () => {
    if (worldId) {
      navigate(`/worlds/${worldId}/feed`);
    }
  };

  const handleCharacterClick = (characterId: string) => {
    navigate(`/characters/${characterId}`);
  };

  const handleViewAllCharacters = () => {
    // TODO: Navigate to characters list page when implemented
    console.log("View all characters for world:", worldId);
  };

  if (loading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="loader"></div>
        </div>
      </Layout>
    );
  }

  if (!currentWorld) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="text-center">
            <h2 className="text-xl font-bold text-gray-800 mb-2">
              World not found
            </h2>
            <p className="text-gray-600 mb-4">
              The world you're looking for doesn't exist.
            </p>
            <button
              onClick={() => navigate("/worlds")}
              className="btn btn-primary"
            >
              Browse Worlds
            </button>
          </div>
        </div>
      </Layout>
    );
  }

  const worldParams = currentWorld.params
    ? JSON.parse(currentWorld.params)
    : {};

  // Function to truncate long text
  const truncateText = (text: string, maxLength: number = 50) => {
    if (!text) return "";
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + "...";
  };

  const worldStats = [
    { label: "Active Characters", value: currentWorld.users_count },
    { label: "Total Posts", value: currentWorld.posts_count },
    { label: "Theme", value: truncateText(worldParams.theme || "", 50) },
  ];

  return (
    <Layout>
      <div className="min-h-screen flex flex-col bg-white">
        {/* World Cover Section */}
        <div
          className="world-cover"
          style={{
            backgroundImage: `linear-gradient(rgba(0, 0, 0, 0.3) 0%, rgba(0, 0, 0, 0.6) 100%), url('${currentWorld.image_url || "/no-image.jpg"}')`,
          }}
        >
          <div className="world-cover-overlay">
            <div className="world-cover-content">
              <div className="world-cover-info">
                {/* World Icon */}
                <div
                  className="world-icon"
                  style={{
                    backgroundImage: `url('${currentWorld.icon_url || "/no-image.jpg"}')`,
                  }}
                ></div>
                <div className="world-details">
                  <h1 className="world-title">{currentWorld.name}</h1>
                  <div className="world-stats">
                    <span>✨ {currentWorld.users_count} Characters</span>
                    <span>📸 {currentWorld.posts_count} Posts</span>
                    <span>❤️ 2.3K Likes</span>
                  </div>
                </div>
              </div>

              {/* Enter World Button */}
              <button
                onClick={handleEnterWorld}
                className="btn btn-primary btn-lg enter-world-btn"
              >
                Enter World
              </button>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="world-about-content">
          <div className="world-about-container">
            <div className="world-about-grid">
              {/* Left Column: World Details */}
              <div className="space-y-8">
                {/* World Description */}
                <div className="world-section">
                  <h2 className="world-section-title">About This World</h2>
                  <p className="world-section-content">
                    {worldParams.description || currentWorld.description}
                  </p>
                  {worldParams.description_short && (
                    <p className="world-section-subcontent">
                      {worldParams.description_short}
                    </p>
                  )}
                </div>

                {/* World History */}
                <div className="world-section">
                  <h2 className="world-section-title">History & Origins</h2>
                  <p className="world-section-content">{worldParams.history}</p>
                </div>

                {/* Theme Details */}
                {worldParams.theme && (
                  <div className="world-section">
                    <h2 className="world-section-title">Theme</h2>
                    <p className="world-section-content">{worldParams.theme}</p>
                  </div>
                )}

                {/* Culture & Society */}
                {worldParams.culture && (
                  <div className="world-section">
                    <h2 className="world-section-title">Culture & Society</h2>
                    <p className="world-section-content">{worldParams.culture}</p>
                  </div>
                )}

                {/* Visual Style */}
                {worldParams.visual_style && (
                  <div className="world-section">
                    <h2 className="world-section-title">Visual Style</h2>
                    <p className="world-section-content">{worldParams.visual_style}</p>
                  </div>
                )}

                {/* World Characteristics */}
                <div className="world-section">
                  <h2 className="world-section-title">World Characteristics</h2>
                  <div className="characteristics-grid">
                    {/* Technology Level */}
                    <div className="characteristic-item">
                      <h3 className="characteristic-title">Technology Level</h3>
                      <p className="characteristic-list-item">
                        {worldParams.technology_level}
                      </p>
                    </div>

                    <div className="characteristic-item">
                      <h3 className="characteristic-title">Magic System</h3>
                      <p className="characteristic-list-item">
                        {worldParams.additional_details?.magic_system}
                      </p>
                    </div>

                    <div className="characteristic-item">
                      <h3 className="characteristic-title">Social Structure</h3>
                      <p className="characteristic-list-item">
                        {worldParams.social_structure}
                      </p>
                    </div>

                    <div className="characteristic-item">
                      <h3 className="characteristic-title">Geography</h3>
                      <p className="characteristic-list-item">
                        {worldParams.geography}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Custom Details */}
                {worldParams.additional_details?.custom_details && 
                 worldParams.additional_details.custom_details.length > 0 && (
                  <div className="world-section">
                    <h2 className="world-section-title">Unique Features</h2>
                    <div className="custom-details-list">
                      {worldParams.additional_details.custom_details.map((detail: string, index: number) => (
                        <div key={index} className="custom-detail-item">
                          <span className="custom-detail-icon">✨</span>
                          <p className="custom-detail-text">{detail}</p>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Right Column: Characters & Stats */}
              <div className="space-y-8">
                {/* Quick Stats */}
                <div className="world-section">
                  <h2 className="world-section-title">World Statistics</h2>
                  <div className="stats-list">
                    {worldStats.map((stat, index) => (
                      <div key={index} className="stats-item">
                        <span className="stats-label">{stat.label}</span>
                        <span className="stats-value">{stat.value}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Featured Characters */}
                <div className="world-section">
                  <h2 className="world-section-title">Featured Characters</h2>
                  <div className="characters-container">
                    <div className="characters-list">
                      {/* {mockWorldDetails.featuredCharacters.map((character) => (
                        <div 
                          key={character.id}
                          className="character-item"
                          onClick={() => handleCharacterClick(character.id)}
                        >
                          <div 
                            className="character-item-avatar" 
                            style={{ backgroundImage: `url('${character.avatar}')` }}
                          ></div>
                          <div className="character-item-info">
                            <p className="character-item-name">{character.name}</p>
                            <p className="character-item-role">{character.role}</p>
                            <p className="character-item-stats">
                              {character.posts} posts • {character.likes} likes
                            </p>
                          </div>
                        </div>
                      ))} */}
                    </div>

                    {/* View All Characters Button */}
                    <button
                      className="btn btn-secondary view-all-btn"
                      onClick={handleViewAllCharacters}
                    >
                      View All {currentWorld.users_count} Characters
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};
