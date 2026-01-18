package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ChatMessageWithSender представляет сообщение с информацией об отправителе
// Используется для запросов, требующих информацию о sender
type ChatMessageWithSender struct {
	ID                    uuid.UUID    `db:"id" json:"id"`
	RoomID                uuid.UUID    `db:"room_id" json:"room_id"`
	SenderID              uuid.UUID    `db:"sender_id" json:"sender_id"`
	MessageText           string       `db:"message_text" json:"message_text"`
	Status                string       `db:"status" json:"status"`
	ModerationCompletedAt sql.NullTime `db:"moderation_completed_at" json:"moderation_completed_at,omitempty"`
	CreatedAt             time.Time    `db:"created_at" json:"created_at"`
	DeletedAt             sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
	SenderName            string       `db:"sender_name" json:"sender_name"`
}

// SendChatMessageRequest представляет запрос на отправку сообщения в чат
type SendChatMessageRequest struct {
	RoomID      uuid.UUID `json:"room_id"`
	MessageText string    `json:"message_text"`
}

// BlockedMessage представляет заблокированное модерацией сообщение
type BlockedMessage struct {
	ID            uuid.UUID              `db:"id" json:"id"`
	MessageID     uuid.UUID              `db:"message_id" json:"message_id"`
	Reason        string                 `db:"reason" json:"reason"`
	AIResponse    map[string]interface{} `db:"ai_response" json:"ai_response"`
	BlockedAt     time.Time              `db:"blocked_at" json:"blocked_at"`
	AdminNotified bool                   `db:"admin_notified" json:"admin_notified"`
	AdminReviewed bool                   `db:"admin_reviewed" json:"admin_reviewed"`
}

const (
	// MessageStatusPendingModeration сообщение ожидает модерации
	MessageStatusPendingModeration = "pending_moderation"
	// MessageStatusDelivered сообщение доставлено
	MessageStatusDelivered = "delivered"
	// MessageStatusBlocked сообщение заблокировано модерацией
	MessageStatusBlocked = "blocked"
)

// Validate выполняет валидацию SendChatMessageRequest
func (r *SendChatMessageRequest) Validate() error {
	// Проверка ID комнаты
	if r.RoomID == uuid.Nil {
		return ErrInvalidChatRoomID
	}

	// Проверка текста сообщения
	if r.MessageText == "" {
		return ErrMessageContentEmpty
	}

	// Проверка максимальной длины (4096 символов как в Telegram)
	if len(r.MessageText) > 4096 {
		return ErrMessageContentTooLong
	}

	return nil
}
