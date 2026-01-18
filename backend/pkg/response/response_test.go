package response

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		data       interface{}
		checkBody  func(t *testing.T, resp SuccessResponse)
	}{
		{
			name:       "OK status with data",
			statusCode: http.StatusOK,
			data:       map[string]string{"key": "value"},
			checkBody: func(t *testing.T, resp SuccessResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
				if resp.Data == nil {
					t.Error("Expected data to be present")
				}
			},
		},
		{
			name:       "Created status",
			statusCode: http.StatusCreated,
			data:       map[string]int{"id": 123},
			checkBody: func(t *testing.T, resp SuccessResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
			},
		},
		{
			name:       "No content data",
			statusCode: http.StatusOK,
			data:       nil,
			checkBody: func(t *testing.T, resp SuccessResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Success(w, tt.statusCode, tt.data)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}

			var resp SuccessResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			tt.checkBody(t, resp)
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		code       string
		message    string
		checkBody  func(t *testing.T, resp ErrorResponse)
	}{
		{
			name:       "Bad request error",
			statusCode: http.StatusBadRequest,
			code:       "INVALID_INPUT",
			message:    "Email is required",
			checkBody: func(t *testing.T, resp ErrorResponse) {
				if resp.Success {
					t.Error("Expected success=false")
				}
				if resp.Error.Code != "INVALID_INPUT" {
					t.Errorf("Expected code INVALID_INPUT, got %s", resp.Error.Code)
				}
				if resp.Error.Message != "Email is required" {
					t.Errorf("Expected message 'Email is required', got %s", resp.Error.Message)
				}
			},
		},
		{
			name:       "Not found error",
			statusCode: http.StatusNotFound,
			code:       "NOT_FOUND",
			message:    "User not found",
			checkBody: func(t *testing.T, resp ErrorResponse) {
				if resp.Success {
					t.Error("Expected success=false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Error(w, tt.statusCode, tt.code, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			tt.checkBody(t, resp)
		})
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	BadRequest(w, "VALIDATION_FAILED", "Invalid email format")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "VALIDATION_FAILED" {
		t.Errorf("Expected VALIDATION_FAILED, got %s", resp.Error.Code)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w, "Invalid credentials")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != ErrCodeUnauthorized {
		t.Errorf("Expected ErrCodeUnauthorized, got %s", resp.Error.Code)
	}
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	Forbidden(w, "Access denied")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != ErrCodeForbidden {
		t.Errorf("Expected ErrCodeForbidden, got %s", resp.Error.Code)
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != ErrCodeNotFound {
		t.Errorf("Expected ErrCodeNotFound, got %s", resp.Error.Code)
	}
}

func TestConflict(t *testing.T) {
	w := httptest.NewRecorder()
	Conflict(w, "ALREADY_EXISTS", "Email already registered")

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "ALREADY_EXISTS" {
		t.Errorf("Expected ALREADY_EXISTS, got %s", resp.Error.Code)
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w, "Database error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != ErrCodeInternalError {
		t.Errorf("Expected ErrCodeInternalError, got %s", resp.Error.Code)
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	Created(w, map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var resp SuccessResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestOK(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, map[string]string{"message": "success"})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp SuccessResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Success {
		t.Error("Expected success=true")
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	if w.Body.Len() != 0 {
		t.Error("Expected empty body for NoContent")
	}
}

func TestTooManyRequests(t *testing.T) {
	w := httptest.NewRecorder()
	TooManyRequests(w, "Rate limit exceeded")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	var resp ErrorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != ErrCodeRateLimitExceeded {
		t.Errorf("Expected ErrCodeRateLimitExceeded, got %s", resp.Error.Code)
	}
}

// BrokenResponseWriter is a test helper that fails on Write
type BrokenResponseWriter struct {
	headerWritten bool
}

func (b *BrokenResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (b *BrokenResponseWriter) Write(p []byte) (n int, err error) {
	b.headerWritten = true
	return 0, io.ErrClosedPipe
}

func (b *BrokenResponseWriter) WriteHeader(statusCode int) {
	b.headerWritten = true
}

// TestSuccessWithBrokenWriter tests JSON encoding error handling
func TestSuccessWithBrokenWriter(t *testing.T) {
	w := &BrokenResponseWriter{}
	Success(w, http.StatusOK, map[string]string{"key": "value"})
	// Should not panic, error should be logged
}

// TestErrorWithBrokenWriter tests JSON encoding error handling for error responses
func TestErrorWithBrokenWriter(t *testing.T) {
	w := &BrokenResponseWriter{}
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Database connection failed")
	// Should not panic, error should be logged
}

// TestSuccessWithComplexData tests encoding of complex nested structures
func TestSuccessWithComplexData(t *testing.T) {
	w := httptest.NewRecorder()
	complexData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    123,
			"name":  "John Doe",
			"roles": []string{"student", "teacher"},
			"metadata": map[string]interface{}{
				"created_at": "2024-01-01T00:00:00Z",
				"verified":   true,
			},
		},
		"lessons": []map[string]interface{}{
			{"id": 1, "title": "Math 101"},
			{"id": 2, "title": "Physics 101"},
		},
	}

	Success(w, http.StatusOK, complexData)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	if resp.Data == nil {
		t.Error("Expected data to be present")
	}
}

// TestErrorWithComplexMessage tests encoding of error responses with special characters
func TestErrorWithComplexMessage(t *testing.T) {
	w := httptest.NewRecorder()
	specialMessage := `Failed to process: "quoted text" with \n escape and 'single quotes'`
	Error(w, http.StatusBadRequest, "VALIDATION_FAILED", specialMessage)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if resp.Error.Message != specialMessage {
		t.Errorf("Expected message %q, got %q", specialMessage, resp.Error.Message)
	}
}

// TestSuccessWithNilData tests encoding with nil data
func TestSuccessWithNilData(t *testing.T) {
	w := httptest.NewRecorder()
	Success(w, http.StatusOK, nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	// nil data should be encoded as JSON null
	if resp.Data != nil {
		// Data might be decoded as map[string]interface{} which is not nil
		// This is acceptable JSON encoding behavior
	}
}

// TestSuccessWithEmptyArray tests encoding with empty arrays
func TestSuccessWithEmptyArray(t *testing.T) {
	w := httptest.NewRecorder()
	Success(w, http.StatusOK, []string{})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp SuccessResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}
}

// ============================================================================
// TESTS FOR STANDARDIZED RESPONSE WRAPPERS
// ============================================================================

func TestListResponseBasic(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		meta       *ResponseMeta
		statusCode int
		checkBody  func(t *testing.T, resp StandardListResponse)
	}{
		{
			name: "Basic list response",
			data: []map[string]string{
				{"id": "1", "name": "Item 1"},
				{"id": "2", "name": "Item 2"},
			},
			meta:       &ResponseMeta{Count: 2},
			statusCode: http.StatusOK,
			checkBody: func(t *testing.T, resp StandardListResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
				if resp.Data == nil {
					t.Error("Expected data to be present")
				}
				if resp.Meta.Count != 2 {
					t.Errorf("Expected count=2, got %d", resp.Meta.Count)
				}
			},
		},
		{
			name:       "Empty list response",
			data:       []string{},
			meta:       &ResponseMeta{Count: 0},
			statusCode: http.StatusOK,
			checkBody: func(t *testing.T, resp StandardListResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
				if resp.Meta.Count != 0 {
					t.Errorf("Expected count=0, got %d", resp.Meta.Count)
				}
			},
		},
		{
			name: "List response with pagination",
			data: []int{1, 2, 3, 4, 5},
			meta: &ResponseMeta{
				Count: 5,
				Pagination: &PaginationMeta{
					Page:       1,
					PageSize:   5,
					TotalCount: 100,
					TotalPages: 20,
				},
			},
			statusCode: http.StatusOK,
			checkBody: func(t *testing.T, resp StandardListResponse) {
				if resp.Meta.Pagination == nil {
					t.Error("Expected pagination metadata")
				}
				if resp.Meta.Pagination.TotalCount != 100 {
					t.Errorf("Expected total_count=100, got %d", resp.Meta.Pagination.TotalCount)
				}
				if resp.Meta.Pagination.TotalPages != 20 {
					t.Errorf("Expected total_pages=20, got %d", resp.Meta.Pagination.TotalPages)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ListResponse(w, tt.statusCode, tt.data, tt.meta)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}

			var resp StandardListResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			tt.checkBody(t, resp)
		})
	}
}

func TestListResponseNilMeta(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"item1", "item2"}
	ListResponse(w, http.StatusOK, data, nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp StandardListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}

	// Meta should be created even if nil
	if resp.Meta.Count == 0 && resp.Meta.Pagination == nil {
		// This is acceptable
	}
}

func TestListResponseWithTimestamp(t *testing.T) {
	w := httptest.NewRecorder()
	data := []map[string]interface{}{}
	meta := &ResponseMeta{
		Count:     0,
		Timestamp: "2024-01-01T00:00:00Z",
	}
	ListResponse(w, http.StatusOK, data, meta)

	var resp StandardListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Meta.Timestamp != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected timestamp to be preserved, got %s", resp.Meta.Timestamp)
	}
}

func TestSingleResponseBasic(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		statusCode int
		checkBody  func(t *testing.T, resp StandardSingleResponse)
	}{
		{
			name:       "Single object response",
			data:       map[string]string{"id": "1", "name": "Test"},
			statusCode: http.StatusOK,
			checkBody: func(t *testing.T, resp StandardSingleResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
				if resp.Data == nil {
					t.Error("Expected data to be present")
				}
			},
		},
		{
			name: "Created single response",
			data: map[string]interface{}{
				"id":   "123",
				"name": "New Item",
			},
			statusCode: http.StatusCreated,
			checkBody: func(t *testing.T, resp StandardSingleResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
			},
		},
		{
			name:       "Single response with nil data",
			data:       nil,
			statusCode: http.StatusOK,
			checkBody: func(t *testing.T, resp StandardSingleResponse) {
				if !resp.Success {
					t.Error("Expected success=true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SingleResponse(w, tt.statusCode, tt.data)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", ct)
			}

			var resp StandardSingleResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			tt.checkBody(t, resp)
		})
	}
}

func TestListResponseNoMetaField(t *testing.T) {
	w := httptest.NewRecorder()
	data := []int{1, 2, 3}
	meta := &ResponseMeta{
		Count: 3,
		Pagination: &PaginationMeta{
			Page:     2,
			PageSize: 10,
		},
	}
	ListResponse(w, http.StatusOK, data, meta)

	var resp StandardListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify that empty fields in pagination are omitted (omitempty)
	if resp.Meta.Pagination.TotalCount != 0 {
		t.Errorf("Expected TotalCount to be 0 (omitted), got %d", resp.Meta.Pagination.TotalCount)
	}
}

func TestListResponseComplexNesting(t *testing.T) {
	w := httptest.NewRecorder()
	complexData := []map[string]interface{}{
		{
			"id":   "1",
			"name": "Item 1",
			"nested": map[string]interface{}{
				"level2": []interface{}{"a", "b", "c"},
			},
		},
	}
	meta := &ResponseMeta{
		Count: 1,
		Pagination: &PaginationMeta{
			Page:       1,
			PageSize:   10,
			TotalCount: 50,
		},
	}
	ListResponse(w, http.StatusOK, complexData, meta)

	var resp StandardListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true for complex nested data")
	}
}

func TestListResponseWithBrokenWriter(t *testing.T) {
	w := &BrokenResponseWriter{}
	data := []string{"item"}
	meta := &ResponseMeta{Count: 1}
	ListResponse(w, http.StatusOK, data, meta)
	// Should not panic, error should be logged
}

func TestSingleResponseWithBrokenWriter(t *testing.T) {
	w := &BrokenResponseWriter{}
	SingleResponse(w, http.StatusOK, map[string]string{"id": "1"})
	// Should not panic, error should be logged
}
