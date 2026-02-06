package service

import (
	"context"
	"fmt"
	"log"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
)

// PaymentSettingsService обрабатывает бизнес-логику управления платежами пользователей
type PaymentSettingsService struct {
	userRepo repository.UserRepository
}

// NewPaymentSettingsService создает новый PaymentSettingsService
func NewPaymentSettingsService(userRepo repository.UserRepository) *PaymentSettingsService {
	return &PaymentSettingsService{
		userRepo: userRepo,
	}
}

// GetPaymentStatus получает статус платежей для пользователя
func (s *PaymentSettingsService) GetPaymentStatus(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return user.PaymentEnabled, nil
}

// UpdatePaymentStatus изменяет статус платежей для студента (только для администраторов)
func (s *PaymentSettingsService) UpdatePaymentStatus(ctx context.Context, adminID, userID uuid.UUID, enabled bool) (*models.User, error) {
	// Проверяем что adminID имеет роль 'admin'
	admin, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	if !admin.IsAdmin() {
		return nil, repository.ErrUnauthorized
	}

	// Проверяем что userID существует
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// Проверяем что userID имеет роль 'student'
	if !user.IsStudent() {
		return nil, repository.ErrInvalidUserRole
	}

	// Обновляем payment_enabled
	updates := map[string]interface{}{
		"payment_enabled": enabled,
	}
	if err := s.userRepo.Update(ctx, userID, updates); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Логируем изменение
	log.Printf("Admin %s changed payment status for student %s to %v", adminID, userID, enabled)

	// Получаем обновленного пользователя
	updatedUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	return updatedUser, nil
}

// ListStudentsPaymentStatus возвращает список всех студентов с их статусом платежей
func (s *PaymentSettingsService) ListStudentsPaymentStatus(ctx context.Context, adminID uuid.UUID, filterByEnabled *bool) ([]*models.StudentPaymentStatus, error) {
	// Проверяем что adminID имеет роль 'admin'
	admin, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	if !admin.IsAdmin() {
		return nil, repository.ErrUnauthorized
	}

	// Получаем всех студентов
	studentRole := models.RoleStudent
	students, err := s.userRepo.List(ctx, &studentRole)
	if err != nil {
		return nil, fmt.Errorf("failed to list students: %w", err)
	}

	// Формируем результат
	result := make([]*models.StudentPaymentStatus, 0, len(students))
	for _, student := range students {
		// Применяем фильтр если указан
		if filterByEnabled != nil && student.PaymentEnabled != *filterByEnabled {
			continue
		}

		status := &models.StudentPaymentStatus{
			ID:             student.ID,
			FirstName:      student.FirstName,
			LastName:       student.LastName,
			FullName:       student.GetFullName(),
			Email:          student.Email,
			PaymentEnabled: student.PaymentEnabled,
			UpdatedAt:      student.UpdatedAt,
		}
		result = append(result, status)
	}

	// Сортируем по имени (в памяти, т.к. репозиторий сортирует по created_at)
	// Простая сортировка пузырьком для небольших объемов
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].FullName > result[j].FullName {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}
