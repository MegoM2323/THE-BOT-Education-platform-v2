package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UUIDArray кастомный тип для работы с массивами UUID в PostgreSQL
type UUIDArray []uuid.UUID

// Scan реализует интерфейс sql.Scanner для чтения массива UUID из БД
func (a *UUIDArray) Scan(src interface{}) error {
	// PostgreSQL возвращает массив как pq.GenericArray
	var arr pq.StringArray
	if err := arr.Scan(src); err != nil {
		return err
	}

	// Преобразуем строки в UUID
	result := make([]uuid.UUID, len(arr))
	for i, s := range arr {
		id, err := uuid.Parse(s)
		if err != nil {
			return fmt.Errorf("invalid UUID in array: %w", err)
		}
		result[i] = id
	}

	*a = result
	return nil
}

// Value реализует интерфейс driver.Valuer для записи массива UUID в БД
func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	// Преобразуем UUID в строки
	strs := make([]string, len(a))
	for i, id := range a {
		strs[i] = id.String()
	}

	// Используем pq.Array для правильной сериализации
	return pq.Array(strs).Value()
}

// MarshalJSON реализует интерфейс json.Marshaler
func (a UUIDArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]uuid.UUID(a))
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler
func (a *UUIDArray) UnmarshalJSON(data []byte) error {
	var arr []uuid.UUID
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}

// TelegramUser представляет привязку пользователя к Telegram аккаунту
type TelegramUser struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	User       *User     `db:"-" json:"user,omitempty"`
	TelegramID int64     `db:"telegram_id" json:"telegram_id"`
	ChatID     int64     `db:"chat_id" json:"chat_id"`
	Username   string    `db:"username" json:"username"`
	Subscribed bool      `db:"subscribed" json:"subscribed"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// LinkTelegramRequest представляет запрос на привязку Telegram аккаунта
type LinkTelegramRequest struct {
	Token      string `json:"token" validate:"required"`
	TelegramID int64  `json:"telegram_id" validate:"required"`
	ChatID     int64  `json:"chat_id" validate:"required"`
	Username   string `json:"username"`
}

// BroadcastList представляет список рассылки для отправки сообщений
type BroadcastList struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	UserIDs     UUIDArray  `db:"user_ids" json:"user_ids"`
	CreatedBy   uuid.UUID  `db:"created_by" json:"created_by"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// CreateBroadcastListRequest представляет запрос на создание списка рассылки
type CreateBroadcastListRequest struct {
	Name        string      `json:"name" validate:"required,min=3"`
	Description string      `json:"description"`
	UserIDs     []uuid.UUID `json:"user_ids" validate:"required,min=1"`
}

// UpdateBroadcastListRequest представляет запрос на обновление списка рассылки
type UpdateBroadcastListRequest struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	UserIDs     []uuid.UUID `json:"user_ids,omitempty"`
}

// Broadcast представляет массовую рассылку сообщений
type Broadcast struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	ListID      *uuid.UUID `db:"list_id" json:"list_id"`
	Message     string     `db:"message" json:"message"`
	SentCount   int        `db:"sent_count" json:"sent_count"`
	FailedCount int        `db:"failed_count" json:"failed_count"`
	Status      string     `db:"status" json:"status"`
	CreatedBy   uuid.UUID  `db:"created_by" json:"created_by"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}

// BroadcastLog представляет лог отправки сообщения конкретному пользователю
type BroadcastLog struct {
	ID          uuid.UUID `db:"id" json:"id"`
	BroadcastID uuid.UUID `db:"broadcast_id" json:"broadcast_id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	TelegramID  int64     `db:"telegram_id" json:"telegram_id"`
	Status      string    `db:"status" json:"status"`
	Error       string    `db:"error" json:"error,omitempty"`
	SentAt      time.Time `db:"sent_at" json:"sent_at"`
}

// SendBroadcastRequest представляет запрос на отправку массовой рассылки
type SendBroadcastRequest struct {
	ListID  *uuid.UUID  `json:"list_id,omitempty"`  // Optional: broadcast list ID
	UserIDs []uuid.UUID `json:"user_ids,omitempty"` // Optional: specific user IDs
	Message string      `json:"message" validate:"required,min=1,max=4096"`
}

