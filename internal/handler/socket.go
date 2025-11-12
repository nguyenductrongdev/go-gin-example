package handler

import (
	"encoding/json"
	"go-gin-example/internal/hub"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // dev only
}

// Handle WebSocket
func WsHandler(c *gin.Context) {

	// Authenticated userId will be injected by gateway
	ctxUserId := c.GetString("user_id")

	// 4. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 5. Tạo client
	client := &hub.Client{Conn: conn, Send: make(chan []byte, 256)}
	hub.Get().Register <- client

	// Gửi welcome
	welcome := map[string]string{
		"type":    "welcome",
		"message": "Connected as " + ctxUserId,
	}
	sendJSON(client, welcome)
}

// Gửi JSON
func sendJSON(client *hub.Client, v interface{}) {
	data, _ := json.Marshal(v)
	client.Send <- data
}
