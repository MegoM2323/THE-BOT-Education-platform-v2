package repository

// SQL snippet constants для переиспользования SELECT field lists
// При добавлении поля в таблицу нужно изменить только в этом файле

const (
	// UserSelectFields - поля таблицы users
	UserSelectFields = `
		id, email, password_hash, full_name, role, payment_enabled, telegram_username,
		parent_telegram_username, parent_chat_id, created_at, updated_at, deleted_at
	`

	// LessonSelectFields - поля таблицы lessons
	LessonSelectFields = `
		id, teacher_id, start_time, end_time,
		max_students, current_students, credits_cost, color, subject, homework_text, report_text, link,
		applied_from_template, template_application_id,
		is_recurring, recurring_group_id, recurring_end_date,
		created_at, updated_at, deleted_at
	`

	// BookingSelectFields - поля таблицы bookings
	BookingSelectFields = `
		id, student_id, lesson_id, status,
		booked_at, cancelled_at, created_at, updated_at
	`

	// CreditSelectFields - поля таблицы credits
	CreditSelectFields = `
		id, user_id, balance, created_at, updated_at
	`

	// MessageSelectFields - поля таблицы messages
	MessageSelectFields = `
		id, room_id, sender_id, message_text, status,
		moderation_completed_at, created_at, deleted_at
	`

	// PaymentSelectFields - поля таблицы payments
	PaymentSelectFields = `
		id, user_id, yookassa_payment_id, amount, credits,
		status, confirmation_url, idempotency_key, processed_at, created_at, updated_at
	`

	// LessonTemplateSelectFields - поля таблицы lesson_templates
	LessonTemplateSelectFields = `
		id, admin_id, name, description,
		created_at, updated_at, deleted_at
	`

	// TemplateLessonSelectFields - поля таблицы template_lessons
	TemplateLessonSelectFields = `
		id, template_id, day_of_week, start_time, end_time,
		teacher_id, lesson_type, max_students, credits_cost, color, subject, description,
		created_at, updated_at
	`

	// ChatRoomSelectFields - поля таблицы chat_rooms
	ChatRoomSelectFields = `
		id, teacher_id, student_id, last_message_at,
		created_at, updated_at, deleted_at
	`

	// FileAttachmentSelectFields - поля таблицы file_attachments
	FileAttachmentSelectFields = `
		id, message_id, file_name, file_path, file_size, mime_type, uploaded_at
	`

	// SwapSelectFields - поля таблицы swaps
	SwapSelectFields = `
		id, student_id, old_lesson_id, new_lesson_id,
		old_booking_id, new_booking_id, created_at
	`

	// TemplateLessonStudentSelectFields - поля таблицы template_lesson_students
	TemplateLessonStudentSelectFields = `
		id, template_lesson_id, student_id, created_at
	`
)
