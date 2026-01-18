package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListResponseIntegration simulates how handlers use ListResponse
func TestListResponseIntegration(t *testing.T) {
	tests := []struct {
		name           string
		items          interface{}
		count          int
		withPagination bool
		expectedStatus int
	}{
		{
			name:           "Integration: Simple list of users",
			items:          []map[string]string{{"id": "1", "email": "a@b.com"}, {"id": "2", "email": "c@d.com"}},
			count:          2,
			withPagination: false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Integration: Paginated lessons list",
			items:          []map[string]interface{}{{"id": "1", "title": "Math"}, {"id": "2", "title": "Physics"}},
			count:          2,
			withPagination: true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Integration: Empty results",
			items:          []interface{}{},
			count:          0,
			withPagination: false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			meta := &ResponseMeta{Count: tt.count}
			if tt.withPagination {
				meta.Pagination = &PaginationMeta{
					Page:       1,
					PageSize:   10,
					TotalCount: 50,
					TotalPages: 5,
				}
			}

			ListResponse(w, tt.expectedStatus, tt.items, meta)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var resp StandardListResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			if !resp.Success {
				t.Error("Expected success=true")
			}

			if resp.Meta.Count != tt.count {
				t.Errorf("Expected count=%d, got %d", tt.count, resp.Meta.Count)
			}

			if tt.withPagination && resp.Meta.Pagination == nil {
				t.Error("Expected pagination metadata")
			}
		})
	}
}

// TestSingleResponseIntegration simulates how handlers use SingleResponse
func TestSingleResponseIntegration(t *testing.T) {
	tests := []struct {
		name        string
		item        interface{}
		statusCode  int
		description string
	}{
		{
			name:        "Integration: Get user",
			item:        map[string]interface{}{"id": "123", "email": "test@test.com", "role": "student"},
			statusCode:  http.StatusOK,
			description: "Retrieve single user",
		},
		{
			name:        "Integration: Create lesson",
			item:        map[string]interface{}{"id": "abc", "title": "New Lesson", "created_at": "2024-01-01T10:00:00Z"},
			statusCode:  http.StatusCreated,
			description: "Resource creation returns 201",
		},
		{
			name:        "Integration: Get booking",
			item:        map[string]interface{}{"id": "booking-1", "lesson_id": "lesson-1", "student_id": "student-1"},
			statusCode:  http.StatusOK,
			description: "Retrieve single booking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SingleResponse(w, tt.statusCode, tt.item)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			var resp StandardSingleResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			if !resp.Success {
				t.Error("Expected success=true")
			}

			if resp.Data == nil {
				t.Error("Expected data field to be present")
			}
		})
	}
}

// TestErrorResponseIntegration simulates how handlers respond with errors
func TestErrorResponseIntegration(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		errorCode     string
		message       string
		validateError func(*ErrorDetail) bool
	}{
		{
			name:       "Integration: Validation error",
			statusCode: http.StatusBadRequest,
			errorCode:  ErrCodeInvalidInput,
			message:    "Email format is invalid",
			validateError: func(ed *ErrorDetail) bool {
				return ed.Code == ErrCodeInvalidInput && ed.Message == "Email format is invalid"
			},
		},
		{
			name:       "Integration: Not found error",
			statusCode: http.StatusNotFound,
			errorCode:  ErrCodeNotFound,
			message:    "User not found",
			validateError: func(ed *ErrorDetail) bool {
				return ed.Code == ErrCodeNotFound
			},
		},
		{
			name:       "Integration: Conflict error",
			statusCode: http.StatusConflict,
			errorCode:  ErrCodeAlreadyExists,
			message:    "Email already registered",
			validateError: func(ed *ErrorDetail) bool {
				return ed.Code == ErrCodeAlreadyExists
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Error(w, tt.statusCode, tt.errorCode, tt.message)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			var resp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode: %v", err)
			}

			if resp.Success {
				t.Error("Expected success=false for error response")
			}

			if !tt.validateError(&resp.Error) {
				t.Errorf("Error validation failed. Got: code=%s, message=%s", resp.Error.Code, resp.Error.Message)
			}
		})
	}
}

// TestResponseConsistency verifies response structure consistency
func TestResponseConsistency(t *testing.T) {
	// Test that list and single responses follow expected JSON structure
	w1 := httptest.NewRecorder()
	ListResponse(w1, http.StatusOK, []string{"a", "b"}, &ResponseMeta{Count: 2})

	var listResp StandardListResponse
	json.NewDecoder(w1.Body).Decode(&listResp)

	// Verify required fields exist
	if listResp.Success != true {
		t.Error("List response missing success field")
	}
	if listResp.Data == nil {
		t.Error("List response missing data field")
	}
	if listResp.Meta.Count == 0 {
		t.Error("List response missing meta.count")
	}

	// Test single response
	w2 := httptest.NewRecorder()
	SingleResponse(w2, http.StatusOK, map[string]string{"id": "1"})

	var singleResp StandardSingleResponse
	if err := json.NewDecoder(w2.Body).Decode(&singleResp); err != nil {
		t.Fatalf("Failed to decode single response: %v", err)
	}

	if singleResp.Success != true {
		t.Error("Single response missing success field")
	}
	if singleResp.Data == nil {
		t.Error("Single response missing data field")
	}
}

// TestPaginationCalculation verifies correct pagination calculations
func TestPaginationCalculation(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		pageSize    int
		totalCount  int
		expectedMax int
	}{
		{
			name:        "First page",
			page:        1,
			pageSize:    10,
			totalCount:  100,
			expectedMax: 10,
		},
		{
			name:        "Middle page",
			page:        5,
			pageSize:    10,
			totalCount:  100,
			expectedMax: 10,
		},
		{
			name:        "Last page partial",
			page:        11,
			pageSize:    10,
			totalCount:  105,
			expectedMax: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalPages := (tt.totalCount + tt.pageSize - 1) / tt.pageSize

			items := make([]int, 0)
			for i := 0; i < tt.expectedMax; i++ {
				items = append(items, i)
			}

			w := httptest.NewRecorder()
			ListResponse(w, http.StatusOK, items, &ResponseMeta{
				Count: len(items),
				Pagination: &PaginationMeta{
					Page:       tt.page,
					PageSize:   tt.pageSize,
					TotalCount: tt.totalCount,
					TotalPages: totalPages,
				},
			})

			var resp StandardListResponse
			json.NewDecoder(w.Body).Decode(&resp)

			if resp.Meta.Count != tt.expectedMax {
				t.Errorf("Expected %d items, got %d", tt.expectedMax, resp.Meta.Count)
			}

			if resp.Meta.Pagination.TotalPages != totalPages {
				t.Errorf("Expected %d total pages, got %d", totalPages, resp.Meta.Pagination.TotalPages)
			}
		})
	}
}
