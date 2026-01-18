package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// TokenStore интерфейс для хранения и валидации токенов привязки
type TokenStore interface {
	ValidateToken(ctx context.Context, token string) (userID uuid.UUID, err error)
	DeleteToken(ctx context.Context, token string) error
}

// BotHandler обработчик команд и обновлений Telegram бота
type BotHandler struct {
	client     *Client
	tokenStore TokenStore
}

// LinkResult результат привязки пользователя к Telegram
type LinkResult struct {
	UserID     uuid.UUID
	TelegramID int64
	ChatID     int64
	Username   string
}

// NewBotHandler создает новый обработчик команд бота
func NewBotHandler(client *Client, tokenStore TokenStore) *BotHandler {
	return &BotHandler{
		client:     client,
		tokenStore: tokenStore,
	}
}

// HandleUpdate обрабатывает входящее обновление от Telegram
func (h *BotHandler) HandleUpdate(ctx context.Context, update *Update) error {
	// Проверяем наличие сообщения
	if update.Message == nil {
		return nil
	}

	message := update.Message

	// Проверяем наличие текста
	if message.Text == "" {
		return nil
	}

	// Парсим команду
	if strings.HasPrefix(message.Text, "/start") {
		return h.handleStartCommand(ctx, message)
	}

	if strings.HasPrefix(message.Text, "/help") {
		return h.handleHelpCommand(ctx, message)
	}

	// Неизвестная команда - отправляем help
	return h.handleHelpCommand(ctx, message)
}

// handleStartCommand обрабатывает команду /start с токеном привязки
func (h *BotHandler) handleStartCommand(ctx context.Context, message *Message) error {
	// Парсим токен из команды /start {token}
	parts := strings.Fields(message.Text)

	// Если нет токена, отправляем справку
	if len(parts) < 2 {
		return h.sendMessage(message.Chat.ID,
			"👋 Добро пожаловать!\n\n"+
				"Для привязки аккаунта используйте ссылку из личного кабинета на платформе.\n\n"+
				"Используйте /help для получения справки.")
	}

	token := parts[1]

	// Валидируем токен через TokenStore
	_, err := h.tokenStore.ValidateToken(ctx, token)
	if err != nil {
		return h.sendMessage(message.Chat.ID,
			"❌ Неверный или истекший токен привязки.\n\n"+
				"Пожалуйста, получите новую ссылку для привязки в личном кабинете.")
	}

	// Извлекаем данные пользователя Telegram для отправки приветственного сообщения
	chatID := message.Chat.ID
	username := message.From.Username
	if username == "" {
		username = message.From.FirstName
	}

	// Отправляем сообщение об успешной привязке
	if err := h.sendWelcomeMessage(chatID, username); err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}

	// Примечание: фактическая привязка должна быть выполнена вызывающим кодом
	// через TelegramUserRepository с использованием метода GetLinkResult(),
	// так как здесь нет доступа к репозиторию

	return nil
}

// GetLinkResult извлекает результат привязки из команды /start
// Этот метод используется для получения данных привязки после обработки команды
func (h *BotHandler) GetLinkResult(ctx context.Context, message *Message) (*LinkResult, error) {
	parts := strings.Fields(message.Text)

	if len(parts) < 2 {
		return nil, fmt.Errorf("no token provided")
	}

	token := parts[1]

	// Валидируем токен
	userID, err := h.tokenStore.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Удаляем токен после использования
	if err := h.tokenStore.DeleteToken(ctx, token); err != nil {
		// Логируем ошибку, но не прерываем процесс
		fmt.Printf("Warning: failed to delete token: %v\n", err)
	}

	// Формируем результат
	username := message.From.Username
	if username == "" {
		username = message.From.FirstName
	}

	return &LinkResult{
		UserID:     userID,
		TelegramID: message.From.ID,
		ChatID:     message.Chat.ID,
		Username:   username,
	}, nil
}

// handleHelpCommand обрабатывает команду /help
func (h *BotHandler) handleHelpCommand(ctx context.Context, message *Message) error {
	helpText := `📚 Справка по боту

Доступные команды:
/start {token} - Привязать аккаунт Telegram к вашему профилю
/help - Показать эту справку

ℹ️ О боте:
Этот бот используется для получения уведомлений о занятиях, бронированиях и других важных событиях на платформе.

Для привязки аккаунта:
1. Войдите в личный кабинет на платформе
2. Перейдите в настройки профиля
3. Нажмите "Привязать Telegram"
4. Перейдите по ссылке или используйте команду /start с токеном

По вопросам обращайтесь к администрации платформы.`

	return h.sendMessage(message.Chat.ID, helpText)
}

// sendWelcomeMessage отправляет приветственное сообщение после успешной привязки
func (h *BotHandler) sendWelcomeMessage(chatID int64, username string) error {
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

	return h.sendMessage(chatID, welcomeText)
}

// sendMessage вспомогательный метод для отправки сообщений
func (h *BotHandler) sendMessage(chatID int64, text string) error {
	return h.client.SendMessage(chatID, text)
}
