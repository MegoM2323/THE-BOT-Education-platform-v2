package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

// ChatRepository управляет операциями с базой данных для чатов и сообщений
type ChatRepository struct {
	db *sqlx.DB
}

// NewChatRepository создает новый ChatRepository
func NewChatRepository(db *sqlx.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// ==================== Chat Rooms ====================

// GetOrCreateRoom получает существующую комнату или создает новую для teacher-student пары
func (r *ChatRepository) GetOrCreateRoom(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error) {
	// Сначала пытаемся найти существующую комнату
	existingRoom, err := r.getRoomByParticipants(ctx, teacherID, studentID)
	if err == nil {
		return existingRoom, nil
	}
	if err != ErrChatRoomNotFound {
		return nil, fmt.Errorf("failed to check existing room: %w", err)
	}

	// Комната не найдена, создаем новую
	query := `
		INSERT INTO chat_rooms (id, teacher_id, student_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, teacher_id, student_id, last_message_at, created_at, updated_at, deleted_at
	`

	room := &models.ChatRoom{
		ID:        uuid.New(),
		TeacherID: teacherID,
		StudentID: studentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = r.db.QueryRowContext(ctx, query,
		room.ID,
		room.TeacherID,
		room.StudentID,
		room.CreatedAt,
		room.UpdatedAt,
	).Scan(
		&room.ID,
		&room.TeacherID,
		&room.StudentID,
		&room.LastMessageAt,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return room, nil
}

// getRoomByParticipants находит комнату по teacher_id и student_id
func (r *ChatRepository) getRoomByParticipants(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error) {
	query := `
		SELECT ` + ChatRoomSelectFields + `
		FROM chat_rooms
		WHERE teacher_id = $1 AND student_id = $2 AND deleted_at IS NULL
	`

	var room models.ChatRoom
	err := r.db.QueryRowContext(ctx, query, teacherID, studentID).Scan(
		&room.ID,
		&room.TeacherID,
		&room.StudentID,
		&room.LastMessageAt,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrChatRoomNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room by participants: %w", err)
	}

	return &room, nil
}

// GetRoomByID получает комнату по ID
func (r *ChatRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.ChatRoom, error) {
	query := `
		SELECT ` + ChatRoomSelectFields + `
		FROM chat_rooms
		WHERE id = $1 AND deleted_at IS NULL
	`

	var room models.ChatRoom
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&room.ID,
		&room.TeacherID,
		&room.StudentID,
		&room.LastMessageAt,
		&room.CreatedAt,
		&room.UpdatedAt,
		&room.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrChatRoomNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room by ID: %w", err)
	}

	return &room, nil
}

// ListRoomsByTeacher получает все комнаты преподавателя с информацией о студентах
func (r *ChatRepository) ListRoomsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*models.ChatRoom, error) {
	query := `
		SELECT
			cr.id, cr.teacher_id, cr.student_id, cr.last_message_at,
			cr.created_at, cr.updated_at, cr.deleted_at,
			u.id AS participant_id,
			CONCAT(u.first_name, ' ', u.last_name) AS participant_name,
			u.role AS participant_role
		FROM chat_rooms cr
		JOIN users u ON u.id = cr.student_id
		WHERE cr.teacher_id = $1 AND cr.deleted_at IS NULL
		ORDER BY cr.last_message_at DESC NULLS LAST, cr.created_at DESC
	`

	rooms := make([]*models.ChatRoom, 0)
	err := r.db.SelectContext(ctx, &rooms, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms by teacher: %w", err)
	}

	return rooms, nil
}

// ListRoomsByStudent получает все комнаты студента с информацией о преподавателях
func (r *ChatRepository) ListRoomsByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.ChatRoom, error) {
	query := `
		SELECT
			cr.id, cr.teacher_id, cr.student_id, cr.last_message_at,
			cr.created_at, cr.updated_at, cr.deleted_at,
			u.id AS participant_id,
			CONCAT(u.first_name, ' ', u.last_name) AS participant_name,
			u.role AS participant_role
		FROM chat_rooms cr
		JOIN users u ON u.id = cr.teacher_id
		WHERE cr.student_id = $1 AND cr.deleted_at IS NULL
		ORDER BY cr.last_message_at DESC NULLS LAST, cr.created_at DESC
	`

	rooms := make([]*models.ChatRoom, 0)
	err := r.db.SelectContext(ctx, &rooms, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms by student: %w", err)
	}

	return rooms, nil
}

