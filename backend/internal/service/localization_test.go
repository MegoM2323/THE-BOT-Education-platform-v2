package service

import (
	"strings"
	"testing"
	"unicode"
)

// TestLocalizationAllMessagesPresent проверяет что все сообщения инициализированы
func TestLocalizationAllMessagesPresent(t *testing.T) {
	loc := NewLocalization()

	tests := []struct {
		name    string
		message string
	}{
		// Booking сообщения
		{"BookingCreated", loc.BookingCreated},
		{"BookingCancelled", loc.BookingCancelled},
		{"BookingCannotReBook", loc.BookingCannotReBook},
		{"BookingAdminCreatedForStudent", loc.BookingAdminCreatedForStudent},
		{"BookingAdminCancelled", loc.BookingAdminCancelled},

		// Broadcast сообщения
		{"BroadcastFromTeacher", loc.BroadcastFromTeacher},
		{"BroadcastFromAdmin", loc.BroadcastFromAdmin},
		{"NewHomework", loc.NewHomework},
		{"HomeworkAdded", loc.HomeworkAdded},

		// Payment сообщения
		{"PaymentDisabled", loc.PaymentDisabled},
		{"PaymentSuccess", loc.PaymentSuccess},
		{"PaymentFailed", loc.PaymentFailed},

		// Error сообщения
		{"ErrorInsufficientCredits", loc.ErrorInsufficientCredits},
		{"ErrorLessonFull", loc.ErrorLessonFull},
		{"ErrorCannotCancelWithin24h", loc.ErrorCannotCancelWithin24h},
		{"ErrorCannotEditPastLesson", loc.ErrorCannotEditPastLesson},
		{"ErrorLessonPreviouslyCancelled", loc.ErrorLessonPreviouslyCancelled},
		{"ErrorUnauthorized", loc.ErrorUnauthorized},
		{"ErrorForbidden", loc.ErrorForbidden},
		{"ErrorNotFound", loc.ErrorNotFound},
		{"ErrorInternalServer", loc.ErrorInternalServer},

		// Success сообщения
		{"SuccessLessonCreated", loc.SuccessLessonCreated},
		{"SuccessTemplateCreated", loc.SuccessTemplateCreated},
		{"SuccessBulkEditApplied", loc.SuccessBulkEditApplied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message == "" {
				t.Errorf("%s is empty", tt.name)
			}
		})
	}
}

// TestLocalizationMessagesInRussian проверяет что сообщения на русском языке
func TestLocalizationMessagesInRussian(t *testing.T) {
	loc := NewLocalization()

	tests := []struct {
		name    string
		message string
	}{
		{"BookingCreated", loc.BookingCreated},
		{"ErrorInsufficientCredits", loc.ErrorInsufficientCredits},
		{"ErrorLessonFull", loc.ErrorLessonFull},
		{"PaymentDisabled", loc.PaymentDisabled},
		{"SuccessLessonCreated", loc.SuccessLessonCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверяем наличие кириллических символов
			hasCyrillic := false
			for _, r := range tt.message {
				if unicode.Is(unicode.Cyrillic, r) {
					hasCyrillic = true
					break
				}
			}

			if !hasCyrillic {
				t.Errorf("%s should contain Cyrillic characters (Russian text): %s", tt.name, tt.message)
			}
		})
	}
}

