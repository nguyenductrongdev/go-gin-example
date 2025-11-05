package server

import "github.com/gofrs/uuid"

const jwtSecret = "super-secret-123"

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
}
