package service

import "fmt"

// Localization содержит все локализованные сообщения платформы
type Localization struct {
	// Booking сообщения
	BookingCreated                string
	BookingCancelled              string
	BookingCannotReBook           string
	BookingAdminCreatedForStudent string
	BookingAdminCancelled         string

	// Broadcast сообщения
	BroadcastFromTeacher string
	BroadcastFromAdmin   string
	NewHomework          string
	HomeworkAdded        string

	// Payment сообщения
	PaymentDisabled string
	PaymentSuccess  string
	PaymentFailed   string
	PaymentRefunded string

	// Credit операции
	CreditDeducted   string
	CreditRefunded   string
	CreditAdded      string
	CreditAdminAdded string

	// Telegram сообщения
	TelegramLinked        string
	TelegramUnlinked      string
	TelegramSubscribed    string
	TelegramUnsubscribed  string
	TelegramBotBlocked    string
	TelegramAlreadyLinked string

	// Error сообщения - Booking
	ErrorInsufficientCredits       string
	ErrorLessonFull                string
	ErrorCannotCancelWithin24h     string
	ErrorCannotEditPastLesson      string
	ErrorLessonPreviouslyCancelled string
	ErrorAlreadyBooked             string
	ErrorScheduleConflict          string
	ErrorBookingTooLate            string

	// Error сообщения - Authorization
	ErrorUnauthorized       string
	ErrorForbidden          string
	ErrorInvalidCredentials string
	ErrorSessionExpired     string
	ErrorInvalidCSRF        string

	// Error сообщения - Validation
	ErrorValidationFailed  string
	ErrorInvalidInput      string
	ErrorMissingField      string
	ErrorInvalidEmail      string
	ErrorInvalidLessonType string
	ErrorInvalidTime       string

	// Error сообщения - Resources
	ErrorNotFound        string
	ErrorAlreadyExists   string
	ErrorConflict        string
	ErrorUserNotFound    string
	ErrorLessonNotFound  string
	ErrorBookingNotFound string

	// Error сообщения - Business Logic
	ErrorCannotChangeToIndividual   string
	ErrorInvalidLessonTypeUpdate    string
	ErrorInvalidSwap                string
	ErrorPaymentDisabledForUser     string
	ErrorUsersNotLinkedToTelegram   string
	ErrorBroadcastAlreadyInProgress string

	// Error сообщения - Server
	ErrorInternalServer    string
	ErrorDatabaseError     string
	ErrorRateLimitExceeded string

	// Success сообщения
	SuccessLessonCreated    string
	SuccessLessonUpdated    string
	SuccessLessonDeleted    string
	SuccessBulkEditApplied  string
	SuccessCreditsAdded     string
	SuccessBroadcastCreated string
	SuccessBroadcastSent    string
	SuccessHomeworkUploaded string
	SuccessHomeworkDeleted  string
}

