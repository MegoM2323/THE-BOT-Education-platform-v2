package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"tutoring-platform/internal/models"
	"tutoring-platform/pkg/pagination"

	"github.com/google/uuid"
)

// MockUserService для тестирования
type MockUserService struct {
	users []*models.User
	total int
}

func (m *MockUserService) ListUsersWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	return m.users, m.total, nil
}

// TestGetUsersPaginationParams проверяет правильный парсинг параметров пагинации
func TestGetUsersPaginationParams(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantPage    int
		wantPerPage int
		wantOffset  int
	}{
		{
			name:        "default pagination",
			query:       "",
			wantPage:    1,
			wantPerPage: pagination.DefaultPerPage,
			wantOffset:  0,
		},
		{
			name:        "page 2 with default per_page",
			query:       "page=2",
			wantPage:    2,
			wantPerPage: pagination.DefaultPerPage,
			wantOffset:  pagination.DefaultPerPage,
		},
		{
			name:        "custom per_page",
			query:       "per_page=50",
			wantPage:    1,
			wantPerPage: 50,
			wantOffset:  0,
		},
		{
			name:        "per_page exceeds max",
			query:       "per_page=200",
			wantPage:    1,
			wantPerPage: pagination.MaxPerPage,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/users?"+tt.query, nil)
			params := pagination.ParseParams(req)

			if params.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.wantPage)
			}
			if params.PerPage != tt.wantPerPage {
				t.Errorf("PerPage = %d, want %d", params.PerPage, tt.wantPerPage)
			}
			if params.Offset != tt.wantOffset {
				t.Errorf("Offset = %d, want %d", params.Offset, tt.wantOffset)
			}
		})
	}
}

// TestPaginationResponseStructure проверяет структуру пагинированного ответа
func TestPaginationResponseStructure(t *testing.T) {
	users := []*models.User{
		{
			ID:       uuid.New(),
			Email:    "test1@example.com",
			FirstName: "User", LastName: "Test User 1",
		},
		{
			ID:       uuid.New(),
			Email:    "test2@example.com",
			FirstName: "User", LastName: "Test User 2",
		},
	}

	data := map[string]interface{}{
		"users": users,
	}

	resp := pagination.NewResponse(data, 1, 20, 50)

	if resp.Meta.Page != 1 {
		t.Errorf("Meta.Page = %d, want 1", resp.Meta.Page)
	}
	if resp.Meta.PerPage != 20 {
		t.Errorf("Meta.PerPage = %d, want 20", resp.Meta.PerPage)
	}
	if resp.Meta.Total != 50 {
		t.Errorf("Meta.Total = %d, want 50", resp.Meta.Total)
	}
	if resp.Meta.TotalPages != 3 {
		t.Errorf("Meta.TotalPages = %d, want 3", resp.Meta.TotalPages)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Data is not map[string]interface{}")
	}

	usersData, exists := dataMap["users"]
	if !exists {
		t.Errorf("users key not found in data")
	}

	usersSlice, ok := usersData.([]*models.User)
	if !ok {
		t.Errorf("users is not []*models.User")
	}

	if len(usersSlice) != len(users) {
		t.Errorf("users count = %d, want %d", len(usersSlice), len(users))
	}
}

// TestPaginationLimitsBoundary проверяет граничные значения пагинации
func TestPaginationLimitsBoundary(t *testing.T) {
	tests := []struct {
		name     string
		perPage  int
		total    int
		expected int
	}{
		{
			name:     "single page with 10 items",
			perPage:  20,
			total:    10,
			expected: 1,
		},
		{
			name:     "exact pages",
			perPage:  20,
			total:    60,
			expected: 3,
		},
		{
			name:     "partial last page",
			perPage:  20,
			total:    65,
			expected: 4,
		},
		{
			name:     "empty results",
			perPage:  20,
			total:    0,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := pagination.NewMeta(1, tt.perPage, tt.total)
			if meta.TotalPages != tt.expected {
				t.Errorf("TotalPages = %d, want %d", meta.TotalPages, tt.expected)
			}
		})
	}
}

// TestOffsetCalculation проверяет правильный расчет offset из page
func TestOffsetCalculation(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		perPage  int
		expected int
	}{
		{
			name:     "page 1",
			page:     1,
			perPage:  20,
			expected: 0,
		},
		{
			name:     "page 2",
			page:     2,
			perPage:  20,
			expected: 20,
		},
		{
			name:     "page 3 with custom per_page",
			page:     3,
			perPage:  50,
			expected: 100,
		},
		{
			name:     "page 10 with per_page 25",
			page:     10,
			perPage:  25,
			expected: 225,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := pagination.Params{
				Page:    tt.page,
				PerPage: tt.perPage,
			}
			offset := (params.Page - 1) * params.PerPage
			if offset != tt.expected {
				t.Errorf("offset = %d, want %d", offset, tt.expected)
			}
		})
	}
}
