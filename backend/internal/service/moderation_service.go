package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
	"tutoring-platform/internal/models"
)

// chatRepository интерфейс для работы с чат сообщениями
type chatRepository interface {
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*models.Message, error)
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) error
	CreateBlockedMessage(ctx context.Context, blockedMsg *models.BlockedMessage) error
}

// ModerationService управляет модерацией сообщений
type ModerationService struct {
	openRouterClient *OpenRouterClient
	chatRepo         chatRepository
	telegramService  *TelegramService
	regexFallback    *RegexModerator
	circuitBreaker   *CircuitBreaker
}

// CircuitBreaker управляет состоянием circuit breaker для защиты от cascading failures
type CircuitBreaker struct {
	mu                sync.RWMutex
	failureCount      int
	lastFailureTime   time.Time
	state             string // "closed", "open", "half-open"
	failureThreshold  int
	recoveryTimeout   time.Duration
	halfOpenSuccesses int
	halfOpenThreshold int
}

const (
	circuitStateClosed   = "closed"
	circuitStateOpen     = "open"
	circuitStateHalfOpen = "half-open"

	// После 5 последовательных ошибок переключаемся на regex-only
	defaultFailureThreshold = 5
	// Через 5 минут пробуем восстановить соединение
	defaultRecoveryTimeout = 5 * time.Minute
	// Требуется 2 успешных запроса для восстановления
	defaultHalfOpenThreshold = 2
)

// NewCircuitBreaker создает новый circuit breaker
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:             circuitStateClosed,
		failureThreshold:  defaultFailureThreshold,
		recoveryTimeout:   defaultRecoveryTimeout,
		halfOpenThreshold: defaultHalfOpenThreshold,
	}
}

// IsOpen проверяет открыт ли circuit breaker
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == circuitStateOpen {
		// Проверяем не истек ли timeout для восстановления
		if time.Since(cb.lastFailureTime) >= cb.recoveryTimeout {
			return false // Переходим в half-open
		}
		return true
	}

	return false
}

// RecordSuccess записывает успешный вызов
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == circuitStateHalfOpen {
		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.halfOpenThreshold {
			// Восстанавливаем circuit
			cb.state = circuitStateClosed
			cb.failureCount = 0
			cb.halfOpenSuccesses = 0
			log.Println("log")
		}
	} else if cb.state == circuitStateClosed {
		// Сбрасываем счетчик ошибок при успехе
		cb.failureCount = 0
	}
}

// RecordFailure записывает неудачный вызов
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == circuitStateClosed && cb.failureCount >= cb.failureThreshold {
		// Открываем circuit
		cb.state = circuitStateOpen
		log.Println("log")
	} else if cb.state == circuitStateHalfOpen {
		// Если ошибка в half-open состоянии, возвращаемся в open
		cb.state = circuitStateOpen
		cb.halfOpenSuccesses = 0
		log.Println("log")
	}
}

// TryTransitionToHalfOpen пытается перейти в half-open состояние
func (cb *CircuitBreaker) TryTransitionToHalfOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == circuitStateOpen && time.Since(cb.lastFailureTime) >= cb.recoveryTimeout {
		cb.state = circuitStateHalfOpen
		cb.halfOpenSuccesses = 0
		log.Println("log")
		return true
	}

	return false
}

// NewModerationService создает новый сервис модерации
func NewModerationService(
	openRouterClient *OpenRouterClient,
	chatRepo chatRepository,
	telegramService *TelegramService,
) *ModerationService {
	return &ModerationService{
		openRouterClient: openRouterClient,
		chatRepo:         chatRepo,
		telegramService:  telegramService,
		regexFallback:    NewRegexModerator(),
		circuitBreaker:   NewCircuitBreaker(),
	}
}

