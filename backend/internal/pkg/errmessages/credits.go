package errmessages

const (
	// Credits errors
	ErrMsgInsufficientCredits  = "Недостаточно кредитов для выполнения операции"
	ErrMsgBalanceExceeded      = "Баланс кредитов не может превышать 10000"
	ErrMsgCreditNotInitialized = "Учетная запись кредитов не инициализирована. Свяжитесь с администратором"
	ErrMsgCreditDuplicate      = "Учетная запись кредитов уже существует"
	ErrMsgInvalidAmount        = "Некорректная сумма (должна быть от -100 до 100)"
	ErrMsgInvalidReason        = "Причина должна быть не менее 3 символов"
	ErrMsgUserNotFound         = "Пользователь не найден"
)

const (
	// Booking errors
	ErrMsgAlreadyBooked             = "Вы уже записаны на это занятие"
	ErrMsgCannotCancelWithin24Hours = "Запись нельзя отменить менее чем за 24 часа до начала занятия"
	ErrMsgLessonFull                = "Это занятие полностью заполнено"
	ErrMsgLessonInPast              = "Нельзя записаться на занятие которое уже прошло"
	ErrMsgBookingNotFound           = "Бронирование не найдено"
	ErrMsgLessonNotFound            = "Занятие не найдено"
	ErrMsgScheduleConflict          = "У вас уже есть запись на это время"
	ErrMsgLessonNotAvailable        = "Это занятие недоступно для записи"
	ErrMsgBookingNotActive          = "Это бронирование неактивно"
	ErrMsgLessonPreviouslyCancelled = "Вы отписались от этого занятия и больше не можете на него записаться"
	ErrMsgUnauthorized              = "Вам не разрешено выполнять эту операцию"
	ErrMsgOperationFailed           = "Операция не выполнена. Попробуйте позже или свяжитесь с поддержкой"
)
