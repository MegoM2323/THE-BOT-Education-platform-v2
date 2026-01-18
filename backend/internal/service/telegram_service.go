package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/telegram"
)

var (
	// ErrInvalidToken возвращается при невалидном или истекшем токене привязки
	ErrInvalidToken = errors.New("invalid or expired link token")
	// ErrTelegramAlreadyLinked возвращается когда Telegram уже привязан к пользователю
	ErrTelegramAlreadyLinked = errors.New("telegram account already linked")
	// ErrTelegramIDAlreadyLinked возвращается когда telegram_id уже привязан к другому пользователю
	ErrTelegramIDAlreadyLinked = errors.New("telegram account already linked to another user")
	// ErrUserNotLinked возвращается когда пользователь не привязан к Telegram
	ErrUserNotLinked = errors.New("user not linked to telegram")
	// ErrTelegramUserNotFound возвращается когда Telegram привязка не найдена
	ErrTelegramUserNotFound = errors.New("telegram user not found")
)

// TokenData содержит данные токена привязки
type TokenData struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// TokenStore управляет хранением и валидацией токенов привязки
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]TokenData
}

// NewTokenStore создает новый TokenStore
func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]TokenData),
	}
}

// GenerateToken генерирует криптографически стойкий токен привязки
func (ts *TokenStore) GenerateToken(userID uuid.UUID, duration time.Duration) (string, error) {
	// Генерируем 32 байта случайных данных
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Кодируем в base64 URL encoding
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Сохраняем токен с данными
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tokens[token] = TokenData{
		UserID:    userID,
		ExpiresAt: time.Now().Add(duration),
	}

	return token, nil
}

// ValidateToken проверяет валидность токена и возвращает userID
func (ts *TokenStore) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	data, exists := ts.tokens[token]
	if !exists {
		return uuid.Nil, ErrInvalidToken
	}

	// Проверяем истечение срока действия
	if time.Now().After(data.ExpiresAt) {
		return uuid.Nil, ErrInvalidToken
	}

	return data.UserID, nil
}

// DeleteToken удаляет использованный токен
func (ts *TokenStore) DeleteToken(ctx context.Context, token string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	delete(ts.tokens, token)
	return nil
}

// CleanExpired удаляет истекшие токены
func (ts *TokenStore) CleanExpired() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	for token, data := range ts.tokens {
		if now.After(data.ExpiresAt) {
			delete(ts.tokens, token)
		}
	}
}

// TelegramService обрабатывает бизнес-логику для работы с Telegram
type TelegramService struct {
	telegramUserRepo  repository.TelegramUserRepository
	telegramTokenRepo repository.TelegramTokenRepository
	userRepo          repository.UserRepository
	telegramClient    *telegram.Client
	botHandler        *telegram.BotHandler
	adminTelegramID   int64
	tokenStore        *TokenStore // Deprecated: kept for backwards compatibility, use telegramTokenRepo
	stopCleanup       chan struct{}
	cleanupDone       chan struct{}
}

// NewTelegramService создает новый TelegramService с запуском фоновой очистки токенов
func NewTelegramService(
	telegramUserRepo repository.TelegramUserRepository,
	telegramTokenRepo repository.TelegramTokenRepository,
	userRepo repository.UserRepository,
	telegramClient *telegram.Client,
	adminTelegramID int64,
) *TelegramService {
	tokenStore := NewTokenStore()

	// Создаем BotHandler с TokenStore
	botHandler := telegram.NewBotHandler(telegramClient, tokenStore)

	service := &TelegramService{
		telegramUserRepo:  telegramUserRepo,
		telegramTokenRepo: telegramTokenRepo,
		userRepo:          userRepo,
		telegramClient:    telegramClient,
		botHandler:        botHandler,
		adminTelegramID:   adminTelegramID,
		tokenStore:        tokenStore,
		stopCleanup:       make(chan struct{}),
		cleanupDone:       make(chan struct{}),
	}

	// Запускаем горутину для периодической очистки истекших токенов
	go service.cleanupExpiredTokens()

	return service
}

