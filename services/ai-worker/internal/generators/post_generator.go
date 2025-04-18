package generators

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdshorin/generia/pkg/logger"
	"go.uber.org/zap"
)

// PostGenerator generates AI posts based on a world prompt
type PostGenerator struct {
	captionTemplates []string
	imagePromptTemplates []string
}

// NewPostGenerator creates a new PostGenerator
func NewPostGenerator() *PostGenerator {
	return &PostGenerator{
		captionTemplates: []string{
			"Just another day in %s",
			"Can't believe I'm seeing %s right now",
			"This is what %s looks like up close",
			"Amazing view of %s today",
			"Exploring %s with friends",
			"Found this hidden spot in %s",
			"The beauty of %s never ceases to amaze me",
			"First time visiting %s",
			"What do you think about %s?",
			"My favorite part of %s",
		},
		imagePromptTemplates: []string{
			"A photo of %s, detailed, realistic",
			"High quality image of %s, 4K, detailed",
			"Cinematic view of %s, dramatic lighting",
			"Wide angle shot of %s, photorealistic",
			"Detailed scene from %s, professional photography",
		},
	}
}

// GenerateCaption creates a post caption based on the world theme
func (g *PostGenerator) GenerateCaption(worldPrompt string, userID string) (string, error) {
	// Seed random with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Get a random template
	template := g.captionTemplates[r.Intn(len(g.captionTemplates))]
	
	// Extract a subject from the world prompt
	// In a real implementation, this would use NLP to extract relevant subjects
	promptTokens := strings.Fields(worldPrompt)
	var subject string
	
	if len(promptTokens) > 3 {
		startIdx := r.Intn(len(promptTokens) - 3)
		words := promptTokens[startIdx : startIdx+3]
		subject = strings.Join(words, " ")
	} else {
		subject = worldPrompt
	}
	
	// Generate caption
	caption := fmt.Sprintf(template, subject)
	
	// Sometimes add a few hashtags
	if r.Intn(10) > 5 {
		hashtags := []string{}
		for i := 0; i < r.Intn(4)+1; i++ {
			if len(promptTokens) > 0 {
				word := promptTokens[r.Intn(len(promptTokens))]
				if len(word) > 3 {
					hashtags = append(hashtags, "#"+strings.ToLower(word))
				}
			}
		}
		
		if len(hashtags) > 0 {
			caption += "\n\n" + strings.Join(hashtags, " ")
		}
	}
	
	return caption, nil
}

// GenerateImagePrompt creates an image generation prompt based on the world theme
func (g *PostGenerator) GenerateImagePrompt(worldPrompt string) string {
	// Seed random with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Get a random template
	template := g.imagePromptTemplates[r.Intn(len(g.imagePromptTemplates))]
	
	// In a real implementation, this would use an LLM to extract the most visual aspects of the world prompt
	// For now, we'll just use the prompt as is
	return fmt.Sprintf(template, worldPrompt)
}

// GeneratePost generates a complete AI post based on the world prompt
func (g *PostGenerator) GeneratePost(ctx context.Context, worldID, worldPrompt, userID string) (string, string, string, error) {
	// In a real implementation, this would use an LLM to generate posts based on the prompt
	// And would use an image generation model to create images based on the prompt
	
	// Generate post ID
	postID := uuid.New().String()
	
	// Generate caption
	caption, err := g.GenerateCaption(worldPrompt, userID)
	if err != nil {
		return "", "", "", err
	}
	
	// Generate image prompt for visualization
	imagePrompt := g.GenerateImagePrompt(worldPrompt)
	
	logger.Logger.Info("Generated AI post", 
		zap.String("post_id", postID), 
		zap.String("user_id", userID),
		zap.String("world_id", worldID))
		
	return postID, caption, imagePrompt, nil
}