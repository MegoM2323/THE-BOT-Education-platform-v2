package hash

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost - стандартная стоимость вычислений для bcrypt
	DefaultCost = bcrypt.DefaultCost
)

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword проверяет, соответствует ли предоставленный пароль хешу
// bcrypt.CompareHashAndPassword использует constant-time сравнение по умолчанию
// что предотвращает timing атаки
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// IsHashValid проверяет, является ли хеш валидным без сравнения с паролем
func IsHashValid(hash string) bool {
	// Пытаемся получить стоимость для валидации формата хеша
	_, err := bcrypt.Cost([]byte(hash))
	return err == nil
}
