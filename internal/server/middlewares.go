package server

import (
	"fmt"
	"go-gin-example/internal/constants"
	"go-gin-example/internal/helper"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != "" {
			tokenParts := strings.Split(token, " ")

			claims, err := helper.ExtractJwtClaim[constants.Claims](tokenParts[1], constants.JwtSecret)
			if err != nil {
				fmt.Printf("Validate token error %v", err)
			}

			c.Set("user_id", claims.UserID.String())
		}

		c.Next()
	}
}
