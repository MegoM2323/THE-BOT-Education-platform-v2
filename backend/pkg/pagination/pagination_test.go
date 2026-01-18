package pagination

import (
	"net/http"
	"testing"
)

func TestParseParams(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantPage    int
		wantPerPage int
		wantOffset  int
	}{
		{
			name:        "default values",
			query:       "",
			wantPage:    1,
			wantPerPage: DefaultPerPage,
			wantOffset:  0,
		},
		{
			name:        "page 2 with default per_page",
			query:       "page=2",
			wantPage:    2,
			wantPerPage: DefaultPerPage,
			wantOffset:  DefaultPerPage,
		},
		{
			name:        "page 1 with per_page 50",
			query:       "per_page=50",
			wantPage:    1,
			wantPerPage: 50,
			wantOffset:  0,
		},
		{
			name:        "page 2 with per_page 10",
			query:       "page=2&per_page=10",
			wantPage:    2,
			wantPerPage: 10,
			wantOffset:  10,
		},
		{
			name:        "per_page exceeds max",
			query:       "per_page=200",
			wantPage:    1,
			wantPerPage: MaxPerPage,
			wantOffset:  0,
		},
		{
			name:        "limit/offset style",
			query:       "limit=25&offset=50",
			wantPage:    3, // (50 / 25) + 1
			wantPerPage: 25,
			wantOffset:  50,
		},
		{
			name:        "limit/offset with zero offset",
			query:       "limit=30&offset=0",
			wantPage:    1,
			wantPerPage: 30,
			wantOffset:  0,
		},
		{
			name:        "invalid page number",
			query:       "page=abc",
			wantPage:    1,
			wantPerPage: DefaultPerPage,
			wantOffset:  0,
		},
		{
			name:        "negative page number",
			query:       "page=-5",
			wantPage:    1,
			wantPerPage: DefaultPerPage,
			wantOffset:  0,
		},
		{
			name:        "zero per_page defaults to DefaultPerPage",
			query:       "per_page=0",
			wantPage:    1,
			wantPerPage: DefaultPerPage,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com/?"+tt.query, nil)
			params := ParseParams(req)

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

func TestNewMeta(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		perPage   int
		total     int
		wantTotal int
		wantPages int
	}{
		{
			name:      "single page",
			page:      1,
			perPage:   20,
			total:     15,
			wantTotal: 15,
			wantPages: 1,
		},
		{
			name:      "multiple pages",
			page:      2,
			perPage:   20,
			total:     100,
			wantTotal: 100,
			wantPages: 5,
		},
		{
			name:      "partial last page",
			page:      3,
			perPage:   30,
			total:     80,
			wantTotal: 80,
			wantPages: 3,
		},
		{
			name:      "zero total",
			page:      1,
			perPage:   20,
			total:     0,
			wantTotal: 0,
			wantPages: 1,
		},
		{
			name:      "total equals per_page",
			page:      1,
			perPage:   50,
			total:     50,
			wantTotal: 50,
			wantPages: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewMeta(tt.page, tt.perPage, tt.total)

			if meta.Page != tt.page {
				t.Errorf("Page = %d, want %d", meta.Page, tt.page)
			}
			if meta.PerPage != tt.perPage {
				t.Errorf("PerPage = %d, want %d", meta.PerPage, tt.perPage)
			}
			if meta.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", meta.Total, tt.wantTotal)
			}
			if meta.TotalPages != tt.wantPages {
				t.Errorf("TotalPages = %d, want %d", meta.TotalPages, tt.wantPages)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	data := map[string]interface{}{"items": []string{"item1", "item2"}}
	resp := NewResponse(data, 1, 20, 50)

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Data is not map[string]interface{}")
	}
	items, exists := dataMap["items"]
	if !exists {
		t.Errorf("items key not found in data")
	}
	if items == nil {
		t.Errorf("items value is nil")
	}

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
}
