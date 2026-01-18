package service

import (
	"context"
	"errors"
	"fmt"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/hash"

	"github.com/google/uuid"
)

// UserService обрабатывает бизнес-логику для пользователей
type UserService struct {
	userRepo    repository.UserRepository
	creditRepo  *repository.CreditRepository
	sessionRepo SessionRepositoryInterface
}

// NewUserService создает новый UserService
func NewUserService(userRepo repository.UserRepository, creditRepo *repository.CreditRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		creditRepo: creditRepo,
	}
}

// NewUserServiceWithSession создает новый UserService с поддержкой инвалидации сессий при удалении
func NewUserServiceWithSession(userRepo repository.UserRepository, creditRepo *repository.CreditRepository, sessionRepo SessionRepositoryInterface) *UserService {
	return &UserService{
		userRepo:    userRepo,
		creditRepo:  creditRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateUser создает нового пользователя и инициализирует кредиты
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверяем, существует ли уже пользователь
	exists, err := s.userRepo.Exists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists: %w", err)
	}
	if exists {
		return nil, repository.ErrUserExists
	}

	// Хешируем пароль
	passwordHash, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user := &models.User{
		Email:          req.Email,
		PasswordHash:   passwordHash,
		FullName:       req.FullName,
		Role:           req.Role,
		PaymentEnabled: true, // По умолчанию оплата включена
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Создаем кредиты для всех новых пользователей (триггер БД создает только для студентов)
	// Для администраторов и учителей требуется явный вызов CreateCredit в сервисе
	if err := s.creditRepo.CreateCredit(ctx, user.ID, 0); err != nil {
		// Игнорируем ошибку дублирования ключа, так как кредиты могут быть уже созданы
		// (например, если повторный запрос содержит данные о пользователе)
		if !errors.Is(err, repository.ErrDuplicateCredit) {
			return nil, fmt.Errorf("failed to create credit account for user: %w", err)
		}
	}

	return user, nil
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateUser обновляет пользователя
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	// Санитизация входных данных
	req.Sanitize()

	updates := make(map[string]interface{})

	if req.Email != nil {
		// Валидация формата email
		if !isValidEmail(*req.Email) {
			return nil, models.ErrInvalidEmail
		}

		// Проверяем, существует ли пользователь с таким email (для других пользователей)
		existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}

		// Если пользователь с таким email найден, проверяем, что это не сам пользователь
		if existingUser != nil && existingUser.ID != userID {
			return nil, repository.ErrUserExists
		}

		updates["email"] = *req.Email
	}
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Role != nil {
		if *req.Role != models.RoleStudent &&
			*req.Role != models.RoleTeacher &&
			*req.Role != models.RoleMethodologist &&
			*req.Role != models.RoleAdmin {
			return nil, models.ErrInvalidRole
		}
		updates["role"] = *req.Role
	}
	if req.PaymentEnabled != nil {
		updates["payment_enabled"] = *req.PaymentEnabled
	}
	if req.TelegramUsername != nil {
		// Валидация telegram_username если он не пустой
		if *req.TelegramUsername != "" {
			if !isValidTelegramUsername(*req.TelegramUsername) {
				return nil, models.ErrInvalidTelegramHandle
			}
			updates["telegram_username"] = *req.TelegramUsername
		} else {
			// Если передана пустая строка - удаляем telegram_username
			updates["telegram_username"] = nil
		}
	}

	// Обработка изменения пароля (для админа)
	if req.Password != nil && *req.Password != "" {
		// Валидация минимальной длины пароля
		if len(*req.Password) < 8 {
			return nil, models.ErrPasswordTooShort
		}
		// Хешируем новый пароль
		passwordHash, err := hash.HashPassword(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		// Обновляем пароль отдельным методом
		if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
			return nil, fmt.Errorf("failed to update password: %w", err)
		}
	}

	if err := s.userRepo.Update(ctx, userID, updates); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}
	return user, nil
}

// isValidEmail выполняет базовую валидацию email адреса
// Используется при обновлении пользователя для проверки формата email
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	// Проверяем наличие @ и точки
	atIndex := -1
	lastDotIndex := -1

	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Несколько @
			}
			atIndex = i
		}
		if char == '.' {
			lastDotIndex = i
		}
	}

	// Должна быть ровно одна @
	if atIndex == -1 {
		return false
	}

	// Проверяем структуру: user@domain.ext
	if atIndex == 0 || atIndex == len(email)-1 {
		return false // @ не может быть в начале или конце
	}

	if lastDotIndex == -1 || lastDotIndex <= atIndex {
		return false // Должна быть точка после @
	}

	if lastDotIndex == len(email)-1 {
		return false // Точка не может быть в конце
	}

	return true
}

// isValidTelegramUsername проверяет корректность имени пользователя Telegram
// Telegram username: 5-32 символа, начинается с буквы, содержит буквы, цифры, подчёркивание
func isValidTelegramUsername(username string) bool {
	if len(username) < 5 || len(username) > 32 {
		return false
	}

	// Telegram username должен начинаться с буквы
	if len(username) > 0 && !((username[0] >= 'a' && username[0] <= 'z') ||
		(username[0] >= 'A' && username[0] <= 'Z')) {
		return false
	}

	// Telegram username может содержать только буквы, цифры и подчеркивание
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}

	return true
}

// DeleteUser выполняет мягкое удаление пользователя и инвалидирует все его сессии
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Удаляем пользователя (мягкое удаление)
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	// Инвалидируем все сессии пользователя для немедленного выхода
	// Это гарантирует что удаленный пользователь не сможет использовать существующие сессии
	if s.sessionRepo != nil {
		if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
			// Логируем ошибку но не прерываем процесс удаления
			fmt.Printf("[WARN] Failed to delete sessions for deleted user %s: %v\n", userID, err)
		}
	}

	return nil
}

// ListUsers получает список всех пользователей
func (s *UserService) ListUsers(ctx context.Context, role *models.UserRole) ([]*models.User, error) {
	return s.userRepo.List(ctx, role)
}

// ListUsersWithPagination получает список пользователей с пагинацией
func (s *UserService) ListUsersWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	return s.userRepo.ListWithPagination(ctx, role, offset, limit)
}
