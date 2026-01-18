package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/telegram"
)

var (
	// ErrBroadcastAlreadyInProgress возвращается при попытке повторного запуска рассылки
	ErrBroadcastAlreadyInProgress = errors.New("broadcast already in progress or completed")
	// ErrBroadcastCannotCancel возвращается при попытке отмены завершенной рассылки
	ErrBroadcastCannotCancel = errors.New("cannot cancel broadcast in current status")
	// ErrUsersNotLinkedToTelegram возвращается когда пользователи не имеют привязки Telegram
	ErrUsersNotLinkedToTelegram = errors.New("some users are not linked to telegram")
	// ErrTelegramNotConfigured возвращается при попытке отправки без настроенного Telegram бота
	ErrTelegramNotConfigured = errors.New("telegram bot is not configured")
	// ErrInvalidFilePath возвращается при попытке использовать небезопасный путь к файлу
	ErrInvalidFilePath = errors.New("invalid file path: contains path traversal attempts or absolute paths")
)

// BroadcastService управляет массовыми рассылками через Telegram
type BroadcastService struct {
	broadcastRepo     repository.BroadcastRepository
	broadcastListRepo repository.BroadcastListRepository
	telegramUserRepo  repository.TelegramUserRepository
	userRepo          repository.UserRepository
	telegramClient    *telegram.Client
	rateLimiter       *time.Ticker
	mu                sync.Mutex
	// Отслеживаем контексты запущенных горутин для возможности отмены
	activeContexts map[string]context.CancelFunc
	contextsMu     sync.RWMutex
}

// NewBroadcastService создает новый BroadcastService с настройкой rate limiting
func NewBroadcastService(
	broadcastRepo repository.BroadcastRepository,
	broadcastListRepo repository.BroadcastListRepository,
	telegramUserRepo repository.TelegramUserRepository,
	userRepo repository.UserRepository,
	telegramClient *telegram.Client,
) *BroadcastService {
	// Telegram лимит: 30 сообщений в секунду
	// Используем ~35ms интервал (28 сообщений в секунду) для безопасности
	rateLimiter := time.NewTicker(35 * time.Millisecond)

	return &BroadcastService{
		broadcastRepo:     broadcastRepo,
		broadcastListRepo: broadcastListRepo,
		telegramUserRepo:  telegramUserRepo,
		userRepo:          userRepo,
		telegramClient:    telegramClient,
		rateLimiter:       rateLimiter,
		activeContexts:    make(map[string]context.CancelFunc),
	}
}

// Shutdown останавливает rate limiter и отменяет все активные рассылки
func (s *BroadcastService) Shutdown() {
	s.rateLimiter.Stop()

	// Отменяем все активные контексты горутин рассылок
	s.contextsMu.Lock()
	for _, cancel := range s.activeContexts {
		cancel()
	}
	s.activeContexts = make(map[string]context.CancelFunc)
	s.contextsMu.Unlock()

	log.Println("log")
}

// CreateBroadcastList создает новый список рассылки
func (s *BroadcastService) CreateBroadcastList(
	ctx context.Context,
	name, description string,
	userIDs []uuid.UUID,
	createdBy uuid.UUID,
) (*models.BroadcastList, error) {
	// Валидация входных данных
	if len(name) < 3 {
		return nil, models.ErrInvalidBroadcastName
	}
	if len(userIDs) < 1 {
		return nil, models.ErrInvalidBroadcastUsers
	}

	// Проверяем что все пользователи существуют и имеют привязанный Telegram
	for _, userID := range userIDs {
		// Проверяем существование пользователя
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return nil, fmt.Errorf("user %s not found", userID)
			}
			return nil, fmt.Errorf("failed to check user %s: %w", userID, err)
		}

		// Проверяем что пользователь не удален
		if user.IsDeleted() {
			return nil, fmt.Errorf("user %s is deleted", userID)
		}

		// Проверяем привязку к Telegram
		_, err = s.telegramUserRepo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrTelegramUserNotFound) {
				return nil, fmt.Errorf("user %s is not linked to telegram: %w", userID, ErrUsersNotLinkedToTelegram)
			}
			return nil, fmt.Errorf("failed to check telegram link for user %s: %w", userID, err)
		}
	}

	// Создаем список рассылки
	list := &models.BroadcastList{
		Name:        name,
		Description: description,
		UserIDs:     userIDs,
		CreatedBy:   createdBy,
	}

	createdList, err := s.broadcastListRepo.Create(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast list: %w", err)
	}

	log.Println("log")
	return createdList, nil
}

