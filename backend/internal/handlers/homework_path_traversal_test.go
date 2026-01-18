package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== Validation Tests ====================

func TestValidateFilePath_ValidPath(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		allowedDir string
		wantValid  bool
	}{
		{
			name:       "normal file in uploads directory",
			filePath:   "uploads/homework/test-file.pdf",
			allowedDir: "uploads/homework",
			wantValid:  true,
		},
		{
			name:       "uuid filename",
			filePath:   "uploads/homework/550e8400-e29b-41d4-a716-446655440000.pdf",
			allowedDir: "uploads/homework",
			wantValid:  true,
		},
		{
			name:       "nested relative path",
			filePath:   "uploads/homework/subdir/file.txt",
			allowedDir: "uploads/homework",
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, errMsg := validateFilePath(tt.filePath, tt.allowedDir)
			assert.Equal(t, tt.wantValid, isValid, "validation result mismatch")
			if !isValid {
				t.Logf("error: %s", errMsg)
			}
		})
	}
}

func TestValidateFilePath_PathTraversalDetection(t *testing.T) {
	tests := []struct {
		name       string
		filePath   string
		allowedDir string
		wantValid  bool
		desc       string
	}{
		{
			name:       "double dot path traversal",
			filePath:   "uploads/homework/../../etc/passwd",
			allowedDir: "uploads/homework",
			wantValid:  false,
			desc:       "should reject .. for directory traversal",
		},
		{
			name:       "double dot in middle",
			filePath:   "uploads/../homework/file.txt",
			allowedDir: "uploads/homework",
			wantValid:  false,
			desc:       "should reject .. even in middle",
		},
		{
			name:       "absolute path",
			filePath:   "/etc/passwd",
			allowedDir: "uploads/homework",
			wantValid:  false,
			desc:       "should reject absolute paths",
		},
		{
			name:       "null byte injection",
			filePath:   "uploads/homework/file.txt\x00.pdf",
			allowedDir: "uploads/homework",
			wantValid:  false,
			desc:       "should reject null bytes",
		},
		{
			name:       "empty path",
			filePath:   "",
			allowedDir: "uploads/homework",
			wantValid:  false,
			desc:       "should reject empty path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, errMsg := validateFilePath(tt.filePath, tt.allowedDir)
			assert.Equal(t, tt.wantValid, isValid, "validation result mismatch: %s", tt.desc)
			if !isValid {
				t.Logf("error: %s", errMsg)
			}
		})
	}
}

func TestSanitizeFileName_ValidNames(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantOK   bool
	}{
		{
			name:     "simple filename",
			fileName: "document.pdf",
			wantOK:   true,
		},
		{
			name:     "filename with spaces",
			fileName: "my document.pdf",
			wantOK:   true,
		},
		{
			name:     "filename with dashes",
			fileName: "my-document.pdf",
			wantOK:   true,
		},
		{
			name:     "filename with underscores",
			fileName: "my_document.pdf",
			wantOK:   true,
		},
		{
			name:     "cyrillic filename",
			fileName: "документ.pdf",
			wantOK:   true,
		},
		{
			name:     "filename with multiple dots",
			fileName: "my.test.document.pdf",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _ := sanitizeFileName(tt.fileName)
			assert.Equal(t, tt.wantOK, isValid)
		})
	}
}

func TestSanitizeFileName_InvalidNames(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantOK   bool
	}{
		{
			name:     "filename with slash",
			fileName: "path/to/file.pdf",
			wantOK:   false,
		},
		{
			name:     "filename with backslash",
			fileName: "path\\to\\file.pdf",
			wantOK:   false,
		},
		{
			name:     "filename with semicolon",
			fileName: "file;.pdf",
			wantOK:   false,
		},
		{
			name:     "filename with pipe",
			fileName: "file|name.pdf",
			wantOK:   false,
		},
		{
			name:     "filename with shell metachar",
			fileName: "file$(whoami).pdf",
			wantOK:   false,
		},
		{
			name:     "filename with backtick",
			fileName: "file`command`.pdf",
			wantOK:   false,
		},
		{
			name:     "empty filename",
			fileName: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _ := sanitizeFileName(tt.fileName)
			assert.Equal(t, tt.wantOK, isValid)
		})
	}
}