// UpdateLastMessageAt обновляет временную метку последнего сообщения в комнате
func (r *ChatRepository) UpdateLastMessageAt(ctx context.Context, roomID uuid.UUID, messageTime time.Time) error {
	query := `
		UPDATE chat_rooms
		SET last_message_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, messageTime, time.Now(), roomID)
	if err != nil {
		return fmt.Errorf("failed to update last_message_at: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrChatRoomNotFound
	}

	return nil
}

// ==================== Messages ====================

// CreateMessage создает новое сообщение
func (r *ChatRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, message_text, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.RoomID,
		msg.SenderID,
		msg.MessageText,
		msg.Status,
		msg.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// UpdateMessageStatus обновляет статус сообщения
func (r *ChatRepository) UpdateMessageStatus(ctx context.Context, msgID uuid.UUID, status string) error {
	var moderationCompletedAt *time.Time
	now := time.Now()

	// Если статус delivered или blocked, устанавливаем moderation_completed_at
	if status == string(models.MessageStatusDelivered) || status == string(models.MessageStatusBlocked) {
		moderationCompletedAt = &now
	}

	query := `
		UPDATE messages
		SET status = $1, moderation_completed_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, status, moderationCompletedAt, msgID)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrMessageNotFound
	}

	return nil
}

// GetMessagesByRoom получает сообщения из комнаты с пагинацией
// Возвращает только delivered сообщения (не показывает blocked)
// Сортировка: старые сообщения первые, новые последние (как в Telegram)
func (r *ChatRepository) GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT
			m.id, m.room_id, m.sender_id, m.message_text, m.status,
			m.moderation_completed_at, m.created_at, m.deleted_at,
			CASE
				WHEN u.first_name IS NOT NULL AND u.first_name != '' THEN
					CASE u.role
						WHEN 'methodologist' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Преподаватель)'
						WHEN 'student' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Студент)'
						WHEN 'admin' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Администратор)'
						ELSE CONCAT(u.first_name, ' ', u.last_name)
					END
				ELSE 'Пользователь'
			END as sender_name
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.room_id = $1
		  AND m.status = $2
		  AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC
		LIMIT $3 OFFSET $4
	`

	var messages []*models.Message
	err := r.db.SelectContext(ctx, &messages, query, roomID, models.MessageStatusDelivered, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by room: %w", err)
	}

	return messages, nil
}

// GetPendingMessages получает все сообщения со статусом pending_moderation
// Используется для обработки очереди модерации
func (r *ChatRepository) GetPendingMessages(ctx context.Context) ([]*models.Message, error) {
	query := `
		SELECT
			m.id, m.room_id, m.sender_id, m.message_text, m.status,
			m.moderation_completed_at, m.created_at, m.deleted_at,
			CASE
				WHEN u.first_name IS NOT NULL AND u.first_name != '' THEN
					CASE u.role
						WHEN 'methodologist' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Преподаватель)'
						WHEN 'student' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Студент)'
						WHEN 'admin' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Администратор)'
						ELSE CONCAT(u.first_name, ' ', u.last_name)
					END
				ELSE 'Пользователь'
			END as sender_name
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.status = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC
		LIMIT 100
	`

	var messages []*models.Message
	err := r.db.SelectContext(ctx, &messages, query, models.MessageStatusPendingModeration)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", err)
	}

	return messages, nil
}

// SoftDeleteMessage мягко удаляет сообщение
func (r *ChatRepository) SoftDeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	query := `
		UPDATE messages
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), msgID)
	if err != nil {
		return fmt.Errorf("failed to soft delete message: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrMessageNotFound
	}

	return nil
}