// GetBroadcastLists получает все активные списки рассылки
func (s *BroadcastService) GetBroadcastLists(ctx context.Context) ([]*models.BroadcastList, error) {
	lists, err := s.broadcastListRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast lists: %w", err)
	}

	return lists, nil
}

// GetBroadcastListByID получает список рассылки по ID
func (s *BroadcastService) GetBroadcastListByID(ctx context.Context, id uuid.UUID) (*models.BroadcastList, error) {
	list, err := s.broadcastListRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastListNotFound) {
			return nil, repository.ErrBroadcastListNotFound
		}
		return nil, fmt.Errorf("failed to get broadcast list: %w", err)
	}

	return list, nil
}

// UpdateBroadcastList обновляет список рассылки
func (s *BroadcastService) UpdateBroadcastList(
	ctx context.Context,
	id uuid.UUID,
	name, description string,
	userIDs []uuid.UUID,
) error {
	// Проверяем существование списка
	_, err := s.broadcastListRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastListNotFound) {
			return repository.ErrBroadcastListNotFound
		}
		return fmt.Errorf("failed to check broadcast list: %w", err)
	}

	// Валидация
	if len(name) < 3 {
		return models.ErrInvalidBroadcastName
	}

	// Проверяем что все новые пользователи имеют Telegram
	if len(userIDs) > 0 {
		for _, userID := range userIDs {
			// Проверяем существование пользователя
			user, err := s.userRepo.GetByID(ctx, userID)
			if err != nil {
				if errors.Is(err, repository.ErrUserNotFound) {
					return fmt.Errorf("user %s not found", userID)
				}
				return fmt.Errorf("failed to check user %s: %w", userID, err)
			}

			// Проверяем что пользователь не удален
			if user.IsDeleted() {
				return fmt.Errorf("user %s is deleted", userID)
			}

			// Проверяем привязку к Telegram
			_, err = s.telegramUserRepo.GetByUserID(ctx, userID)
			if err != nil {
				if errors.Is(err, repository.ErrTelegramUserNotFound) {
					return fmt.Errorf("user %s is not linked to telegram: %w", userID, ErrUsersNotLinkedToTelegram)
				}
				return fmt.Errorf("failed to check telegram link for user %s: %w", userID, err)
			}
		}
	}

	// Обновляем список
	if err := s.broadcastListRepo.Update(ctx, id, name, description, userIDs); err != nil {
		return fmt.Errorf("failed to update broadcast list: %w", err)
	}

	log.Println("log")
	return nil
}

// DeleteBroadcastList выполняет мягкое удаление списка рассылки
func (s *BroadcastService) DeleteBroadcastList(ctx context.Context, id uuid.UUID) error {
	if err := s.broadcastListRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrBroadcastListNotFound) {
			return repository.ErrBroadcastListNotFound
		}
		return fmt.Errorf("failed to delete broadcast list: %w", err)
	}

	log.Println("log")
	return nil
}

// CreateBroadcast создает новую рассылку со статусом "pending" для списка рассылки
func (s *BroadcastService) CreateBroadcast(
	ctx context.Context,
	listID uuid.UUID,
	message string,
	createdBy uuid.UUID,
) (*models.Broadcast, error) {
	// Валидация сообщения
	if message == "" {
		return nil, models.ErrInvalidBroadcastMessage
	}
	if len(message) > 4096 {
		return nil, models.ErrBroadcastMessageTooLong
	}

	// Проверяем существование списка
	_, err := s.broadcastListRepo.GetByID(ctx, listID)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastListNotFound) {
			return nil, repository.ErrBroadcastListNotFound
		}
		return nil, fmt.Errorf("failed to check broadcast list: %w", err)
	}

	// Создаем запись broadcast
	broadcast := &models.Broadcast{
		ListID:    &listID,
		Message:   message,
		Status:    models.BroadcastStatusPending,
		CreatedBy: createdBy,
	}

	createdBroadcast, err := s.broadcastRepo.Create(ctx, broadcast)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	log.Println("log")
	return createdBroadcast, nil
}