// TeacherLessonBroadcastRequest представляет запрос на отправку рассылки студентам занятия (от преподавателя)
type TeacherLessonBroadcastRequest struct {
	Message string `json:"message" validate:"required,min=1,max=4096"`
}

// Константы статусов рассылки
const (
	BroadcastStatusPending    = "pending"
	BroadcastStatusInProgress = "in_progress"
	BroadcastStatusCompleted  = "completed"
	BroadcastStatusFailed     = "failed"
	BroadcastStatusCancelled  = "cancelled"

	// Статусы логов доставки (должны соответствовать CHECK constraint в БД: 'sent', 'failed', 'skipped')
	BroadcastLogStatusSuccess = "sent"
	BroadcastLogStatusFailed  = "failed"
	BroadcastLogStatusSkipped = "skipped"
)

// Validate выполняет валидацию LinkTelegramRequest
func (r *LinkTelegramRequest) Validate() error {
	if r.Token == "" {
		return ErrInvalidToken
	}
	if r.TelegramID == 0 {
		return ErrInvalidTelegramID
	}
	if r.ChatID == 0 {
		return ErrInvalidChatID
	}
	return nil
}

// Validate выполняет валидацию CreateBroadcastListRequest
func (r *CreateBroadcastListRequest) Validate() error {
	if len(r.Name) < 3 {
		return ErrInvalidBroadcastName
	}
	if len(r.UserIDs) < 1 {
		return ErrInvalidBroadcastUsers
	}
	return nil
}

// Validate выполняет валидацию UpdateBroadcastListRequest
func (r *UpdateBroadcastListRequest) Validate() error {
	if r.Name != "" && len(r.Name) < 3 {
		return ErrInvalidBroadcastName
	}
	return nil
}

// Validate выполняет валидацию SendBroadcastRequest
func (r *SendBroadcastRequest) Validate() error {
	// Требуем ЛИБО list_id ЛИБО user_ids (не оба, не ни один)
	hasListID := r.ListID != nil && *r.ListID != uuid.Nil
	hasUserIDs := len(r.UserIDs) > 0

	if !hasListID && !hasUserIDs {
		return errors.New("either list_id or user_ids must be provided")
	}
	if hasListID && hasUserIDs {
		return errors.New("cannot specify both list_id and user_ids")
	}

	// Валидация сообщения
	if r.Message == "" {
		return ErrInvalidBroadcastMessage
	}
	if len(r.Message) > 4096 {
		return ErrBroadcastMessageTooLong
	}

	return nil
}

// Validate выполняет валидацию TeacherLessonBroadcastRequest
func (r *TeacherLessonBroadcastRequest) Validate() error {
	// Валидация сообщения
	if r.Message == "" {
		return ErrInvalidBroadcastMessage
	}
	if len(r.Message) > 4096 {
		return ErrBroadcastMessageTooLong
	}

	return nil
}

// IsDeleted проверяет, удален ли список рассылки (мягкое удаление)
func (bl *BroadcastList) IsDeleted() bool {
	return bl.DeletedAt != nil
}

// IsPending проверяет, находится ли рассылка в состоянии ожидания
func (b *Broadcast) IsPending() bool {
	return b.Status == BroadcastStatusPending
}

// IsInProgress проверяет, выполняется ли рассылка в данный момент
func (b *Broadcast) IsInProgress() bool {
	return b.Status == BroadcastStatusInProgress
}

// IsCompleted проверяет, завершена ли рассылка
func (b *Broadcast) IsCompleted() bool {
	return b.Status == BroadcastStatusCompleted
}

// IsFailed проверяет, провалилась ли рассылка
func (b *Broadcast) IsFailed() bool {
	return b.Status == BroadcastStatusFailed
}

// IsCancelled проверяет, отменена ли рассылка
func (b *Broadcast) IsCancelled() bool {
	return b.Status == BroadcastStatusCancelled
}
