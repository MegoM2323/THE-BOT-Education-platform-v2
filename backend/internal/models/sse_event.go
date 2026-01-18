package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	SSEEventNewMessage           = "new_message"
	SSEEventMessageDeleted       = "message_deleted"
	SSEEventMessageStatusUpdated = "message_status_updated"
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
	ChatID      uuid.UUID `json:"chat_id"`
	ID          uuid.UUID `json:"id"`
	SenderID    uuid.UUID `json:"sender_id"`
	MessageText string    `json:"message_text"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type MessageDeletedPayload struct {
	ChatID    uuid.UUID `json:"chat_id"`
	MessageID uuid.UUID `json:"message_id"`
}

type MessageStatusUpdatedPayload struct {
	ChatID    uuid.UUID `json:"chat_id"`
	MessageID uuid.UUID `json:"message_id"`
	Status    string    `json:"status"`
}

func NewMessageEvent(chatID uuid.UUID, message *ChatMessageWithSender) SSEEvent {
	return SSEEvent{
		Type: SSEEventNewMessage,
		Data: NewMessagePayload{
			ChatID: chatID,
			Message: SSEMessagePayload{
				ChatID:      chatID,
				ID:          message.ID,
				SenderID:    message.SenderID,
				MessageText: message.MessageText,
				Status:      "delivered",
				CreatedAt:   message.CreatedAt,
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
				ChatID:      chatID,
				ID:          message.ID,
				SenderID:    message.SenderID,
				MessageText: message.MessageText,
				Status:      "delivered",
				CreatedAt:   message.CreatedAt,
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

func MessageStatusUpdatedEvent(chatID uuid.UUID, messageID uuid.UUID, status string) SSEEvent {
	return SSEEvent{
		Type: SSEEventMessageStatusUpdated,
		Data: MessageStatusUpdatedPayload{
			ChatID:    chatID,
			MessageID: messageID,
			Status:    status,
		},
	}
}

func (e *SSEEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