// cleanupExpiredTokens периодически очищает истекшие токены
func (s *TelegramService) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	defer close(s.cleanupDone)

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Очищаем истекшие токены
			_, err := s.telegramTokenRepo.DeleteExpiredTokens(ctx)
			if err != nil {
				log.Info().Msg("ERROR: Failed to delete expired tokens")
			} else {
				log.Info().Msg("Cleaned expired telegram tokens")
			}

			// Очищаем невалидные записи привязок (telegram_id = 0 или NULL)
			// Это решает проблему накопления "мусорных" записей после неудачных попыток привязки
			cleaned, err := s.telegramUserRepo.CleanupInvalidLinks(ctx)
			if err != nil {
				log.Info().Msg("ERROR: Failed to cleanup invalid telegram links")
			} else if cleaned > 0 {
				log.Info().Msgf("Cleaned %d invalid telegram links", cleaned)
			}

			cancel()

			// Also clean in-memory store for backwards compatibility
			s.tokenStore.CleanExpired()
		case <-s.stopCleanup:
			// Graceful shutdown
			log.Info().Msg("log")
			return
		}
	}
}

// Shutdown останавливает фоновую очистку токенов (для graceful shutdown)
func (s *TelegramService) Shutdown() {
	close(s.stopCleanup)
	<-s.cleanupDone
	log.Info().Msg("log")
}

// GenerateLinkToken генерирует токен для привязки пользователя к Telegram
func (s *TelegramService) GenerateLinkToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Проверяем существование пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", repository.ErrUserNotFound
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем, не удален ли пользователь
	if user.IsDeleted() {
		return "", repository.ErrUserNotFound
	}

	// Очищаем невалидные записи (telegram_id = 0 или NULL) перед проверкой
	// Это решает проблему false positive после неудачной попытки привязки
	_ = s.telegramUserRepo.DeleteByUserID(ctx, userID)

	// Проверяем, не привязан ли уже Telegram валидно (telegram_id > 0)
	isLinked, err := s.telegramUserRepo.IsValidlyLinked(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to check telegram link status: %w", err)
	}
	if isLinked {
		// Пользователь уже валидно привязан к Telegram
		log.Info().Msgf("User %s already linked to Telegram", userID)
		return "", ErrTelegramAlreadyLinked
	}

	// Удаляем ВСЕ старые токены пользователя перед генерацией нового
	// Это решает проблему "уже привязан" после неудачной попытки
	if err := s.telegramTokenRepo.DeleteByUserID(ctx, userID); err != nil {
		log.Info().Msgf("Failed to delete old tokens for user %s: %v (proceeding anyway)", userID, err)
		// Продолжаем, это не критичная ошибка
	}

	// Также удаляем из in-memory хранилища (для backward compatibility)
	s.tokenStore.mu.Lock()
	for token, data := range s.tokenStore.tokens {
		if data.UserID == userID {
			delete(s.tokenStore.tokens, token)
		}
	}
	s.tokenStore.mu.Unlock()

	// Генерируем токен на 15 минут
	token, err := s.tokenStore.GenerateToken(userID, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Сохраняем токен в базу данных для персистентности
	expiresAt := time.Now().Add(15 * time.Minute)
	if err := s.telegramTokenRepo.SaveToken(ctx, token, userID, expiresAt); err != nil {
		log.Info().Msgf("Failed to save token to DB for user %s: %v", userID, err)
		// Не прерываем процесс, токен все еще работает через in-memory store
	}

	log.Info().Msgf("Generated new link token for user %s", userID)
	return token, nil
}

// LinkUserAccount привязывает Telegram аккаунт к пользователю
func (s *TelegramService) LinkUserAccount(ctx context.Context, token string, telegramID, chatID int64, username string) error {
	// Валидируем токен (сначала проверяем БД, потом in-memory)
	userID, err := s.telegramTokenRepo.GetTokenUser(ctx, token)
	if err != nil {
		// Если в БД не найдено, пробуем in-memory store для backwards compatibility
		var inMemErr error
		userID, inMemErr = s.tokenStore.ValidateToken(ctx, token)
		if inMemErr != nil {
			return ErrInvalidToken
		}
	}

	// Проверяем, не привязан ли уже этот Telegram к другому пользователю
	_, err = s.telegramUserRepo.GetByTelegramID(ctx, telegramID)
	if err == nil {
		// Telegram уже привязан к другому пользователю
		return repository.ErrTelegramUserAlreadyLinked
	}
	if !errors.Is(err, repository.ErrTelegramUserNotFound) {
		return fmt.Errorf("failed to check telegram ID: %w", err)
	}

	// Привязываем пользователя к Telegram
	if err := s.telegramUserRepo.LinkUserToTelegram(ctx, userID, telegramID, chatID, username); err != nil {
		return fmt.Errorf("failed to link user to telegram: %w", err)
	}

	// Синхронизируем telegram_username в таблице users
	if err := s.userRepo.UpdateTelegramUsername(ctx, userID, username); err != nil {
		// Логируем ошибку, но не прерываем процесс (основная привязка уже успешна)
		log.Info().Msgf("Failed to sync telegram username for user %s: %v", userID, err)
	}

	// Удаляем использованный токен из обеих хранилищ
	if err := s.telegramTokenRepo.DeleteToken(ctx, token); err != nil {
		// Логируем ошибку, но не прерываем процесс
		log.Info().Msg("log")
	}
	if err := s.tokenStore.DeleteToken(ctx, token); err != nil {
		// Логируем ошибку, но не прерываем процесс
		log.Info().Msg("log")
	}

	log.Info().Msg("log")
	return nil
}

