package models

import "github.com/google/uuid"

// CreditReconciliation детальная информация по кредитам студентов
// Используется для проверки корректности кредитного баланса после операций замены недели
type CreditReconciliation struct {
	StudentID       uuid.UUID `json:"student_id"`
	StudentName     string    `json:"student_name"`
	PreviousBalance int       `json:"previous_balance"`
	RefundedAmount  int       `json:"refunded_amount"`
	DeductedAmount  int       `json:"deducted_amount"`
	FinalBalance    int       `json:"final_balance"`
}
