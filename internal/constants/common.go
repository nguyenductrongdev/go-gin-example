package constants

import "github.com/gofrs/uuid"

const JwtSecret = "IAMSOMEONE"

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
}
