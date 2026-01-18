package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BulkEditService handles bulk editing operations on lessons and subsequent occurrences
type BulkEditService struct {
	pool                   *pgxpool.Pool
	lessonRepo             *repository.LessonRepository
	lessonModificationRepo *repository.LessonModificationRepository
	userRepo               repository.UserRepository
	creditRepo             *repository.CreditRepository
}

// NewBulkEditService creates a new BulkEditService
func NewBulkEditService(
	pool *pgxpool.Pool,
	lessonRepo *repository.LessonRepository,
	lessonModificationRepo *repository.LessonModificationRepository,
	userRepo repository.UserRepository,
	creditRepo *repository.CreditRepository,
) *BulkEditService {
	return &BulkEditService{
		pool:                   pool,
		lessonRepo:             lessonRepo,
		lessonModificationRepo: lessonModificationRepo,
		userRepo:               userRepo,
		creditRepo:             creditRepo,
	}
}

// ApplyToAllSubsequent applies a modification to a lesson and all subsequent matching lessons
func (s *BulkEditService) ApplyToAllSubsequent(ctx context.Context, adminID uuid.UUID, req *models.ApplyToAllSubsequentRequest) (*models.LessonModification, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Route to appropriate handler based on modification type
	switch req.ModificationType {
	case "add_student":
		return s.AddStudentToAllSubsequent(ctx, adminID, req.LessonID, *req.StudentID)
	case "remove_student":
		return s.RemoveStudentFromAllSubsequent(ctx, adminID, req.LessonID, *req.StudentID)
	case "change_teacher":
		return s.ChangeTeacherForAllSubsequent(ctx, adminID, req.LessonID, *req.TeacherID)
	case "change_time":
		return s.ChangeTimeForAllSubsequent(ctx, adminID, req.LessonID, *req.NewStartTime)
	case "change_capacity":
		return s.ChangeCapacityForAllSubsequent(ctx, adminID, req.LessonID, *req.NewMaxStudents)
	default:
		return nil, fmt.Errorf("unsupported modification type: %s", req.ModificationType)
	}
}

// FindSubsequentMatchingLessons finds all lessons matching the source lesson pattern after a given date
func (s *BulkEditService) FindSubsequentMatchingLessons(ctx context.Context, sourceLessonID uuid.UUID) ([]*models.Lesson, error) {
	// Load source lesson
	sourceLesson, err := s.lessonRepo.GetByID(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source lesson: %w", err)
	}

	// Extract time pattern from source lesson
	dayOfWeek := int(sourceLesson.StartTime.Weekday()) // Go time.Weekday: 0=Sunday, 1=Monday, ..., 6=Saturday
	sourceHour := sourceLesson.StartTime.Hour()
	sourceMinute := sourceLesson.StartTime.Minute()

	// Find all matching future lessons
	lessons, err := s.lessonRepo.GetLessonsByTimePattern(
		ctx,
		sourceLesson.TeacherID,
		dayOfWeek,
		sourceHour,
		sourceMinute,
		sourceLesson.StartTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching lessons: %w", err)
	}

	return lessons, nil
}

