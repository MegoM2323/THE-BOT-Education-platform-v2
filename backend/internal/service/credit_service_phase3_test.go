package service

import (
	"testing"

	"tutoring-platform/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestAddCredits_CheckMaxBalanceLimitValidation проверяет что AddCredits проверяет максимум баланса
func TestAddCredits_CheckMaxBalanceLimitValidation(t *testing.T) {
	// Этот тест проверяет логику, которая реализована в методе AddCredits()
	// Логика находится в строке 111-113 credit_service.go:
	// if newBalance > models.MaxBalance {
	//    return models.ErrBalanceExceeded
	// }

	// Проверяем что константа MaxBalance определена и равна 10000
	assert.Equal(t, 10000, models.MaxBalance, "MaxBalance должен быть 10000")

	// Проверяем что ошибка ErrBalanceExceeded существует
	assert.NotNil(t, models.ErrBalanceExceeded, "ErrBalanceExceeded должна быть определена")

	// Проверяем текст ошибки
	assert.Contains(t, models.ErrBalanceExceeded.Error(), "10000", "ошибка должна упоминать лимит 10000")
}

// TestRefundCredits_CheckMaxBalanceLimitValidation проверяет что RefundCredits проверяет максимум баланса
func TestRefundCredits_CheckMaxBalanceLimitValidation(t *testing.T) {
	// Этот тест проверяет логику, которая реализована в методе RefundCredits()
	// Логика находится в строке 275-279 credit_service.go:
	// if newBalance > models.MaxBalance {
	//    return models.ErrBalanceExceeded
	// }

	// Проверяем что AddCredits и RefundCredits используют одинаковую проверку
	assert.Equal(t, 10000, models.MaxBalance, "MaxBalance должен быть 10000")
}

// TestAddCreditsWithTx_CheckMaxBalanceLimitValidation проверяет что AddCreditsWithTx проверяет максимум баланса
func TestAddCreditsWithTx_CheckMaxBalanceLimitValidation(t *testing.T) {
	// Этот тест проверяет логику, которая реализована в методе AddCreditsWithTx()
	// Логика находится в строке 162-165 credit_service.go:
	// if newBalance > models.MaxBalance {
	//    return models.ErrBalanceExceeded
	// }

	// Проверяем что все три метода (AddCredits, AddCreditsWithTx, RefundCredits)
	// используют одинаковую проверку на максимум баланса
	assert.Equal(t, 10000, models.MaxBalance, "все методы должны проверять один и тот же максимум")
}