// TestLocalizationFormatFunctions проверяет функции форматирования
func TestLocalizationFormatFunctions(t *testing.T) {
	loc := NewLocalization()

	tests := []struct {
		name     string
		fn       func() string
		expected []string // Подстроки которые должны присутствовать
	}{
		{
			name:     "FormatBookingCreated",
			fn:       func() string { return loc.FormatBookingCreated("Математика", "2024-03-15 10:00") },
			expected: []string{"Математика", "2024-03-15 10:00"},
		},
		{
			name:     "FormatBookingCancelled",
			fn:       func() string { return loc.FormatBookingCancelled("Физика", "2024-03-16 14:00") },
			expected: []string{"Физика", "2024-03-16 14:00"},
		},
		{
			name: "FormatBroadcastFromTeacher",
			fn: func() string {
				return loc.FormatBroadcastFromTeacher("Химия", "Завтра контрольная")
			},
			expected: []string{"Химия", "Завтра контрольная"},
		},
		{
			name:     "FormatPaymentSuccess",
			fn:       func() string { return loc.FormatPaymentSuccess(10) },
			expected: []string{"10", "кредит"},
		},
		{
			name:     "FormatBulkEditApplied",
			fn:       func() string { return loc.FormatBulkEditApplied(5) },
			expected: []string{"5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()

			if result == "" {
				t.Error("Formatted message is empty")
			}

			for _, substr := range tt.expected {
				if !strings.Contains(result, substr) {
					t.Errorf("Result should contain '%s', got: %s", substr, result)
				}
			}
		})
	}
}

// TestLocalizationTemplateValidation проверяет правильность шаблонов
func TestLocalizationTemplateValidation(t *testing.T) {
	loc := NewLocalization()

	tests := []struct {
		name           string
		template       string
		expectedParams int // Ожидаемое количество %s или %d
	}{
		{"BookingCreated", loc.BookingCreated, 2},                 // %s, %s
		{"BroadcastFromTeacher", loc.BroadcastFromTeacher, 2},     // %s, %s
		{"PaymentSuccess", loc.PaymentSuccess, 1},                 // %d
		{"SuccessBulkEditApplied", loc.SuccessBulkEditApplied, 1}, // %d
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подсчитываем количество параметров форматирования
			countS := strings.Count(tt.template, "%s")
			countD := strings.Count(tt.template, "%d")
			countV := strings.Count(tt.template, "%v")
			total := countS + countD + countV

			if total != tt.expectedParams {
				t.Errorf("Template %s has %d parameters (%d %%s, %d %%d, %d %%v), expected %d",
					tt.name, total, countS, countD, countV, tt.expectedParams)
			}
		})
	}
}

// TestFormatCreditsWithDeclension проверяет правильное склонение слова "кредит"
func TestFormatCreditsWithDeclension(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{1, "1 кредит"},
		{2, "2 кредита"},
		{3, "3 кредита"},
		{4, "4 кредита"},
		{5, "5 кредитов"},
		{10, "10 кредитов"},
		{11, "11 кредитов"},
		{12, "12 кредитов"},
		{13, "13 кредитов"},
		{14, "14 кредитов"},
		{15, "15 кредитов"},
		{20, "20 кредитов"},
		{21, "21 кредит"},
		{22, "22 кредита"},
		{23, "23 кредита"},
		{24, "24 кредита"},
		{25, "25 кредитов"},
		{100, "100 кредитов"},
		{101, "101 кредит"},
		{102, "102 кредита"},
		{111, "111 кредитов"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatCreditsWithDeclension(tt.count)
			if result != tt.expected {
				t.Errorf("FormatCreditsWithDeclension(%d) = %s, want %s", tt.count, result, tt.expected)
			}
		})
	}
}

// TestFormatLessonsWithDeclension проверяет правильное склонение слова "занятие"
func TestFormatLessonsWithDeclension(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{1, "1 занятие"},
		{2, "2 занятия"},
		{3, "3 занятия"},
		{4, "4 занятия"},
		{5, "5 занятий"},
		{10, "10 занятий"},
		{11, "11 занятий"},
		{12, "12 занятий"},
		{13, "13 занятий"},
		{14, "14 занятий"},
		{15, "15 занятий"},
		{20, "20 занятий"},
		{21, "21 занятие"},
		{22, "22 занятия"},
		{25, "25 занятий"},
		{101, "101 занятие"},
		{102, "102 занятия"},
		{111, "111 занятий"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatLessonsWithDeclension(tt.count)
			if result != tt.expected {
				t.Errorf("FormatLessonsWithDeclension(%d) = %s, want %s", tt.count, result, tt.expected)
			}
		})
	}
}