// GetUserLinkStatus получает статус привязки пользователя к Telegram
func (s *TelegramService) GetUserLinkStatus(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error) {
	telegramUser, err := s.telegramUserRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			// Не привязан - это не ошибка, возвращаем nil
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get telegram link status: %w", err)
	}

	return telegramUser, nil
}

// UnlinkUser отвязывает Telegram от пользователя
func (s *TelegramService) UnlinkUser(ctx context.Context, userID uuid.UUID) error {
	// Проверяем наличие привязки
	_, err := s.telegramUserRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			return ErrUserNotLinked
		}
		return fmt.Errorf("failed to check telegram link: %w", err)
	}

	// Удаляем привязку
	if err := s.telegramUserRepo.UnlinkTelegram(ctx, userID); err != nil {
		return fmt.Errorf("failed to unlink telegram: %w", err)
	}

	// Очищаем telegram_username в таблице users (десинхронизация при отвязке)
	if err := s.userRepo.UpdateTelegramUsername(ctx, userID, ""); err != nil {
		// Логируем ошибку, но не прерываем процесс (основная отвязка уже успешна)
		log.Info().Msgf("Failed to clear telegram username for user %s: %v", userID, err)
	}

	log.Info().Msg("log")
	return nil
}

// GetLinkedUsers получает список привязанных пользователей с полной информацией, опционально отфильтрованных по роли
// Использует оптимизированный запрос с JOIN вместо N+1 запросов
func (s *TelegramService) GetLinkedUsers(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	var telegramUsers []*models.TelegramUser
	var err error

	if role == "" {
		// Получаем всех привязанных пользователей с полной информацией через JOIN (один запрос)
		telegramUsers, err = s.telegramUserRepo.GetAllWithUserInfo(ctx)
	} else {
		// Фильтруем по роли с полной информацией через JOIN (один запрос)
		telegramUsers, err = s.telegramUserRepo.GetByRoleWithUserInfo(ctx, role)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get linked users: %w", err)
	}

	return telegramUsers, nil
}

// SubscribeToNotifications подписывает пользователя на Telegram уведомления
func (s *TelegramService) SubscribeToNotifications(ctx context.Context, userID uuid.UUID) error {
	// Проверяем наличие привязки
	telegramUser, err := s.telegramUserRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			return ErrTelegramUserNotFound
		}
		return fmt.Errorf("failed to get telegram user: %w", err)
	}

	// Если уже подписан, просто возвращаем успех
	if telegramUser.Subscribed {
		return nil
	}

	// Подписываем пользователя
	if err := s.telegramUserRepo.UpdateSubscription(ctx, userID, true); err != nil {
		return fmt.Errorf("failed to subscribe to notifications: %w", err)
	}

	log.Info().Msg("log")
	return nil
}

// UnsubscribeFromNotifications отписывает пользователя от Telegram уведомлений
func (s *TelegramService) UnsubscribeFromNotifications(ctx context.Context, userID uuid.UUID) error {
	// Проверяем наличие привязки
	telegramUser, err := s.telegramUserRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			return ErrTelegramUserNotFound
		}
		return fmt.Errorf("failed to get telegram user: %w", err)
	}

	// Если уже отписан, просто возвращаем успех
	if !telegramUser.Subscribed {
		return nil
	}

	// Отписываем пользователя
	if err := s.telegramUserRepo.UpdateSubscription(ctx, userID, false); err != nil {
		return fmt.Errorf("failed to unsubscribe from notifications: %w", err)
	}

	log.Info().Msg("log")
	return nil
}

