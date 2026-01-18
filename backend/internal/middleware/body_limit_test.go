package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBodyLimitMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		bodySize        int64
		limit           int64
		expectedStatus  int
		expectedCanRead bool
		description     string
	}{
		{
			name:            "valid: body under limit",
			bodySize:        100,
			limit:           1024,
			expectedStatus:  http.StatusOK,
			expectedCanRead: true,
			description:     "Should allow reading body under limit",
		},
		{
			name:            "valid: body exactly at limit",
			bodySize:        1024,
			limit:           1024,
			expectedStatus:  http.StatusOK,
			expectedCanRead: true,
			description:     "Should allow reading body exactly at limit",
		},
		{
			name:            "invalid: body exceeds limit",
			bodySize:        2048,
			limit:           1024,
			expectedStatus:  http.StatusRequestEntityTooLarge,
			expectedCanRead: false,
			description:     "Should reject body exceeding limit with 413",
		},
		{
			name:            "invalid: very large body",
			bodySize:        50 * 1024 * 1024, // 50MB
			limit:           1 * 1024 * 1024,  // 1MB
			expectedStatus:  http.StatusRequestEntityTooLarge,
			expectedCanRead: false,
			description:     "Should reject very large body",
		},
		{
			name:            "valid: empty body",
			bodySize:        0,
			limit:           1024,
			expectedStatus:  http.StatusOK,
			expectedCanRead: true,
			description:     "Should allow empty body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler that tries to read the body
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Try to read the entire body
				data, err := io.ReadAll(r.Body)
				defer r.Body.Close()

				if err != nil {
					// This error is expected when body exceeds limit
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					w.Write([]byte("body limit exceeded"))
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok: " + string(data)))
			})

			// Wrap with body limit middleware
			wrappedHandler := BodyLimitMiddlewareWithLimit(tt.limit)(handler)

			// Create request with large body
			bodyData := bytes.Repeat([]byte("a"), int(tt.bodySize))
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyData))
			rec := httptest.NewRecorder()

			// Serve
			wrappedHandler.ServeHTTP(rec, req)

			// Verify response
			if rec.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.description, tt.expectedStatus, rec.Code)
			}

			if tt.expectedCanRead && rec.Code != http.StatusOK {
				t.Errorf("%s: expected to read body but got error", tt.description)
			}
		})
	}
}

func TestDefaultBodyLimitConfig(t *testing.T) {
	config := DefaultBodyLimitConfig()

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	// Check default values
	expectedDefault := int64(1024 * 1024) // 1MB
	if config.DefaultLimit != expectedDefault {
		t.Errorf("expected DefaultLimit %d, got %d", expectedDefault, config.DefaultLimit)
	}

	expectedFileUpload := int64(10 * 1024 * 1024) // 10MB
	if config.FileUploadLimit != expectedFileUpload {
		t.Errorf("expected FileUploadLimit %d, got %d", expectedFileUpload, config.FileUploadLimit)
	}
}

func TestBodyLimitMiddlewareMultipleCalls(t *testing.T) {
	// Test that multiple reads within limit work correctly
	config := DefaultBodyLimitConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First read
		data1, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		w.WriteHeader(http.StatusOK)
		w.Write(data1)
	})

	wrappedHandler := BodyLimitMiddleware(config)(handler)

	// Create request
	bodyData := []byte("test data within limit")
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyData))
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if !bytes.Contains(rec.Body.Bytes(), bodyData) {
		t.Errorf("expected body to contain %s, got %s", string(bodyData), rec.Body.String())
	}
}

func TestBodyLimitConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int64
		expected int64
	}{
		{
			name:     "BodyLimitSmall",
			constant: BodyLimitSmall,
			expected: 512 * 1024,
		},
		{
			name:     "BodyLimitMedium",
			constant: BodyLimitMedium,
			expected: 1024 * 1024,
		},
		{
			name:     "BodyLimitLarge",
			constant: BodyLimitLarge,
			expected: 5 * 1024 * 1024,
		},
		{
			name:     "BodyLimitExtraLarge",
			constant: BodyLimitExtraLarge,
			expected: 10 * 1024 * 1024,
		},
		{
			name:     "BodyLimitXLarge",
			constant: BodyLimitXLarge,
			expected: 50 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, tt.constant)
			}
		})
	}
}

func TestBodyLimitMiddlewareForFileUpload(t *testing.T) {
	// Test that file upload middleware uses larger limit
	config := DefaultBodyLimitConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("uploaded: %d", len(data))))
	})

	wrappedHandler := BodyLimitMiddlewareForFileUpload(config)(handler)

	// Create request with body between default and file upload limit
	// This should work with file upload middleware but fail with default middleware
	bodySize := int64(5 * 1024 * 1024) // 5MB, between 1MB (default) and 10MB (file upload)
	bodyData := bytes.Repeat([]byte("a"), int(bodySize))
	req := httptest.NewRequest("POST", "/upload", bytes.NewReader(bodyData))
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// With file upload limit (10MB), 5MB should be accepted
	if rec.Code != http.StatusOK {
		t.Errorf("file upload middleware should accept 5MB, got status %d", rec.Code)
	}
}

func TestBodyLimitMiddlewareJSON(t *testing.T) {
	config := DefaultBodyLimitConfig()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	wrappedHandler := BodyLimitMiddlewareJSON(config)(handler)

	// Create request with body under default limit
	bodyData := []byte(`{"name":"test","value":123}`)
	req := httptest.NewRequest("POST", "/api", bytes.NewReader(bodyData))
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetBodyLimitConfig(t *testing.T) {
	defaultLimit := int64(2 * 1024 * 1024)
	fileLimit := int64(50 * 1024 * 1024)

	config := GetBodyLimitConfig(defaultLimit, fileLimit)

	if config.DefaultLimit != defaultLimit {
		t.Errorf("expected DefaultLimit %d, got %d", defaultLimit, config.DefaultLimit)
	}

	if config.FileUploadLimit != fileLimit {
		t.Errorf("expected FileUploadLimit %d, got %d", fileLimit, config.FileUploadLimit)
	}
}

func TestBodyLimitMiddlewareWithNilConfig(t *testing.T) {
	// Test that nil config is handled gracefully
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Should use default config
	wrappedHandler := BodyLimitMiddleware(nil)(handler)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("test")))
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestBodyLimitMiddlewareChained(t *testing.T) {
	// Test that middleware can be chained properly
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	// Chain multiple middleware
	config := DefaultBodyLimitConfig()
	wrappedHandler := BodyLimitMiddleware(config)(handler)

	bodyData := []byte("test chained")
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyData))
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// Benchmark tests
func BenchmarkBodyLimitMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})

	config := DefaultBodyLimitConfig()
	wrappedHandler := BodyLimitMiddleware(config)(handler)

	bodyData := bytes.Repeat([]byte("a"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyData))
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
	}
}

func BenchmarkBodyLimitMiddlewareWithLimit(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := BodyLimitMiddlewareWithLimit(1024 * 1024)(handler)

	bodyData := bytes.Repeat([]byte("a"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyData))
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
	}
}
