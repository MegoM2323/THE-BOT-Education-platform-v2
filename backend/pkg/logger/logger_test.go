package logger

import (
	"testing"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{
			name:  "Debug level",
			level: "debug",
		},
		{
			name:  "Info level",
			level: "info",
		},
		{
			name:  "Warn level",
			level: "warn",
		},
		{
			name:  "Error level",
			level: "error",
		},
		{
			name:  "Invalid level defaults",
			level: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup should not panic
			Setup(tt.level)
		})
	}
}

func TestSetupWithConsole(t *testing.T) {
	// Test that Setup initializes logger without panicking
	Setup("info")

	// If we reach here, Setup worked correctly
	t.Log("Logger setup completed successfully")
}
