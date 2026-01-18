package validator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	// ErrLessonNotAvailable возвращается, когда урок недоступен для бронирования
	ErrLessonNotAvailable = errors.New("lesson is not available for booking")
	// ErrScheduleConflict возвращается при конфликте времени
	ErrScheduleConflict = errors.New("schedule conflict with existing booking")
	// ErrLessonInPast возвращается при попытке забронировать прошедший урок
	ErrLessonInPast = errors.New("cannot book a lesson in the past")
	// ErrBookingNotActive возвращается при попытке отменить неактивное бронирование
	ErrBookingNotActive = errors.New("booking is not active")
	// ErrCannotCancelWithin24Hours возвращается при попытке отменить бронирование менее чем за 24 часа
	ErrCannotCancelWithin24Hours = errors.New("cannot cancel booking within 24 hours of lesson start")
)

// LessonGetter интерфейс для получения урока (для тестирования)
type LessonGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
}

// ConflictChecker интерфейс для проверки конфликтов расписания (для тестирования)
type ConflictChecker interface {
	HasScheduleConflict(ctx context.Context, studentID uuid.UUID, startTime, endTime time.Time) (bool, error)
}

// CreditGetter интерфейс для получения баланса кредитов (для тестирования)
type CreditGetter interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*models.Credit, error)
}

// BookingValidator обрабатывает логику валидации бронирований
type BookingValidator struct {
	lessonRepo  LessonGetter
	bookingRepo ConflictChecker
	creditRepo  CreditGetter
}

// NewBookingValidator создает новый BookingValidator
func NewBookingValidator(lessonRepo *repository.LessonRepository, bookingRepo *repository.BookingRepository, creditRepo *repository.CreditRepository) *BookingValidator {
	return &BookingValidator{
		lessonRepo:  lessonRepo,
		bookingRepo: bookingRepo,
		creditRepo:  creditRepo,
	}
}

// ValidateBooking проверяет, можно ли создать бронирование
// NOTE: Capacity check (IsFull) is intentionally NOT done here to avoid race condition.
// The authoritative capacity check happens inside the transaction with row-level locking.
func (v *BookingValidator) ValidateBooking(ctx context.Context, studentID, lessonID uuid.UUID, isAdmin bool) error {
	// Получаем детали урока
	lesson, err := v.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Проверяем, не в прошлом ли урок (администраторы могут добавлять студентов в прошедшие занятия)
	if !isAdmin && lesson.IsInPast() {
		return ErrLessonInPast
	}

	// REMOVED: IsFull() check - this is done inside transaction with FOR UPDATE lock
	// to prevent race condition where multiple concurrent requests pass this check
	// but only one should succeed.

	// No 24-hour advance requirement - students can book any time before lesson starts
	// The IsInPast check above already ensures lesson is in future

	// Проверяем конфликты расписания (пропускаем для прошлых занятий - админ добавляет ретроактивно)
	if !lesson.IsInPast() {
		hasConflict, err := v.bookingRepo.HasScheduleConflict(ctx, studentID, lesson.StartTime, lesson.EndTime)
		if err != nil {
			return fmt.Errorf("failed to check schedule conflict: %w", err)
		}
		if hasConflict {
			return ErrScheduleConflict
		}
	}

	// Проверяем достаточность кредитов (ранняя валидация перед транзакцией)
	// Если creditsCost = 0, урок бесплатный - пропускаем проверку баланса
	creditsCost := lesson.CreditsCost
	if creditsCost < 0 {
		creditsCost = 0 // Исправляем отрицательные значения на 0
	}
	// Проверяем баланс только если урок платный (creditsCost > 0)
	if creditsCost > 0 {
		credit, err := v.creditRepo.GetBalance(ctx, studentID)
		if err != nil {
			return fmt.Errorf("failed to check credits: %w", err)
		}
		if credit.Balance < creditsCost {
			return repository.ErrInsufficientCredits
		}
	}

	return nil
}

// ValidateCancellation проверяет, можно ли отменить бронирование
func (v *BookingValidator) ValidateCancellation(ctx context.Context, booking *models.Booking, isAdmin bool) error {
	if !booking.IsActive() {
		return ErrBookingNotActive
	}

	// Получаем урок для проверки времени
	lesson, err := v.lessonRepo.GetByID(ctx, booking.LessonID)
	if err != nil {
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Only students cannot cancel within 24 hours. Admins can cancel at any time.
	if !isAdmin && lesson.IsWithin24Hours() {
		return ErrCannotCancelWithin24Hours
	}

	return nil
}