// GetMessageByID получает сообщение по ID
func (r *ChatRepository) GetMessageByID(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
	query := `
		SELECT
			m.id, m.room_id, m.sender_id, m.message_text, m.status,
			m.moderation_completed_at, m.created_at, m.deleted_at,
			CASE
				WHEN u.first_name IS NOT NULL AND u.first_name != '' THEN
					CASE u.role
						WHEN 'methodologist' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Преподаватель)'
						WHEN 'student' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Студент)'
						WHEN 'admin' THEN CONCAT(u.first_name, ' ', u.last_name) || ' (Администратор)'
						ELSE CONCAT(u.first_name, ' ', u.last_name)
					END
				ELSE 'Пользователь'
			END as sender_name
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`

	var msg models.Message
	err := r.db.QueryRowContext(ctx, query, msgID).Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.SenderID,
		&msg.MessageText,
		&msg.Status,
		&msg.ModerationCompletedAt,
		&msg.CreatedAt,
		&msg.DeletedAt,
		&msg.SenderName,
	)

	if err == sql.ErrNoRows {
		return nil, ErrMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID: %w", err)
	}

	return &msg, nil
}

// ==================== File Attachments ====================

// CreateAttachment создает запись о вложенном файле
func (r *ChatRepository) CreateAttachment(ctx context.Context, att *models.FileAttachment) error {
	query := `
		INSERT INTO file_attachments (id, message_id, file_name, file_path, file_size, mime_type, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	att.ID = uuid.New()
	att.UploadedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		att.ID,
		att.MessageID,
		att.FileName,
		att.FilePath,
		att.FileSize,
		att.MimeType,
		att.UploadedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

// GetAttachmentsByMessage получает все вложения для сообщения
func (r *ChatRepository) GetAttachmentsByMessage(ctx context.Context, msgID uuid.UUID) ([]*models.FileAttachment, error) {
	query := `
		SELECT ` + FileAttachmentSelectFields + `
		FROM file_attachments
		WHERE message_id = $1
		ORDER BY uploaded_at ASC
	`

	var attachments []*models.FileAttachment
	err := r.db.SelectContext(ctx, &attachments, query, msgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments by message: %w", err)
	}

	return attachments, nil
}

// GetAttachmentByID получает вложение по ID
func (r *ChatRepository) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error) {
	query := `
		SELECT ` + FileAttachmentSelectFields + `
		FROM file_attachments
		WHERE id = $1
	`

	var att models.FileAttachment
	err := r.db.QueryRowContext(ctx, query, attachmentID).Scan(
		&att.ID,
		&att.MessageID,
		&att.FileName,
		&att.FilePath,
		&att.FileSize,
		&att.MimeType,
		&att.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrAttachmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment by ID: %w", err)
	}

	return &att, nil
}

// ==================== Transaction Methods ====================

// CreateMessageTx создает сообщение в рамках транзакции
func (r *ChatRepository) CreateMessageTx(ctx context.Context, tx pgx.Tx, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, message_text, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()

	_, err := tx.Exec(ctx, query,
		msg.ID,
		msg.RoomID,
		msg.SenderID,
		msg.MessageText,
		msg.Status,
		msg.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message in transaction: %w", err)
	}

	return nil
}

// UpdateLastMessageAtTx обновляет last_message_at в транзакции
func (r *ChatRepository) UpdateLastMessageAtTx(ctx context.Context, tx pgx.Tx, roomID uuid.UUID, messageTime time.Time) error {
	query := `
		UPDATE chat_rooms
		SET last_message_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.Exec(ctx, query, messageTime, time.Now(), roomID)
	if err != nil {
		return fmt.Errorf("failed to update last_message_at in transaction: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrChatRoomNotFound
	}

	return nil
}

// ==================== Blocked Messages ====================

// CreateBlockedMessage сохраняет заблокированное сообщение
func (r *ChatRepository) CreateBlockedMessage(ctx context.Context, blockedMsg *models.BlockedMessage) error {
	query := `
		INSERT INTO blocked_messages (id, message_id, reason, ai_response, blocked_at, admin_notified, admin_reviewed)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		blockedMsg.ID,
		blockedMsg.MessageID,
		blockedMsg.Reason,
		blockedMsg.AIResponse,
		blockedMsg.BlockedAt,
		blockedMsg.AdminNotified,
		blockedMsg.AdminReviewed,
	)
	if err != nil {
		return fmt.Errorf("failed to create blocked message: %w", err)
	}

	return nil
}

// MarkAdminNotified помечает что админ был уведомлён о заблокированном сообщении
func (r *ChatRepository) MarkAdminNotified(ctx context.Context, messageID uuid.UUID) error {
	query := `
		UPDATE blocked_messages
		SET admin_notified = true
		WHERE message_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to mark admin notified: %w", err)
	}

	return nil
}

