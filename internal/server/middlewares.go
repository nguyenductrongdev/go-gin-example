package server

import (
	"fmt"
	"go-gin-example/internal/constants"
	"go-gin-example/internal/helper"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUserId := "00000000-0000-0000-0000-000000000000"

		gatewayWsUserId := c.GetHeader("X-Ws-User-Id")
		gatewayRestUserId := c.GetHeader("X-User-Id")

		log.Printf("--- X-Ws-User-Id: %v", gatewayWsUserId)
		log.Printf("--- X-User-Id: %v", gatewayRestUserId)

		if gatewayWsUserId != "" {
			log.Printf("Gateway ws authenticated user %v", gatewayWsUserId)
			currentUserId = gatewayWsUserId
		} else if gatewayRestUserId != "" {
			log.Printf("Gateway rest authenticated user %v", gatewayRestUserId)
			currentUserId = gatewayRestUserId
		} else {
			token := c.GetHeader("Authorization")

			if token != "" {
				tokenParts := strings.Split(token, " ")

				claims, err := helper.ExtractJwtClaim[constants.Claims](tokenParts[1], constants.JwtSecret)
				if err != nil {
					fmt.Printf("Validate token error %v", err)
				}

				currentUserId = claims.UserID.String()

			}
		}

		c.Set("user_id", currentUserId)

		c.Next()
	}
}