// AddStudentToAllSubsequent adds a student to the source lesson and all subsequent matching lessons
func (s *BulkEditService) AddStudentToAllSubsequent(ctx context.Context, adminID uuid.UUID, sourceLessonID uuid.UUID, studentID uuid.UUID) (*models.LessonModification, error) {
	// Validate student exists and has student role
	student, err := s.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load student: %w", err)
	}
	if !student.IsStudent() {
		return nil, fmt.Errorf("user %s is not a student (role: %s)", studentID, student.Role)
	}

	// Load source lesson
	sourceLesson, err := s.lessonRepo.GetByID(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source lesson: %w", err)
	}

	// Find all matching future lessons
	futureMatches, err := s.FindSubsequentMatchingLessons(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching lessons: %w", err)
	}

	// Start SERIALIZABLE transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			// Log rollback error
		}
	}()

	// Set isolation level to SERIALIZABLE
	_, err = tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	// Pre-check: Validate all lessons and student not already booked
	allLessons := append([]*models.Lesson{sourceLesson}, futureMatches...)
	for _, lesson := range allLessons {
		if lesson.CurrentStudents >= lesson.MaxStudents {
			return nil, fmt.Errorf("lesson %s (at %s) is full, cannot add student", lesson.ID, lesson.StartTime.Format("2006-01-02 15:04"))
		}

		isBooked, err := s.lessonRepo.IsStudentBookedForLessonTx(ctx, tx, lesson.ID, studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing booking: %w", err)
		}
		if isBooked {
			return nil, fmt.Errorf("student is already booked for lesson at %s", lesson.StartTime.Format("2006-01-02 15:04"))
		}
	}

	// Check student has enough credits for all lessons
	creditsNeeded := len(allLessons)
	credit, err := s.creditRepo.GetBalanceForUpdate(ctx, tx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student credit balance: %w", err)
	}
	if credit.Balance < creditsNeeded {
		return nil, fmt.Errorf("insufficient credits: student has %d, needs %d for %d lessons", credit.Balance, creditsNeeded, len(allLessons))
	}

	// Deduct credits from student
	newBalance := credit.Balance - creditsNeeded
	if err := s.creditRepo.UpdateBalance(ctx, tx, studentID, newBalance); err != nil {
		return nil, fmt.Errorf("failed to deduct credits: %w", err)
	}

	// Add student to all lessons (source + future matches)
	now := time.Now()
	for _, lesson := range allLessons {
		booking := &models.Booking{
			ID:        uuid.New(),
			StudentID: studentID,
			LessonID:  lesson.ID,
			Status:    models.BookingStatusActive,
			BookedAt:  now,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.lessonRepo.CreateBookingTx(ctx, tx, booking); err != nil {
			return nil, fmt.Errorf("failed to add student to lesson %s: %w", lesson.ID, err)
		}

		// Increment student count
		if err := s.lessonRepo.IncrementStudents(ctx, tx, lesson.ID); err != nil {
			return nil, fmt.Errorf("failed to increment student count for lesson %s: %w", lesson.ID, err)
		}

		// Record credit transaction for each booking
		creditTx := &models.CreditTransaction{
			UserID:        studentID,
			Amount:        1,
			OperationType: models.OperationTypeDeduct,
			Reason:        fmt.Sprintf("Admin bulk booking for lesson at %s", lesson.StartTime.Format("2006-01-02 15:04")),
			BookingID:     uuid.NullUUID{UUID: booking.ID, Valid: true},
			PerformedBy:   uuid.NullUUID{UUID: adminID, Valid: true},
			BalanceBefore: credit.Balance - (creditsNeeded - len(allLessons)) + 1, // Calculate balance before this specific deduction
			BalanceAfter:  newBalance,
		}
		// Simplified: use the final balance since all credits are deducted atomically
		creditTx.BalanceBefore = credit.Balance
		creditTx.BalanceAfter = newBalance
		if err := s.creditRepo.CreateTransaction(ctx, tx, creditTx); err != nil {
			return nil, fmt.Errorf("failed to record credit transaction: %w", err)
		}
	}

	// Create modification record for audit trail
	changesJSON, err := json.Marshal(map[string]interface{}{
		"student_id":   studentID.String(),
		"student_name": student.FullName,
		"action":       "add",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes: %w", err)
	}

	modification := &models.LessonModification{
		OriginalLessonID:     sourceLessonID,
		ModificationType:     "add_student",
		AppliedByID:          adminID,
		AffectedLessonsCount: len(allLessons),
		ChangesJSON:          json.RawMessage(changesJSON),
	}

	if err := s.lessonModificationRepo.LogModificationTx(ctx, tx, modification); err != nil {
		return nil, fmt.Errorf("failed to log modification: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return modification, nil
}

// RemoveStudentFromAllSubsequent removes a student from the source lesson and all subsequent matching lessons
func (s *BulkEditService) RemoveStudentFromAllSubsequent(ctx context.Context, adminID uuid.UUID, sourceLessonID uuid.UUID, studentID uuid.UUID) (*models.LessonModification, error) {
	// Validate student exists
	student, err := s.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load student: %w", err)
	}

	// Load source lesson
	sourceLesson, err := s.lessonRepo.GetByID(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source lesson: %w", err)
	}

	// Find all matching future lessons
	futureMatches, err := s.FindSubsequentMatchingLessons(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching lessons: %w", err)
	}

	// Start SERIALIZABLE transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			// Log rollback error
		}
	}()

	_, err = tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	// Remove student from all lessons (source + future matches)
	allLessons := append([]*models.Lesson{sourceLesson}, futureMatches...)
	removedCount := 0

	// Get student's credit balance for refunds
	credit, err := s.creditRepo.GetBalanceForUpdate(ctx, tx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student credit balance: %w", err)
	}

	for _, lesson := range allLessons {
		// Check if student is booked
		isBooked, err := s.lessonRepo.IsStudentBookedForLessonTx(ctx, tx, lesson.ID, studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to check booking: %w", err)
		}

		if isBooked {
			// Cancel the booking
			if err := s.lessonRepo.CancelBookingTx(ctx, tx, lesson.ID, studentID); err != nil {
				return nil, fmt.Errorf("failed to cancel booking for lesson %s: %w", lesson.ID, err)
			}

			// Decrement student count
			if err := s.lessonRepo.DecrementStudents(ctx, tx, lesson.ID); err != nil {
				return nil, fmt.Errorf("failed to decrement student count for lesson %s: %w", lesson.ID, err)
			}

			removedCount++
		}
	}

	if removedCount == 0 {
		return nil, fmt.Errorf("student %s is not booked for any of the matching lessons", studentID)
	}

	// Refund credits for cancelled bookings
	newBalance := credit.Balance + removedCount
	if err := s.creditRepo.UpdateBalance(ctx, tx, studentID, newBalance); err != nil {
		return nil, fmt.Errorf("failed to refund credits: %w", err)
	}

	// Record credit transaction for refund
	creditTx := &models.CreditTransaction{
		UserID:        studentID,
		Amount:        removedCount,
		OperationType: models.OperationTypeRefund,
		Reason:        fmt.Sprintf("Admin bulk removal: %d lesson(s) cancelled", removedCount),
		PerformedBy:   uuid.NullUUID{UUID: adminID, Valid: true},
		BalanceBefore: credit.Balance,
		BalanceAfter:  newBalance,
	}
	if err := s.creditRepo.CreateTransaction(ctx, tx, creditTx); err != nil {
		return nil, fmt.Errorf("failed to record credit refund transaction: %w", err)
	}

	// Create modification record
	changesJSON, err := json.Marshal(map[string]interface{}{
		"student_id":    studentID.String(),
		"student_name":  student.FullName,
		"action":        "remove",
		"removed_count": removedCount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes: %w", err)
	}

	modification := &models.LessonModification{
		OriginalLessonID:     sourceLessonID,
		ModificationType:     "remove_student",
		AppliedByID:          adminID,
		AffectedLessonsCount: removedCount,
		ChangesJSON:          json.RawMessage(changesJSON),
	}

	if err := s.lessonModificationRepo.LogModificationTx(ctx, tx, modification); err != nil {
		return nil, fmt.Errorf("failed to log modification: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return modification, nil
}

// ChangeTeacherForAllSubsequent changes the teacher for the source lesson and all subsequent matching lessons
func (s *BulkEditService) ChangeTeacherForAllSubsequent(ctx context.Context, adminID uuid.UUID, sourceLessonID uuid.UUID, newTeacherID uuid.UUID) (*models.LessonModification, error) {
	// Validate new teacher exists and has teacher or admin role
	newTeacher, err := s.userRepo.GetByID(ctx, newTeacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to load new teacher: %w", err)
	}
	if !newTeacher.CanBeAssignedAsTeacher() {
		return nil, fmt.Errorf("user %s cannot be assigned as teacher (role: %s)", newTeacherID, newTeacher.Role)
	}

	// Load source lesson
	sourceLesson, err := s.lessonRepo.GetByID(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source lesson: %w", err)
	}

	oldTeacherID := sourceLesson.TeacherID

	// Find all matching future lessons
	futureMatches, err := s.FindSubsequentMatchingLessons(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching lessons: %w", err)
	}

	// Start SERIALIZABLE transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			// Log rollback error
		}
	}()

	_, err = tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	// Update teacher for all lessons (source + future matches)
	allLessons := append([]*models.Lesson{sourceLesson}, futureMatches...)
	for _, lesson := range allLessons {
		if err := s.lessonRepo.UpdateTeacherTx(ctx, tx, lesson.ID, newTeacherID); err != nil {
			return nil, fmt.Errorf("failed to update teacher for lesson %s: %w", lesson.ID, err)
		}
	}

	// Create modification record
	changesJSON, err := json.Marshal(map[string]interface{}{
		"old_teacher_id":   oldTeacherID.String(),
		"new_teacher_id":   newTeacherID.String(),
		"new_teacher_name": newTeacher.FullName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes: %w", err)
	}

	modification := &models.LessonModification{
		OriginalLessonID:     sourceLessonID,
		ModificationType:     "change_teacher",
		AppliedByID:          adminID,
		AffectedLessonsCount: len(allLessons),
		ChangesJSON:          json.RawMessage(changesJSON),
	}

	if err := s.lessonModificationRepo.LogModificationTx(ctx, tx, modification); err != nil {
		return nil, fmt.Errorf("failed to log modification: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return modification, nil
}

// ChangeTimeForAllSubsequent changes the time for the source lesson and all subsequent matching lessons
func (s *BulkEditService) ChangeTimeForAllSubsequent(ctx context.Context, adminID uuid.UUID, sourceLessonID uuid.UUID, newStartTimeStr string) (*models.LessonModification, error) {
	// Parse new start time
	newStartTime, err := time.Parse(time.RFC3339, newStartTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid time format (expected RFC3339): %w", err)
	}

	// Load source lesson
	sourceLesson, err := s.lessonRepo.GetByID(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source lesson: %w", err)
	}

	oldStartTime := sourceLesson.StartTime

	// Find all matching future lessons
	futureMatches, err := s.FindSubsequentMatchingLessons(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching lessons: %w", err)
	}

	// Start SERIALIZABLE transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			// Log rollback error
		}
	}()

	_, err = tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	// Update time for all lessons (source + future matches)
	// Note: This changes the time but preserves the day, so it may create scheduling conflicts
	// A production implementation would need conflict checking
	allLessons := append([]*models.Lesson{sourceLesson}, futureMatches...)
	for _, lesson := range allLessons {
		// Calculate new time preserving the original date
		lessonDate := lesson.StartTime.Truncate(24 * time.Hour)
		adjustedStartTime := lessonDate.Add(time.Duration(newStartTime.Hour())*time.Hour + time.Duration(newStartTime.Minute())*time.Minute)
		adjustedEndTime := adjustedStartTime.Add(2 * time.Hour)

		if err := s.lessonRepo.UpdateTimeTx(ctx, tx, lesson.ID, adjustedStartTime, adjustedEndTime); err != nil {
			return nil, fmt.Errorf("failed to update time for lesson %s: %w", lesson.ID, err)
		}
	}

	// Create modification record
	changesJSON, err := json.Marshal(map[string]interface{}{
		"old_start_time": oldStartTime.Format(time.RFC3339),
		"new_start_time": newStartTime.Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes: %w", err)
	}

	modification := &models.LessonModification{
		OriginalLessonID:     sourceLessonID,
		ModificationType:     "change_time",
		AppliedByID:          adminID,
		AffectedLessonsCount: len(allLessons),
		ChangesJSON:          json.RawMessage(changesJSON),
	}

	if err := s.lessonModificationRepo.LogModificationTx(ctx, tx, modification); err != nil {
		return nil, fmt.Errorf("failed to log modification: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return modification, nil
}

// ChangeCapacityForAllSubsequent changes max_students for the source lesson and all subsequent matching lessons
func (s *BulkEditService) ChangeCapacityForAllSubsequent(ctx context.Context, adminID uuid.UUID, sourceLessonID uuid.UUID, newMaxStudents int) (*models.LessonModification, error) {
	if newMaxStudents < 1 {
		return nil, fmt.Errorf("new_max_students must be >= 1")
	}

	// Load source lesson
	sourceLesson, err := s.lessonRepo.GetByID(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to load source lesson: %w", err)
	}

	oldMaxStudents := sourceLesson.MaxStudents

	// Find all matching future lessons
	futureMatches, err := s.FindSubsequentMatchingLessons(ctx, sourceLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching lessons: %w", err)
	}

	// Start SERIALIZABLE transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			// Log rollback error
		}
	}()

	_, err = tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		return nil, fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	// Pre-check: ensure new capacity doesn't conflict with current_students
	allLessons := append([]*models.Lesson{sourceLesson}, futureMatches...)
	for _, lesson := range allLessons {
		if lesson.CurrentStudents > newMaxStudents {
			return nil, fmt.Errorf("cannot set max_students to %d for lesson %s: already has %d students enrolled", newMaxStudents, lesson.ID, lesson.CurrentStudents)
		}
	}

	// Update max_students for all lessons
	for _, lesson := range allLessons {
		if err := s.lessonRepo.UpdateMaxStudentsTx(ctx, tx, lesson.ID, newMaxStudents); err != nil {
			return nil, fmt.Errorf("failed to update capacity for lesson %s: %w", lesson.ID, err)
		}
	}

	// Create modification record
	changesJSON, err := json.Marshal(map[string]interface{}{
		"old_max_students": oldMaxStudents,
		"new_max_students": newMaxStudents,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes: %w", err)
	}

	modification := &models.LessonModification{
		OriginalLessonID:     sourceLessonID,
		ModificationType:     "change_capacity",
		AppliedByID:          adminID,
		AffectedLessonsCount: len(allLessons),
		ChangesJSON:          json.RawMessage(changesJSON),
	}

	if err := s.lessonModificationRepo.LogModificationTx(ctx, tx, modification); err != nil {
		return nil, fmt.Errorf("failed to log modification: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return modification, nil
}

// ValidateModificationApplicability validates if a modification can be applied to target lessons
func (s *BulkEditService) ValidateModificationApplicability(ctx context.Context, sourceLessonID uuid.UUID, modificationType string, targetLessons []*models.Lesson) error {
	switch modificationType {
	case "add_student":
		// Validate all lessons have capacity
		for _, lesson := range targetLessons {
			if lesson.CurrentStudents >= lesson.MaxStudents {
				return fmt.Errorf("lesson %s is full", lesson.ID)
			}
		}
	case "remove_student":
		// No specific validation needed - removal will skip lessons where student isn't booked
		return nil
	case "change_teacher":
		// Teacher validation happens in the main method
		return nil
	case "change_time":
		// Time conflict checking would go here in production
		return nil
	case "change_capacity":
		// Capacity validation happens in the main method
		return nil
	default:
		return fmt.Errorf("unsupported modification type: %s", modificationType)
	}
	return nil
}
