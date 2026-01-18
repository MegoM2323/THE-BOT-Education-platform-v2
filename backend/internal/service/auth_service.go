package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/auth"
	"tutoring-platform/pkg/hash"

	"github.com/google/uuid"
)

var (
	// ErrInvalidCredentials возвращается при неверных учетных данных для входа
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserNotActive возвращается когда учетная запись пользователя неактивна
	ErrUserNotActive = errors.New("user account is not active")
)

// SessionExpiryBuffer - буфер для определения "близости к истечению"
// Используется для предотвращения race condition в refresh логике
// НЕ используется для отклонения валидных сессий - только для принятия решения о refresh
const SessionExpiryBuffer = 30 * time.Second

// SessionRefreshThreshold - порог для автоматического продления сессии
// Если до истечения осталось меньше этого времени, сессия будет продлена
// При 7-дневной сессии продлеваем когда остается меньше 3 дней
const SessionRefreshThreshold = 72 * time.Hour

// SessionRepositoryInterface интерфейс для работы с сессиями (для поддержки mock'ов в тестах)
type SessionRepositoryInterface interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error)
	GetWithUser(ctx context.Context, sessionID uuid.UUID) (*models.SessionWithUser, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	UpdateExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error
}

// AuthService обрабатывает аутентификацию и управление сессиями
type AuthService struct {
	userRepo      repository.UserRepository
	sessionRepo   SessionRepositoryInterface
	sessionMgr    *auth.SessionManager
	sessionMaxAge time.Duration
}

// NewAuthService создает новый AuthService
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo SessionRepositoryInterface,
	sessionMgr *auth.SessionManager,
	sessionMaxAge time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		sessionMgr:    sessionMgr,
		sessionMaxAge: sessionMaxAge,
	}
}

// isSessionValid проверяет, не истекла ли сессия
// Используется когда нужна точная проверка (без буфера)
// ПРИМЕЧАНИЕ: Эту функцию нужно использовать с осторожностью
// Предпочитайте isSessionNearExpiry() для проверки в middleware
func (s *AuthService) isSessionValid(session *models.Session) bool {
	return time.Now().Before(session.ExpiresAt)
}

// isSessionNearExpiry проверяет, близка ли сессия к истечению
// Использует буфер SessionExpiryBuffer для предотвращения race condition
// Возвращает true если сессия истекла ИЛИ близка к истечению
// Это единственная функция для проверки истечения сессии!
func (s *AuthService) isSessionNearExpiry(session *models.Session) bool {
	return time.Now().Add(SessionExpiryBuffer).After(session.ExpiresAt)
}

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse представляет ответ на вход
type LoginResponse struct {
	User         *models.User    `json:"user"`
	SessionToken string          `json:"-"` // Не включается в JSON, используется для cookie
	Session      *models.Session `json:"-"` // Сессия для внутреннего использования (например, для CSRF токена)
}

// Login аутентифицирует пользователя и создает сессию
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// Получаем пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Проверяем, не удален ли пользователь
	if user.IsDeleted() {
		return nil, ErrUserNotActive
	}

	// Проверяем пароль
	if err := hash.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Создаем сессию и получаем объект сессии для CSRF токена
	sessionToken, session, err := s.CreateSessionWithData(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		User:         user,
		SessionToken: sessionToken,
		Session:      session,
	}, nil
}

