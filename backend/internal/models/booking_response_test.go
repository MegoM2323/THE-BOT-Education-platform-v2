package models

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBookingResponse_ToBookingResponse_FromBooking(t *testing.T) {
	tests := []struct {
		name    string
		booking *Booking
		verify  func(*BookingResponse)
	}{
		{
			name: "basic booking conversion",
			booking: &Booking{
				ID:        uuid.New(),
				StudentID: uuid.New(),
				LessonID:  uuid.New(),
				Status:    BookingStatusActive,
				BookedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			verify: func(resp *BookingResponse) {
				if resp.ID == uuid.Nil {
					t.Error("ID should be set")
				}
				if resp.StudentID == uuid.Nil {
					t.Error("StudentID should be set")
				}
				if resp.LessonID == uuid.Nil {
					t.Error("LessonID should be set")
				}
				if resp.Status != BookingStatusActive {
					t.Errorf("Status = %v, want %v", resp.Status, BookingStatusActive)
				}
				if resp.CancelledAt != nil {
					t.Error("CancelledAt should be nil for active booking")
				}
			},
		},
		{
			name: "cancelled booking with null date",
			booking: &Booking{
				ID:        uuid.New(),
				StudentID: uuid.New(),
				LessonID:  uuid.New(),
				Status:    BookingStatusCancelled,
				BookedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CancelledAt: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			},
			verify: func(resp *BookingResponse) {
				if resp.Status != BookingStatusCancelled {
					t.Errorf("Status = %v, want %v", resp.Status, BookingStatusCancelled)
				}
				if resp.CancelledAt == nil {
					t.Error("CancelledAt should not be nil for cancelled booking")
				}
			},
		},
		{
			name: "booking with invalid cancel time",
			booking: &Booking{
				ID:        uuid.New(),
				StudentID: uuid.New(),
				LessonID:  uuid.New(),
				Status:    BookingStatusActive,
				BookedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CancelledAt: sql.NullTime{
					Valid: false,
				},
			},
			verify: func(resp *BookingResponse) {
				if resp.CancelledAt != nil {
					t.Error("CancelledAt should be nil when sql.NullTime is invalid")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.booking.ToBookingResponse()
			if resp == nil {
				t.Fatal("ToBookingResponse returned nil")
			}
			tt.verify(resp)
		})
	}
}

func TestBookingResponse_ToBookingResponse_FromBookingWithDetails(t *testing.T) {
	teacherID := uuid.New()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(1 * time.Hour)
	bookingCreatedAt := time.Now()

	tests := []struct {
		name           string
		bookingDetails *BookingWithDetails
		verify         func(*BookingResponse)
	}{
		{
			name: "full booking with all details",
			bookingDetails: &BookingWithDetails{
				Booking: Booking{
					ID:        uuid.New(),
					StudentID: uuid.New(),
					LessonID:  uuid.New(),
					Status:    BookingStatusActive,
					BookedAt:  time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				StartTime:        startTime,
				EndTime:          endTime,
				TeacherID:        teacherID,
				TeacherName:      "John Doe",
				StudentFullName:  "Jane Smith",
				StudentEmail:     "jane@example.com",
				BookingCreatedAt: bookingCreatedAt,
			},
			verify: func(resp *BookingResponse) {
				if resp.StartTime == nil || *resp.StartTime != startTime {
					t.Error("StartTime not properly converted")
				}
				if resp.EndTime == nil || *resp.EndTime != endTime {
					t.Error("EndTime not properly converted")
				}
				if resp.TeacherID == nil || *resp.TeacherID != teacherID {
					t.Error("TeacherID not properly converted")
				}
				if resp.TeacherName != "John Doe" {
					t.Errorf("TeacherName = %v, want %v", resp.TeacherName, "John Doe")
				}
				if resp.StudentFullName != "Jane Smith" {
					t.Errorf("StudentFullName = %v, want %v", resp.StudentFullName, "Jane Smith")
				}
				if resp.StudentEmail != "jane@example.com" {
					t.Errorf("StudentEmail = %v, want %v", resp.StudentEmail, "jane@example.com")
				}
				if resp.BookingCreatedAt == nil || *resp.BookingCreatedAt != bookingCreatedAt {
					t.Error("BookingCreatedAt not properly converted")
				}
			},
		},
		{
			name: "booking with zero times",
			bookingDetails: &BookingWithDetails{
				Booking: Booking{
					ID:        uuid.New(),
					StudentID: uuid.New(),
					LessonID:  uuid.New(),
					Status:    BookingStatusActive,
					BookedAt:  time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				StartTime:        time.Time{},
				EndTime:          time.Time{},
				TeacherID:        uuid.Nil,
				TeacherName:      "",
				StudentFullName:  "",
				StudentEmail:     "",
				BookingCreatedAt: time.Time{},
			},
			verify: func(resp *BookingResponse) {
				if resp.StartTime != nil {
					t.Error("StartTime should be nil for zero value")
				}
				if resp.EndTime != nil {
					t.Error("EndTime should be nil for zero value")
				}
				if resp.TeacherID != nil {
					t.Error("TeacherID should be nil for nil uuid")
				}
				if resp.BookingCreatedAt != nil {
					t.Error("BookingCreatedAt should be nil for zero value")
				}
			},
		},
		{
			name: "booking with cancelled date",
			bookingDetails: &BookingWithDetails{
				Booking: Booking{
					ID:        uuid.New(),
					StudentID: uuid.New(),
					LessonID:  uuid.New(),
					Status:    BookingStatusCancelled,
					BookedAt:  time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					CancelledAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				},
				StartTime:        startTime,
				EndTime:          endTime,
				TeacherID:        teacherID,
				TeacherName:      "John Doe",
				StudentFullName:  "Jane Smith",
				StudentEmail:     "jane@example.com",
				BookingCreatedAt: bookingCreatedAt,
			},
			verify: func(resp *BookingResponse) {
				if resp.Status != BookingStatusCancelled {
					t.Error("Status should be cancelled")
				}
				if resp.CancelledAt == nil {
					t.Error("CancelledAt should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.bookingDetails.ToBookingResponse()
			if resp == nil {
				t.Fatal("ToBookingResponse returned nil")
			}
			tt.verify(resp)
		})
	}
}

func TestBookingResponse_JSONMarshal(t *testing.T) {
	resp := &BookingResponse{
		ID:        uuid.New(),
		StudentID: uuid.New(),
		LessonID:  uuid.New(),
		Status:    BookingStatusActive,
		BookedAt:  time.Now().Round(time.Second),
		CreatedAt: time.Now().Round(time.Second),
		UpdatedAt: time.Now().Round(time.Second),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Check that required fields are present
	if _, ok := unmarshalled["id"]; !ok {
		t.Error("id field missing from JSON")
	}
	if _, ok := unmarshalled["student_id"]; !ok {
		t.Error("student_id field missing from JSON")
	}
	if _, ok := unmarshalled["lesson_id"]; !ok {
		t.Error("lesson_id field missing from JSON")
	}
	if _, ok := unmarshalled["status"]; !ok {
		t.Error("status field missing from JSON")
	}

	// Check that omitted null fields are not present
	if _, ok := unmarshalled["cancelled_at"]; ok {
		t.Error("cancelled_at field should be omitted when nil")
	}
	if _, ok := unmarshalled["start_time"]; ok {
		t.Error("start_time field should be omitted when nil")
	}
}

func TestBookingResponse_JSONMarshal_WithOptionalFields(t *testing.T) {
	teacherID := uuid.New()
	startTime := time.Now().Round(time.Second)
	endTime := startTime.Add(time.Hour)
	cancelledAt := time.Now().Round(time.Second)

	resp := &BookingResponse{
		ID:              uuid.New(),
		StudentID:       uuid.New(),
		LessonID:        uuid.New(),
		Status:          BookingStatusCancelled,
		BookedAt:        time.Now().Round(time.Second),
		CancelledAt:     &cancelledAt,
		CreatedAt:       time.Now().Round(time.Second),
		UpdatedAt:       time.Now().Round(time.Second),
		StartTime:       &startTime,
		EndTime:         &endTime,
		TeacherID:       &teacherID,
		TeacherName:     "Teacher Name",
		StudentFullName: "Student Name",
		StudentEmail:    "student@example.com",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Check that optional fields are present
	if _, ok := unmarshalled["cancelled_at"]; !ok {
		t.Error("cancelled_at field should be present when not nil")
	}
	if _, ok := unmarshalled["start_time"]; !ok {
		t.Error("start_time field should be present when not nil")
	}
	if _, ok := unmarshalled["teacher_id"]; !ok {
		t.Error("teacher_id field should be present when not nil")
	}
	if _, ok := unmarshalled["teacher_name"]; !ok {
		t.Error("teacher_name field should be present")
	}
	if _, ok := unmarshalled["student_name"]; !ok {
		t.Error("student_name field should be present")
	}
}

func TestBookingResponse_ConversionRoundtrip(t *testing.T) {
	// Create original booking
	teacherID := uuid.New()
	originalBooking := &BookingWithDetails{
		Booking: Booking{
			ID:        uuid.New(),
			StudentID: uuid.New(),
			LessonID:  uuid.New(),
			Status:    BookingStatusActive,
			BookedAt:  time.Now().Round(time.Microsecond),
			CreatedAt: time.Now().Round(time.Microsecond),
			UpdatedAt: time.Now().Round(time.Microsecond),
		},
		StartTime:        time.Now().Add(24 * time.Hour).Round(time.Microsecond),
		EndTime:          time.Now().Add(25 * time.Hour).Round(time.Microsecond),
		TeacherID:        teacherID,
		TeacherName:      "Test Teacher",
		StudentFullName:  "Test Student",
		StudentEmail:     "test@example.com",
		BookingCreatedAt: time.Now().Round(time.Microsecond),
	}

	// Convert to response
	resp := originalBooking.ToBookingResponse()

	// Verify all fields are preserved
	if resp.ID != originalBooking.ID {
		t.Error("ID not preserved in conversion")
	}
	if resp.StudentID != originalBooking.StudentID {
		t.Error("StudentID not preserved in conversion")
	}
	if resp.LessonID != originalBooking.LessonID {
		t.Error("LessonID not preserved in conversion")
	}
	if resp.Status != originalBooking.Status {
		t.Error("Status not preserved in conversion")
	}
	if resp.StartTime != nil && *resp.StartTime != originalBooking.StartTime {
		t.Error("StartTime not preserved in conversion")
	}
	if resp.EndTime != nil && *resp.EndTime != originalBooking.EndTime {
		t.Error("EndTime not preserved in conversion")
	}
	if resp.TeacherID != nil && *resp.TeacherID != originalBooking.TeacherID {
		t.Error("TeacherID not preserved in conversion")
	}
	if resp.TeacherName != originalBooking.TeacherName {
		t.Error("TeacherName not preserved in conversion")
	}
	if resp.StudentFullName != originalBooking.StudentFullName {
		t.Error("StudentFullName not preserved in conversion")
	}
	if resp.StudentEmail != originalBooking.StudentEmail {
		t.Error("StudentEmail not preserved in conversion")
	}
}

func TestBookingResponse_NoSensitiveFields(t *testing.T) {
	// Verify that BookingResponse doesn't expose sensitive fields
	resp := &BookingResponse{
		ID:        uuid.New(),
		StudentID: uuid.New(),
		LessonID:  uuid.New(),
		Status:    BookingStatusActive,
		BookedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(data)

	// Fields that should NOT appear
	sensitiveFields := []string{"password", "secret", "token", "credit", "balance", "admin"}

	for _, field := range sensitiveFields {
		if _, ok := jsonToMap(data)[field]; ok {
			t.Errorf("Sensitive field %q should not be in response", field)
		}
	}

	// Ensure we're not exposing the raw struct content
	if len(jsonStr) == 0 {
		t.Error("Response should not be empty")
	}
}

// Helper function to convert JSON to map
func jsonToMap(data []byte) map[string]interface{} {
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	return m
}

func TestBookingResponse_EmptyBooking(t *testing.T) {
	emptyBooking := &Booking{}
	resp := emptyBooking.ToBookingResponse()

	if resp == nil {
		t.Fatal("ToBookingResponse should not return nil for empty booking")
	}

	if resp.ID != uuid.Nil {
		t.Error("Empty booking should have nil UUID")
	}
	if resp.Status != "" {
		t.Error("Empty booking should have empty status")
	}
	if resp.CancelledAt != nil {
		t.Error("Empty booking should have nil CancelledAt")
	}
}

func TestBookingResponse_MultipleConversions(t *testing.T) {
	// Test that multiple conversions from same data are consistent
	bookingDetails := &BookingWithDetails{
		Booking: Booking{
			ID:        uuid.New(),
			StudentID: uuid.New(),
			LessonID:  uuid.New(),
			Status:    BookingStatusActive,
			BookedAt:  time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		StartTime:       time.Now().Add(24 * time.Hour),
		EndTime:         time.Now().Add(25 * time.Hour),
		TeacherID:       uuid.New(),
		TeacherName:     "Teacher",
		StudentFullName: "Student",
		StudentEmail:    "student@example.com",
	}

	resp1 := bookingDetails.ToBookingResponse()
	resp2 := bookingDetails.ToBookingResponse()

	if resp1.ID != resp2.ID {
		t.Error("Multiple conversions produce different IDs")
	}
	if resp1.TeacherName != resp2.TeacherName {
		t.Error("Multiple conversions produce different TeacherName")
	}
}

func TestBookingWithDetails_MarshalJSON_SubjectAndHomework(t *testing.T) {
	tests := []struct {
		name           string
		booking        *BookingWithDetails
		expectSubject  string
		expectHomework string
	}{
		{
			name: "with valid subject and homework",
			booking: &BookingWithDetails{
				Booking: Booking{
					ID:        uuid.New(),
					StudentID: uuid.New(),
					LessonID:  uuid.New(),
					Status:    BookingStatusActive,
					BookedAt:  time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				StartTime:    time.Now().Add(24 * time.Hour),
				EndTime:      time.Now().Add(25 * time.Hour),
				TeacherID:    uuid.New(),
				TeacherName:  "Teacher",
				Subject:      sql.NullString{String: "Математика", Valid: true},
				HomeworkText: sql.NullString{String: "Решить задачи 1-10", Valid: true},
			},
			expectSubject:  "Математика",
			expectHomework: "Решить задачи 1-10",
		},
		{
			name: "with null subject and homework",
			booking: &BookingWithDetails{
				Booking: Booking{
					ID:        uuid.New(),
					StudentID: uuid.New(),
					LessonID:  uuid.New(),
					Status:    BookingStatusActive,
					BookedAt:  time.Now(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				StartTime:    time.Now().Add(24 * time.Hour),
				EndTime:      time.Now().Add(25 * time.Hour),
				TeacherID:    uuid.New(),
				TeacherName:  "Teacher",
				Subject:      sql.NullString{Valid: false},
				HomeworkText: sql.NullString{Valid: false},
			},
			expectSubject:  "",
			expectHomework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.booking)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Check subject field
			if subject, ok := result["subject"].(string); ok {
				if subject != tt.expectSubject {
					t.Errorf("subject = %q, want %q", subject, tt.expectSubject)
				}
			} else if tt.expectSubject != "" {
				t.Errorf("subject field missing, expected %q", tt.expectSubject)
			}

			// Check homework_text field
			if homework, ok := result["homework_text"].(string); ok {
				if homework != tt.expectHomework {
					t.Errorf("homework_text = %q, want %q", homework, tt.expectHomework)
				}
			} else if tt.expectHomework != "" {
				t.Errorf("homework_text field missing, expected %q", tt.expectHomework)
			}

			// Verify that NullString fields are NOT serialized as objects
			if subjectObj, ok := result["subject"].(map[string]interface{}); ok {
				t.Errorf("subject should be string, got object: %v", subjectObj)
			}
			if homeworkObj, ok := result["homework_text"].(map[string]interface{}); ok {
				t.Errorf("homework_text should be string, got object: %v", homeworkObj)
			}
		})
	}
}