// NewLocalization создает новую локализацию с русскими сообщениями
func NewLocalization() *Localization {
	return &Localization{
		// Booking сообщения
		BookingCreated:                "Вы записаны на занятие %s (%s)",
		BookingCancelled:              "Вы отписались от занятия %s (%s)",
		BookingCannotReBook:           "Вы отписались от этого занятия и больше не можете на него записаться",
		BookingAdminCreatedForStudent: "Администратор записал вас на занятие %s (%s)",
		BookingAdminCancelled:         "Администратор отменил вашу запись на занятие %s (%s)",

		// Broadcast сообщения
		BroadcastFromTeacher: "Сообщение от преподавателя по занятию %s:\n\n%s",
		BroadcastFromAdmin:   "Объявление от администрации:\n\n%s",
		NewHomework:          "Новое домашнее задание по занятию %s",
		HomeworkAdded:        "Домашнее задание добавлено к занятию %s",

		// Payment сообщения
		PaymentDisabled: "Платежи временно недоступны. Обратитесь к администратору.",
		PaymentSuccess:  "Оплата успешно проведена. Зачислено %d кредитов.",
		PaymentFailed:   "Ошибка при обработке платежа. Пожалуйста, попробуйте позже.",
		PaymentRefunded: "Платеж возвращен. Зачислено %d кредитов.",

		// Credit операции
		CreditDeducted:   "Списан 1 кредит за запись на занятие",
		CreditRefunded:   "Возврат 1 кредита за отмену записи",
		CreditAdded:      "Пополнение счета на %d кредитов",
		CreditAdminAdded: "Администратор добавил вам %d кредитов",

		// Telegram сообщения
		TelegramLinked:        "Ваш Telegram аккаунт успешно привязан",
		TelegramUnlinked:      "Ваш Telegram аккаунт отвязан",
		TelegramSubscribed:    "Вы подписаны на уведомления в Telegram",
		TelegramUnsubscribed:  "Вы отписаны от уведомлений в Telegram",
		TelegramBotBlocked:    "Бот заблокирован пользователем. Отписываем от уведомлений.",
		TelegramAlreadyLinked: "Этот Telegram аккаунт уже привязан к другому пользователю платформы",

		// Error сообщения - Booking
		ErrorInsufficientCredits:       "Недостаточно кредитов для записи на занятие",
		ErrorLessonFull:                "На занятии нет свободных мест",
		ErrorCannotCancelWithin24h:     "Нельзя отменить запись менее чем за 24 часа до начала занятия",
		ErrorCannotEditPastLesson:      "Вы не можете редактировать занятия которые уже прошли",
		ErrorLessonPreviouslyCancelled: "Вы отписались от этого занятия и больше не можете на него записаться",
		ErrorAlreadyBooked:             "Вы уже записаны на это занятие",
		ErrorScheduleConflict:          "У вас уже есть запись на другое занятие в это время",
		ErrorBookingTooLate:            "Запись на занятие должна быть сделана не менее чем за 24 часа до начала",

		// Error сообщения - Authorization
		ErrorUnauthorized:       "У вас нет доступа к этому ресурсу",
		ErrorForbidden:          "Доступ запрещен. Недостаточно прав.",
		ErrorInvalidCredentials: "Неверный email или пароль",
		ErrorSessionExpired:     "Ваша сессия истекла. Пожалуйста, войдите снова.",
		ErrorInvalidCSRF:        "Неверный CSRF токен. Обновите страницу и попробуйте снова.",

		// Error сообщения - Validation
		ErrorValidationFailed:  "Ошибка валидации данных",
		ErrorInvalidInput:      "Некорректные входные данные",
		ErrorMissingField:      "Обязательное поле не заполнено",
		ErrorInvalidEmail:      "Некорректный email адрес",
		ErrorInvalidLessonType: "Некорректный тип занятия. Допустимые значения: individual, group",
		ErrorInvalidTime:       "Некорректное время занятия. Конец занятия должен быть позже начала.",

		// Error сообщения - Resources
		ErrorNotFound:        "Ресурс не найден",
		ErrorAlreadyExists:   "Ресурс с такими данными уже существует",
		ErrorConflict:        "Конфликт данных. Ресурс был изменен другим пользователем.",
		ErrorUserNotFound:    "Пользователь не найден",
		ErrorLessonNotFound:  "Занятие не найдено",
		ErrorBookingNotFound: "Запись на занятие не найдена",

		// Error сообщения - Business Logic
		ErrorCannotChangeToIndividual:   "Нельзя изменить тип занятия на индивидуальное, пока на нем несколько студентов. Удалите лишних студентов.",
		ErrorInvalidLessonTypeUpdate:    "Некорректное изменение типа занятия или количества мест",
		ErrorInvalidSwap:                "Невозможно выполнить обмен занятий",
		ErrorPaymentDisabledForUser:     "Платежи отключены для вашего аккаунта. Обратитесь к администратору.",
		ErrorUsersNotLinkedToTelegram:   "Некоторые пользователи не привязаны к Telegram",
		ErrorBroadcastAlreadyInProgress: "Рассылка уже выполняется или завершена",

		// Error сообщения - Server
		ErrorInternalServer:    "Внутренняя ошибка сервера. Пожалуйста, попробуйте позже.",
		ErrorDatabaseError:     "Ошибка при работе с базой данных",
		ErrorRateLimitExceeded: "Превышен лимит запросов. Пожалуйста, подождите немного.",

		// Success сообщения
		SuccessLessonCreated:    "Занятие успешно создано",
		SuccessLessonUpdated:    "Занятие успешно обновлено",
		SuccessLessonDeleted:    "Занятие успешно удалено",
		SuccessBulkEditApplied:  "Изменения применены к %d занятиям",
		SuccessCreditsAdded:     "Добавлено %d кредитов",
		SuccessBroadcastCreated: "Рассылка создана и запущена",
		SuccessBroadcastSent:    "Рассылка успешно отправлена",
		SuccessHomeworkUploaded: "Домашнее задание успешно загружено",
		SuccessHomeworkDeleted:  "Домашнее задание удалено",
	}
}

// --- Вспомогательные функции форматирования ---