// SendAdminNotification отправляет уведомление администратору
func (s *TelegramService) SendAdminNotification(ctx context.Context, message string) error {
	if s.adminTelegramID == 0 {
		// Админский Telegram ID не настроен - пропускаем
		log.Info().Msg("log")
		return nil
	}

	if s.telegramClient == nil {
		// Telegram клиент не инициализирован
		log.Info().Msg("log")
		return nil
	}

	// Отправляем сообщение
	if err := s.telegramClient.SendMessage(s.adminTelegramID, message); err != nil {
		log.Info().Msg("log")

		// Проверяем, не заблокирован ли бот
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			if telegramErr.ErrorCode == 403 {
				log.Info().Msg("log")
				return fmt.Errorf("bot is blocked by admin user")
			}
			if telegramErr.ErrorCode == 400 {
				log.Info().Msg("log")
				return fmt.Errorf("invalid admin telegram ID")
			}
		}

		// Логируем ошибку, но не возвращаем её для критичных операций
		return fmt.Errorf("failed to send admin notification: %w", err)
	}

	log.Info().Msg("log")
	return nil
}

// SendUserNotification отправляет уведомление пользователю
func (s *TelegramService) SendUserNotification(ctx context.Context, userID uuid.UUID, message string) error {
	// Получаем привязку пользователя
	telegramUser, err := s.telegramUserRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			return ErrUserNotLinked
		}
		return fmt.Errorf("failed to get telegram user: %w", err)
	}

	// Проверяем подписку на уведомления
	if !telegramUser.Subscribed {
		log.Info().Msg("log")
		return nil
	}

	// Отправляем сообщение
	if err := s.telegramClient.SendMessage(telegramUser.ChatID, message); err != nil {
		// Проверяем, не заблокирован ли бот
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			if telegramErr.ErrorCode == 403 {
				log.Info().Msg("log")
				// Можно автоматически отписать пользователя от уведомлений
				if updateErr := s.telegramUserRepo.UpdateSubscription(ctx, userID, false); updateErr != nil {
					log.Info().Msg("log")
				}
			}
		}
		return fmt.Errorf("failed to send notification to user: %w", err)
	}

	log.Info().Msg("log")
	return nil
}

// HandleWebhook обрабатывает webhook от Telegram
func (s *TelegramService) HandleWebhook(ctx context.Context, update *telegram.Update) error {
	// Проверяем наличие сообщения
	if update.Message == nil {
		// Не сообщение - пропускаем
		return nil
	}

	message := update.Message

	// Проверяем, является ли это командой /start с токеном
	if message.Text != "" && len(message.Text) > 7 && message.Text[:6] == "/start" {
		// Получаем результат привязки (токен уже валидирован и удален внутри GetLinkResult)
		linkResult, err := s.botHandler.GetLinkResult(ctx, message)
		if err != nil {
			// Ошибка валидации токена - отправляем сообщение об ошибке
			log.Info().Msg("log")
			if sendErr := s.telegramClient.SendMessage(message.Chat.ID,
				"❌ Неверный или истекший токен привязки.\n\n"+
					"Пожалуйста, получите новую ссылку для привязки в личном кабинете."); sendErr != nil {
				log.Info().Msg("log")
			}
			return nil
		}

		// Проверяем, не привязан ли уже этот Telegram к другому пользователю
		_, err = s.telegramUserRepo.GetByTelegramID(ctx, linkResult.TelegramID)
		if err == nil {
			// Telegram уже привязан к другому пользователю
			log.Info().Msg("log")
			if sendErr := s.telegramClient.SendMessage(linkResult.ChatID,
				"❌ Этот Telegram аккаунт уже привязан к другому пользователю платформы."); sendErr != nil {
				log.Info().Msg("log")
			}
			return nil
		}
		if !errors.Is(err, repository.ErrTelegramUserNotFound) {
			return fmt.Errorf("failed to check telegram ID: %w", err)
		}

		// Выполняем привязку напрямую через репозиторий
		if err := s.telegramUserRepo.LinkUserToTelegramAtomic(
			ctx,
			linkResult.UserID,
			linkResult.TelegramID,
			linkResult.ChatID,
			linkResult.Username,
		); err != nil {
			log.Info().Msg("log")
			if sendErr := s.telegramClient.SendMessage(linkResult.ChatID,
				"❌ Произошла ошибка при привязке аккаунта. Пожалуйста, попробуйте позже."); sendErr != nil {
				log.Info().Msg("log")
			}
			return fmt.Errorf("failed to link user to telegram: %w", err)
		}

		// Синхронизируем telegram_username в таблице users
		if err := s.userRepo.UpdateTelegramUsername(ctx, linkResult.UserID, linkResult.Username); err != nil {
			// Логируем ошибку, но не прерываем процесс (основная привязка уже успешна)
			log.Info().Msgf("Failed to sync telegram username for user %s: %v", linkResult.UserID, err)
		}

		log.Info().Msg("log")

		// Отправляем приветственное сообщение после успешной привязки
		username := linkResult.Username
		if username == "" {
			username = "пользователь"
		}
		welcomeText := fmt.Sprintf(
			"✅ Аккаунт успешно привязан!\n\n"+
				"Привет, %s! 👋\n\n"+
				"Теперь вы будете получать уведомления о:\n"+
				"• Предстоящих занятиях\n"+
				"• Новых бронированиях\n"+
				"• Отменах и переносах\n"+
				"• Изменениях в расписании\n"+
				"• Важных объявлениях\n\n"+
				"Вы можете управлять уведомлениями в настройках профиля на платформе.\n\n"+
				"Используйте /help для получения справки.",
			username,
		)

		if sendErr := s.telegramClient.SendMessage(linkResult.ChatID, welcomeText); sendErr != nil {
			log.Info().Msg("log")
		}

		// ✅ Успешная привязка - выходим, не обрабатываем дальше
		return nil
	}

	// Передаем обновление в обработчик (для других команд кроме /start с токеном)
	if err := s.botHandler.HandleUpdate(ctx, update); err != nil {
		return fmt.Errorf("failed to handle update: %w", err)
	}

	return nil
}

