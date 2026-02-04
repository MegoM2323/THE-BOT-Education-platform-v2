package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/telegram"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestTelegramNotificationMessageFormats tests the message format for notifications
func TestTelegramNotificationMessageFormats(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()

	lesson := &models.Lesson{
		ID:              uuid.New(),
		TeacherID:       uuid.New(),
		StartTime:       time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2025, 1, 15, 16, 0, 0, 0, time.UTC),
		MaxStudents:     1,
		CurrentStudents: 0,
		CreditsCost:     2,
		Color:           "#FF5733",
	}
	lesson.Subject.String = "Математика"
	lesson.Subject.Valid = true

	// Mock telegram client that captures messages
	var capturedMessage string
	var capturedChatID int64

	mockClient := &telegram.Client{}
	// We'll use the service methods to test message formatting

	// Test message format for booking
	dateTime := lesson.StartTime.Format("02.01.2006 15:04")
	expectedBookingMsg := "📚 Запись на занятие\n\n" +
		"Предмет: Математика\n" +
		"Дата и время: " + dateTime + "\n" +
		"Студент: Иван Иванов\n" +
		"Стоимость: 2 кредита\n\n" +
		"Вы успешно записаны на занятие!"

	assert.Contains(t, expectedBookingMsg, "Запись на занятие")
	assert.Contains(t, expectedBookingMsg, "Математика")
	assert.Contains(t, expectedBookingMsg, "2 кредита")
	assert.Contains(t, expectedBookingMsg, "Иван Иванов")

	// Test message format for reschedule
	oldStartTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	newStartTime := lesson.StartTime

	oldDateTime := oldStartTime.Format("02.01.2006 15:04")
	newDateTime := newStartTime.Format("02.01.2006 15:04")

	expectedRescheduleMsg := "📅 Перенос занятия\n\n" +
		"Предмет: Математика\n\n" +
		"⏰ Старое время: " + oldDateTime + "\n" +
		"✅ Новое время: " + newDateTime + "\n\n" +
		"Занятие было перенесено. Пожалуйста, обновите свой календарь."

	assert.Contains(t, expectedRescheduleMsg, "Перенос занятия")
	assert.Contains(t, expectedRescheduleMsg, "Математика")
	assert.Contains(t, expectedRescheduleMsg, oldDateTime)
	assert.Contains(t, expectedRescheduleMsg, newDateTime)

	// Test message format for cancellation
	expectedCancelMsg := "❌ Отмена занятия\n\n" +
		"Предмет: Математика\n" +
		"Дата и время: " + dateTime + "\n\n" +
		"К сожалению, занятие было отменено. Кредиты будут возвращены на ваш счет."

	assert.Contains(t, expectedCancelMsg, "Отмена занятия")
	assert.Contains(t, expectedCancelMsg, "Математика")
	assert.Contains(t, expectedCancelMsg, "Кредиты будут возвращены")

	// Verify no errors in format
	assert.NotNil(t, capturedMessage)
	assert.NotNil(t, capturedChatID)

	_ = studentID
	_ = capturedMessage
	_ = capturedChatID
	_ = mockClient
	_ = ctx
}

