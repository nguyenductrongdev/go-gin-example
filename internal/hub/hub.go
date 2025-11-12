package hub

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go-gin-example/internal/models" // adjust path

	"github.com/go-stomp/stomp/v3"
	"github.com/gorilla/websocket"
)

// ======================
// 1. WebSocket Client
// ======================

type Client struct {
	ID     string
	UserID string
	RoomID string
	Conn   *websocket.Conn
	Send   chan []byte
}

// ======================
// 2. Hub (Singleton)
// ======================

type Hub struct {
	clients map[string][]*Client // userID → []Client
	mu      sync.RWMutex

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *models.Message

	stompConn *stomp.Conn
	sub       *stomp.Subscription

	done      chan struct{}
	closeOnce sync.Once
}

var (
	globalHub *Hub
	once      sync.Once
)

func Get() *Hub {
	once.Do(func() {
		globalHub = &Hub{
			clients:    make(map[string][]*Client),
			Register:   make(chan *Client, 256),
			Unregister: make(chan *Client, 256),
			Broadcast:  make(chan *models.Message, 1024),
			done:       make(chan struct{}),
		}
		go globalHub.Run(context.Background())
	})
	return globalHub
}

// ======================
// 3. Run Loop
// ======================

func (h *Hub) Run(ctx context.Context) {
	log.Println("Hub started")
	for {
		select {
		case <-ctx.Done():
			h.cleanup()
			return
		case c := <-h.Register:
			h.registerClient(c)
		case c := <-h.Unregister:
			h.unregisterClient(c)
		case msg := <-h.Broadcast:
			h.broadcastMessage(msg)
		case <-h.done:
			h.cleanup()
			return
		}
	}
}

func (h *Hub) registerClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.UserID]; !ok {
		h.clients[c.UserID] = make([]*Client, 0)
	}
	h.clients[c.UserID] = append(h.clients[c.UserID], c)
	log.Printf("Registered: %s (UserID=%s)", c.ID, c.UserID)
}

func (h *Hub) unregisterClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if list, ok := h.clients[c.UserID]; ok {
		for i, cl := range list {
			if cl.ID == c.ID {
				h.clients[c.UserID] = append(list[:i], list[i+1:]...)
				break
			}
		}
		if len(h.clients[c.UserID]) == 0 {
			delete(h.clients, c.UserID)
		}
	}
	select {
	case <-c.Send:
	default:
		close(c.Send)
	}
	log.Printf("Unregistered: %s (UserID=%s)", c.ID, c.UserID)
}

func (h *Hub) broadcastMessage(msg *models.Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	recipients := make(map[string]bool)
	if msg.RecipientID != "" {
		recipients[msg.RecipientID] = true
	}
	// Add group logic here if needed

	data, _ := json.Marshal(msg)
	sent := 0
	for uid := range recipients {
		for _, c := range h.clients[uid] {
			select {
			case c.Send <- data:
				sent++
			default:
				log.Printf("Buffer full: dropping msg for %s", uid)
			}
		}
	}
	log.Printf("Sent to %d clients", sent)
}

// ======================
// 4. STOMP Integration
// ======================

func (h *Hub) InitSTOMP(broker, dest string, ack stomp.AckMode) error {
	conn, err := stomp.Dial("tcp", broker,
		stomp.ConnOpt.HeartBeat(10*time.Second, 10*time.Second),
	)
	if err != nil {
		return err
	}
	sub, err := conn.Subscribe(dest, ack)
	if err != nil {
		conn.Disconnect()
		return err
	}
	h.stompConn = conn
	h.sub = sub
	go h.stompForwarder(ack)
	go h.handleSignals()
	return nil
}

func (h *Hub) stompForwarder(ack stomp.AckMode) {
	for {
		select {
		case <-h.done:
			return
		case msg, ok := <-h.sub.C:
			if !ok {
				return
			}
			if ack != stomp.AckAuto {
				_ = h.stompConn.Ack(msg)
			}
			var chatMsg models.Message
			if json.Unmarshal(msg.Body, &chatMsg) != nil {
				continue
			}
			select {
			case h.Broadcast <- &chatMsg:
			case <-h.done:
				return
			}
		}
	}
}

// ======================
// 5. WebSocket Handler
// ======================

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func validateToken(token string) (string, error) {
	// TODO: real JWT parse
	if token == "valid-token" {
		return "user123", nil
	}
	return "", http.ErrNoCookie
}

// ======================
// 6. Client Pumps
// ======================

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer c.Conn.Close()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump(h *Hub) {
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// optional: process client → server messages
	}

	// Unregister **before** closing the connection
	h.Unregister <- c
	c.Conn.Close()
}

// ======================
// 7. Public Broadcast
// ======================

func Broadcast(msg *models.Message) {
	Get().Broadcast <- msg
}

// ======================
// 8. Cleanup & Signals
// ======================

func (h *Hub) cleanup() {
	h.mu.Lock()
	for _, list := range h.clients {
		for _, c := range list {
			close(c.Send)
			c.Conn.Close()
		}
	}
	h.clients = make(map[string][]*Client)
	h.mu.Unlock()

	if h.sub != nil {
		_ = h.sub.Unsubscribe()
	}
	if h.stompConn != nil {
		_ = h.stompConn.Disconnect()
	}
	log.Println("Hub cleaned")
}

func (h *Hub) handleSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down...")
	h.closeOnce.Do(func() { close(h.done) })
}

// GetClientsByUser returns a copy of the client slice for a user
func (h *Hub) GetClientsByUser(userID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[userID]; ok {
		out := make([]*Client, len(clients))
		copy(out, clients)
		return out
	}
	return nil
}
