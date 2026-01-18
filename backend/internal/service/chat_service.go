package service

import (
	"context"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
)

// ChatService обрабатывает бизнес-логику для чатов и сообщений
type ChatService struct {
	chatRepo          chatServiceRepository
	userRepo          chatServiceUserRepository
	moderationService *ModerationService // Используется существующий ModerationService из moderation_service.go
}

// chatServiceRepository - интерфейс для dependency injection в тестах
type chatServiceRepository interface {
	GetOrCreateRoom(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error)
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.ChatRoom, error)
	ListRoomsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*models.ChatRoom, error)
	ListRoomsByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.ChatRoom, error)
	CreateMessage(ctx context.Context, msg *models.Message) error
	UpdateMessageStatus(ctx context.Context, msgID uuid.UUID, status string) error
	GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*models.Message, error)
	GetMessageByID(ctx context.Context, msgID uuid.UUID) (*models.Message, error)
	GetAttachmentsByMessage(ctx context.Context, msgID uuid.UUID) ([]*models.FileAttachment, error)
	UpdateLastMessageAt(ctx context.Context, roomID uuid.UUID, messageTime time.Time) error
	GetPendingMessages(ctx context.Context) ([]*models.Message, error)
	CreateAttachment(ctx context.Context, att *models.FileAttachment) error
	GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error)
}

// chatServiceUserRepository - интерфейс для dependency injection
type chatServiceUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// NewChatService создает новый ChatService
func NewChatService(
	chatRepo chatServiceRepository,
	userRepo chatServiceUserRepository,
	moderationService *ModerationService,
) *ChatService {
	return &ChatService{
		chatRepo:          chatRepo,
		userRepo:          userRepo,
		moderationService: moderationService,
	}
}

// ==================== Chat Room Methods ====================

// GetOrCreateRoom получает существующую комнату или создает новую для текущего пользователя и другого участника
// Автоматически определяет кто teacher а кто student
func (s *ChatService) GetOrCreateRoom(ctx context.Context, currentUserID, otherUserID uuid.UUID) (*models.ChatRoom, error) {
	// Валидация: пользователь не может создать чат с собой
	if currentUserID == otherUserID {
		return nil, models.ErrCannotChatWithSelf
	}

	// Получаем информацию о пользователях
	currentUser, err := s.userRepo.GetByID(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	otherUser, err := s.userRepo.GetByID(ctx, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get other user: %w", err)
	}

	// Определяем кто teacher а кто student
	var teacherID, studentID uuid.UUID

	if currentUser.IsTeacher() || currentUser.IsAdmin() {
		teacherID = currentUserID
		studentID = otherUserID
	} else if otherUser.IsTeacher() || otherUser.IsAdmin() {
		teacherID = otherUserID
		studentID = currentUserID
	} else {
		// Оба студента — невозможно создать комнату (студенты могут общаться только с преподавателями)
		return nil, fmt.Errorf("chat rooms can only be created between teachers and students")
	}

	// Проверяем что другой пользователь действительно student (если currentUser — teacher)
	if currentUser.IsTeacher() && !otherUser.IsStudent() {
		return nil, fmt.Errorf("teachers can only chat with students")
	}

	// Получаем или создаем комнату
	room, err := s.chatRepo.GetOrCreateRoom(ctx, teacherID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create room: %w", err)
	}

	return room, nil
}

// GetUserChats получает список комнат для текущего пользователя
// Возвращает комнаты с информацией о последнем сообщении и непрочитанных сообщениях
func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID, role string) ([]*models.ChatRoom, error) {
	var rooms []*models.ChatRoom
	var err error

	switch role {
	case string(models.RoleTeacher), string(models.RoleAdmin):
		rooms, err = s.chatRepo.ListRoomsByTeacher(ctx, userID)
	case string(models.RoleStudent):
		rooms, err = s.chatRepo.ListRoomsByStudent(ctx, userID)
	default:
		return nil, models.ErrInvalidRole
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list user chats: %w", err)
	}

	return rooms, nil
}

// GetRoomByID получает комнату по ID с проверкой доступа
func (s *ChatService) GetRoomByID(ctx context.Context, roomID, userID uuid.UUID) (*models.ChatRoom, error) {
	room, err := s.chatRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Проверяем что пользователь — участник комнаты
	if !room.IsParticipant(userID) {
		return nil, repository.ErrUnauthorized
	}

	return room, nil
}

// ==================== Message Methods ====================

// SendMessage отправляет сообщение в комнату с асинхронной модерацией
// Workflow:
// 1. Валидация: проверить что sender - участник комнаты
// 2. Создать сообщение со статусом pending_moderation
// 3. Запустить асинхронную модерацию (goroutine)
// 4. Вернуть сообщение
func (s *ChatService) SendMessage(ctx context.Context, senderID uuid.UUID, req *models.SendMessageRequest) (*models.Message, error) {
	// Валидация запроса
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Получаем комнату
	room, err := s.chatRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Проверяем что отправитель — участник комнаты
	if !room.IsParticipant(senderID) {
		return nil, repository.ErrUnauthorized
	}

	// Создаем сообщение со статусом pending_moderation
	message := &models.Message{
		RoomID:      req.RoomID,
		SenderID:    senderID,
		MessageText: req.MessageText,
		Status:      string(models.MessageStatusPendingModeration),
	}

	if err := s.chatRepo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Обновляем last_message_at в комнате
	if err := s.chatRepo.UpdateLastMessageAt(ctx, req.RoomID, message.CreatedAt); err != nil {
		// Логируем ошибку но не прерываем процесс
		fmt.Printf("[WARN] Failed to update last_message_at for room %s: %v\n", req.RoomID, err)
	}

	// Запускаем асинхронную модерацию
	go s.moderateMessageAsync(message.ID, message.MessageText)

	return message, nil
}

