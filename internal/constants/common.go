package constants

import "github.com/gofrs/uuid"

const JwtSecret = "super-secret-123"

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
}
