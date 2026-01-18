package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

// Mock chatRepository для тестирования
type mockChatRepository struct {
	getMessage    func(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
	updateStatus  func(ctx context.Context, messageID uuid.UUID, status string) error
	createBlocked func(ctx context.Context, blockedMsg *models.BlockedMessage) error
}

func (m *mockChatRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error) {
	if m.getMessage != nil {
		return m.getMessage(ctx, messageID)
	}
	return nil, nil
}

func (m *mockChatRepository) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) error {
	if m.updateStatus != nil {
		return m.updateStatus(ctx, messageID, status)
	}
	return nil
}

func (m *mockChatRepository) CreateBlockedMessage(ctx context.Context, blockedMsg *models.BlockedMessage) error {
	if m.createBlocked != nil {
		return m.createBlocked(ctx, blockedMsg)
	}
	return nil
}

// TestRegexModerator_CheckPhoneNumbers проверяет обнаружение телефонных номеров
func TestRegexModerator_CheckPhoneNumbers(t *testing.T) {
	moderator := NewRegexModerator()

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Phone with +7",
			message:  "Мой номер +79991234567",
			expected: true,
		},
		{
			name:     "Phone with 8",
			message:  "Звони 89991234567",
			expected: true,
		},
		{
			name:     "Phone with spaces",
			message:  "Телефон: +7 999 123 45 67",
			expected: true,
		},
		{
			name:     "Phone with dashes",
			message:  "Номер: 8-999-123-45-67",
			expected: true,
		},
		{
			name:     "Phone with parentheses",
			message:  "Позвони +7(999)123-45-67",
			expected: true,
		},
		{
			name:     "Clean message",
			message:  "Привет, как дела?",
			expected: false,
		},
		{
			name:     "Message with numbers but not phone",
			message:  "У меня 5 яблок",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, _ := moderator.Check(tt.message)
			if blocked != tt.expected {
				t.Errorf("Check(%q) = %v, want %v", tt.message, blocked, tt.expected)
			}
		})
	}
}

// TestRegexModerator_CheckEmail проверяет обнаружение email адресов
func TestRegexModerator_CheckEmail(t *testing.T) {
	moderator := NewRegexModerator()

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Simple email",
			message:  "Пиши мне на test@example.com",
			expected: true,
		},
		{
			name:     "Email with dots",
			message:  "Контакт: first.last@company.org",
			expected: true,
		},
		{
			name:     "Email in text",
			message:  "Связаться можно по адресу user123@mail.ru хорошо?",
			expected: true,
		},
		{
			name:     "Clean message",
			message:  "Привет! Как настроение?",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, _ := moderator.Check(tt.message)
			if blocked != tt.expected {
				t.Errorf("Check(%q) = %v, want %v", tt.message, blocked, tt.expected)
			}
		})
	}
}

// TestRegexModerator_CheckSocialMedia проверяет обнаружение соцсетей
func TestRegexModerator_CheckSocialMedia(t *testing.T) {
	moderator := NewRegexModerator()

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "VK mention",
			message:  "Добавляйся в вконтакте",
			expected: true,
		},
		{
			name:     "VK link",
			message:  "Мой профиль vk.com/username",
			expected: true,
		},
		{
			name:     "Instagram",
			message:  "Подпишись на мой instagram",
			expected: true,
		},
		{
			name:     "Telegram username",
			message:  "Пиши в телеграм @username",
			expected: true,
		},
		{
			name:     "WhatsApp",
			message:  "Есть whatsapp?",
			expected: true,
		},
		{
			name:     "Clean message",
			message:  "Давай обсудим домашнее задание",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, _ := moderator.Check(tt.message)
			if blocked != tt.expected {
				t.Errorf("Check(%q) = %v, want %v", tt.message, blocked, tt.expected)
			}
		})
	}
}

// TestRegexModerator_CheckPlatforms проверяет обнаружение платформ для видеозвонков
func TestRegexModerator_CheckPlatforms(t *testing.T) {
	moderator := NewRegexModerator()

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Zoom",
			message:  "Давай созвонимся в Zoom",
			expected: true,
		},
		{
			name:     "Skype",
			message:  "Есть скайп?",
			expected: true,
		},
		{
			name:     "Discord",
			message:  "Пиши в discord",
			expected: true,
		},
		{
			name:     "Google Meet",
			message:  "Встретимся в google meet",
			expected: true,
		},
		{
			name:     "Clean message",
			message:  "Когда следующее занятие?",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, _ := moderator.Check(tt.message)
			if blocked != tt.expected {
				t.Errorf("Check(%q) = %v, want %v", tt.message, blocked, tt.expected)
			}
		})
	}
}

// TestRegexModerator_CheckObfuscatedPhone проверяет обнаружение обфусцированных телефонов
func TestRegexModerator_CheckObfuscatedPhone(t *testing.T) {
	moderator := NewRegexModerator()

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Obfuscated phone",
			message:  "Мой номер восемь девять девять один...",
			expected: true,
		},
		{
			name:     "Written digits",
			message:  "Позвони семь девять два три...",
			expected: true,
		},
		{
			name:     "Clean message with numbers",
			message:  "Сегодня среда, завтра четверг",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocked, _ := moderator.Check(tt.message)
			if blocked != tt.expected {
				t.Errorf("Check(%q) = %v, want %v", tt.message, blocked, tt.expected)
			}
		})
	}
}