// ==================== Admin Methods ====================

// ChatParticipant представляет участника чата
type ChatParticipant struct {
	ID        uuid.UUID `db:"id" json:"id"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	Role      string    `db:"role" json:"role"`
}

// GetFullName возвращает полное имя участника
func (cp *ChatParticipant) GetFullName() string {
	return cp.FirstName + " " + cp.LastName
}

// ChatRoomWithDetails представляет комнату чата с расширенной информацией для админов
type ChatRoomWithDetails struct {
	ID              uuid.UUID         `db:"id" json:"id"`
	Participants    []ChatParticipant `json:"participants"`
	MessagesCount   int               `db:"messages_count" json:"messages_count"`
	LastMessageAt   *time.Time        `db:"last_message_at" json:"last_message_at,omitempty"`
	ParticipantName string            `json:"participant_name,omitempty"`
}

// ListAllRooms возвращает все чаты с информацией о участниках для админ-панели
func (r *ChatRepository) ListAllRooms(ctx context.Context) ([]ChatRoomWithDetails, error) {
	query := `
		SELECT
			cr.id,
			cr.teacher_id,
			cr.student_id,
			t.first_name AS teacher_first_name,
			t.last_name AS teacher_last_name,
			t.role AS teacher_role,
			s.first_name AS student_first_name,
			s.last_name AS student_last_name,
			s.role AS student_role,
			COALESCE(mc.messages_count, 0) AS messages_count,
			cr.last_message_at
		FROM chat_rooms cr
		JOIN users t ON t.id = cr.teacher_id
		JOIN users s ON s.id = cr.student_id
		LEFT JOIN (
			SELECT room_id, COUNT(*) AS messages_count
			FROM messages
			WHERE deleted_at IS NULL
			GROUP BY room_id
		) mc ON mc.room_id = cr.id
		WHERE cr.deleted_at IS NULL
		ORDER BY cr.last_message_at DESC NULLS LAST, cr.created_at DESC
	`

	type roomRow struct {
		ID                 uuid.UUID    `db:"id"`
		TeacherID          uuid.UUID    `db:"teacher_id"`
		StudentID          uuid.UUID    `db:"student_id"`
		TeacherFirstName   string       `db:"teacher_first_name"`
		TeacherLastName    string       `db:"teacher_last_name"`
		TeacherRole        string       `db:"teacher_role"`
		StudentFirstName   string       `db:"student_first_name"`
		StudentLastName    string       `db:"student_last_name"`
		StudentRole        string       `db:"student_role"`
		MessagesCount      int          `db:"messages_count"`
		LastMessageAt      sql.NullTime `db:"last_message_at"`
	}

	var rows []roomRow
	err := r.db.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all rooms: %w", err)
	}

	result := make([]ChatRoomWithDetails, 0, len(rows))
	for _, row := range rows {
		teacherFullName := row.TeacherFirstName + " " + row.TeacherLastName
		studentFullName := row.StudentFirstName + " " + row.StudentLastName
		room := ChatRoomWithDetails{
			ID:              row.ID,
			MessagesCount:   row.MessagesCount,
			ParticipantName: teacherFullName + " ↔ " + studentFullName,
			Participants: []ChatParticipant{
				{
					ID:        row.TeacherID,
					FirstName: row.TeacherFirstName,
					LastName:  row.TeacherLastName,
					Role:      row.TeacherRole,
				},
				{
					ID:        row.StudentID,
					FirstName: row.StudentFirstName,
					LastName:  row.StudentLastName,
					Role:      row.StudentRole,
				},
			},
		}
		if row.LastMessageAt.Valid {
			room.LastMessageAt = &row.LastMessageAt.Time
		}
		result = append(result, room)
	}

	return result, nil
}