// ModerateMessageAsync асинхронно модерирует сообщение
func (s *ModerationService) ModerateMessageAsync(ctx context.Context, messageID uuid.UUID) {
	go func() {
		// Используем background context для асинхронной операции
		bgCtx := context.Background()

		// Получить сообщение из БД
		message, err := s.chatRepo.GetMessageByID(bgCtx, messageID)
		if err != nil {
			log.Println("log")
			return
		}

		var blocked bool
		var reason string
		usedFallback := false

		// Проверить circuit breaker или отсутствие OpenRouter клиента
		if s.openRouterClient == nil || s.circuitBreaker.IsOpen() {
			if s.openRouterClient == nil {
				log.Println("log")
			} else {
				log.Println("log")
			}
			blocked, reason = s.regexFallback.Check(message.MessageText)
			usedFallback = true
		} else {
			// Попробовать перейти в half-open если нужно
			s.circuitBreaker.TryTransitionToHalfOpen()

			// Попытка AI модерации через OpenRouter
			result, err := s.openRouterClient.ModerateMessage(bgCtx, message.MessageText)
			if err != nil {
				log.Println("log")
				s.circuitBreaker.RecordFailure()

				// Fallback на regex модерацию
				blocked, reason = s.regexFallback.Check(message.MessageText)
				usedFallback = true
				reason = fmt.Sprintf("Regex fallback: %s (OpenRouter unavailable)", reason)
			} else {
				s.circuitBreaker.RecordSuccess()
				blocked = result.Blocked
				reason = result.Reason
			}
		}

		now := time.Now()

		if blocked {
			// Обновить статус сообщения на 'blocked'
			if err := s.chatRepo.UpdateMessageStatus(bgCtx, messageID, models.MessageStatusBlocked); err != nil {
				log.Println("log")
				return
			}

			// Сохранить в blocked_messages
			blockedMsg := &models.BlockedMessage{
				ID:        uuid.New(),
				MessageID: messageID,
				Reason:    reason,
				AIResponse: map[string]interface{}{
					"blocked":       blocked,
					"reason":        reason,
					"used_fallback": usedFallback,
					"moderated_at":  now,
					"circuit_state": s.getCircuitState(),
				},
				BlockedAt:     now,
				AdminNotified: false,
				AdminReviewed: false,
			}

			if err := s.chatRepo.CreateBlockedMessage(bgCtx, blockedMsg); err != nil {
				log.Println("log")
			}

			// Уведомить админа через Telegram
			s.notifyAdminAboutBlockedMessage(message, reason)

		} else {
			// Обновить статус на 'delivered'
			if err := s.chatRepo.UpdateMessageStatus(bgCtx, messageID, models.MessageStatusDelivered); err != nil {
				log.Println("log")
			}
		}
	}()
}

// notifyAdminAboutBlockedMessage отправляет уведомление админу о заблокированном сообщении
func (s *ModerationService) notifyAdminAboutBlockedMessage(message *models.Message, reason string) {
	if s.telegramService == nil {
		log.Println("log")
		return
	}

	// Формируем сообщение для админа
	text := fmt.Sprintf(
		"🚫 *Сообщение заблокировано модерацией*\n\n"+
			"*ID сообщения:* %s\n"+
			"*Отправитель:* %s\n"+
			"*Текст:* %s\n\n"+
			"*Причина блокировки:* %s",
		message.ID,
		message.SenderID,
		truncateText(message.MessageText, 100),
		reason,
	)

	ctx := context.Background()
	if err := s.telegramService.SendAdminNotification(ctx, text); err != nil {
		log.Println("log")
	}
}

// truncateText обрезает текст до указанной длины
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// getCircuitState возвращает текущее состояние circuit breaker
func (s *ModerationService) getCircuitState() string {
	s.circuitBreaker.mu.RLock()
	defer s.circuitBreaker.mu.RUnlock()
	return s.circuitBreaker.state
}

// GetCircuitBreakerStatus возвращает детальную информацию о circuit breaker (для мониторинга)
func (s *ModerationService) GetCircuitBreakerStatus() map[string]interface{} {
	s.circuitBreaker.mu.RLock()
	defer s.circuitBreaker.mu.RUnlock()

	return map[string]interface{}{
		"state":               s.circuitBreaker.state,
		"failure_count":       s.circuitBreaker.failureCount,
		"last_failure_time":   s.circuitBreaker.lastFailureTime,
		"half_open_successes": s.circuitBreaker.halfOpenSuccesses,
	}
}
