package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestMaskUserID(t *testing.T) {
	tests := []struct {
		name     string
		id       uuid.UUID
		expected string
	}{
		{
			name:     "Standard UUID",
			id:       uuid.MustParse("d3c8c7a6-1234-5678-abcd-ef1234567890"),
			expected: "d3c8c7a6***",
		},
		{
			name:     "All zeros UUID",
			id:       uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			expected: "00000000***",
		},
		{
			name:     "All ones UUID",
			id:       uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			expected: "ffffffff***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskUserID(tt.id)
			if result != tt.expected {
				t.Errorf("MaskUserID() = %q, want %q", result, tt.expected)
			}
			if len(result) != 11 { // 8 chars + 3 asterisks
				t.Errorf("MaskUserID() length = %d, want 11", len(result))
			}
		})
	}
}

func TestMaskAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   int
		expected string
	}{
		{
			name:     "Zero amount",
			amount:   0,
			expected: "0-100",
		},
		{
			name:     "Small amount (50)",
			amount:   50,
			expected: "0-100",
		},
		{
			name:     "Boundary (100)",
			amount:   100,
			expected: "100-1000",
		},
		{
			name:     "Medium amount (500)",
			amount:   500,
			expected: "100-1000",
		},
		{
			name:     "Boundary (1000)",
			amount:   1000,
			expected: "1000+",
		},
		{
			name:     "Large amount (5000)",
			amount:   5000,
			expected: "1000+",
		},
		{
			name:     "Negative amount (-50)",
			amount:   -50,
			expected: "0-100",
		},
		{
			name:     "Negative large amount (-5000)",
			amount:   -5000,
			expected: "1000+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskAmount(tt.amount)
			if result != tt.expected {
				t.Errorf("MaskAmount(%d) = %q, want %q", tt.amount, result, tt.expected)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "Standard email",
			email:    "user@example.com",
			expected: "u***@example.com",
		},
		{
			name:     "Single letter before @",
			email:    "a@example.com",
			expected: "a***@example.com",
		},
		{
			name:     "No @ symbol",
			email:    "userexample.com",
			expected: "***",
		},
		{
			name:     "Empty email",
			email:    "",
			expected: "***",
		},
		{
			name:     "Only @ symbol",
			email:    "@",
			expected: "***",
		},
		{
			name:     "Long email",
			email:    "verylongemailaddress@subdomain.example.co.uk",
			expected: "v***@subdomain.example.co.uk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.email)
			if result != tt.expected {
				t.Errorf("MaskEmail(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}
