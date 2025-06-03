import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import '../../styles/components/world-facts-carousel.css';

interface WorldFact {
  title: string;
  content: string;
}

interface WorldFactsCarouselProps {
  worldParams: any;
  className?: string;
}

export const WorldFactsCarousel: React.FC<WorldFactsCarouselProps> = ({
  worldParams,
  className = ''
}) => {
  const [currentFactIndex, setCurrentFactIndex] = useState(0);
  const [facts, setFacts] = useState<WorldFact[]>([]);

  // Extract facts from world parameters
  useEffect(() => {
    if (!worldParams) return;

    const extractedFacts: WorldFact[] = [];

    // Map of keys to user-friendly titles
    const factMappings: Record<string, string> = {
      theme: 'Тема мира',
      geography: 'География',
      culture: 'Культура',
      visual_style: 'Визуальный стиль',
      technology_level: 'Уровень технологий',
      social_structure: 'Социальная структура',
      history: 'История мира'
    };

    // Add main facts
    Object.entries(factMappings).forEach(([key, title]) => {
      if (worldParams[key] && typeof worldParams[key] === 'string') {
        extractedFacts.push({
          title,
          content: worldParams[key]
        });
      }
    });

    // Add additional details
    if (worldParams.additional_details) {
      const additionalMappings: Record<string, string> = {
        climate: 'Климат',
        resources: 'Ресурсы',
        conflicts: 'Конфликты',
        traditions: 'Традиции',
        technology: 'Технологии',
        magic_system: 'Система магии',
        time_period: 'Временной период',
        language: 'Язык'
      };

      Object.entries(additionalMappings).forEach(([key, title]) => {
        if (worldParams.additional_details[key] && typeof worldParams.additional_details[key] === 'string') {
          extractedFacts.push({
            title,
            content: worldParams.additional_details[key]
          });
        }
      });

      // Add custom details if they exist
      if (worldParams.additional_details.custom_details && Array.isArray(worldParams.additional_details.custom_details)) {
        worldParams.additional_details.custom_details.forEach((detail: string, index: number) => {
          extractedFacts.push({
            title: `Особенность ${index + 1}`,
            content: detail
          });
        });
      }
    }

    // Add common activities as facts
    if (worldParams.common_activities && Array.isArray(worldParams.common_activities)) {
      const activitiesText = worldParams.common_activities.slice(0, 5).join(', ');
      if (activitiesText) {
        extractedFacts.push({
          title: 'Популярные занятия',
          content: activitiesText
        });
      }
    }

    setFacts(extractedFacts);
  }, [worldParams]);

  // Auto-rotate facts every 5 seconds
  useEffect(() => {
    if (facts.length === 0) return;

    const interval = setInterval(() => {
      setCurrentFactIndex((prev) => (prev + 1) % facts.length);
    }, 5000);

    return () => clearInterval(interval);
  }, [facts.length]);

  if (facts.length === 0) return null;

  const currentFact = facts[currentFactIndex];

  return (
    <div className={`world-facts-carousel ${className}`}>
      <div className="facts-container">
        <AnimatePresence mode="wait">
          <motion.div
            key={currentFactIndex}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.5 }}
            className="fact-content"
          >
            <h3 className="fact-title">{currentFact.title}</h3>
            <p className="fact-text">{currentFact.content}</p>
          </motion.div>
        </AnimatePresence>
        
        {/* Fact indicators */}
        <div className="fact-indicators">
          {facts.map((_, index) => (
            <button
              key={index}
              className={`fact-indicator ${index === currentFactIndex ? 'active' : ''}`}
              onClick={() => setCurrentFactIndex(index)}
              aria-label={`Перейти к факту ${index + 1}`}
            />
          ))}
        </div>
      </div>
    </div>
  );
};