// CreateBroadcastForUsers создает новую рассылку для конкретных пользователей (без списка рассылки)
// Пользователи без привязки к Telegram будут пропущены при отправке, но рассылка будет создана
func (s *BroadcastService) CreateBroadcastForUsers(
	ctx context.Context,
	userIDs []uuid.UUID,
	message string,
	createdBy uuid.UUID,
) (*models.Broadcast, error) {
	// Валидация сообщения
	if message == "" {
		return nil, models.ErrInvalidBroadcastMessage
	}
	if len(message) > 4096 {
		return nil, models.ErrBroadcastMessageTooLong
	}

	// Валидация user_ids
	if len(userIDs) == 0 {
		return nil, models.ErrInvalidBroadcastUsers
	}

	// Фильтруем только пользователей с привязанным Telegram
	var validUserIDs []uuid.UUID
	for _, userID := range userIDs {
		// Проверяем существование пользователя
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				log.Printf("[WARN] User %s not found, skipping", userID)
				continue
			}
			return nil, fmt.Errorf("failed to check user %s: %w", userID, err)
		}

		// Проверяем что пользователь не удален
		if user.IsDeleted() {
			log.Printf("[WARN] User %s is deleted, skipping", userID)
			continue
		}

		// Проверяем привязку к Telegram
		_, err = s.telegramUserRepo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrTelegramUserNotFound) {
				log.Printf("[WARN] User %s is not linked to Telegram, skipping", userID)
				continue
			}
			return nil, fmt.Errorf("failed to check telegram link for user %s: %w", userID, err)
		}

		validUserIDs = append(validUserIDs, userID)
	}

	// Если ни один пользователь не имеет привязки к Telegram
	if len(validUserIDs) == 0 {
		return nil, ErrUsersNotLinkedToTelegram
	}

	// Создаем временный список рассылки для этих пользователей
	// Это нужно чтобы использовать существующую логику отправки
	tempList := &models.BroadcastList{
		Name:        fmt.Sprintf("Direct broadcast %s", time.Now().Format("2006-01-02 15:04:05")),
		Description: "Temporary list for direct user broadcast",
		UserIDs:     validUserIDs,
		CreatedBy:   createdBy,
	}

	createdList, err := s.broadcastListRepo.Create(ctx, tempList)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary broadcast list: %w", err)
	}

	// Создаем запись broadcast с временным списком
	broadcast := &models.Broadcast{
		ListID:    &createdList.ID,
		Message:   message,
		Status:    models.BroadcastStatusPending,
		CreatedBy: createdBy,
	}

	createdBroadcast, err := s.broadcastRepo.Create(ctx, broadcast)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	log.Printf("[INFO] Broadcast created for %d users (out of %d requested)", len(validUserIDs), len(userIDs))
	return createdBroadcast, nil
}

// SendBroadcast запускает процесс отправки рассылки в горутине
func (s *BroadcastService) SendBroadcast(ctx context.Context, broadcastID uuid.UUID) error {
	// Проверяем что Telegram клиент настроен
	if s.telegramClient == nil {
		return ErrTelegramNotConfigured
	}

	// Получаем broadcast по ID
	broadcast, err := s.broadcastRepo.GetByID(ctx, broadcastID)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastNotFound) {
			return repository.ErrBroadcastNotFound
		}
		return fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Проверяем статус (должен быть pending)
	if !broadcast.IsPending() {
		return ErrBroadcastAlreadyInProgress
	}

	// Получаем список пользователей
	if broadcast.ListID == nil {
		return fmt.Errorf("broadcast has no list ID")
	}

	list, err := s.broadcastListRepo.GetByID(ctx, *broadcast.ListID)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastListNotFound) {
			return repository.ErrBroadcastListNotFound
		}
		return fmt.Errorf("failed to get broadcast list: %w", err)
	}

	log.Printf("[INFO] SendBroadcast: got list %s with %d users", list.ID, len(list.UserIDs))

	// Обновляем статус на "in_progress"
	if err := s.broadcastRepo.UpdateStatus(ctx, broadcastID, models.BroadcastStatusInProgress); err != nil {
		return fmt.Errorf("failed to update broadcast status: %w", err)
	}

	// Создаем отменяемый контекст для горутины рассылки на основе Background
	// НЕ наследуем от request context, чтобы горутина продолжала работать после завершения HTTP запроса
	broadcastCtx, cancel := context.WithCancel(context.Background())

	// Сохраняем cancel функцию для дальнейшей отмены
	s.contextsMu.Lock()
	s.activeContexts[broadcastID.String()] = cancel
	s.contextsMu.Unlock()

	// Запускаем отправку в горутине с отменяемым контекстом
	go s.processBroadcast(broadcastCtx, broadcast, list)

	log.Println("log")
	return nil
}

