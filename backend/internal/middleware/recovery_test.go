package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRecoveryMiddlewareCatchesPanic tests that panic recovery middleware catches panics
func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic error")
	})

	// Wrap it with recovery middleware
	wrappedHandler := RecoveryMiddleware(panicHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call the handler (should not crash)
	wrappedHandler.ServeHTTP(w, req)

	// Verify response status is 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify response is JSON with error message
	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Code != http.StatusInternalServerError {
		t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, errResp.Code)
	}

	if errResp.Error != "Internal server error" {
		t.Errorf("expected error message 'Internal server error', got '%s'", errResp.Error)
	}
}

// TestRecoveryMiddlewareAllowsNormalExecution tests that normal handlers work fine
func TestRecoveryMiddlewareAllowsNormalExecution(t *testing.T) {
	// Create a normal handler
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap it with recovery middleware
	wrappedHandler := RecoveryMiddleware(normalHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call the handler
	wrappedHandler.ServeHTTP(w, req)

	// Verify response status is 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify response body
	if w.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got '%s'", w.Body.String())
	}
}

// TestRecoveryMiddlewareWithNilPanic tests recovery from nil panic
func TestRecoveryMiddlewareWithNilPanic(t *testing.T) {
	// Create a handler that panics with nil
	nilPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p *int
		panic(p) // panic with nil pointer
	})

	// Wrap it with recovery middleware
	wrappedHandler := RecoveryMiddleware(nilPanicHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call the handler (should not crash)
	wrappedHandler.ServeHTTP(w, req)

	// Verify response status is 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify response is valid JSON
	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Code != http.StatusInternalServerError {
		t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, errResp.Code)
	}
}

// TestRecoveryMiddlewarePreservesContentType tests that Content-Type is set
func TestRecoveryMiddlewarePreservesContentType(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap it with recovery middleware
	wrappedHandler := RecoveryMiddleware(panicHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Call the handler
	wrappedHandler.ServeHTTP(w, req)

	// Verify Content-Type is application/json
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestRecoveryMiddlewareWithStringPanic tests recovery from string panic
func TestRecoveryMiddlewareWithStringPanic(t *testing.T) {
	// Create a handler that panics with a string
	stringPanicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("string panic message")
	})

	// Wrap it with recovery middleware
	wrappedHandler := RecoveryMiddleware(stringPanicHandler)

	// Create a test request
	req := httptest.NewRequest("POST", "/api/test", nil)
	w := httptest.NewRecorder()

	// Call the handler (should not crash)
	wrappedHandler.ServeHTTP(w, req)

	// Verify response status is 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify response is valid JSON
	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Code != http.StatusInternalServerError {
		t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, errResp.Code)
	}
}

// TestRecoveryMiddlewareHandlesPanicInDifferentMethods tests all HTTP methods
func TestRecoveryMiddlewareHandlesPanicInDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Create a handler that panics
			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("test panic for " + method)
			})

			// Wrap it with recovery middleware
			wrappedHandler := RecoveryMiddleware(panicHandler)

			// Create a test request
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			// Call the handler
			wrappedHandler.ServeHTTP(w, req)

			// Verify response status is 500
			if w.Code != http.StatusInternalServerError {
				t.Errorf("expected status %d, got %d for method %s", http.StatusInternalServerError, w.Code, method)
			}

			// Verify response is valid JSON
			var errResp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response for method %s: %v", method, err)
			}

			if errResp.Code != http.StatusInternalServerError {
				t.Errorf("expected error code %d, got %d for method %s", http.StatusInternalServerError, errResp.Code, method)
			}
		})
	}
}
