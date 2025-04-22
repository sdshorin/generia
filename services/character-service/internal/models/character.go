package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Character struct {
	ID            string
	WorldID       string
	RealUserID    sql.NullString // NULL for AI characters
	IsAI          bool
	DisplayName   string
	AvatarMediaID sql.NullString
	Meta          json.RawMessage
	CreatedAt     time.Time
}

type CreateCharacterParams struct {
	WorldID       string
	RealUserID    sql.NullString
	DisplayName   string
	AvatarMediaID sql.NullString
	Meta          json.RawMessage
}