package service

import (
	"context"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/validator"
)

// TrialRequestService обрабатывает бизнес-логику для запросов на пробные уроки
type TrialRequestService struct {
	repo            *repository.TrialRequestRepository
	validator       *validator.TrialRequestValidator
	telegramService *TelegramService
}

// NewTrialRequestService создает новый TrialRequestService
func NewTrialRequestService(
	repo *repository.TrialRequestRepository,
	validator *validator.TrialRequestValidator,
	telegramService *TelegramService,
) *TrialRequestService {
	return &TrialRequestService{
		repo:            repo,
		validator:       validator,
		telegramService: telegramService,
	}
}

// SetTelegramService устанавливает TelegramService после инициализации
func (s *TrialRequestService) SetTelegramService(telegramService *TelegramService) {
	s.telegramService = telegramService
}

// CreateTrialRequest создает новый запрос на пробный урок
func (s *TrialRequestService) CreateTrialRequest(ctx context.Context, input *models.CreateTrialRequestInput) (*models.TrialRequest, error) {
	// Проверяем входные данные
	if err := s.validator.Validate(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Создаем запрос на пробный урок
	trialRequest, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create trial request: %w", err)
	}

	// Отправляем уведомление админу в горутине (не блокируем основной поток)
	go s.sendAdminNotification(trialRequest)

	return trialRequest, nil
}

// sendAdminNotification отправляет уведомление админу о новой заявке
func (s *TrialRequestService) sendAdminNotification(trialRequest *models.TrialRequest) {
	// Проверяем, что TelegramService инициализирован
	if s.telegramService == nil {
		fmt.Printf("Info: Telegram service not configured, skipping admin notification for trial request %d\n", trialRequest.ID)
		return
	}

	// Используем новый контекст с таймаутом для отправки уведомления
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Форматируем сообщение
	message := s.formatTrialRequestMessage(trialRequest)

	// Отправляем уведомление
	if err := s.telegramService.SendAdminNotification(ctx, message); err != nil {
		// Логируем ошибку, но не прерываем процесс
		// SendAdminNotification уже логирует ошибку внутри
		fmt.Printf("Warning: Failed to send admin notification for trial request %d: %v\n", trialRequest.ID, err)
	} else {
		fmt.Printf("Successfully sent admin notification for trial request %d\n", trialRequest.ID)
	}
}

// formatTrialRequestMessage форматирует сообщение о новой заявке
func (s *TrialRequestService) formatTrialRequestMessage(tr *models.TrialRequest) string {
	message := "🆕 Новая заявка на пробное занятие\n\n"
	message += fmt.Sprintf("👤 Имя: %s\n", tr.Name)

	if tr.Email != nil && *tr.Email != "" {
		message += fmt.Sprintf("📧 Email: %s\n", *tr.Email)
	}

	message += fmt.Sprintf("📱 Телефон: %s\n", tr.Phone)
	message += fmt.Sprintf("💬 Telegram: @%s\n", tr.Telegram)
	message += fmt.Sprintf("📅 Дата заявки: %s\n", tr.CreatedAt.Format("02.01.2006 15:04"))
	message += fmt.Sprintf("\n🆔 ID заявки: %d", tr.ID)

	return message
}

// GetAllTrialRequests получает все запросы на пробные уроки
func (s *TrialRequestService) GetAllTrialRequests(ctx context.Context) ([]*models.TrialRequest, error) {
	requests, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trial requests: %w", err)
	}

	return requests, nil
}

// GetTrialRequestByID получает запрос на пробный урок по ID
func (s *TrialRequestService) GetTrialRequestByID(ctx context.Context, id int64) (*models.TrialRequest, error) {
	request, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get trial request: %w", err)
	}

	return request, nil
}
