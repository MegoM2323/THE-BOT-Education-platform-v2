package validator

import (
	"errors"
	"testing"
	"time"
)

func TestTimeValidator_ValidateDateString(t *testing.T) {
	validator := NewTimeValidator()

	tests := []struct {
		name      string
		dateStr   string
		wantErr   error
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "Valid date within range - 2025-12-29",
			dateStr:   "2025-12-29",
			wantErr:   nil,
			wantYear:  2025,
			wantMonth: time.December,
			wantDay:   29,
		},
		{
			name:      "Valid date at minimum year - 2020-01-01",
			dateStr:   "2020-01-01",
			wantErr:   nil,
			wantYear:  2020,
			wantMonth: time.January,
			wantDay:   1,
		},
		{
			name:      "Valid date at maximum year - 2030-12-31",
			dateStr:   "2030-12-31",
			wantErr:   nil,
			wantYear:  2030,
			wantMonth: time.December,
			wantDay:   31,
		},
		{
			name:      "Valid leap year date - 2024-02-29",
			dateStr:   "2024-02-29",
			wantErr:   nil,
			wantYear:  2024,
			wantMonth: time.February,
			wantDay:   29,
		},
		{
			name:     "Invalid format - too short",
			dateStr:  "2025-12-2",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid format - too long",
			dateStr:  "2025-12-299",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid format - wrong separator",
			dateStr:  "2025/12/29",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid format - wrong order",
			dateStr:  "29-12-2025",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid format - letters in date",
			dateStr:  "202a-12-29",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid format - empty string",
			dateStr:  "",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid format - spaces",
			dateStr:  "2025-12-29 ",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Out of range - year before 2020",
			dateStr:  "2019-12-31",
			wantErr:  ErrDateOutOfRange,
			wantYear: 0,
		},
		{
			name:     "Out of range - year after 2030",
			dateStr:  "2031-01-01",
			wantErr:  ErrDateOutOfRange,
			wantYear: 0,
		},
		{
			name:     "Out of range - far future year",
			dateStr:  "2099-12-31",
			wantErr:  ErrDateOutOfRange,
			wantYear: 0,
		},
		{
			name:     "Out of range - far past year",
			dateStr:  "1999-01-01",
			wantErr:  ErrDateOutOfRange,
			wantYear: 0,
		},
		{
			name:     "Invalid - non-existent date (Feb 30)",
			dateStr:  "2025-02-30",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid - non-existent date (month 13)",
			dateStr:  "2025-13-01",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
		{
			name:     "Invalid - non-existent date (day 0)",
			dateStr:  "2025-12-00",
			wantErr:  ErrInvalidDateFormat,
			wantYear: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.ValidateDateString(tt.dateStr)

			// Check error
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateDateString() error = nil, want %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateDateString() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateDateString() unexpected error = %v", err)
					return
				}

				// Check date components
				if got.Year() != tt.wantYear {
					t.Errorf("ValidateDateString() year = %d, want %d", got.Year(), tt.wantYear)
				}
				if got.Month() != tt.wantMonth {
					t.Errorf("ValidateDateString() month = %v, want %v", got.Month(), tt.wantMonth)
				}
				if got.Day() != tt.wantDay {
					t.Errorf("ValidateDateString() day = %d, want %d", got.Day(), tt.wantDay)
				}
			}
		})
	}
}

func TestTimeValidator_ValidateDateRange(t *testing.T) {
	validator := NewTimeValidator()

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		maxDays   int
		wantErr   bool
		errType   error
	}{
		{
			name:      "Valid range - same day",
			startDate: time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			maxDays:   365,
			wantErr:   false,
		},
		{
			name:      "Valid range - 7 days",
			startDate: time.Date(2025, 12, 22, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			maxDays:   365,
			wantErr:   false,
		},
		{
			name:      "Valid range - exactly 365 days",
			startDate: time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			maxDays:   365,
			wantErr:   false,
		},
		{
			name:      "Valid range - 30 days with max 100",
			startDate: time.Date(2025, 11, 29, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			maxDays:   100,
			wantErr:   false,
		},
		{
			name:      "Invalid - end before start",
			startDate: time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 22, 0, 0, 0, 0, time.UTC),
			maxDays:   365,
			wantErr:   true,
			errType:   ErrInvalidTimeRange,
		},
		{
			name:      "Invalid - exceeds max days",
			startDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			maxDays:   365,
			wantErr:   true,
		},
		{
			name:      "Invalid - exceeds small max days limit",
			startDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 29, 0, 0, 0, 0, time.UTC),
			maxDays:   20,
			wantErr:   true,
		},
		{
			name:      "Edge case - just within limit (1 day before limit)",
			startDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			maxDays:   31,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDateRange(tt.startDate, tt.endDate, tt.maxDays)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errType != nil && err != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("ValidateDateRange() error = %v, want %v", err, tt.errType)
				}
			}
		})
	}
}

func TestTimeValidator_ValidateDateStringWithRange(t *testing.T) {
	// Integration test combining date string validation with range validation
	validator := NewTimeValidator()

	tests := []struct {
		name         string
		startDateStr string
		endDateStr   string
		maxDays      int
		wantErr      bool
	}{
		{
			name:         "Valid range - both valid dates",
			startDateStr: "2025-12-22",
			endDateStr:   "2025-12-29",
			maxDays:      365,
			wantErr:      false,
		},
		{
			name:         "Invalid - start date bad format",
			startDateStr: "2025/12/22",
			endDateStr:   "2025-12-29",
			maxDays:      365,
			wantErr:      true,
		},
		{
			name:         "Invalid - end date out of range",
			startDateStr: "2025-12-22",
			endDateStr:   "2031-12-29",
			maxDays:      365,
			wantErr:      true,
		},
		{
			name:         "Valid - minimum year dates",
			startDateStr: "2020-01-01",
			endDateStr:   "2020-01-08",
			maxDays:      365,
			wantErr:      false,
		},
		{
			name:         "Valid - maximum year dates",
			startDateStr: "2030-12-24",
			endDateStr:   "2030-12-31",
			maxDays:      365,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDate, errStart := validator.ValidateDateString(tt.startDateStr)
			if errStart != nil {
				if !tt.wantErr {
					t.Errorf("ValidateDateString(start) error = %v, want nil", errStart)
				}
				return
			}

			endDate, errEnd := validator.ValidateDateString(tt.endDateStr)
			if errEnd != nil {
				if !tt.wantErr {
					t.Errorf("ValidateDateString(end) error = %v, want nil", errEnd)
				}
				return
			}

			rangeErr := validator.ValidateDateRange(startDate, endDate, tt.maxDays)
			if (rangeErr != nil) != tt.wantErr {
				t.Errorf("ValidateDateRange() error = %v, wantErr %v", rangeErr, tt.wantErr)
			}
		})
	}
}
