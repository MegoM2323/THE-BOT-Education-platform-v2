package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/validator"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBookingRaceCondition tests that concurrent booking attempts for the last spot
// result in exactly one success and one failure with ErrLessonFull
func TestBookingRaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pool := database.GetTestPool(t)
	db := database.GetTestSqlxDB(t)

	// Create repositories
	bookingRepo := repository.NewBookingRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(db)
	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)

	// Create service
	bookingService := NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		cancelledBookingRepo,
		bookingValidator,
		nil,
		nil,
	)

	// Setup test data
	// Create teacher with unique email
	teacherID := uuid.New()
	teacherEmail := "teacher_race_" + teacherID.String()[:8] + "@test.com"
	teacherSQL := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := db.ExecContext(ctx, teacherSQL,
		teacherID, teacherEmail, "hash", "Teacher", "Race", "methodologist",
		time.Now(), time.Now(),
	)
	require.NoError(t, err, "Failed to create teacher")
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)

	// Create 2 students with unique emails
	student1ID := uuid.New()
	student2ID := uuid.New()

	students := []struct {
		id    uuid.UUID
		email string
	}{
		{student1ID, "student1_race_" + student1ID.String()[:8] + "@test.com"},
		{student2ID, "student2_race_" + student2ID.String()[:8] + "@test.com"},
	}

	for _, s := range students {
		_, err := db.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, s.id, s.email, "hash", "Student", "Race", "student", time.Now(), time.Now())
		require.NoError(t, err, "Failed to create student")
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", s.id)

		// Give students credits
		_, err = db.ExecContext(ctx, `
			INSERT INTO credits (id, user_id, balance, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id) DO UPDATE SET balance = 10
		`, uuid.New(), s.id, 10, time.Now(), time.Now())
		require.NoError(t, err, "Failed to create credits")
		defer db.ExecContext(ctx, "DELETE FROM credits WHERE user_id = $1", s.id)
	}

	// Create lesson with max_students = 1 (only ONE spot)
	lessonID := uuid.New()
	startTime := time.Now().Add(48 * time.Hour) // Future lesson
	endTime := startTime.Add(2 * time.Hour)

	lessonSQL := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = db.ExecContext(ctx, lessonSQL,
		lessonID, teacherID, startTime, endTime, 1, 0,
		time.Now(), time.Now(),
	)
	require.NoError(t, err, "Failed to create lesson")
	defer db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", lessonID)
	defer db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", lessonID)
	defer db.ExecContext(ctx, "DELETE FROM credit_transactions WHERE user_id IN ($1, $2)", student1ID, student2ID)

	// Run concurrent booking attempts
	var wg sync.WaitGroup
	results := make([]error, 2)

	// Create booking requests
	req1 := &models.CreateBookingRequest{
		StudentID: student1ID,
		LessonID:  lessonID,
		IsAdmin:   false,
		AdminID:   uuid.Nil,
	}

	req2 := &models.CreateBookingRequest{
		StudentID: student2ID,
		LessonID:  lessonID,
		IsAdmin:   false,
		AdminID:   uuid.Nil,
	}

	// Start concurrent goroutines
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := bookingService.CreateBooking(ctx, req1)
		results[0] = err
	}()

	go func() {
		defer wg.Done()
		_, err := bookingService.CreateBooking(ctx, req2)
		results[1] = err
	}()

	wg.Wait()

	// Assertions: Exactly ONE success, ONE failure
	successCount := 0
	failCount := 0
	var failedError error

	for i, err := range results {
		if err == nil {
			successCount++
			t.Logf("Booking %d: SUCCESS", i+1)
		} else {
			failCount++
			failedError = err
			t.Logf("Booking %d: FAILED with %v", i+1, err)
		}
	}

	assert.Equal(t, 1, successCount, "Expected exactly 1 successful booking")
	assert.Equal(t, 1, failCount, "Expected exactly 1 failed booking")
	assert.ErrorIs(t, failedError, repository.ErrLessonFull, "Failed booking should return ErrLessonFull")

	// Verify database state
	var currentStudents int
	err = db.GetContext(ctx, &currentStudents,
		"SELECT current_students FROM lessons WHERE id = $1", lessonID)
	require.NoError(t, err)
	assert.Equal(t, 1, currentStudents, "Lesson should have exactly 1 student")

	// Verify exactly 1 active booking exists
	var bookingCount int
	err = db.GetContext(ctx, &bookingCount,
		"SELECT COUNT(*) FROM bookings WHERE lesson_id = $1 AND status = 'active'", lessonID)
	require.NoError(t, err)
	assert.Equal(t, 1, bookingCount, "Should have exactly 1 active booking")
}