// processBroadcast обрабатывает отправку рассылки с rate limiting
func (s *BroadcastService) processBroadcast(ctx context.Context, broadcast *models.Broadcast, list *models.BroadcastList) {
	var sentCount, failedCount int64

	log.Printf("[INFO] processBroadcast started: broadcast_id=%s, list_id=%s, user_count=%d", broadcast.ID, list.ID, len(list.UserIDs))

	// Убеждаемся что контекст очищен после завершения горутины
	defer func() {
		s.contextsMu.Lock()
		delete(s.activeContexts, broadcast.ID.String())
		s.contextsMu.Unlock()
	}()

	// Отправляем сообщения с rate limiting
	for i, userID := range list.UserIDs {
		// Проверяем отмену контекста
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Broadcast %s cancelled after sending %d messages\n", broadcast.ID, atomic.LoadInt64(&sentCount))
			// Используем контекст с timeout для finalizing, так как основной контекст был отменён
			// Это гарантирует что мы сможем записать результаты отмены в БД
			finCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			s.finalizeBroadcast(finCtx, broadcast.ID, int(atomic.LoadInt64(&sentCount)), int(atomic.LoadInt64(&failedCount)), models.BroadcastStatusCancelled)
			cancel()
			return
		default:
		}

		// Проверяем идемпотентность: если сообщение уже успешно доставлено, пропускаем
		alreadyDelivered, err := s.broadcastRepo.HasSuccessfulDelivery(ctx, broadcast.ID, userID)
		if err != nil {
			// Логируем ошибку проверки, но продолжаем обработку
			log.Printf("[WARN] Failed to check delivery status for user %s: %v\n", userID, err)
		} else if alreadyDelivered {
			// Сообщение уже доставлено этому пользователю в этой рассылке
			log.Printf("[INFO] Skipping duplicate delivery for user %s in broadcast %s\n", userID, broadcast.ID)
			continue
		}

		// Получаем привязку Telegram для пользователя
		telegramUser, err := s.telegramUserRepo.GetByUserID(ctx, userID)
		if err != nil {
			atomic.AddInt64(&failedCount, 1)

			// Логируем ошибку
			s.logBroadcastMessage(ctx, broadcast.ID, userID, 0, models.BroadcastLogStatusFailed, err.Error())

			log.Println("log")
			continue
		}

		// Проверяем подписку на уведомления
		if !telegramUser.Subscribed {
			log.Println("log")
			continue
		}

		// Ждем rate limiter
		<-s.rateLimiter.C

		// Отправляем сообщение с retry логикой с поддержкой идемпотентности
		if err := s.sendMessageWithIdempotency(ctx, broadcast.ID, userID, telegramUser.ChatID, telegramUser.TelegramID, broadcast.Message, 3); err != nil {
			atomic.AddInt64(&failedCount, 1)

			// Логируем ошибку
			s.logBroadcastMessage(ctx, broadcast.ID, userID, telegramUser.TelegramID, models.BroadcastLogStatusFailed, err.Error())

			log.Println("log")

			// Проверяем код ошибки
			if telegramErr, ok := err.(*telegram.TelegramError); ok {
				if telegramErr.ErrorCode == 403 {
					// Бот заблокирован пользователем - отписываем от уведомлений
					log.Println("log")
					if updateErr := s.telegramUserRepo.UpdateSubscription(ctx, userID, false); updateErr != nil {
						log.Println("log")
					}
				}
			}
		} else {
			atomic.AddInt64(&sentCount, 1)

			// Логируем успех
			s.logBroadcastMessage(ctx, broadcast.ID, userID, telegramUser.TelegramID, models.BroadcastLogStatusSuccess, "")

			// Логируем прогресс каждые 100 сообщений
			if (i+1)%100 == 0 {
				log.Println("log")
			}
		}
	}

	// Завершаем рассылку
	status := models.BroadcastStatusCompleted
	sentVal := atomic.LoadInt64(&sentCount)
	failedVal := atomic.LoadInt64(&failedCount)
	if failedVal > 0 && sentVal == 0 {
		status = models.BroadcastStatusFailed
	}

	s.finalizeBroadcast(ctx, broadcast.ID, int(sentVal), int(failedVal), status)
	log.Println("log")
}

