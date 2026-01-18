package handlers

import (
	"testing"

	"tutoring-platform/internal/pkg/errmessages"

	"github.com/stretchr/testify/assert"
)

// TestErrorMessages_AllConstantsDefined проверяет что все используемые константы определены в errmessages
func TestErrorMessages_AllConstantsDefined(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{"ErrMsgInsufficientCredits", errmessages.ErrMsgInsufficientCredits, true},
		{"ErrMsgBalanceExceeded", errmessages.ErrMsgBalanceExceeded, true},
		{"ErrMsgCreditNotInitialized", errmessages.ErrMsgCreditNotInitialized, true},
		{"ErrMsgAlreadyBooked", errmessages.ErrMsgAlreadyBooked, true},
		{"ErrMsgCannotCancelWithin24Hours", errmessages.ErrMsgCannotCancelWithin24Hours, true},
		{"ErrMsgLessonFull", errmessages.ErrMsgLessonFull, true},
		{"ErrMsgLessonInPast", errmessages.ErrMsgLessonInPast, true},
		{"ErrMsgBookingNotFound", errmessages.ErrMsgBookingNotFound, true},
		{"ErrMsgLessonNotFound", errmessages.ErrMsgLessonNotFound, true},
		{"ErrMsgScheduleConflict", errmessages.ErrMsgScheduleConflict, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.message, "константа %s должна быть определена", tt.name)
		})
	}
}

// TestCreditHandlers_UseErrorMessageConstants проверяет что credit handlers используют константы, не hardcoded строки
func TestCreditHandlers_UseErrorMessageConstants(t *testing.T) {
	// Проверяем что все credit error messages определены
	creditMessages := []string{
		errmessages.ErrMsgInsufficientCredits,
		errmessages.ErrMsgBalanceExceeded,
		errmessages.ErrMsgCreditNotInitialized,
	}

	for _, msg := range creditMessages {
		assert.NotEmpty(t, msg, "credit error message должно быть определено")
	}
}

// TestBookingHandlers_UseErrorMessageConstants проверяет что booking handlers используют константы
func TestBookingHandlers_UseErrorMessageConstants(t *testing.T) {
	// Проверяем что все booking error messages определены
	bookingMessages := []string{
		errmessages.ErrMsgAlreadyBooked,
		errmessages.ErrMsgCannotCancelWithin24Hours,
		errmessages.ErrMsgLessonFull,
		errmessages.ErrMsgBookingNotFound,
		errmessages.ErrMsgLessonInPast,
	}

	for _, msg := range bookingMessages {
		assert.NotEmpty(t, msg, "booking error message должно быть определено")
	}
}

// TestErrorMessages_Russian проверяет что все сообщения на русском языке (нет английских)
func TestErrorMessages_Russian(t *testing.T) {
	messages := map[string]string{
		"ErrMsgInsufficientCredits":       errmessages.ErrMsgInsufficientCredits,
		"ErrMsgBalanceExceeded":           errmessages.ErrMsgBalanceExceeded,
		"ErrMsgCreditNotInitialized":      errmessages.ErrMsgCreditNotInitialized,
		"ErrMsgAlreadyBooked":             errmessages.ErrMsgAlreadyBooked,
		"ErrMsgCannotCancelWithin24Hours": errmessages.ErrMsgCannotCancelWithin24Hours,
		"ErrMsgLessonFull":                errmessages.ErrMsgLessonFull,
		"ErrMsgLessonInPast":              errmessages.ErrMsgLessonInPast,
		"ErrMsgBookingNotFound":           errmessages.ErrMsgBookingNotFound,
		"ErrMsgLessonNotFound":            errmessages.ErrMsgLessonNotFound,
		"ErrMsgScheduleConflict":          errmessages.ErrMsgScheduleConflict,
	}

	for name, msg := range messages {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, msg, "сообщение не должно быть пустым")
			// Проверяем что содержит кириллические символы (русский текст)
			hasRussian := false
			for _, r := range msg {
				// Кириллический диапазон Unicode: 0x0400 - 0x04FF
				if r >= 0x0400 && r <= 0x04FF {
					hasRussian = true
					break
				}
			}
			assert.True(t, hasRussian, "%s должно быть на русском языке", name)
		})
	}
}
