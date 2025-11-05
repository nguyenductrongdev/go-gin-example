package server

import (
	"net/http"
	"time"

	_ "my_project/docs"
	"my_project/internal/helper"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(Authenticated())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.POST("/signin", s.SignInHandler)

	r.GET("/me", s.WhoamiHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// SignInHandler godoc
// @Summary      Sign in
func (s *Server) SignInHandler(c *gin.Context) {
	userId, _ := uuid.NewV4()
	jwtBody := Claims{UserID: userId}
	if token, err := helper.SignJwt(jwtBody, jwtSecret, 5*time.Minute); err == nil {
		resp := make(map[string]string)
		resp["access_token"] = token

		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusInternalServerError, err)
	}

}

// SignInHandler godoc
// @Summary      Get self user info
func (s *Server) WhoamiHandler(c *gin.Context) {
	userId := c.MustGet("user_id").(string)

	resp := make(map[string]string)
	resp["user_id"] = userId

	c.JSON(http.StatusOK, resp)
}