// sendMessageWithRetry отправляет сообщение с retry логикой для обработки 429
func (s *BroadcastService) sendMessageWithRetry(ctx context.Context, chatID int64, message string, maxRetries int) error {
	// Проверяем что Telegram клиент инициализирован
	if s.telegramClient == nil {
		return errors.New("telegram client is not configured")
	}

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Проверяем контекст перед каждой попыткой
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during message send retry: %w", ctx.Err())
		default:
		}

		err := s.telegramClient.SendMessage(chatID, message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Проверяем тип ошибки
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			switch telegramErr.ErrorCode {
			case 403:
				// Forbidden - бот заблокирован, не ретраим
				return err
			case 400:
				// Bad Request - невалидные данные, не ретраим
				return err
			case 429:
				// Too Many Requests - exponential backoff с поддержкой отмены контекста
				if attempt < maxRetries-1 {
					backoff := time.Duration(1<<attempt) * time.Second
					log.Println("log")
					// Используем context-aware sleep
					select {
					case <-time.After(backoff):
						continue
					case <-ctx.Done():
						return fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
					}
				}
			}
		}

		// Для других ошибок (network errors) - retry с context-aware sleep
		if attempt < maxRetries-1 {
			select {
			case <-time.After(time.Second):
				continue
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry sleep: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// sendMessageWithIdempotency отправляет сообщение с поддержкой идемпотентности
// Идемпотентность гарантирует, что сообщение не будет отправлено дважды при повторе неудачной попытки
// Parameters:
// - ctx: контекст для отмены
// - broadcastID: ID рассылки для отслеживания
// - userID: ID пользователя (для идемпотентности)
// - chatID: Telegram chat ID для отправки
// - telegramID: Telegram user ID для логирования
// - message: текст сообщения
// - maxRetries: максимальное количество попыток
func (s *BroadcastService) sendMessageWithIdempotency(
	ctx context.Context,
	broadcastID uuid.UUID,
	userID uuid.UUID,
	chatID int64,
	telegramID int64,
	message string,
	maxRetries int,
) error {
	// Проверяем что Telegram клиент инициализирован
	if s.telegramClient == nil {
		return errors.New("telegram client is not configured")
	}

	// Получаем или создаем лог доставки для идемпотентности
	existingLog, err := s.broadcastRepo.GetLogByBroadcastAndUser(ctx, broadcastID, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing log: %w", err)
	}

	var logID uuid.UUID
	if existingLog != nil {
		// Лог уже существует - обновляем его при повторной попытке
		logID = existingLog.ID
		// Если уже успешно отправлено, это не должно здесь вызываться
		// (проверка в processBroadcast должна это перехватить)
		if existingLog.Status == models.BroadcastLogStatusSuccess {
			return fmt.Errorf("message already successfully delivered to user %s", userID)
		}
	} else {
		// Создаем новый лог
		newLog := &models.BroadcastLog{
			BroadcastID: broadcastID,
			UserID:      userID,
			TelegramID:  telegramID,
			Status:      models.BroadcastLogStatusFailed, // Изначально неудача
			Error:       "pending",
		}

		if err := s.broadcastRepo.CreateLog(ctx, newLog); err != nil {
			return fmt.Errorf("failed to create broadcast log for idempotency tracking: %w", err)
		}
		logID = newLog.ID
	}

	// Отправляем сообщение с retry логикой
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Проверяем контекст перед каждой попыткой
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during message send: %w", ctx.Err())
		default:
		}

		err := s.telegramClient.SendMessage(chatID, message)
		if err == nil {
			// Успешно отправлено - обновляем статус
			if updateErr := s.broadcastRepo.UpdateLogStatus(ctx, logID, models.BroadcastLogStatusSuccess, ""); updateErr != nil {
				log.Printf("[WARN] Failed to update success status: %v\n", updateErr)
				// Продолжаем несмотря на ошибку обновления (сообщение отправлено)
			}
			return nil
		}

		lastErr = err

		// Проверяем тип ошибки
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			switch telegramErr.ErrorCode {
			case 403:
				// Forbidden - бот заблокирован, не ретраим
				errMsg := fmt.Sprintf("Telegram bot blocked by user: %v", telegramErr)
				if updateErr := s.broadcastRepo.UpdateLogStatus(ctx, logID, models.BroadcastLogStatusFailed, errMsg); updateErr != nil {
					log.Printf("[WARN] Failed to update error status: %v\n", updateErr)
				}
				return err
			case 400:
				// Bad Request - невалидные данные, не ретраим
				errMsg := fmt.Sprintf("Bad request to Telegram API: %v", telegramErr)
				if updateErr := s.broadcastRepo.UpdateLogStatus(ctx, logID, models.BroadcastLogStatusFailed, errMsg); updateErr != nil {
					log.Printf("[WARN] Failed to update error status: %v\n", updateErr)
				}
				return err
			case 429:
				// Too Many Requests - exponential backoff
				if attempt < maxRetries-1 {
					backoff := time.Duration(1<<attempt) * time.Second
					log.Printf("[INFO] Rate limited, retrying after %v\n", backoff)
					select {
					case <-time.After(backoff):
						continue
					case <-ctx.Done():
						return fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
					}
				}
			}
		}

		// Для других ошибок (network errors) - retry
		if attempt < maxRetries-1 {
			log.Printf("[WARN] Failed to send message (attempt %d/%d): %v\n", attempt+1, maxRetries, err)
			select {
			case <-time.After(time.Second):
				continue
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			}
		}
	}

	// Все попытки исчерпаны - обновляем статус с ошибкой
	errMsg := fmt.Sprintf("max retries exceeded: %v", lastErr)
	if updateErr := s.broadcastRepo.UpdateLogStatus(ctx, logID, models.BroadcastLogStatusFailed, errMsg); updateErr != nil {
		log.Printf("[WARN] Failed to update final error status: %v\n", updateErr)
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// logBroadcastMessage логирует отправку сообщения в БД
func (s *BroadcastService) logBroadcastMessage(
	ctx context.Context,
	broadcastID, userID uuid.UUID,
	telegramID int64,
	status, errorMsg string,
) {
	logEntry := &models.BroadcastLog{
		BroadcastID: broadcastID,
		UserID:      userID,
		TelegramID:  telegramID,
		Status:      status,
		Error:       errorMsg,
	}

	if err := s.broadcastRepo.CreateLog(ctx, logEntry); err != nil {
		log.Println("log")
	}
}

// finalizeBroadcast завершает рассылку, обновляя счетчики и статус
func (s *BroadcastService) finalizeBroadcast(
	ctx context.Context,
	broadcastID uuid.UUID,
	sentCount, failedCount int,
	status string,
) {
	// Обновляем счетчики
	if err := s.broadcastRepo.UpdateCounts(ctx, broadcastID, sentCount, failedCount); err != nil {
		log.Println("log")
	}

	// Обновляем статус (completed_at будет установлен автоматически)
	if err := s.broadcastRepo.UpdateStatus(ctx, broadcastID, status); err != nil {
		log.Println("log")
	}
}

// GetBroadcasts получает список рассылок с пагинацией
func (s *BroadcastService) GetBroadcasts(ctx context.Context, limit, offset int) ([]*models.Broadcast, int, error) {
	// Validate pagination params
	if limit < 1 {
		limit = 20 // default
	}
	if limit > 100 {
		limit = 100 // max
	}
	if offset < 0 {
		offset = 0
	}

	broadcasts, total, err := s.broadcastRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get broadcasts: %w", err)
	}

	return broadcasts, total, nil
}

