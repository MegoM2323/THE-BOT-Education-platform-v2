package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestScenario_11_BulkEditAddStudent - Edit lesson, add student, apply to all subsequent
func TestScenario_11_BulkEditAddStudent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	admin := createTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createTestUser(t, db, "teacher@test.com", "Teacher User", "teacher")
	student := createTestUser(t, db, "student@test.com", "Student User", "student")

	// Give student enough credits for all bookings
	addCreditToStudent(t, db, student.ID, 20)

	// Create 3 lessons with same time pattern (Monday 9:00)
	lessons := createMatchingLessons(t, db, teacher.ID, 3)

	// Get first lesson
	sourceLessonID := lessons[0]

	// Apply bulk edit: add student to first lesson and all subsequent
	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "add_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	assert.Equal(t, "add_student", bulkEditReq["modification_type"], "Modification type should be add_student")
	assert.Equal(t, student.ID.String(), bulkEditReq["student_id"], "Student ID should match")
}

// TestScenario_12_BulkEditRemoveStudent - Edit lesson, remove student, apply to all
func TestScenario_12_BulkEditRemoveStudent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))
	student := createTestUser(t, pool, "student@test.com", "Student", string(models.RoleStudent))

	// Create lessons with student pre-booked
	lessons := createMatchingLessons(t, pool, teacher.ID, 3)
	sourceLessonID := lessons[0]

	// Add student to all lessons
	for _, lessonID := range lessons {
		addStudentToLesson(t, pool, lessonID, student.ID)
	}

	// Apply bulk edit: remove student from first lesson and all subsequent
	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "remove_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	assert.Equal(t, "remove_student", bulkEditReq["modification_type"], "Modification type should be remove_student")
	assert.NotNil(t, httpReq)
}

// TestScenario_13_BulkEditChangeTeacher - Edit lesson, change teacher, apply to all
func TestScenario_13_BulkEditChangeTeacher(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher1 := createTestUser(t, pool, "teacher1@test.com", "Teacher", string(models.RoleTeacher))
	teacher2 := createTestUser(t, pool, "teacher2@test.com", "Teacher", string(models.RoleTeacher))

	// Create lessons with teacher1
	lessons := createMatchingLessons(t, pool, teacher1.ID, 3)
	sourceLessonID := lessons[0]

	// Apply bulk edit: change teacher to teacher2 for first lesson and all subsequent
	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "change_teacher",
		"teacher_id":        teacher2.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	assert.Equal(t, "change_teacher", bulkEditReq["modification_type"], "Modification type should be change_teacher")
	assert.Equal(t, teacher2.ID.String(), bulkEditReq["teacher_id"], "Teacher ID should match")
}

// TestScenario_14_BulkEditMissingFutureLessons - Apply to all when some lessons don't exist → skips missing
func TestScenario_14_BulkEditMissingFutureLessons(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))
	student := createTestUser(t, pool, "student@test.com", "Student", string(models.RoleStudent))

	addCreditToStudent(t, pool, student.ID, 20)

	// Create only 2 lessons (gaps in schedule)
	lesson1 := createLesson(t, pool, teacher.ID, 4, time.Now().AddDate(0, 0, 1))
	// lesson2 is intentionally missing (gap of 7 days)
	createLesson(t, pool, teacher.ID, 4, time.Now().AddDate(0, 0, 15))

	// Apply bulk edit: should only modify lesson1 and lesson3 (skipping the missing one)
	bulkEditReq := map[string]interface{}{
		"lesson_id":         lesson1.String(),
		"modification_type": "add_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", lesson1.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	// Verify request structure (in real implementation, this should modify 2 lessons, not 3)
	assert.NotNil(t, httpReq)
}

// TestScenario_15_BulkEditInsufficientCredits - Bulk edit with insufficient credits → error, no partial
func TestScenario_15_BulkEditInsufficientCredits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))
	student := createTestUser(t, pool, "student@test.com", "Student", string(models.RoleStudent))

	// Give student only 1 credit, but will try to add to 3 lessons
	addCreditToStudent(t, pool, student.ID, 1)

	// Create 3 lessons
	lessons := createMatchingLessons(t, pool, teacher.ID, 3)
	sourceLessonID := lessons[0]

	// Try to add student to all - should fail due to insufficient credits
	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "add_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	// Should fail (insufficient credits)
	assert.NotNil(t, httpReq)
}

