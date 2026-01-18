package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	SSEEventNewMessage      = "new_message"
	SSEEventMessageDeleted  = "message_deleted"
)

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type NewMessagePayload struct {
	ChatID  uuid.UUID         `json:"chat_id"`
	Message SSEMessagePayload `json:"message"`
}

type SSEMessagePayload struct {
	ID        uuid.UUID `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageDeletedPayload struct {
	ChatID    uuid.UUID `json:"chat_id"`
	MessageID uuid.UUID `json:"message_id"`
}

func NewMessageEvent(chatID uuid.UUID, message *ChatMessageWithSender) SSEEvent {
	return SSEEvent{
		Type: SSEEventNewMessage,
		Data: NewMessagePayload{
			ChatID: chatID,
			Message: SSEMessagePayload{
				ID:        message.ID,
				SenderID:  message.SenderID,
				Content:   message.MessageText,
				CreatedAt: message.CreatedAt,
			},
		},
	}
}

func NewMessageEventFromMessage(chatID uuid.UUID, message *Message) SSEEvent {
	return SSEEvent{
		Type: SSEEventNewMessage,
		Data: NewMessagePayload{
			ChatID: chatID,
			Message: SSEMessagePayload{
				ID:        message.ID,
				SenderID:  message.SenderID,
				Content:   message.MessageText,
				CreatedAt: message.CreatedAt,
			},
		},
	}
}

func MessageDeletedEvent(chatID uuid.UUID, messageID uuid.UUID) SSEEvent {
	return SSEEvent{
		Type: SSEEventMessageDeleted,
		Data: MessageDeletedPayload{
			ChatID:    chatID,
			MessageID: messageID,
		},
	}
}

func (e *SSEEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
