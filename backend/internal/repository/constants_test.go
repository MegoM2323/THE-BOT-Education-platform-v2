package repository

import "testing"

func TestNormalizeLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "Zero limit returns default",
			input:    0,
			expected: DefaultQueryLimit,
		},
		{
			name:     "Negative limit returns default",
			input:    -10,
			expected: DefaultQueryLimit,
		},
		{
			name:     "Valid limit below max",
			input:    100,
			expected: 100,
		},
		{
			name:     "Limit equals max",
			input:    MaxQueryLimit,
			expected: MaxQueryLimit,
		},
		{
			name:     "Limit exceeds max returns max",
			input:    999999,
			expected: MaxQueryLimit,
		},
		{
			name:     "Small valid limit",
			input:    1,
			expected: 1,
		},
		{
			name:     "Large but valid limit",
			input:    500,
			expected: 500,
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

func TestConstants(t *testing.T) {
	// Проверяем что константы имеют правильные значения
	if DefaultQueryLimit != 50 {
		t.Errorf("DefaultQueryLimit = %d, expected 50", DefaultQueryLimit)
	}

	if MaxQueryLimit != 1000 {
		t.Errorf("MaxQueryLimit = %d, expected 1000", MaxQueryLimit)
	}

	// Проверяем что max больше default
	if MaxQueryLimit <= DefaultQueryLimit {
		t.Error("MaxQueryLimit must be greater than DefaultQueryLimit")
	}
}
