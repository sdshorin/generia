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

// UserGenerator generates AI users based on a world prompt
type UserGenerator struct {
	namesPrefixes []string
	namesSuffixes []string
}

// NewUserGenerator creates a new UserGenerator
func NewUserGenerator() *UserGenerator {
	return &UserGenerator{
		namesPrefixes: []string{
			"cyber", "neon", "flux", "pixel", "synth", "vapor", "retro", "future", "space", "star", 
			"astro", "cosmo", "lunar", "solar", "galaxy", "nebula", "orbit", "quantum", "atom", "data",
			"glitch", "bit", "byte", "tech", "digital", "virtual", "electric", "spark", "surge", "wave",
			"echo", "pulse", "signal", "code", "algo", "crypto", "nexus", "matrix", "vector", "grid",
		},
		namesSuffixes: []string{
			"runner", "racer", "rider", "walker", "drifter", "hunter", "seeker", "finder", "watcher", "gazer",
			"dreamer", "thinker", "maker", "builder", "crafter", "weaver", "smith", "wright", "mind", "soul",
			"heart", "spirit", "ghost", "phantom", "shadow", "light", "flame", "spark", "glow", "shine",
			"wave", "storm", "cloud", "rain", "wind", "breeze", "quake", "shock", "blast", "burst",
		},
	}
}

// GenerateUsername creates a username based on the world theme
func (g *UserGenerator) GenerateUsername(worldPrompt string) string {
	// Seed random with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Generate a random combined username from prefixes and suffixes
	prefix := g.namesPrefixes[r.Intn(len(g.namesPrefixes))]
	suffix := g.namesSuffixes[r.Intn(len(g.namesSuffixes))]
	
	// Add a random number sometimes
	if r.Intn(10) > 6 {
		return fmt.Sprintf("%s_%s%d", prefix, suffix, r.Intn(999))
	}
	
	return fmt.Sprintf("%s_%s", prefix, suffix)
}

// GenerateUser generates a complete AI user based on the world prompt
func (g *UserGenerator) GenerateUser(ctx context.Context, worldID, worldPrompt string) (string, string, string, error) {
	// In a real implementation, this would use an LLM to generate users based on the prompt
	// For now, we'll just use a simple template-based approach
	
	// Generate ID
	userID := uuid.New().String()
	
	// Generate username
	username := g.GenerateUsername(worldPrompt)
	
	// In a real implementation, you would pass the world prompt to an LLM to generate a user description
	// For now, we'll just use a very basic template
	promptTokens := strings.Fields(worldPrompt)
	var description string
	if len(promptTokens) > 10 {
		// Take random phrases from the prompt
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		startIdx := r.Intn(len(promptTokens) - 10)
		words := promptTokens[startIdx : startIdx+10]
		description = "A user from the world of " + strings.Join(words, " ")
	} else {
		description = "A user from " + worldPrompt
	}
	
	logger.Logger.Info("Generated AI user", 
		zap.String("user_id", userID), 
		zap.String("username", username),
		zap.String("world_id", worldID))
		
	return userID, username, description, nil
}