// FormatBookingCreated форматирует сообщение о создании записи
func (l *Localization) FormatBookingCreated(lessonName, dateTime string) string {
	return fmt.Sprintf(l.BookingCreated, lessonName, dateTime)
}

// FormatBookingCancelled форматирует сообщение об отмене записи
func (l *Localization) FormatBookingCancelled(lessonName, dateTime string) string {
	return fmt.Sprintf(l.BookingCancelled, lessonName, dateTime)
}

// FormatBookingAdminCreated форматирует сообщение о записи администратором
func (l *Localization) FormatBookingAdminCreated(lessonName, dateTime string) string {
	return fmt.Sprintf(l.BookingAdminCreatedForStudent, lessonName, dateTime)
}

// FormatBookingAdminCancelled форматирует сообщение об отмене записи администратором
func (l *Localization) FormatBookingAdminCancelled(lessonName, dateTime string) string {
	return fmt.Sprintf(l.BookingAdminCancelled, lessonName, dateTime)
}

// FormatBroadcastFromTeacher форматирует сообщение от преподавателя
func (l *Localization) FormatBroadcastFromTeacher(lessonName, message string) string {
	return fmt.Sprintf(l.BroadcastFromTeacher, lessonName, message)
}

// FormatBroadcastFromAdmin форматирует объявление от администрации
func (l *Localization) FormatBroadcastFromAdmin(message string) string {
	return fmt.Sprintf(l.BroadcastFromAdmin, message)
}

// FormatNewHomework форматирует уведомление о новом домашнем задании
func (l *Localization) FormatNewHomework(lessonName string) string {
	return fmt.Sprintf(l.NewHomework, lessonName)
}

// FormatHomeworkAdded форматирует сообщение о добавлении домашнего задания
func (l *Localization) FormatHomeworkAdded(lessonName string) string {
	return fmt.Sprintf(l.HomeworkAdded, lessonName)
}

// FormatPaymentSuccess форматирует сообщение об успешной оплате
func (l *Localization) FormatPaymentSuccess(credits int) string {
	return fmt.Sprintf(l.PaymentSuccess, credits)
}

// FormatPaymentRefunded форматирует сообщение о возврате платежа
func (l *Localization) FormatPaymentRefunded(credits int) string {
	return fmt.Sprintf(l.PaymentRefunded, credits)
}

// FormatCreditAdded форматирует сообщение о пополнении кредитов
func (l *Localization) FormatCreditAdded(amount int) string {
	return fmt.Sprintf(l.CreditAdded, amount)
}

// FormatCreditAdminAdded форматирует сообщение о добавлении кредитов администратором
func (l *Localization) FormatCreditAdminAdded(amount int) string {
	return fmt.Sprintf(l.CreditAdminAdded, amount)
}

// FormatBulkEditApplied форматирует сообщение о применении массовых изменений
func (l *Localization) FormatBulkEditApplied(count int) string {
	return fmt.Sprintf(l.SuccessBulkEditApplied, count)
}

// FormatCreditsAdded форматирует сообщение об успешном добавлении кредитов
func (l *Localization) FormatCreditsAdded(amount int) string {
	return fmt.Sprintf(l.SuccessCreditsAdded, amount)
}

// --- Вспомогательные функции для правильного склонения ---

// FormatCreditsWithDeclension форматирует количество кредитов с правильным склонением
// 1 кредит, 2 кредита, 5 кредитов
func FormatCreditsWithDeclension(count int) string {
	lastDigit := count % 10
	lastTwoDigits := count % 100

	if lastTwoDigits >= 11 && lastTwoDigits <= 14 {
		return fmt.Sprintf("%d кредитов", count)
	}

	switch lastDigit {
	case 1:
		return fmt.Sprintf("%d кредит", count)
	case 2, 3, 4:
		return fmt.Sprintf("%d кредита", count)
	default:
		return fmt.Sprintf("%d кредитов", count)
	}
}

// FormatLessonsWithDeclension форматирует количество занятий с правильным склонением
// 1 занятие, 2 занятия, 5 занятий
func FormatLessonsWithDeclension(count int) string {
	lastDigit := count % 10
	lastTwoDigits := count % 100

	if lastTwoDigits >= 11 && lastTwoDigits <= 14 {
		return fmt.Sprintf("%d занятий", count)
	}

	switch lastDigit {
	case 1:
		return fmt.Sprintf("%d занятие", count)
	case 2, 3, 4:
		return fmt.Sprintf("%d занятия", count)
	default:
		return fmt.Sprintf("%d занятий", count)
	}
}