// SetUserTelegram устанавливает или обновляет Telegram для пользователя (для админов)
// Использует атомарную операцию для защиты от race condition при одновременных запросах.
// Гарантирует, что только один пользователь может привязать данный telegram_id.
func (s *TelegramService) SetUserTelegram(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	// Используем атомарную операцию с SELECT FOR UPDATE для защиты от race condition
	// При одновременных запросах от разных пользователей с одинаковым telegram_id,
	// только первый успешно привяжет, остальные получат ErrTelegramIDAlreadyLinked
	if err := s.telegramUserRepo.LinkUserToTelegramAtomic(ctx, userID, telegramID, chatID, username); err != nil {
		// Обрабатываем все возможные ошибки привязки
		if errors.Is(err, repository.ErrTelegramIDAlreadyLinked) {
			return ErrTelegramIDAlreadyLinked
		}
		if errors.Is(err, repository.ErrTelegramUserAlreadyLinked) {
			return ErrTelegramIDAlreadyLinked
		}
		return fmt.Errorf("failed to link user to telegram: %w", err)
	}

	// Синхронизируем telegram_username в таблице users
	if err := s.userRepo.UpdateTelegramUsername(ctx, userID, username); err != nil {
		// Логируем ошибку, но не прерываем процесс (основная привязка уже успешна)
		log.Info().Msgf("Failed to sync telegram username for user %s: %v", userID, err)
	}

	log.Info().Msg("log")
	return nil
}

// SendMessage отправляет сообщение в Telegram чат (для админ операций)
func (s *TelegramService) SendMessage(ctx context.Context, chatID int64, message string) error {
	if s.telegramClient == nil {
		return fmt.Errorf("telegram client not configured")
	}

	if chatID == 0 {
		return fmt.Errorf("invalid chat ID")
	}

	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	// Отправляем сообщение
	if err := s.telegramClient.SendMessage(chatID, message); err != nil {
		log.Info().Msg("log")

		// Проверяем специфичные ошибки Telegram API
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			if telegramErr.ErrorCode == 403 {
				return telegramErr
			}
			if telegramErr.ErrorCode == 400 {
				log.Info().Msg("log")
				return fmt.Errorf("invalid chat ID or message format")
			}
		}

		return fmt.Errorf("failed to send message: %w", err)
	}

	// Обновляем метрики успешной отправки

	log.Info().Msg("log")
	return nil
}