// moderateMessageAsync выполняет асинхронную модерацию сообщения
// Вызывается в горутине из SendMessage
func (s *ChatService) moderateMessageAsync(messageID uuid.UUID, messageText string) {
	ctx := context.Background()

	// Используем новый ModerationService с автоматическим fallback и circuit breaker
	if s.moderationService != nil {
		// Вызываем асинхронную модерацию (она сама управляет fallback и уведомлениями)
		s.moderationService.ModerateMessageAsync(ctx, messageID)
		return
	}

	// Если ModerationService не инициализирован, просто доставляем сообщение
	var blocked bool
	var reason string
	blocked = false
	reason = ""

	// Обновляем статус сообщения
	var newStatus string
	if blocked {
		newStatus = string(models.MessageStatusBlocked)
	} else {
		newStatus = string(models.MessageStatusDelivered)
	}

	if err := s.chatRepo.UpdateMessageStatus(ctx, messageID, newStatus); err != nil {
		fmt.Printf("[ERROR] Failed to update message status for %s: %v\n", messageID, err)
		return
	}

	if blocked {
		fmt.Printf("[INFO] Message %s blocked: %s\n", messageID, reason)
	}
}

// GetChatHistory получает историю сообщений в комнате
// Возвращает только delivered сообщения (не показывает blocked)
func (s *ChatService) GetChatHistory(ctx context.Context, userID uuid.UUID, req *models.GetMessagesRequest) ([]*models.Message, error) {
	// Валидация запроса
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Получаем комнату
	room, err := s.chatRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Проверяем что пользователь — участник комнаты
	if !room.IsParticipant(userID) {
		return nil, repository.ErrUnauthorized
	}

	// Получаем сообщения (только delivered)
	messages, err := s.chatRepo.GetMessagesByRoom(ctx, req.RoomID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Для каждого сообщения загружаем вложения (если есть)
	for _, msg := range messages {
		attachments, err := s.chatRepo.GetAttachmentsByMessage(ctx, msg.ID)
		if err != nil {
			// Логируем ошибку но продолжаем
			fmt.Printf("[WARN] Failed to get attachments for message %s: %v\n", msg.ID, err)
			continue
		}

		// Конвертируем []*FileAttachment в []FileAttachment
		for _, att := range attachments {
			if att != nil {
				msg.Attachments = append(msg.Attachments, *att)
			}
		}
	}

	return messages, nil
}

// GetPendingMessages получает все сообщения на модерации
// Используется для batch обработки очереди модерации
func (s *ChatService) GetPendingMessages(ctx context.Context) ([]*models.Message, error) {
	messages, err := s.chatRepo.GetPendingMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", err)
	}

	return messages, nil
}

// ==================== File Attachment Methods ====================

// CreateAttachment создает запись о вложенном файле
// Validation: проверяет что пользователь — участник комнаты сообщения
func (s *ChatService) CreateAttachment(ctx context.Context, userID uuid.UUID, attachment *models.FileAttachment) error {
	// Получаем сообщение
	message, err := s.chatRepo.GetMessageByID(ctx, attachment.MessageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Получаем комнату
	room, err := s.chatRepo.GetRoomByID(ctx, message.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}

	// Проверяем что пользователь — участник комнаты
	if !room.IsParticipant(userID) {
		return repository.ErrUnauthorized
	}

	// Создаем вложение
	if err := s.chatRepo.CreateAttachment(ctx, attachment); err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

// GetAttachmentsByMessage получает вложения для сообщения
// Validation: проверяет что пользователь — участник комнаты
func (s *ChatService) GetAttachmentsByMessage(ctx context.Context, userID, messageID uuid.UUID) ([]*models.FileAttachment, error) {
	// Получаем сообщение
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Получаем комнату
	room, err := s.chatRepo.GetRoomByID(ctx, message.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Проверяем что пользователь — участник комнаты
	if !room.IsParticipant(userID) {
		return nil, repository.ErrUnauthorized
	}

	// Получаем вложения
	attachments, err := s.chatRepo.GetAttachmentsByMessage(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}

	return attachments, nil
}

// GetMessageByIDInternal получает сообщение по ID без проверки авторизации
// Используется для внутренних операций (например, проверки авторизации для скачивания файлов)
func (s *ChatService) GetMessageByIDInternal(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	return s.chatRepo.GetMessageByID(ctx, messageID)
}

// GetMessageWithAttachments получает сообщение с вложениями
// Проверяет что пользователь - участник комнаты
func (s *ChatService) GetMessageWithAttachments(ctx context.Context, userID, messageID uuid.UUID) (*models.Message, error) {
	// Получаем сообщение
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Получаем комнату
	room, err := s.chatRepo.GetRoomByID(ctx, message.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Проверяем что пользователь - участник комнаты
	if !room.IsParticipant(userID) {
		return nil, repository.ErrUnauthorized
	}

	// Загружаем вложения
	attachments, err := s.chatRepo.GetAttachmentsByMessage(ctx, messageID)
	if err != nil {
		// Логируем ошибку но продолжаем (вложения необязательны)
		fmt.Printf("[WARN] Failed to get attachments for message %s: %v\n", messageID, err)
		return message, nil
	}

	// Конвертируем []*FileAttachment в []FileAttachment
	for _, att := range attachments {
		if att != nil {
			message.Attachments = append(message.Attachments, *att)
		}
	}

	return message, nil
}