// createSession создает новую сессию для пользователя
// DEPRECATED: Используйте CreateSessionWithData вместо этого. Эта функция оставлена для обратной совместимости.
func (s *AuthService) createSession(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (string, error) {
	expiresAt := time.Now().Add(s.sessionMaxAge)

	session := &models.Session{
		UserID:    userID,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Сохраняем сессию в базу данных
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("failed to save session to database: %w", err)
	}

	// Создаем токен сессии
	token, err := s.sessionMgr.CreateSessionToken(session.ID, userID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session token: %w", err)
	}

	return token, nil
}

// CreateSessionWithData создает новую сессию и возвращает токен и объект сессии
func (s *AuthService) CreateSessionWithData(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (string, *models.Session, error) {
	expiresAt := time.Now().Add(s.sessionMaxAge)

	session := &models.Session{
		UserID:    userID,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	// Сохраняем сессию в базу данных
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", nil, fmt.Errorf("failed to save session to database: %w", err)
	}

	// Создаем токен сессии
	token, err := s.sessionMgr.CreateSessionToken(session.ID, userID, expiresAt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session token: %w", err)
	}

	return token, session, nil
}

// validateSession проверяет токен сессии и возвращает сессию с информацией о пользователе
// DEPRECATED: Используйте ValidateSessionWithBuffer вместо этого. Эта функция может привести к race condition.
func (s *AuthService) validateSession(ctx context.Context, token string) (*models.SessionWithUser, error) {
	// Проверяем токен и извлекаем данные сессии
	sessionData, err := s.sessionMgr.ValidateSessionToken(token)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredSession) && sessionData != nil {
			// Очищаем истекшую сессию
			s.sessionRepo.Delete(ctx, sessionData.SessionID)
		}
		return nil, err
	}

	// Получаем сессию с информацией о пользователе из базы данных
	session, err := s.sessionRepo.GetWithUser(ctx, sessionData.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from database: %w", err)
	}

	// Дополнительная проверка истечения
	if session.IsExpired() {
		s.sessionRepo.Delete(ctx, session.ID)
		return nil, auth.ErrExpiredSession
	}

	return session, nil
}

// ValidateSessionWithBuffer проверяет токен сессии
// ВАЖНО: Проверка истечения происходит по данным из БД, а не из токена
// Это позволяет сессии продолжать работать после refresh (когда токен обновлен)
func (s *AuthService) ValidateSessionWithBuffer(ctx context.Context, token string) (*models.SessionWithUser, error) {
	// Проверяем токен и извлекаем данные сессии
	// ВАЖНО: Игнорируем ErrExpiredSession из токена - проверяем по БД
	sessionData, err := s.sessionMgr.ValidateSessionToken(token)
	if err != nil {
		// Если токен невалиден (не expired, а именно невалиден - плохая подпись и т.д.)
		if !errors.Is(err, auth.ErrExpiredSession) {
			return nil, err
		}
		// Если токен expired, пробуем достать sessionData для проверки по БД
		// ValidateSessionToken возвращает sessionData даже при expired
		if sessionData == nil {
			return nil, err
		}
	}

	// Получаем сессию из базы данных - это источник истины
	session, err := s.sessionRepo.GetWithUser(ctx, sessionData.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from database: %w", err)
	}

	// Проверяем истечение ТОЛЬКО по данным из БД (не по токену!)
	// Сессия может быть продлена через RefreshSessionToken, и БД содержит актуальное время
	if !s.isSessionValid(&session.Session) {
		// Сессия реально истекла - удаляем асинхронно
		sessionID := session.ID
		go func(id uuid.UUID) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.sessionRepo.Delete(ctx, id); err != nil {
				fmt.Printf("ERROR: failed to delete expired session %s: %v\n", id, err)
			}
		}(sessionID)
		return nil, auth.ErrExpiredSession
	}

	return session, nil
}

// Logout выполняет выход пользователя, удаляя сессию
func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// LogoutAll выполняет выход пользователя со всех устройств
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// GetCurrentUser получает текущего пользователя из сессии
func (s *AuthService) GetCurrentUser(ctx context.Context, token string) (*models.User, error) {
	session, err := s.ValidateSessionWithBuffer(ctx, token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByID получает пользователя по ID
// Это вспомогательный метод для работы с методом ValidateSessionWithBuffer
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// ChangePassword изменяет пароль пользователя
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for password change: %w", err)
	}

	// Проверяем старый пароль
	if err := hash.CheckPassword(oldPassword, user.PasswordHash); err != nil {
		return errors.New("current password is incorrect")
	}

	// Хешируем новый пароль
	newHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Обновляем пароль
	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Выходим из всех сессий для безопасности
	s.LogoutAll(ctx, userID)

	return nil
}

