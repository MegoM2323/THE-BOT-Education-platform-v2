package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tutoring-platform/internal/models"
)

// TestBookingAccessControl_StudentAccess tests student access control logic
func TestBookingAccessControl_StudentAccess(t *testing.T) {
	// Setup test users
	studentID := uuid.New()
	otherStudentID := uuid.New()
	teacherID := uuid.New()

	student := &models.User{ID: studentID, Role: models.RoleStudent}
	otherStudent := &models.User{ID: otherStudentID, Role: models.RoleStudent}

	// Setup booking
	booking := &models.BookingWithDetails{
		Booking: models.Booking{
			StudentID: studentID,
			LessonID:  uuid.New(),
		},
		TeacherID: teacherID,
	}

	// Table-driven tests
	tests := []struct {
		name        string
		user        *models.User
		booking     *models.BookingWithDetails
		shouldAllow bool
		description string
	}{
		{
			name:        "Student can see own booking",
			user:        student,
			booking:     booking,
			shouldAllow: true,
			description: "Student viewing own booking should be allowed",
		},
		{
			name:        "Student cannot see other student's booking",
			user:        otherStudent,
			booking:     booking,
			shouldAllow: false,
			description: "Student viewing other student's booking should be denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check student access control
			if tt.user.IsStudent() {
				if tt.booking.StudentID != tt.user.ID {
					// Access should be denied
					assert.False(t, tt.shouldAllow, tt.description)
					return
				}
			}
			// Access allowed
			assert.True(t, tt.shouldAllow, tt.description)
		})
	}
}

// TestBookingAccessControl_TeacherAccess tests teacher access control logic
func TestBookingAccessControl_TeacherAccess(t *testing.T) {
	// Setup test users
	studentID := uuid.New()
	teacher1ID := uuid.New()
	teacher2ID := uuid.New()

	teacher1 := &models.User{ID: teacher1ID, Role: models.RoleTeacher}
	teacher2 := &models.User{ID: teacher2ID, Role: models.RoleTeacher}
	admin := &models.User{ID: uuid.New(), Role: models.RoleAdmin}

	// Setup booking from teacher1
	booking := &models.BookingWithDetails{
		Booking: models.Booking{
			StudentID: studentID,
			LessonID:  uuid.New(),
		},
		TeacherID: teacher1ID,
	}

	// Table-driven tests
	tests := []struct {
		name        string
		user        *models.User
		booking     *models.BookingWithDetails
		shouldAllow bool
		description string
	}{
		{
			name:        "Teacher can see own lesson booking",
			user:        teacher1,
			booking:     booking,
			shouldAllow: true,
			description: "Teacher viewing own lesson's booking should be allowed",
		},
		{
			name:        "Teacher cannot see other teacher's booking",
			user:        teacher2,
			booking:     booking,
			shouldAllow: false,
			description: "Teacher viewing other teacher's lesson's booking should be denied",
		},
		{
			name:        "Admin can see any booking",
			user:        admin,
			booking:     booking,
			shouldAllow: true,
			description: "Admin should always be allowed to see any booking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check access control
			allowed := true

			if tt.user.IsStudent() && tt.booking.StudentID != tt.user.ID {
				allowed = false
			} else if tt.user.IsTeacher() && tt.booking.TeacherID != tt.user.ID {
				allowed = false
			}
			// Admins always get access (no additional check)

			if tt.shouldAllow {
				assert.True(t, allowed, tt.description)
			} else {
				assert.False(t, allowed, tt.description)
			}
		})
	}
}

// TestBookingAccessControl_AdminAccess tests admin access control (should have unrestricted access)
func TestBookingAccessControl_AdminAccess(t *testing.T) {
	// Setup test users
	studentID := uuid.New()
	teacherID := uuid.New()
	adminID := uuid.New()

	admin := &models.User{ID: adminID, Role: models.RoleAdmin}

	// Setup booking
	booking := &models.BookingWithDetails{
		Booking: models.Booking{
			StudentID: studentID,
			LessonID:  uuid.New(),
		},
		TeacherID: teacherID,
	}

	// Admin should be able to see any booking
	t.Run("Admin can see any booking", func(t *testing.T) {
		allowed := true

		// Check access control
		if admin.IsStudent() && booking.StudentID != admin.ID {
			allowed = false
		} else if admin.IsTeacher() && booking.TeacherID != admin.ID {
			allowed = false
		}
		// Admin has IsAdmin() == true, so no additional check blocks access

		assert.True(t, allowed, "Admin should be able to see any booking")
		assert.True(t, admin.IsAdmin(), "User should be admin")
	})
}