// TestLocalizationNoEnglishErrorMessages проверяет что нет английских ошибок
func TestLocalizationNoEnglishErrorMessages(t *testing.T) {
	loc := NewLocalization()

	// Список английских ключевых слов которых не должно быть в русских сообщениях
	englishKeywords := []string{
		"error", "failed", "not found", "invalid", "unauthorized",
		"forbidden", "conflict", "already exists", "too late",
		"insufficient", "full", "cannot cancel", "cannot edit",
	}

	errorMessages := []struct {
		name    string
		message string
	}{
		{"ErrorInsufficientCredits", loc.ErrorInsufficientCredits},
		{"ErrorLessonFull", loc.ErrorLessonFull},
		{"ErrorCannotCancelWithin24h", loc.ErrorCannotCancelWithin24h},
		{"ErrorUnauthorized", loc.ErrorUnauthorized},
		{"ErrorForbidden", loc.ErrorForbidden},
		{"ErrorNotFound", loc.ErrorNotFound},
		{"ErrorInternalServer", loc.ErrorInternalServer},
	}

	for _, em := range errorMessages {
		t.Run(em.name, func(t *testing.T) {
			lowerMsg := strings.ToLower(em.message)
			for _, keyword := range englishKeywords {
				if strings.Contains(lowerMsg, keyword) {
					t.Errorf("%s contains English keyword '%s': %s", em.name, keyword, em.message)
				}
			}
		})
	}
}

// TestLocalizationConsistency проверяет консистентность сообщений
func TestLocalizationConsistency(t *testing.T) {
	loc := NewLocalization()

	// Проверяем что похожие сообщения используют одинаковую терминологию
	t.Run("Consistency_Credits", func(t *testing.T) {
		messages := []string{
			loc.ErrorInsufficientCredits,
			loc.CreditDeducted,
			loc.CreditRefunded,
			loc.PaymentSuccess,
		}

		// Все должны содержать слово "кредит"
		for _, msg := range messages {
			if !strings.Contains(strings.ToLower(msg), "кредит") {
				t.Errorf("Credit-related message should contain 'кредит': %s", msg)
			}
		}
	})

	t.Run("Consistency_Lessons", func(t *testing.T) {
		messages := []string{
			loc.ErrorLessonFull,
			loc.ErrorLessonNotFound,
			loc.SuccessLessonCreated,
			loc.ErrorCannotEditPastLesson,
		}

		// Все должны содержать слово "занятие" или "занятия"
		for _, msg := range messages {
			lowerMsg := strings.ToLower(msg)
			if !strings.Contains(lowerMsg, "занят") {
				t.Errorf("Lesson-related message should contain 'занятие': %s", msg)
			}
		}
	})

	t.Run("Consistency_Booking", func(t *testing.T) {
		messages := []string{
			loc.BookingCreated,
			loc.BookingCancelled,
			loc.ErrorAlreadyBooked,
		}

		// Все должны содержать слово "запис" или "писал" (записаны, отписались, записаны)
		for _, msg := range messages {
			lowerMsg := strings.ToLower(msg)
			if !strings.Contains(lowerMsg, "запис") && !strings.Contains(lowerMsg, "писал") {
				t.Errorf("Booking-related message should contain 'запис' or 'писал': %s", msg)
			}
		}
	})
}

// TestLocalizationUserFriendlyMessages проверяет что сообщения понятны пользователю
func TestLocalizationUserFriendlyMessages(t *testing.T) {
	loc := NewLocalization()

	// Сообщения об ошибках не должны содержать технические термины
	technicalTerms := []string{"sql", "database", "null", "undefined", "exception", "stack"}

	errorMessages := []string{
		loc.ErrorInsufficientCredits,
		loc.ErrorLessonFull,
		loc.ErrorNotFound,
		loc.ErrorInternalServer,
	}

	for _, msg := range errorMessages {
		lowerMsg := strings.ToLower(msg)
		for _, term := range technicalTerms {
			if strings.Contains(lowerMsg, term) {
				t.Errorf("User-facing error message should not contain technical term '%s': %s", term, msg)
			}
		}
	}
}