// CleanupExpiredSessions удаляет истекшие сессии
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}

// Register регистрирует нового пользователя через email и пароль
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Валидация входных данных
	if req.Email == "" {
		return nil, models.ErrInvalidEmail
	}
	if req.Password == "" || len(req.Password) < 8 {
		return nil, models.ErrPasswordTooShort
	}
	if req.FullName == "" || len(req.FullName) < 2 {
		return nil, models.ErrInvalidFullName
	}

	// Проверяем, что пользователь с таким email еще не существует
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil && !existing.IsDeleted() {
		return nil, errors.New("user with this email already exists")
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Хешируем пароль
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем нового пользователя со статусом студента
	user := &models.User{
		Email:          req.Email,
		FullName:       req.FullName,
		PasswordHash:   hashedPassword,
		Role:           models.RoleStudent,
		PaymentEnabled: true, // По умолчанию оплата включена
	}

	// Сохраняем пользователя в БД
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// RegisterViaTelegram регистрирует нового пользователя через Telegram username
func (s *AuthService) RegisterViaTelegram(ctx context.Context, req *models.RegisterViaTelegramRequest) (*models.User, error) {
	// Валидация
	if req.TelegramUsername == "" {
		return nil, errors.New("telegram username is required")
	}

	// Генерируем уникальный email на основе Telegram username
	email := fmt.Sprintf("%s@telegram.local", req.TelegramUsername)

	// Генерируем случайный пароль
	randomPassword := uuid.New().String()

	// Создаем пользователя со статусом студента
	user := &models.User{
		Email:          email,
		FullName:       req.TelegramUsername, // Используем Telegram username как имя
		Role:           models.RoleStudent,
		PaymentEnabled: true, // По умолчанию оплата включена
	}

	// Хешируем пароль
	hashedPassword, err := hash.HashPassword(randomPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hashedPassword

	// Создаем пользователя в БД
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// RefreshSessionToken продлевает сессию пользователя, если она близка к истечению
// Проверка истечения происходит по БД, а не по токену
func (s *AuthService) RefreshSessionToken(ctx context.Context, token string) (string, error) {
	// Извлекаем данные из токена (игнорируем expired - проверяем по БД)
	sessionData, err := s.sessionMgr.ValidateSessionToken(token)
	if err != nil && !errors.Is(err, auth.ErrExpiredSession) {
		return "", fmt.Errorf("failed to validate session token: %w", err)
	}
	if sessionData == nil {
		return "", fmt.Errorf("failed to extract session data from token")
	}

	// Получаем сессию из БД - источник истины
	session, err := s.sessionRepo.GetByID(ctx, sessionData.SessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	// Проверяем валидность по БД
	if !s.isSessionValid(session) {
		return "", auth.ErrExpiredSession
	}

	// Проверяем, нужно ли продлевать сессию
	timeUntilExpiry := time.Until(session.ExpiresAt)

	if timeUntilExpiry > SessionRefreshThreshold {
		// Сессия еще свежая, продлевать не нужно
		return token, nil
	}

	// Продлеваем сессию на полный sessionMaxAge
	newExpiresAt := time.Now().Add(s.sessionMaxAge)

	// Обновляем время истечения в БД
	if err := s.sessionRepo.UpdateExpiry(ctx, session.ID, newExpiresAt); err != nil {
		return "", fmt.Errorf("failed to update session expiry: %w", err)
	}

	// Создаем новый токен с обновленным временем истечения
	newToken, err := s.sessionMgr.CreateSessionToken(session.ID, session.UserID, newExpiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create new session token: %w", err)
	}

	return newToken, nil
}
