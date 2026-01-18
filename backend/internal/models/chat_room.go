package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageStatus constants moved to chat_message.go to avoid duplication

// ChatRoom представляет комнату для 1-на-1 общения между преподавателем и студентом
type ChatRoom struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	TeacherID     uuid.UUID    `db:"teacher_id" json:"teacher_id"`
	StudentID     uuid.UUID    `db:"student_id" json:"student_id"`
	LastMessageAt sql.NullTime `db:"last_message_at" json:"last_message_at,omitempty"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt     sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`

	// Информация о другом участнике (teacher для student, student для teacher)
	ParticipantID   uuid.UUID `db:"participant_id" json:"participant_id,omitempty"`
	ParticipantName string    `db:"participant_name" json:"participant_name,omitempty"`
	ParticipantRole string    `db:"participant_role" json:"participant_role,omitempty"`
}

// Message представляет сообщение в чате с модерацией и статусами.
// Это единственный исходный тип для всех операций с сообщениями.
// Status может быть: pending_moderation, delivered или blocked
type Message struct {
	ID                    uuid.UUID        `db:"id" json:"id"`
	RoomID                uuid.UUID        `db:"room_id" json:"room_id"`
	SenderID              uuid.UUID        `db:"sender_id" json:"sender_id"`
	MessageText           string           `db:"message_text" json:"message_text"`
	Status                string           `db:"status" json:"status"` // pending_moderation, delivered, blocked
	ModerationCompletedAt sql.NullTime     `db:"moderation_completed_at" json:"moderation_completed_at,omitempty"`
	CreatedAt             time.Time        `db:"created_at" json:"created_at"`
	DeletedAt             sql.NullTime     `db:"deleted_at" json:"deleted_at,omitempty"`
	Attachments           []FileAttachment `db:"-" json:"attachments,omitempty"` // Загружаются отдельно
}

// FileAttachment представляет файл, прикрепленный к сообщению
type FileAttachment struct {
	ID         uuid.UUID `db:"id" json:"id"`
	MessageID  uuid.UUID `db:"message_id" json:"message_id"`
	FileName   string    `db:"file_name" json:"file_name"`
	FilePath   string    `db:"file_path" json:"-"` // НЕ выдавать в API ответе
	FileSize   int64     `db:"file_size" json:"file_size"`
	MimeType   string    `db:"mime_type" json:"mime_type"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}

// BlockedMessage type moved to chat_message.go to avoid duplication

// ==================== Request Models ====================

// GetOrCreateRoomRequest представляет запрос на получение или создание комнаты
type GetOrCreateRoomRequest struct {
	OtherUserID uuid.UUID `json:"other_user_id"`
}

// SendMessageRequest представляет запрос на отправку сообщения
type SendMessageRequest struct {
	RoomID      uuid.UUID `json:"room_id"`
	MessageText string    `json:"message_text"`
}

// GetMessagesRequest представляет запрос на получение сообщений
type GetMessagesRequest struct {
	RoomID uuid.UUID `json:"room_id"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// ==================== Validation Methods ====================

// Validate валидирует GetOrCreateRoomRequest
func (r *GetOrCreateRoomRequest) Validate() error {
	if r.OtherUserID == uuid.Nil {
		return ErrInvalidUserID
	}
	return nil
}

// Validate валидирует SendMessageRequest
func (r *SendMessageRequest) Validate() error {
	if r.RoomID == uuid.Nil {
		return ErrInvalidChatID
	}

	if r.MessageText == "" {
		return ErrMessageContentEmpty
	}

	// Максимальная длина сообщения 4096 символов (как в Telegram)
	if len(r.MessageText) > 4096 {
		return ErrMessageContentTooLong
	}

	return nil
}

// Validate валидирует GetMessagesRequest
func (r *GetMessagesRequest) Validate() error {
	if r.RoomID == uuid.Nil {
		return ErrInvalidChatID
	}

	// Лимит должен быть положительным
	if r.Limit <= 0 {
		return ErrInvalidMessageLimit
	}

	// Максимальный лимит 100 сообщений за раз
	if r.Limit > 100 {
		r.Limit = 100
	}

	// Offset не может быть отрицательным
	if r.Offset < 0 {
		return ErrInvalidMessageOffset
	}

	return nil
}

// ==================== Helper Methods ====================

// IsParticipant проверяет, является ли пользователь участником комнаты
func (r *ChatRoom) IsParticipant(userID uuid.UUID) bool {
	return r.TeacherID == userID || r.StudentID == userID
}

// GetOtherParticipant возвращает ID другого участника комнаты
func (r *ChatRoom) GetOtherParticipant(userID uuid.UUID) uuid.UUID {
	if r.TeacherID == userID {
		return r.StudentID
	}
	return r.TeacherID
}

// IsPending проверяет, находится ли сообщение на модерации
func (m *Message) IsPending() bool {
	return m.Status == MessageStatusPendingModeration
}

// IsDelivered проверяет, доставлено ли сообщение
func (m *Message) IsDelivered() bool {
	return m.Status == MessageStatusDelivered
}

// IsBlocked проверяет, заблокировано ли сообщение
func (m *Message) IsBlocked() bool {
	return m.Status == MessageStatusBlocked
}

// IsDeleted проверяет, удалено ли сообщение
func (m *Message) IsDeleted() bool {
	return m.DeletedAt.Valid
}

// CanDelete проверяет, может ли пользователь удалить сообщение
// Только отправитель может удалить сообщение
func (m *Message) CanDelete(userID uuid.UUID) bool {
	return m.SenderID == userID
}

// GetStatusDescription возвращает человеко-читаемое описание статуса
func (m *Message) GetStatusDescription() string {
	switch m.Status {
	case MessageStatusPendingModeration:
		return "Проверяется модерацией"
	case MessageStatusDelivered:
		return "Доставлено"
	case MessageStatusBlocked:
		return "Заблокировано модерацией"
	default:
		return "Неизвестный статус"
	}
}

// MarshalJSON customizes JSON marshaling for Message
// Ensures proper serialization of nullable time fields
func (m *Message) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"id":           m.ID,
		"room_id":      m.RoomID,
		"sender_id":    m.SenderID,
		"message_text": m.MessageText,
		"status":       m.Status,
		"created_at":   m.CreatedAt,
	}

	// Handle ModerationCompletedAt
	if m.ModerationCompletedAt.Valid {
		data["moderation_completed_at"] = m.ModerationCompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	// Handle DeletedAt
	if m.DeletedAt.Valid {
		data["deleted_at"] = m.DeletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	// Handle Attachments
	if len(m.Attachments) > 0 {
		data["attachments"] = m.Attachments
	}

	return json.Marshal(data)
}
