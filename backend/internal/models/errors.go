package models

import "errors"

// Доменные ошибки для валидации моделей
var (
	// Ошибки пользователя
	ErrInvalidEmail          = errors.New("некорректный адрес email")
	ErrPasswordTooShort      = errors.New("пароль должен быть не менее 8 символов")
	ErrInvalidFullName       = errors.New("полное имя обязательно")
	ErrInvalidRole           = errors.New("некорректная роль пользователя")
	ErrInvalidUserID         = errors.New("некорректный ID пользователя")
	ErrInvalidTeacherID      = errors.New("некорректный ID преподавателя")
	ErrInvalidStudentID      = errors.New("некорректный ID студента")
	ErrInvalidPerformedBy    = errors.New("некорректный ID исполнителя (performed_by)")
	ErrInvalidTelegramHandle = errors.New("некорректное имя пользователя Telegram (должно быть от 3 до 32 символов, только буквы, цифры и подчеркивание)")

	// Ошибки урока
	ErrInvalidLessonID             = errors.New("некорректный ID урока")
	ErrInvalidLessonTime           = errors.New("некорректное время урока")
	ErrInvalidMaxStudents          = errors.New("максимальное количество студентов должно быть больше 0")
	ErrInvalidColor                = errors.New("цвет должен быть в формате #RRGGBB (например, #3B82F6)")
	ErrSubjectTooLong              = errors.New("тема урока не должна превышать 200 символов")
	ErrIndividualLessonMaxStudents = errors.New("индивидуальный урок должен иметь ровно 1 студента")
	ErrGroupLessonMinStudents      = errors.New("групповой урок должен иметь минимум 4 студента")
	ErrCannotChangeIndividualMax   = errors.New("невозможно изменить максимальное количество студентов для индивидуального урока")
	ErrCannotChangeToIndividual    = errors.New("невозможно изменить групповой урок с несколькими студентами на индивидуальный")

	// Ошибки бронирования
	ErrInvalidBookingID = errors.New("некорректный ID бронирования")

	// Ошибки кредитов
	ErrInvalidCreditAmount = errors.New("количество кредитов должно быть от 1 до 100")
	ErrInvalidReason       = errors.New("причина обязательна")
	ErrBalanceExceeded     = errors.New("баланс превысит максимальный лимит (10000 кредитов)")

	// Ошибки обмена
	ErrInvalidOldLessonID = errors.New("некорректный ID старого урока")
	ErrInvalidNewLessonID = errors.New("некорректный ID нового урока")
	ErrSameLessonSwap     = errors.New("старый и новый урок не могут быть одинаковыми")

	// Ошибки запроса на пробный урок
	ErrInvalidName        = errors.New("имя должно быть от 2 до 100 символов")
	ErrInvalidPhone       = errors.New("номер телефона обязателен")
	ErrInvalidTelegram    = errors.New("telegram должен быть от 3 до 50 символов")
	ErrInvalidEmailFormat = errors.New("некорректный формат email")

	// Ошибки Telegram
	ErrInvalidToken            = errors.New("некорректный токен привязки")
	ErrInvalidTelegramID       = errors.New("некорректный Telegram ID")
	ErrInvalidChatID           = errors.New("некорректный Chat ID")
	ErrInvalidBroadcastName    = errors.New("название списка рассылки должно быть не менее 3 символов")
	ErrInvalidBroadcastUsers   = errors.New("список рассылки должен содержать хотя бы одного пользователя")
	ErrInvalidBroadcastListID  = errors.New("некорректный ID списка рассылки")
	ErrInvalidBroadcastMessage = errors.New("сообщение рассылки обязательно")
	ErrBroadcastMessageTooLong = errors.New("сообщение рассылки не должно превышать 4096 символов")

	// Ошибки предметов
	ErrInvalidSubject            = errors.New("некорректный предмет")
	ErrInvalidSubjectID          = errors.New("некорректный ID предмета")
	ErrInvalidSubjectName        = errors.New("название предмета должно быть от 2 до 100 символов")
	ErrSubjectDescriptionTooLong = errors.New("описание предмета не должно превышать 500 символов")

	// Ошибки сообщений
	ErrInvalidMessageID      = errors.New("некорректный ID сообщения")
	ErrInvalidRecipientID    = errors.New("некорректный ID получателя")
	ErrMessageContentEmpty   = errors.New("содержимое сообщения не может быть пустым")
	ErrMessageContentTooLong = errors.New("сообщение не должно превышать 4096 символов")
	ErrInvalidMessageLimit   = errors.New("лимит сообщений должен быть положительным числом")
	ErrMessageLimitTooHigh   = errors.New("максимальный лимит 100 сообщений за раз")
	ErrInvalidMessageOffset  = errors.New("смещение сообщений не может быть отрицательным")

	// Ошибки чатов
	ErrCannotChatWithSelf = errors.New("невозможно создать чат с самим собой")
	ErrInvalidChatRoomID  = errors.New("некорректный ID комнаты чата")

	// Ошибки статуса сообщений
	ErrInvalidMessageStatus = errors.New("некорректный статус сообщения (разрешены: delivered, blocked)")

	// Ошибки файловых вложений
	ErrInvalidFileName = errors.New("имя файла обязательно")
	ErrInvalidFileSize = errors.New("размер файла должен быть от 1 байта до 10MB")
	ErrInvalidMimeType = errors.New("MIME тип файла обязателен")

	// Ошибки блокировки сообщений
	ErrInvalidBlockReason = errors.New("причина блокировки обязательна")

	// Ошибки платежей
	ErrInvalidPaymentStatus = errors.New("некорректный статус платежа")
	ErrInvalidPaymentID     = errors.New("некорректный ID платежа")

	// Ошибки рассылок по урокам
	ErrInvalidBroadcastStatus = errors.New("некорректный статус рассылки")
	ErrInvalidBroadcastID     = errors.New("некорректный ID рассылки")
	ErrTooManyFiles           = errors.New("превышено максимальное количество файлов (максимум 10)")

	// Ошибки домашних заданий
	ErrInvalidHomeworkID      = errors.New("некорректный ID домашнего задания")
	ErrFileNameTooLong        = errors.New("имя файла не должно превышать 255 символов")
	ErrFileTooLarge           = errors.New("размер файла превышает максимально допустимый (10MB)")
	ErrMimeTypeNotAllowed     = errors.New("данный тип файла не разрешен для загрузки")
	ErrFileStorageFailed      = errors.New("не удалось сохранить файл на сервере")
	ErrHomeworkContentTooLong = errors.New("описание домашнего задания не должно превышать 10000 символов")
)
