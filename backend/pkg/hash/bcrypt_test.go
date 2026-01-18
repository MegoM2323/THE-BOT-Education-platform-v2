package hash

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "Simple password",
			password: "simplepass123",
		},
		{
			name:     "Complex password",
			password: "P@ssw0rd!#$%^&*()_+{}[]|:;<>?,./-=",
		},
		{
			name:     "Very long password",
			password: strings.Repeat("a", 72), // bcrypt ограничен 72 байтами
		},
		{
			name:     "Unicode password",
			password: "пароль密码🔒",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			if hash == "" {
				t.Error("Hash is empty")
			}

			if hash == tt.password {
				t.Error("Hash should not equal password")
			}

			// Check it's a valid bcrypt hash (starts with $2a$, $2b$, or $2y$)
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
				t.Error("Hash doesn't look like bcrypt hash")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"
	hash, _ := HashPassword(password)

	tests := []struct {
		name        string
		password    string
		hash        string
		expectMatch bool
	}{
		{
			name:        "Correct password",
			password:    password,
			hash:        hash,
			expectMatch: true,
		},
		{
			name:        "Wrong password",
			password:    "wrongPassword123",
			hash:        hash,
			expectMatch: false,
		},
		{
			name:        "Empty password",
			password:    "",
			hash:        hash,
			expectMatch: false,
		},
		{
			name:        "Empty hash",
			password:    password,
			hash:        "",
			expectMatch: false,
		},
		{
			name:        "Case sensitive",
			password:    "testpassword123",
			hash:        hash,
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.password, tt.hash)
			matches := err == nil

			if matches != tt.expectMatch {
				t.Errorf("CheckPassword: expected match=%v, got match=%v (error=%v)", tt.expectMatch, matches, err)
			}
		})
	}
}

func TestIsHashValid(t *testing.T) {
	validHash, _ := HashPassword("test")

	tests := []struct {
		name    string
		hash    string
		isValid bool
	}{
		{
			name:    "Valid bcrypt hash",
			hash:    validHash,
			isValid: true,
		},
		{
			name:    "Empty hash",
			hash:    "",
			isValid: false,
		},
		{
			name:    "Invalid hash format",
			hash:    "not-a-bcrypt-hash",
			isValid: false,
		},
		{
			name:    "Too short",
			hash:    "$2a$10$short",
			isValid: false,
		},
		{
			name:    "Wrong prefix",
			hash:    "$1a$10$somehashhere",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHashValid(tt.hash)
			if result != tt.isValid {
				t.Errorf("IsHashValid: expected %v, got %v for hash %q", tt.isValid, result, tt.hash)
			}
		})
	}
}

func TestPasswordRoundTrip(t *testing.T) {
	// Test that hash -> check works correctly
	passwords := []string{
		"simple",
		"P@ssw0rd!",
		"verylongpasswordwithmanycharacters1234567890",
	}

	for _, pwd := range passwords {
		hash, err := HashPassword(pwd)
		if err != nil {
			t.Fatalf("Failed to hash %q: %v", pwd, err)
		}

		if err := CheckPassword(pwd, hash); err != nil {
			t.Errorf("Failed to verify password %q: %v", pwd, err)
		}
	}
}

func TestHashUniqueness(t *testing.T) {
	// Test that same password creates different hashes (due to salt)
	password := "testPassword"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes (different salts)")
	}

	// But both should match the password
	if err := CheckPassword(password, hash1); err != nil {
		t.Errorf("Hash1 should match password: %v", err)
	}
	if err := CheckPassword(password, hash2); err != nil {
		t.Errorf("Hash2 should match password: %v", err)
	}
}
