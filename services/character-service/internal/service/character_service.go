package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	pb "github.com/sdshorin/generia/api/grpc/character"
	"github.com/sdshorin/generia/pkg/logger"
	"github.com/sdshorin/generia/services/character-service/internal/models"
	"github.com/sdshorin/generia/services/character-service/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type CharacterService struct {
	pb.UnimplementedCharacterServiceServer
	repo repository.CharacterRepository
}

func NewCharacterService(repo repository.CharacterRepository) *CharacterService {
	return &CharacterService{
		repo: repo,
	}
}

func (s *CharacterService) CreateCharacter(ctx context.Context, req *pb.CreateCharacterRequest) (*pb.Character, error) {
	logger.Logger.Info("Creating character",
		zap.String("world_id", req.WorldId),
		zap.String("display_name", req.DisplayName))

	var meta json.RawMessage = []byte("{}")
	if req.Meta != nil && *req.Meta != "" {
		meta = json.RawMessage(*req.Meta)
	}

	var realUserID sql.NullString
	if req.RealUserId != nil && *req.RealUserId != "" {
		realUserID = sql.NullString{String: *req.RealUserId, Valid: true}
	}

	var avatarMediaID sql.NullString
	if req.AvatarMediaId != nil && *req.AvatarMediaId != "" {
		avatarMediaID = sql.NullString{String: *req.AvatarMediaId, Valid: true}
	}

	params := models.CreateCharacterParams{
		WorldID:       req.WorldId,
		RealUserID:    realUserID,
		DisplayName:   req.DisplayName,
		AvatarMediaID: avatarMediaID,
		Meta:          meta,
	}

	character, err := s.repo.CreateCharacter(ctx, params)
	if err != nil {
		logger.Logger.Error("Failed to create character", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create character")
	}

	return characterModelToProto(character), nil
}

func (s *CharacterService) GetCharacter(ctx context.Context, req *pb.GetCharacterRequest) (*pb.Character, error) {
	logger.Logger.Info("Getting character", zap.String("id", req.CharacterId))

	character, err := s.repo.GetCharacter(ctx, req.CharacterId)
	if err != nil {
		logger.Logger.Error("Failed to get character", zap.Error(err))
		return nil, status.Error(codes.NotFound, "Character not found")
	}

	return characterModelToProto(character), nil
}

func (s *CharacterService) GetUserCharactersInWorld(ctx context.Context, req *pb.GetUserCharactersInWorldRequest) (*pb.CharacterList, error) {
	logger.Logger.Info("Getting user characters in world",
		zap.String("user_id", req.UserId),
		zap.String("world_id", req.WorldId))

	characters, err := s.repo.GetUserCharactersInWorld(ctx, req.UserId, req.WorldId)
	if err != nil {
		logger.Logger.Error("Failed to get user characters in world", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get characters")
	}

	protoCharacters := make([]*pb.Character, 0, len(characters))
	for _, character := range characters {
		protoCharacters = append(protoCharacters, characterModelToProto(character))
	}

	return &pb.CharacterList{Characters: protoCharacters}, nil
}

func (s *CharacterService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *CharacterService) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func (s *CharacterService) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: "SERVING"}, nil
}

// Helper functions
func characterModelToProto(character *models.Character) *pb.Character {
	protoCharacter := &pb.Character{
		Id:          character.ID,
		WorldId:     character.WorldID,
		IsAi:        character.IsAI,
		DisplayName: character.DisplayName,
		CreatedAt:   character.CreatedAt.Format(time.RFC3339),
	}

	if character.RealUserID.Valid {
		protoCharacter.RealUserId = &character.RealUserID.String
	}

	if character.AvatarMediaID.Valid {
		protoCharacter.AvatarMediaId = &character.AvatarMediaID.String
	}

	if len(character.Meta) > 0 {
		metaStr := string(character.Meta)
		protoCharacter.Meta = &metaStr
	}

	return protoCharacter
}