// GetBroadcastDetails получает рассылку и все её логи
func (s *BroadcastService) GetBroadcastDetails(ctx context.Context, id uuid.UUID) (*models.Broadcast, []*models.BroadcastLog, error) {
	// Получаем рассылку
	broadcast, err := s.broadcastRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastNotFound) {
			return nil, nil, repository.ErrBroadcastNotFound
		}
		return nil, nil, fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Получаем логи
	logs, err := s.broadcastRepo.GetLogsByBroadcastID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get broadcast logs: %w", err)
	}

	return broadcast, logs, nil
}

// CancelBroadcast отменяет рассылку и останавливает её горутину
func (s *BroadcastService) CancelBroadcast(ctx context.Context, id uuid.UUID) error {
	// Получаем рассылку
	broadcast, err := s.broadcastRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBroadcastNotFound) {
			return repository.ErrBroadcastNotFound
		}
		return fmt.Errorf("failed to get broadcast: %w", err)
	}

	// Проверяем статус - можно отменить только pending или in_progress
	if !broadcast.IsPending() && !broadcast.IsInProgress() {
		return ErrBroadcastCannotCancel
	}

	// Отменяем контекст если горутина сейчас работает
	s.contextsMu.Lock()
	if cancel, exists := s.activeContexts[id.String()]; exists {
		cancel()
		delete(s.activeContexts, id.String())
	}
	s.contextsMu.Unlock()

	// Обновляем статус на cancelled
	if err := s.broadcastRepo.UpdateStatus(ctx, id, models.BroadcastStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel broadcast: %w", err)
	}

	log.Println("log")
	return nil
}

