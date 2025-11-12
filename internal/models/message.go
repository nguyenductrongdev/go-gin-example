package models

// Message represents a chat message from Kafka
type Message struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	SenderID       string                 `json:"sender_id"`
	RecipientID    string                 `json:"recipient_id,omitempty"`
	GroupID        string                 `json:"group_id,omitempty"`
	Content        string                 `json:"content"`
	MessageType    string                 `json:"message_type"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      string                 `json:"created_at"`
	EventType      string                 `json:"event_type"` // message.sent, message.edited, message.deleted, etc.
}

// EventType constants
const (
	EventTypeSent    = "message.sent"
	EventTypeEdited  = "message.edited"
	EventTypeDeleted = "message.deleted"
	EventTypeRead    = "message.read"
	EventTypeTyping  = "typing.start"
	EventTypeStopTyping = "typing.stop"
)

// WelcomeMessage represents a welcome message sent to newly connected clients
type WelcomeMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Time    string `json:"time"`
}

