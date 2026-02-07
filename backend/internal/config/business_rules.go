package config

import "time"

// BusinessRules содержит бизнес-правила и константы для приложения
// Все magic numbers из кода вынесены сюда для централизованного управления
type BusinessRules struct {
	// File Upload Rules
	MaxFileSize int64 // Максимальный размер загружаемого файла в байтах (10MB)

	// Message Rules
	MaxMessageLength int // Максимальная длина сообщения в символах (4096 как в Telegram)

	// Lesson Rules
	DefaultLessonDuration time.Duration // Продолжительность урока по умолчанию (2 часа)
	MinGroupStudents      int           // Минимальное количество студентов для группового урока (4)
	DefaultGroupStudents  int           // Количество студентов по умолчанию для группового урока (4)
	IndividualMaxStudents int           // Максимальное количество студентов для индивидуального урока (всегда 1)
}

// DefaultBusinessRules возвращает бизнес-правила по умолчанию
func DefaultBusinessRules() BusinessRules {
	return BusinessRules{
		// File Upload: 10MB максимум
		MaxFileSize: 10 * 1024 * 1024,

		// Messages: 4096 символов максимум (Telegram limit)
		MaxMessageLength: 4096,

		// Lessons: 2 часа по умолчанию
		DefaultLessonDuration: 2 * time.Hour,

		// Group Lessons: минимум 4 студента
		MinGroupStudents:      4,
		DefaultGroupStudents:  4,
		IndividualMaxStudents: 1,
	}
}

// LoadBusinessRules загружает бизнес-правила из окружения (с дефолтами)
// В будущем можно добавить переменные окружения для кастомизации
func LoadBusinessRules() BusinessRules {
	// Пока просто возвращаем дефолты
	// См. BACKLOG.md: ENV переменные для бизнес-правил (MAX_FILE_SIZE_MB, MAX_MESSAGE_LENGTH, etc.)
	return DefaultBusinessRules()
}

// Recurring lessons rules
const (
	// DefaultRecurringMonths - длительность повторяющихся занятий (24 месяца = 2 года)
	// Это обеспечивает автоматическое продолжение серии надолго вперёд
	DefaultRecurringMonths = 24
)