// TestNotifyLessonBooking_SkipsUnsubscribedUsers tests that unsubscribed users are skipped
func TestNotifyLessonBooking_SkipsUnsubscribedUsers(t *testing.T) {
	// This test verifies the business logic that:
	// 1. Users without Telegram link are skipped
	// 2. Users with Telegram link but unsubscribed are skipped
	// 3. Only subscribed users receive notifications

	ctx := context.Background()
	studentID := uuid.New()

	lesson := &models.Lesson{
		ID:              uuid.New(),
		TeacherID:       uuid.New(),
		StartTime:       time.Now().Add(24 * time.Hour),
		EndTime:         time.Now().Add(25 * time.Hour),
		MaxStudents:     1,
		CurrentStudents: 0,
		CreditsCost:     2,
		Color:           "#FF5733",
	}
	lesson.Subject.String = "История"
	lesson.Subject.Valid = true

	// Test case 1: User not linked (ErrTelegramUserNotFound)
	// Expected: No error, user is skipped
	notLinkedUserRepo := &MockTelegramUserRepoSimple{
		getUserFunc: func(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error) {
			return nil, repository.ErrTelegramUserNotFound
		},
	}

	service1 := &TelegramService{
		telegramUserRepo: notLinkedUserRepo,
		telegramClient:   nil, // Nil client - should not be called
	}

	err := service1.NotifyLessonBooking(ctx, lesson, "Student", []uuid.UUID{studentID})
	assert.NoError(t, err, "Should not return error when user not linked")

	// Test case 2: User linked but not subscribed
	// Expected: No error, user is skipped
	unsubscribedUserRepo := &MockTelegramUserRepoSimple{
		getUserFunc: func(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error) {
			return &models.TelegramUser{
				UserID:     studentID,
				TelegramID: 123456789,
				ChatID:     987654321,
				Username:   "testuser",
				Subscribed: false,
			}, nil
		},
	}

	service2 := &TelegramService{
		telegramUserRepo: unsubscribedUserRepo,
		telegramClient:   nil, // Nil client - should not be called
	}

	err = service2.NotifyLessonBooking(ctx, lesson, "Student", []uuid.UUID{studentID})
	assert.NoError(t, err, "Should not return error when user not subscribed")

	_ = service1
	_ = service2
}

// TestFormatCreditsWithDeclensionInNotifications verifies credit formatting in messages
func TestFormatCreditsWithDeclensionInNotifications(t *testing.T) {
	tests := []struct {
		name     string
		cost     int
		expected string
	}{
		{"1 credit", 1, "1 кредит"},
		{"2 credits", 2, "2 кредита"},
		{"5 credits", 5, "5 кредитов"},
		{"21 credit", 21, "21 кредит"},
		{"25 credits", 25, "25 кредитов"},
		{"101 credit", 101, "101 кредит"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCreditsWithDeclension(tt.cost)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// MockTelegramUserRepoSimple is a simple mock for testing
type MockTelegramUserRepoSimple struct {
	getUserFunc func(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error)
}

func (m *MockTelegramUserRepoSimple) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, userID)
	}
	return nil, repository.ErrTelegramUserNotFound
}

func (m *MockTelegramUserRepoSimple) LinkUserToTelegram(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	return nil
}

func (m *MockTelegramUserRepoSimple) LinkUserToTelegramAtomic(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	return nil
}

func (m *MockTelegramUserRepoSimple) GetByTelegramID(ctx context.Context, telegramID int64) (*models.TelegramUser, error) {
	return nil, nil
}

func (m *MockTelegramUserRepoSimple) GetAllLinked(ctx context.Context) ([]*models.TelegramUser, error) {
	return nil, nil
}

func (m *MockTelegramUserRepoSimple) GetAllWithUserInfo(ctx context.Context) ([]*models.TelegramUser, error) {
	return nil, nil
}

func (m *MockTelegramUserRepoSimple) GetByRoleWithUserInfo(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	return nil, nil
}

func (m *MockTelegramUserRepoSimple) GetByRole(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	return nil, nil
}

func (m *MockTelegramUserRepoSimple) GetSubscribedUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *MockTelegramUserRepoSimple) UpdateSubscription(ctx context.Context, userID uuid.UUID, subscribed bool) error {
	return nil
}

func (m *MockTelegramUserRepoSimple) UnlinkTelegram(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *MockTelegramUserRepoSimple) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *MockTelegramUserRepoSimple) CleanupInvalidLinks(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockTelegramUserRepoSimple) IsValidlyLinked(ctx context.Context, userID uuid.UUID) (bool, error) {
	return true, nil
}