// ValidateBroadcastFilePath проверяет безопасность пути к файлу для рассылок
// Предотвращает path traversal атаки и доступ к файлам вне разрешенной директории
// Параметры:
// - filePath: путь к файлу для проверки
// - allowedBaseDir: базовая директория, в пределах которой разрешены файлы
// Возвращает ошибку если:
// - Путь содержит ".." (path traversal)
// - Путь является абсолютным
// - Нормализованный путь находится вне allowedBaseDir
func ValidateBroadcastFilePath(filePath, allowedBaseDir string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	if allowedBaseDir == "" {
		return fmt.Errorf("allowed base directory cannot be empty")
	}

	// Проверяем на абсолютные пути - отклоняем их для безопасности
	if filepath.IsAbs(filePath) {
		return ErrInvalidFilePath
	}

	// Проверяем на Windows-стиль абсолютные пути (C:, D:, и т.д.)
	// и UNC пути (\\server\share)
	if len(filePath) >= 2 && filePath[1] == ':' {
		// Windows drive letter (C:, D:, etc.)
		return ErrInvalidFilePath
	}
	if strings.HasPrefix(filePath, "\\\\") {
		// Windows UNC path (\\server\share)
		return ErrInvalidFilePath
	}

	// Проверяем на явные попытки path traversal ("..")
	// Это должно быть отдельным path component, а не просто подстрока (например "..." - это валидное имя файла)
	pathParts := strings.FieldsFunc(filePath, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	for _, part := range pathParts {
		if part == ".." {
			return ErrInvalidFilePath
		}
	}

	// Check for URL-encoded path traversal attempts (e.g., ..%2F..%2Fetc%2Fpasswd)
	decodedPath, err := url.QueryUnescape(filePath)
	if err == nil && decodedPath != filePath {
		// If decoding changed the path, check the decoded version too
		decodedParts := strings.FieldsFunc(decodedPath, func(r rune) bool {
			return r == '/' || r == '\\'
		})
		for _, part := range decodedParts {
			if part == ".." {
				return ErrInvalidFilePath
			}
		}
	}

	// Нормализуем пути для сравнения
	cleanFilePath := filepath.Clean(filePath)
	cleanBaseDir := filepath.Clean(allowedBaseDir)

	// Снова проверяем нормализованный путь на ".."
	// (в некоторых случаях filepath.Clean может преобразовать пути)
	cleanPathParts := strings.FieldsFunc(cleanFilePath, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	for _, part := range cleanPathParts {
		if part == ".." {
			return ErrInvalidFilePath
		}
	}

	// Абсолютный путь может быть результатом Clean() если filePath был абсолютным
	if filepath.IsAbs(cleanFilePath) {
		return ErrInvalidFilePath
	}

	// Проверяем что нормализованный путь находится внутри базовой директории
	fullPath := filepath.Join(cleanBaseDir, cleanFilePath)
	fullPathAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	baseDirAbs, err := filepath.Abs(cleanBaseDir)
	if err != nil {
		return fmt.Errorf("failed to resolve base directory absolute path: %w", err)
	}

	// Нормализуем абсолютные пути для последовательного сравнения
	fullPathAbs = filepath.Clean(fullPathAbs)
	baseDirAbs = filepath.Clean(baseDirAbs)

	// Проверяем что полный путь начинается с базовой директории
	// Используем filepath.Join и filepath.Rel для безопасного сравнения
	rel, err := filepath.Rel(baseDirAbs, fullPathAbs)
	if err != nil {
		return ErrInvalidFilePath
	}

	// Check if rel starts with ".." as a path component (not just as prefix)
	// For example "..." should be allowed, but "../" should not
	relParts := strings.FieldsFunc(rel, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(relParts) > 0 && relParts[0] == ".." {
		return ErrInvalidFilePath
	}

	return nil
}
