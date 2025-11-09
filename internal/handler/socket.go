package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // dev only
}

// WebSocket Hub đơn giản
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)

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
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	clients[client] = true
	defer delete(clients, client)

	// Gửi welcome
	welcome := map[string]string{
		"type":    "welcome",
		"message": "Connected as " + ctxUserId,
	}
	sendJSON(client, welcome)

	// Read loop
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			broadcast <- msg // echo to all
		}
	}()

	// Write loop
	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

// Gửi JSON
func sendJSON(client *Client, v interface{}) {
	data, _ := json.Marshal(v)
	client.send <- data
}

// Broadcast worker
func HandleMessages() {
	for msg := range broadcast {
		for client := range clients {
			client.send <- msg
		}
	}
}