// TestBookingRaceConditionMultipleSpots tests race condition with 2 spots and 3 concurrent requests
func TestBookingRaceConditionMultipleSpots(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pool := database.GetTestPool(t)
	db := database.GetTestSqlxDB(t)

	bookingRepo := repository.NewBookingRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(db)
	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)

	bookingService := NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		cancelledBookingRepo,
		bookingValidator,
		nil,
		nil,
	)

	// Create teacher with unique email
	teacherID := uuid.New()
	teacherEmail := "teacher_multi_" + teacherID.String()[:8] + "@test.com"
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, teacherEmail, "hash", "Teacher Multi", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)

	// Create 3 students with unique emails
	students := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		studentID := uuid.New()
		students[i] = studentID

		email := "student_multi_" + studentID.String()[:8] + "@test.com"
		_, err := db.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, studentID, email, "hash", "Student", "Multi", "student", time.Now(), time.Now())
		require.NoError(t, err)
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", studentID)

		_, err = db.ExecContext(ctx, `
			INSERT INTO credits (id, user_id, balance, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id) DO UPDATE SET balance = 10
		`, uuid.New(), studentID, 10, time.Now(), time.Now())
		require.NoError(t, err)
		defer db.ExecContext(ctx, "DELETE FROM credits WHERE user_id = $1", studentID)
	}

	// Create lesson with max_students = 2 (TWO spots, THREE requests)
	lessonID := uuid.New()
	startTime := time.Now().Add(48 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	_, err = db.ExecContext(ctx, `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, lessonID, teacherID, startTime, endTime, 2, 0, time.Now(), time.Now())
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", lessonID)
	defer db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", lessonID)
	defer db.ExecContext(ctx, "DELETE FROM credit_transactions WHERE user_id = ANY($1)", students)

	// Run 3 concurrent bookings
	var wg sync.WaitGroup
	results := make([]error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		index := i
		go func(studentID uuid.UUID, idx int) {
			defer wg.Done()
			req := &models.CreateBookingRequest{
				StudentID: studentID,
				LessonID:  lessonID,
				IsAdmin:   false,
				AdminID:   uuid.Nil,
			}
			_, err := bookingService.CreateBooking(ctx, req)
			results[idx] = err
		}(students[index], index)
	}

	wg.Wait()

	// Count results
	successCount := 0
	failCount := 0

	for i, err := range results {
		if err == nil {
			successCount++
			t.Logf("Booking %d: SUCCESS", i+1)
		} else {
			failCount++
			t.Logf("Booking %d: FAILED with %v", i+1, err)
		}
	}

	// Assertions: 2 success, 1 failure
	assert.Equal(t, 2, successCount, "Expected exactly 2 successful bookings")
	assert.Equal(t, 1, failCount, "Expected exactly 1 failed booking")

	// Verify database state
	var currentStudents int
	err = db.GetContext(ctx, &currentStudents,
		"SELECT current_students FROM lessons WHERE id = $1", lessonID)
	require.NoError(t, err)
	assert.Equal(t, 2, currentStudents, "Lesson should have exactly 2 students")

	var bookingCount int
	err = db.GetContext(ctx, &bookingCount,
		"SELECT COUNT(*) FROM bookings WHERE lesson_id = $1 AND status = 'active'", lessonID)
	require.NoError(t, err)
	assert.Equal(t, 2, bookingCount, "Should have exactly 2 active bookings")
}

// TestBookingNoRaceWithAvailableSpots tests that when there are enough spots,
// all concurrent bookings succeed
func TestBookingNoRaceWithAvailableSpots(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pool := database.GetTestPool(t)
	db := database.GetTestSqlxDB(t)

	bookingRepo := repository.NewBookingRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(db)
	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)

	bookingService := NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		cancelledBookingRepo,
		bookingValidator,
		nil,
		nil,
	)

	// Create teacher with unique email
	teacherID := uuid.New()
	teacherEmail := "teacher_available_" + teacherID.String()[:8] + "@test.com"
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, teacherEmail, "hash", "Teacher Available", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)

	// Create 3 students with unique emails
	students := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		studentID := uuid.New()
		students[i] = studentID

		email := "student_avail_" + studentID.String()[:8] + "@test.com"
		_, err := db.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, studentID, email, "hash", "Student", "Avail", "student", time.Now(), time.Now())
		require.NoError(t, err)
		defer db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", studentID)

		_, err = db.ExecContext(ctx, `
			INSERT INTO credits (id, user_id, balance, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id) DO UPDATE SET balance = 10
		`, uuid.New(), studentID, 10, time.Now(), time.Now())
		require.NoError(t, err)
		defer db.ExecContext(ctx, "DELETE FROM credits WHERE user_id = $1", studentID)
	}

	// Create lesson with max_students = 10 (plenty of spots)
	lessonID := uuid.New()
	startTime := time.Now().Add(48 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	_, err = db.ExecContext(ctx, `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, lessonID, teacherID, startTime, endTime, 10, 0, time.Now(), time.Now())
	require.NoError(t, err)
	defer db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", lessonID)
	defer db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", lessonID)
	defer db.ExecContext(ctx, "DELETE FROM credit_transactions WHERE user_id = ANY($1)", students)

	// Run 3 concurrent bookings
	var wg sync.WaitGroup
	results := make([]error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		index := i
		go func(studentID uuid.UUID, idx int) {
			defer wg.Done()
			req := &models.CreateBookingRequest{
				StudentID: studentID,
				LessonID:  lessonID,
				IsAdmin:   false,
				AdminID:   uuid.Nil,
			}
			_, err := bookingService.CreateBooking(ctx, req)
			results[idx] = err
		}(students[index], index)
	}

	wg.Wait()

	// All should succeed
	for i, err := range results {
		assert.NoError(t, err, "Booking %d should succeed", i+1)
	}

	// Verify database state
	var currentStudents int
	err = db.GetContext(ctx, &currentStudents,
		"SELECT current_students FROM lessons WHERE id = $1", lessonID)
	require.NoError(t, err)
	assert.Equal(t, 3, currentStudents, "Lesson should have exactly 3 students")
}
