package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Ошибки репозитория
var (
	// Ошибки пользователя
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrUserExists      = errors.New("пользователь уже существует")
	ErrInvalidUserData = errors.New("некорректные данные пользователя")

	// Ошибки урока
	ErrLessonNotFound        = errors.New("урок не найден")
	ErrLessonFull            = errors.New("урок заполнен")
	ErrLessonOverlapConflict = errors.New("расписание преподавателя конфликтует с существующим уроком")
	ErrLessonTimeInvalid     = errors.New("некорректное время урока")

	// Ошибки бронирования
	ErrBookingNotFound           = errors.New("бронирование не найдено")
	ErrBookingNotActive          = errors.New("бронирование не активно")
	ErrScheduleConflict          = errors.New("обнаружен конфликт расписания")
	ErrAlreadyBooked             = errors.New("урок уже забронирован")
	ErrDuplicateBooking          = errors.New("вы уже забронировали это занятие")
	ErrLessonPreviouslyCancelled = errors.New("вы отписались от этого занятия и больше не можете на него записаться")
	ErrUnauthorized              = errors.New("доступ запрещен")

	// Ошибки кредитов
	ErrCreditNotFound      = errors.New("счет кредитов не найден")
	ErrInsufficientCredits = errors.New("недостаточно кредитов")
	ErrDuplicateCredit     = errors.New("кредиты для этого пользователя уже существуют")

	// Ошибки сессии
	ErrSessionNotFound = errors.New("сессия не найдена")
	ErrSessionExpired  = errors.New("срок действия сессии истек")

	// Ошибки обмена
	ErrSwapNotFound = errors.New("обмен не найден")

	// Ошибки Telegram
	ErrTelegramUserNotFound      = errors.New("привязка Telegram не найдена")
	ErrTelegramUserAlreadyLinked = errors.New("пользователь уже привязан к Telegram")
	ErrTelegramIDAlreadyLinked   = errors.New("этот Telegram ID уже привязан к другому пользователю")
	ErrTelegramUsernameInUse     = errors.New("это Telegram имя уже используется другим пользователем")
	ErrBroadcastListNotFound     = errors.New("список рассылки не найден")
	ErrBroadcastNotFound         = errors.New("рассылка не найдена")

	// Ошибки чата
	ErrChatRoomNotFound   = errors.New("комната чата не найдена")
	ErrMessageNotFound    = errors.New("сообщение не найдено")
	ErrAttachmentNotFound = errors.New("вложение не найдено")

	// Ошибки платежей
	ErrPaymentNotFound        = errors.New("платеж не найден")
	ErrPaymentDisabledForUser = errors.New("платежи отключены для пользователя")
	ErrInvalidUserRole        = errors.New("некорректная роль пользователя")

	// Ошибки отменённых бронирований
	ErrCancelledNotFound = errors.New("отменённое бронирование не найдено")

	// Ошибки домашних заданий
	ErrHomeworkNotFound = errors.New("домашнее задание не найдено")

	// Ошибки рассылок по урокам
	ErrLessonBroadcastNotFound = errors.New("рассылка урока не найдена")

	// Ошибки предметов
	ErrSubjectNotFound = errors.New("предмет не найден")

	// Ошибки удаления пользователей
	ErrUserHasActiveLessons  = errors.New("невозможно удалить преподавателя с активными занятиями: сначала soft-delete занятия")
	ErrUserHasActiveBookings = errors.New("невозможно удалить студента с активными бронированиями: сначала отмените бронирования")
	ErrUserHasTransactions   = errors.New("невозможно удалить пользователя с кредитными транзакциями: используйте soft-delete")
	ErrUserHasPayments       = errors.New("невозможно удалить пользователя с историей платежей: используйте soft-delete")

	// Ошибки удаления уроков
	ErrLessonHasActiveBookings = errors.New("невозможно удалить урок с активными бронированиями: сначала отмените бронирования и верните кредиты")
	ErrLessonHasHomework       = errors.New("невозможно удалить урок с домашним заданием: сначала удалите домашнее задание")
	ErrLessonHasBroadcasts     = errors.New("невозможно удалить урок с рассылками: сначала удалите рассылки")
	ErrHardDeleteRequiresForce = errors.New("жёсткое удаление требует force=true")
)

// IsUniqueViolationError проверяет, вызвана ли ошибка нарушением UNIQUE constraint в PostgreSQL
// Это используется для преобразования database-level ошибок в domain-level ошибки
// Код ошибки 23505 = UNIQUE constraint violation in PostgreSQL
func IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL SQLSTATE 23505 = unique_violation
		// Подробнее: https://www.postgresql.org/docs/current/errcodes-appendix.html
		return pgErr.SQLState() == "23505"
	}
	return false
}

// IsExclusionViolationError проверяет, вызвана ли ошибка нарушением EXCLUDE constraint в PostgreSQL
// Код ошибки 23P01 = exclusion_violation
// Используется для детектирования конфликтов расписания преподавателя
func IsExclusionViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL SQLSTATE 23P01 = exclusion_violation
		// Подробнее: https://www.postgresql.org/docs/current/errcodes-appendix.html
		return pgErr.SQLState() == "23P01"
	}
	return false
}