// TestCircuitBreaker_RecordFailures проверяет работу circuit breaker
func TestCircuitBreaker_RecordFailures(t *testing.T) {
	cb := NewCircuitBreaker()

	// Изначально circuit закрыт
	if cb.IsOpen() {
		t.Error("Circuit breaker should be closed initially")
	}

	// Записываем 4 ошибки - circuit еще закрыт
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
	}

	if cb.IsOpen() {
		t.Error("Circuit breaker should still be closed after 4 failures")
	}

	// 5-я ошибка открывает circuit
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Error("Circuit breaker should be open after 5 failures")
	}

	// Проверяем что состояние Open
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	if state != circuitStateOpen {
		t.Errorf("Circuit state = %s, want %s", state, circuitStateOpen)
	}
}

// TestCircuitBreaker_Recovery проверяет восстановление circuit breaker
func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := NewCircuitBreaker()
	cb.recoveryTimeout = 100 * time.Millisecond // Короткий timeout для теста

	// Открываем circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if !cb.IsOpen() {
		t.Fatal("Circuit should be open")
	}

	// Ждем recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Пытаемся перейти в half-open
	transitioned := cb.TryTransitionToHalfOpen()
	if !transitioned {
		t.Error("Should transition to half-open after recovery timeout")
	}

	// Проверяем что состояние half-open
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	if state != circuitStateHalfOpen {
		t.Errorf("Circuit state = %s, want %s", state, circuitStateHalfOpen)
	}

	// Записываем 2 успешных запроса
	cb.RecordSuccess()
	cb.RecordSuccess()

	// Circuit должен закрыться
	cb.mu.RLock()
	state = cb.state
	cb.mu.RUnlock()

	if state != circuitStateClosed {
		t.Errorf("Circuit state = %s, want %s after recovery", state, circuitStateClosed)
	}

	// IsOpen должен вернуть false
	if cb.IsOpen() {
		t.Error("Circuit should be closed after successful recovery")
	}
}

// TestModerationService_ModerateMessageAsync тестирует асинхронную модерацию
func TestModerationService_ModerateMessageAsync(t *testing.T) {
	messageID := uuid.New()
	senderID := uuid.New()
	roomID := uuid.New()

	testMessage := &models.Message{
		ID:          messageID,
		RoomID:      roomID,
		SenderID:    senderID,
		MessageText: "Привет, как дела? Когда следующий урок?",
		Status:      models.MessageStatusPendingModeration,
		CreatedAt:   time.Now(),
	}

	statusUpdated := false
	var updatedStatus string

	mockRepo := &mockChatRepository{
		getMessage: func(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
			return testMessage, nil
		},
		updateStatus: func(ctx context.Context, msgID uuid.UUID, status string) error {
			statusUpdated = true
			updatedStatus = status
			return nil
		},
	}

	// Создаем service без OpenRouter (используется fallback regex)
	service := NewModerationService(nil, mockRepo, nil)

	// Запускаем асинхронную модерацию
	service.ModerateMessageAsync(context.Background(), messageID)

	// Ждем завершения goroutine
	time.Sleep(100 * time.Millisecond)

	// Проверяем что статус был обновлен
	if !statusUpdated {
		t.Error("Message status should be updated")
	}

	// Проверяем что сообщение было доставлено (нет запрещенного контента)
	if updatedStatus != models.MessageStatusDelivered {
		t.Errorf("Message status = %s, want %s", updatedStatus, models.MessageStatusDelivered)
	}
}

// TestModerationService_BlockMessage тестирует блокировку сообщения
func TestModerationService_BlockMessage(t *testing.T) {
	messageID := uuid.New()
	senderID := uuid.New()
	roomID := uuid.New()

	testMessage := &models.Message{
		ID:          messageID,
		RoomID:      roomID,
		SenderID:    senderID,
		MessageText: "Мой номер +79991234567, позвони",
		Status:      models.MessageStatusPendingModeration,
		CreatedAt:   time.Now(),
	}

	statusUpdated := false
	var updatedStatus string
	blockedCreated := false

	mockRepo := &mockChatRepository{
		getMessage: func(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
			return testMessage, nil
		},
		updateStatus: func(ctx context.Context, msgID uuid.UUID, status string) error {
			statusUpdated = true
			updatedStatus = status
			return nil
		},
		createBlocked: func(ctx context.Context, blockedMsg *models.BlockedMessage) error {
			blockedCreated = true
			return nil
		},
	}

	// Создаем service без OpenRouter (используется fallback regex)
	service := NewModerationService(nil, mockRepo, nil)

	// Запускаем асинхронную модерацию
	service.ModerateMessageAsync(context.Background(), messageID)

	// Ждем завершения goroutine
	time.Sleep(100 * time.Millisecond)

	// Проверяем что статус был обновлен
	if !statusUpdated {
		t.Error("Message status should be updated")
	}

	// Проверяем что сообщение было заблокировано
	if updatedStatus != models.MessageStatusBlocked {
		t.Errorf("Message status = %s, want %s", updatedStatus, models.MessageStatusBlocked)
	}

	// Проверяем что запись о блокировке создана
	if !blockedCreated {
		t.Error("Blocked message record should be created")
	}
}
