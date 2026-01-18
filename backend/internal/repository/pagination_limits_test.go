package repository

import (
	"testing"
)

// TestNormalizeLimitEdgeCases дополнительные тесты для граничных случаев
func TestNormalizeLimitEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "Exactly max limit",
			input:    MaxQueryLimit,
			expected: MaxQueryLimit,
		},
		{
			name:     "One above max limit",
			input:    MaxQueryLimit + 1,
			expected: MaxQueryLimit,
		},
		{
			name:     "Large DoS attack value",
			input:    999999999,
			expected: MaxQueryLimit,
		},
		{
			name:     "Zero should return default",
			input:    0,
			expected: DefaultQueryLimit,
		},
		{
			name:     "Negative DoS value",
			input:    -999999,
			expected: DefaultQueryLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeLimit(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeLimit(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPaginationLimitsSafety проверяет что все значения limit безопасны
func TestPaginationLimitsSafety(t *testing.T) {
	// Симулируем DoS атаку с огромным limit
	attackLimits := []int{
		999999,
		1000000,
		-1,
		-999999,
		0,
	}

	for _, attackLimit := range attackLimits {
		normalized := NormalizeLimit(attackLimit)

		// Проверяем что нормализованное значение безопасно
		if normalized > MaxQueryLimit {
			t.Errorf("Normalized limit %d exceeds MaxQueryLimit %d for input %d",
				normalized, MaxQueryLimit, attackLimit)
		}

		if normalized <= 0 {
			t.Errorf("Normalized limit %d is not positive for input %d",
				normalized, attackLimit)
		}
	}
}