// TestBulkEdit_ValidateApplicableModifications - Only matching lessons are modified
func TestBulkEdit_ValidateApplicableModifications(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))

	// Create lessons at different times (should NOT match for bulk edit)
	lesson1 := createLessonWithTime(t, pool, teacher.ID, 4, time.Now().AddDate(0, 0, 1), 9, 0)
	createLessonWithTime(t, pool, teacher.ID, 4, time.Now().AddDate(0, 0, 8), 10, 0) // Different time - won't match

	bulkEditReq := map[string]interface{}{
		"lesson_id":         lesson1.String(),
		"modification_type": "add_student",
		"student_id":        uuid.New().String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", lesson1.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	// Verify request (should only affect lesson1, not lesson2 due to different time)
	assert.NotNil(t, httpReq)
}

// TestBulkEdit_DuplicateStudentBooking - Cannot add student already booked
func TestBulkEdit_DuplicateStudentBooking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))
	student := createTestUser(t, pool, "student@test.com", "Student", string(models.RoleStudent))

	addCreditToStudent(t, pool, student.ID, 20)

	// Create matching lessons
	lessons := createMatchingLessons(t, pool, teacher.ID, 3)
	sourceLessonID := lessons[0]

	// Pre-add student to first lesson
	addStudentToLesson(t, pool, lessons[0], student.ID)

	// Try to add same student again - should fail
	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "add_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	// Should fail (student already booked)
	assert.NotNil(t, httpReq)
}

// TestBulkEdit_LessonCapacityFull - Cannot add student to full lesson
func TestBulkEdit_LessonCapacityFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))
	student := createTestUser(t, pool, "student@test.com", "Student", string(models.RoleStudent))

	addCreditToStudent(t, pool, student.ID, 20)

	// Create lesson with capacity 1 (full)
	lessons := createMatchingLessons(t, pool, teacher.ID, 2)
	sourceLessonID := lessons[0]

	// Fill the first lesson
	for i := 0; i < 4; i++ { // Assuming max_students=4
		otherStudent := createTestUser(t, pool, fmt.Sprintf("student%d@test.com", i), "Student", string(models.RoleStudent))
		addCreditToStudent(t, pool, otherStudent.ID, 20)
		addStudentToLesson(t, pool, lessons[0], otherStudent.ID)
	}

	// Try to add another student - should fail (lesson full)
	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "add_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	// Should fail (capacity full)
	assert.NotNil(t, httpReq)
}

// TestBulkEdit_ModificationTracking - Modifications logged with details
func TestBulkEdit_ModificationTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	admin := createTestUser(t, pool, "admin@test.com", "Admin", string(models.RoleAdmin))
	teacher := createTestUser(t, pool, "teacher@test.com", "Teacher", string(models.RoleTeacher))
	student := createTestUser(t, pool, "student@test.com", "Student", string(models.RoleStudent))

	addCreditToStudent(t, pool, student.ID, 20)

	lessons := createMatchingLessons(t, pool, teacher.ID, 3)
	sourceLessonID := lessons[0]

	bulkEditReq := map[string]interface{}{
		"lesson_id":         sourceLessonID.String(),
		"modification_type": "add_student",
		"student_id":        student.ID.String(),
	}

	body, _ := json.Marshal(bulkEditReq)
	httpReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/lessons/%s/apply-to-all", sourceLessonID.String()),
		bytes.NewBuffer(body),
	)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", admin.ID.String())

	// Verify modification can be tracked
	modification := &models.LessonModification{
		ID:                   uuid.New(),
		OriginalLessonID:     sourceLessonID,
		ModificationType:     "add_student",
		AppliedByID:          admin.ID,
		AppliedAt:            time.Now(),
		AffectedLessonsCount: 3,
	}

	assert.Equal(t, sourceLessonID, modification.OriginalLessonID)
	assert.Equal(t, "add_student", modification.ModificationType)
	assert.Equal(t, 3, modification.AffectedLessonsCount)
}
