package server

import (
	"my_project/internal/helper"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token != "" {
			tokenParts := strings.Split(token, " ")

			claims, _ := helper.ExtractJwtClaim[Claims](tokenParts[1], jwtSecret)

			c.Set("user_id", claims.UserID.String())
		}

		c.Next()
	}
}
