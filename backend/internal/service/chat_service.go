package service

import (
	"context"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/sse"

	"github.com/google/uuid"
)

// ChatService обрабатывает бизнес-логику для чатов и сообщений
type ChatService struct {
	chatRepo          chatServiceRepository
	userRepo          chatServiceUserRepository
	moderationService *ModerationService
	sseManager        *sse.ConnectionManagerUUID
	telegramService   *TelegramService
}

// chatServiceRepository - интерфейс для dependency injection в тестах
type chatServiceRepository interface {
	GetOrCreateRoom(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error)
	GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.ChatRoom, error)
	ListRoomsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*models.ChatRoom, error)
	ListRoomsByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.ChatRoom, error)
	ListAllRooms(ctx context.Context) ([]repository.ChatRoomWithDetails, error)
	CreateMessage(ctx context.Context, msg *models.Message) error
	UpdateMessageStatus(ctx context.Context, msgID uuid.UUID, status string) error
	GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*models.Message, error)
	GetMessageByID(ctx context.Context, msgID uuid.UUID) (*models.Message, error)
	GetAttachmentsByMessage(ctx context.Context, msgID uuid.UUID) ([]*models.FileAttachment, error)
	UpdateLastMessageAt(ctx context.Context, roomID uuid.UUID, messageTime time.Time) error
	GetPendingMessages(ctx context.Context) ([]*models.Message, error)
	CreateAttachment(ctx context.Context, att *models.FileAttachment) error
	GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error)
	SoftDeleteMessage(ctx context.Context, msgID uuid.UUID) error
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

// SetSSEManager устанавливает SSE менеджер для broadcast сообщений
func (s *ChatService) SetSSEManager(manager *sse.ConnectionManagerUUID) {
	s.sseManager = manager
}

// SetTelegramService устанавливает Telegram сервис для уведомлений
func (s *ChatService) SetTelegramService(service *TelegramService) {
	s.telegramService = service
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

	// Проверяем что оба не студенты
	if currentUser.IsStudent() && otherUser.IsStudent() {
		return nil, fmt.Errorf("students cannot chat with each other")
	}

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

// GetUserByID получает информацию о пользователе по ID
func (s *ChatService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
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

	// Получаем информацию об отправителе для уведомления
	sender, err := s.userRepo.GetByID(ctx, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}

	// Создаем сообщение со статусом delivered (временно без модерации)
	message := &models.Message{
		RoomID:      req.RoomID,
		SenderID:    senderID,
		MessageText: req.MessageText,
		Status:      string(models.MessageStatusDelivered),
	}

	if err := s.chatRepo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Обновляем last_message_at в комнате
	if err := s.chatRepo.UpdateLastMessageAt(ctx, req.RoomID, message.CreatedAt); err != nil {
		fmt.Printf("[WARN] Failed to update last_message_at for room %s: %v\n", req.RoomID, err)
	}

	// SSE broadcast: отправляем участникам чата (кроме отправителя)
	if s.sseManager != nil {
		event := models.NewMessageEventFromMessage(req.RoomID, message)
		sseEvent := sse.EventUUID{
			Type: event.Type,
			Data: event.Data,
		}
		s.sseManager.SendToChat(req.RoomID, sseEvent, senderID)
	}

	// Отправляем уведомление в Telegram получателю
	if s.telegramService != nil {
		var recipientID uuid.UUID
		if room.StudentID == senderID {
			recipientID = room.TeacherID
		} else {
			recipientID = room.StudentID
		}

		senderName := sender.GetFullName()
		notificationText := fmt.Sprintf("💬 Новое сообщение от %s:\n\n%s", senderName, req.MessageText)
		go s.telegramService.SendUserNotification(ctx, recipientID, notificationText)
	}

	// Модерация временно отключена - все сообщения доставляются напрямую
	// TODO: включить модерацию после стабилизации системы
	// go s.moderateMessageAsync(message.ID, message.RoomID)

	return message, nil
}

// moderateMessageAsync выполняет асинхронную модерацию сообщения
// Вызывается в горутине из SendMessage
func (s *ChatService) moderateMessageAsync(messageID uuid.UUID, roomID uuid.UUID) {
	ctx := context.Background()

	// Используем новый ModerationService с автоматическим fallback и circuit breaker
	if s.moderationService != nil {
		// Вызываем асинхронную модерацию (она сама управляет fallback и уведомлениями)
		s.moderationService.ModerateMessageAsync(ctx, messageID)
		return
	}

	// Если ModerationService не инициализирован, просто доставляем сообщение
	newStatus := string(models.MessageStatusDelivered)

	if err := s.chatRepo.UpdateMessageStatus(ctx, messageID, newStatus); err != nil {
		fmt.Printf("[ERROR] Failed to update message status for %s: %v\n", messageID, err)
		return
	}

	// Отправляем SSE уведомление об изменении статуса
	if s.sseManager != nil {
		event := models.MessageStatusUpdatedEvent(roomID, messageID, newStatus)
		sseEvent := sse.EventUUID{
			Type: event.Type,
			Data: event.Data,
		}
		s.sseManager.SendToChat(roomID, sseEvent, uuid.Nil)
	}
}

// GetChatHistory получает историю сообщений в комнате
// Возвращает только delivered сообщения (не показывает blocked)
// Админы могут читать любые чаты без проверки участия
func (s *ChatService) GetChatHistory(ctx context.Context, userID uuid.UUID, role string, req *models.GetMessagesRequest) ([]*models.Message, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	room, err := s.chatRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	isAdmin := role == string(models.RoleAdmin)
	if !isAdmin && !room.IsParticipant(userID) {
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

// GetAttachmentByID получает вложение по ID с проверкой доступа к комнате
// Validation: проверяет что пользователь — участник комнаты, которой принадлежит сообщение с вложением
func (s *ChatService) GetAttachmentByID(ctx context.Context, userID, roomID, attachmentID uuid.UUID) (*models.FileAttachment, error) {
	// Проверяем доступ к комнате
	room, err := s.chatRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	if !room.IsParticipant(userID) {
		return nil, repository.ErrUnauthorized
	}

	// Получаем вложение по ID
	attachment, err := s.chatRepo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	// Проверяем что сообщение принадлежит запрошенной комнате
	message, err := s.chatRepo.GetMessageByID(ctx, attachment.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if message.RoomID != roomID {
		return nil, repository.ErrUnauthorized
	}

	return attachment, nil
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

// ==================== Delete Methods ====================

// DeleteMessage удаляет сообщение и отправляет SSE событие участникам чата
func (s *ChatService) DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) error {
	message, err := s.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	room, err := s.chatRepo.GetRoomByID(ctx, message.RoomID)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}

	if !room.IsParticipant(userID) {
		return repository.ErrUnauthorized
	}

	if err := s.chatRepo.SoftDeleteMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if s.sseManager != nil {
		event := models.MessageDeletedEvent(message.RoomID, messageID)
		sseEvent := sse.EventUUID{
			Type: event.Type,
			Data: event.Data,
		}
		s.sseManager.SendToChat(message.RoomID, sseEvent, uuid.Nil)
	}

	return nil
}

// ==================== Admin Methods ====================

// GetAllChats возвращает все чаты для админ-панели
func (s *ChatService) GetAllChats(ctx context.Context) ([]repository.ChatRoomWithDetails, error) {
	rooms, err := s.chatRepo.ListAllRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list all chats: %w", err)
	}
	return rooms, nil
}
