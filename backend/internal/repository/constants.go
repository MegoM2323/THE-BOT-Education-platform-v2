package repository

const (
	// DefaultQueryLimit - лимит по умолчанию для запросов с пагинацией
	DefaultQueryLimit = 50

	// MaxQueryLimit - максимальный лимит для запросов с пагинацией (защита от DoS)
	MaxQueryLimit = 1000
)

// NormalizeLimit нормализует значение limit для безопасного использования
// Применяет default если limit <= 0, ограничивает максимум если превышен
func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultQueryLimit
	}
	if limit > MaxQueryLimit {
		return MaxQueryLimit
	}
	return limit
}
