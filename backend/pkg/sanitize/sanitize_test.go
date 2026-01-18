package sanitize

import (
	"strings"
	"testing"
)

func TestInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Extra spaces",
			input:    "  Hello   World  ",
			expected: "Hello World",
		},
		{
			name:     "Tabs and newlines stripped",
			input:    "Hello\tWorld\nTest",
			expected: "HelloWorldTest",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   \t\n  ",
			expected: "",
		},
		{
			name:     "With special chars",
			input:    "Test@#$%",
			expected: "Test@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Input(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMultilineInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Preserves newlines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Trims leading/trailing whitespace",
			input:    "  \nLine 1\nLine 2\n  ",
			expected: "Line 1\nLine 2",
		},
		{
			name:     "Replaces tabs",
			input:    "Line1\ttext\nLine2",
			expected: "Line1 text\nLine2",
		},
		{
			name:     "Single line",
			input:    "Just one line",
			expected: "Just one line",
		},
		{
			name:     "Empty lines preserved",
			input:    "Line1\n\nLine3",
			expected: "Line1\n\nLine3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MultilineInput(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal email",
			input:    "user@example.com",
			expected: "user@example.com",
		},
		{
			name:     "Email with spaces",
			input:    "  user@example.com  ",
			expected: "user@example.com",
		},
		{
			name:     "Lowercase conversion",
			input:    "User@Example.COM",
			expected: "user@example.com",
		},
		{
			name:     "Empty email",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Email(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal name",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "Name with extra spaces",
			input:    "  John   Doe  ",
			expected: "John Doe",
		},
		{
			name:     "Hyphenated name",
			input:    "Mary-Anne Smith",
			expected: "Mary-Anne Smith",
		},
		{
			name:     "Empty name",
			input:    "",
			expected: "",
		},
		{
			name:     "Only spaces",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Name(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestReason(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal reason",
			input:    "User was offline",
			expected: "User was offline",
		},
		{
			name:     "Multiline reason",
			input:    "User was offline\nFor two weeks",
			expected: "User was offline\nFor two weeks",
		},
		{
			name:     "With tabs",
			input:    "Reason\twith\ttabs",
			expected: "Reason with tabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reason(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal description",
			input:    "This is a description",
			expected: "This is a description",
		},
		{
			name:     "Multiline description",
			input:    "First line\nSecond line",
			expected: "First line\nSecond line",
		},
		{
			name:     "With tabs",
			input:    "Description\twith\ttabs",
			expected: "Description with tabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Description(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTelegramUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal username",
			input:    "username123",
			expected: "username123",
		},
		{
			name:     "Username with leading @",
			input:    "@username",
			expected: "@username",
		},
		{
			name:     "Username with uppercase and @",
			input:    "@UserName",
			expected: "@username",
		},
		{
			name:     "Spaces trimmed",
			input:    "  @username  ",
			expected: "@username",
		},
		{
			name:     "Empty username",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TelegramUsername(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSanitizationPreservesContent(t *testing.T) {
	// Verify that useful content is preserved
	inputs := []struct {
		name  string
		input string
		fn    func(string) string
	}{
		{
			name:  "Email with plus addressing",
			input: "user+tag@example.com",
			fn:    Email,
		},
		{
			name:  "Name with apostrophe",
			input: "O'Brien",
			fn:    Name,
		},
		{
			name:  "Description with punctuation",
			input: "Hello, world! This is important.",
			fn:    Description,
		},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result == "" {
				t.Errorf("Expected sanitization to preserve content, got empty string")
			}
		})
	}
}

func TestLengthPreservation(t *testing.T) {
	// Verify that typical lengths are within reasonable bounds
	testCases := []struct {
		name  string
		input string
		fn    func(string) string
	}{
		{
			name:  "Email",
			input: "verylongemailaddress@verylongdomainname.com",
			fn:    Email,
		},
		{
			name:  "Name",
			input: "John Alexander Christopher Douglas Michael O'Brien Smith III",
			fn:    Name,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if len(result) == 0 {
				t.Error("Result should not be empty for valid input")
			}
			// Result should not be wildly longer
			if len(result) > len(tt.input)+10 {
				t.Errorf("Result unexpectedly long: %d > %d + 10", len(result), len(tt.input))
			}
		})
	}
}

func TestEmptyStringHandling(t *testing.T) {
	// All functions should handle empty string gracefully
	fns := map[string]func(string) string{
		"Input":            Input,
		"MultilineInput":   MultilineInput,
		"Email":            Email,
		"Name":             Name,
		"Reason":           Reason,
		"Description":      Description,
		"TelegramUsername": TelegramUsername,
	}

	for name, fn := range fns {
		t.Run(name, func(t *testing.T) {
			result := fn("")
			// Should not panic and should return empty or whitespace
			if len(strings.TrimSpace(result)) > 0 {
				t.Errorf("%s should return empty/whitespace for empty input, got %q", name, result)
			}
		})
	}
}
