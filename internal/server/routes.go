package server

import (
	"net/http"
	"time"

	_ "go-gin-example/docs"
	"go-gin-example/internal/constants"
	"go-gin-example/internal/handler"
	"go-gin-example/internal/helper"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ──────────────────────────────────────────────────────────────
// GENERAL API INFO (REQUIRED)
// ──────────────────────────────────────────────────────────────
// @title           My Project API
// @version         1.0
// @description     Simple JWT auth demo
// @host            localhost:8000
// @BasePath       /

// ──────────────────────────────────────────────────────────────
// JWT AUTH DEFINITION (PUT THIS ANYWHERE IN THE FILE)
// ──────────────────────────────────────────────────────────────
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <your-jwt-token>" in the field

// ──────────────────────────────────────────────────────────────
// YOUR HANDLERS
// ──────────────────────────────────────────────────────────────
func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(Authenticated())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.POST("/ws-chat/signin", s.SignInHandler)

	r.GET("/ws-chat/me", s.WhoamiHandler)

	r.GET("/ws-chat/ws", handler.WsHandler)

	r.GET("/ws-chat/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

// SignInHandler godoc
// @Summary      Sign in and get JWT
// @Description  Generates a JWT token with a random user ID (demo only)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "access_token"
// @Failure      500  {object}  map[string]string  "error"
// @Router       /signin [post]
func (s *Server) SignInHandler(c *gin.Context) {
	userId, _ := uuid.NewV4()
	jwtBody := constants.Claims{UserID: userId}
	if token, err := helper.SignJwt(jwtBody, constants.JwtSecret, 30*time.Minute); err == nil {
		resp := make(map[string]string)
		resp["access_token"] = token

		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusInternalServerError, err)
	}

}

// WhoamiHandler godoc
// @Summary      Get current user info
// @Description  Returns the user_id from JWT (requires Authenticated middleware)
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "user_id"
// @Failure      401  {object}  map[string]string  "unauthorized"
// @Router       /me [get]
func (s *Server) WhoamiHandler(c *gin.Context) {
	userId := c.MustGet("user_id").(string)

	resp := make(map[string]string)
	resp["user_id"] = userId

	c.JSON(http.StatusOK, resp)
}
