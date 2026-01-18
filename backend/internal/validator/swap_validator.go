package validator

import (
	"context"
	"errors"
	"fmt"

	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
)

var (
	// ErrNoActiveBooking возвращается, когда у студента нет активного бронирования на старый урок
	ErrNoActiveBooking = errors.New("no active booking found for the old lesson")
	// ErrNewLessonNotAvailable возвращается, когда новый урок недоступен
	ErrNewLessonNotAvailable = errors.New("new lesson is not available for booking")
	// ErrSwapScheduleConflict возвращается, когда замена создаст конфликт расписания
	ErrSwapScheduleConflict = errors.New("swap would create a schedule conflict")
	// ErrSwapTooLate возвращается при попытке замены менее чем за 24 часа
	ErrSwapTooLate = errors.New("swaps must be made at least 24 hours before both lessons")
	// ErrSwapLessonInPast возвращается при попытке замены на урок в прошлом
	ErrSwapLessonInPast = errors.New("cannot swap to a lesson in the past")
)

// SwapValidator обрабатывает логику валидации замены уроков
type SwapValidator struct {
	lessonRepo  *repository.LessonRepository
	bookingRepo *repository.BookingRepository
}

// NewSwapValidator создает новый SwapValidator
func NewSwapValidator(lessonRepo *repository.LessonRepository, bookingRepo *repository.BookingRepository) *SwapValidator {
	return &SwapValidator{
		lessonRepo:  lessonRepo,
		bookingRepo: bookingRepo,
	}
}

// ValidateSwap проверяет, можно ли выполнить замену урока
func (v *SwapValidator) ValidateSwap(ctx context.Context, studentID, oldLessonID, newLessonID uuid.UUID) error {
	// Проверяем, есть ли у студента активное бронирование на старый урок
	oldBooking, err := v.bookingRepo.GetActiveBookingByStudentAndLesson(ctx, studentID, oldLessonID)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return ErrNoActiveBooking
		}
		return fmt.Errorf("failed to get old booking: %w", err)
	}

	// Получаем детали старого урока
	oldLesson, err := v.lessonRepo.GetByID(ctx, oldLessonID)
	if err != nil {
		return fmt.Errorf("failed to get old lesson: %w", err)
	}

	// Проверяем, не менее ли 24 часов до старого урока
	if oldLesson.IsWithin24Hours() {
		return ErrSwapTooLate
	}

	// Получаем детали нового урока
	newLesson, err := v.lessonRepo.GetByID(ctx, newLessonID)
	if err != nil {
		return fmt.Errorf("failed to get new lesson: %w", err)
	}

	// Проверяем, не в прошлом ли новый урок
	if newLesson.IsInPast() {
		return ErrSwapLessonInPast
	}

	// Проверяем, не заполнен ли новый урок
	if newLesson.IsFull() {
		return ErrNewLessonNotAvailable
	}

	// Проверяем, не менее ли 24 часов до нового урока
	if newLesson.IsWithin24Hours() {
		return ErrSwapTooLate
	}

	// Проверяем конфликты расписания (исключая старое бронирование)
	hasConflict, err := v.bookingRepo.HasScheduleConflictExcluding(
		ctx,
		studentID,
		oldBooking.ID,
		newLesson.StartTime,
		newLesson.EndTime,
	)
	if err != nil {
		return fmt.Errorf("failed to check schedule conflict: %w", err)
	}
	if hasConflict {
		return ErrSwapScheduleConflict
	}

	return nil
}
