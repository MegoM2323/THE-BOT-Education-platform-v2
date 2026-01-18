package validator

import (
	"errors"
	"time"
)

const (
	// MinBookingAdvance минимальное время предварительного бронирования (24 часа)
	MinBookingAdvance = 24 * time.Hour
	// MinAllowedYear минимальный допустимый год для дат параметров
	MinAllowedYear = 2020
	// MaxAllowedYear максимальный допустимый год для дат параметров
	MaxAllowedYear = 2030
	// DateLayout формат даты YYYY-MM-DD
	DateLayout = "2006-01-02"
)

var (
	// ErrInvalidTimeRange возвращается, когда временной диапазон недействителен
	ErrInvalidTimeRange = errors.New("invalid time range: end time must be after start time")
	// ErrInsufficientAdvance возвращается, когда бронирование не соответствует минимальному требованию предварительности
	ErrInsufficientAdvance = errors.New("booking must be made at least 24 hours in advance")
	// ErrInvalidDateFormat возвращается, когда формат даты не соответствует YYYY-MM-DD
	ErrInvalidDateFormat = errors.New("invalid date format, expected YYYY-MM-DD")
	// ErrDateOutOfRange возвращается, когда дата выходит за допустимые пределы
	ErrDateOutOfRange = errors.New("date must be between 2020 and 2030")
)

// TimeValidator обрабатывает валидацию, связанную со временем
type TimeValidator struct{}

// NewTimeValidator создает новый TimeValidator
func NewTimeValidator() *TimeValidator {
	return &TimeValidator{}
}

// ValidateBookingTime проверяет, соответствует ли время урока требованиям бронирования
func (v *TimeValidator) ValidateBookingTime(lessonStartTime time.Time) error {
	now := time.Now()

	// Проверяем, не в прошлом ли урок
	if lessonStartTime.Before(now) {
		return errors.New("cannot book a lesson in the past")
	}

	// Проверяем, что бронирование делается минимум за 24 часа
	timeUntilLesson := lessonStartTime.Sub(now)
	if timeUntilLesson < MinBookingAdvance {
		return ErrInsufficientAdvance
	}

	return nil
}

// ValidateTimeRange проверяет, что временной диапазон действителен
func (v *TimeValidator) ValidateTimeRange(startTime, endTime time.Time) error {
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return ErrInvalidTimeRange
	}
	return nil
}

// ValidateLessonDuration проверяет, что продолжительность урока разумна
func (v *TimeValidator) ValidateLessonDuration(startTime, endTime time.Time, minDuration, maxDuration time.Duration) error {
	if err := v.ValidateTimeRange(startTime, endTime); err != nil {
		return err
	}

	duration := endTime.Sub(startTime)

	if duration < minDuration {
		return errors.New("lesson duration is too short")
	}

	if duration > maxDuration {
		return errors.New("lesson duration is too long")
	}

	return nil
}

// IsWithin24Hours проверяет, находится ли время в пределах 24 часов от текущего момента
func (v *TimeValidator) IsWithin24Hours(t time.Time) bool {
	return time.Until(t) < MinBookingAdvance
}

// CanBeCancelled проверяет, можно ли отменить бронирование на основе времени
func (v *TimeValidator) CanBeCancelled(lessonStartTime time.Time) bool {
	return time.Until(lessonStartTime) >= MinBookingAdvance
}

// ValidateDateString проверяет строковый формат даты и возвращает time.Time
// Строка должна быть в формате YYYY-MM-DD и находиться в диапазоне 2020-2030
func (v *TimeValidator) ValidateDateString(dateStr string) (time.Time, error) {
	// Проверяем формат даты строго: YYYY-MM-DD
	if len(dateStr) != len(DateLayout) {
		return time.Time{}, ErrInvalidDateFormat
	}

	t, err := time.Parse(DateLayout, dateStr)
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}

	// Проверяем диапазон лет
	if t.Year() < MinAllowedYear || t.Year() > MaxAllowedYear {
		return time.Time{}, ErrDateOutOfRange
	}

	return t, nil
}

// ValidateDateRange проверяет диапазон дат
// Убеждается, что end_date >= start_date и диапазон не превышает 365 дней
func (v *TimeValidator) ValidateDateRange(startDate, endDate time.Time, maxDays int) error {
	if endDate.Before(startDate) {
		return ErrInvalidTimeRange
	}

	duration := endDate.Sub(startDate)
	maxDuration := time.Duration(maxDays) * 24 * time.Hour
	if duration > maxDuration {
		return errors.New("date range exceeds maximum allowed duration")
	}

	return nil
}
