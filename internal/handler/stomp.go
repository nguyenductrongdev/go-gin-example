package handler

import (
	"go-gin-example/internal/hub"
	"go-gin-example/internal/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func StompHandler(c *gin.Context) {
	userID := c.GetHeader("UserID")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WS upgrade:", err)
		return
	}

	client := &hub.Client{
		ID:     time.Now().Format("150405.000000"),
		UserID: userID,
		RoomID: "", // TODO: empty for personal, need to support for group
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h := hub.Get()
	h.Register <- client
	go client.WritePump()
	go client.ReadPump(h)
}

func SendStompPrivateHandler(c *gin.Context) {
	msg := models.Message{
		RecipientID: "536080c8-3f5e-4471-b8ae-6ed2085f7649",
		Content:     "hello world",
	}
	hub.Get().Broadcast <- &msg
	c.JSON(http.StatusOK, msg)
}
