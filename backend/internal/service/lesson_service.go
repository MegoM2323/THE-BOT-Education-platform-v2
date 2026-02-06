package service

import (
	"context"
	"fmt"
	"time"

	"tutoring-platform/internal/config"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/validator"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// BookingCreator interface for creating bookings (to avoid circular dependency)
type BookingCreator interface {
	CreateBooking(ctx context.Context, req *models.CreateBookingRequest) (*models.Booking, error)
}

// LessonService обрабатывает бизнес-логику для уроков
type LessonService struct {
	lessonRepo      *repository.LessonRepository
	userRepo        repository.UserRepository
	lessonValidator *validator.LessonValidator
	bookingCreator  BookingCreator
	telegramService *TelegramService
}

// NewLessonService создает новый LessonService
func NewLessonService(lessonRepo *repository.LessonRepository, userRepo repository.UserRepository) *LessonService {
	return &LessonService{
		lessonRepo:      lessonRepo,
		userRepo:        userRepo,
		lessonValidator: validator.NewLessonValidator(),
	}
}

// SetBookingCreator sets the booking creator for enrolling students on lesson creation
func (s *LessonService) SetBookingCreator(bc BookingCreator) {
	s.bookingCreator = bc
}

// SetTelegramService sets the telegram service for sending notifications
func (s *LessonService) SetTelegramService(ts *TelegramService) {
	s.telegramService = ts
}

// CreateLesson создает новый урок
func (s *LessonService) CreateLesson(ctx context.Context, req *models.CreateLessonRequest) (*models.Lesson, error) {
	// Apply defaults BEFORE validation
	req.ApplyDefaults()

	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate color and subject
	if err := s.lessonValidator.ValidateCreateLessonRequest(req); err != nil {
		return nil, err
	}

	// Проверяем существование преподавателя
	teacher, err := s.userRepo.GetByID(ctx, req.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("teacher not found: %w", err)
	}

	// Проверяем что пользователь может быть назначен преподавателем
	// Допустимые роли: teacher, admin, methodologist
	if !teacher.CanBeAssignedAsTeacher() {
		return nil, fmt.Errorf("user cannot be assigned as teacher (role: %s): %w", teacher.Role, models.ErrInvalidTeacherID)
	}

	// Создаем урок
	lesson := &models.Lesson{
		TeacherID:   req.TeacherID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		MaxStudents: req.MaxStudents,
		CreditsCost: req.CreditsCost,
		Color:       req.Color,
	}

	// Set subject if provided
	if req.Subject != nil && *req.Subject != "" {
		lesson.Subject.String = *req.Subject
		lesson.Subject.Valid = true
	}

	// Set homework_text if provided
	if req.HomeworkText != nil && *req.HomeworkText != "" {
		lesson.HomeworkText.String = *req.HomeworkText
		lesson.HomeworkText.Valid = true
	}

	// Set link if provided
	if req.Link != nil && *req.Link != "" {
		lesson.Link.String = *req.Link
		lesson.Link.Valid = true
	}

	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	// Enroll students if provided
	if len(req.StudentIDs) > 0 && s.bookingCreator != nil {
		// Validate: number of students must not exceed max_students
		if len(req.StudentIDs) > req.MaxStudents {
			return nil, fmt.Errorf("number of students (%d) exceeds max_students (%d)", len(req.StudentIDs), req.MaxStudents)
		}

		// Validate: check for duplicate student IDs
		seenStudents := make(map[uuid.UUID]bool)
		for _, studentID := range req.StudentIDs {
			if seenStudents[studentID] {
				return nil, fmt.Errorf("duplicate student_id: %s", studentID)
			}
			seenStudents[studentID] = true
		}

		// Validate: verify all students exist and are students
		for _, studentID := range req.StudentIDs {
			student, err := s.userRepo.GetByID(ctx, studentID)
			if err != nil {
				return nil, fmt.Errorf("student not found: %s", studentID)
			}
			if !student.IsStudent() {
				return nil, fmt.Errorf("user %s is not a student", studentID)
			}
		}

		// Create bookings for each student
		for _, studentID := range req.StudentIDs {
			bookingReq := &models.CreateBookingRequest{
				StudentID: studentID,
				LessonID:  lesson.ID,
				IsAdmin:   true, // Admin is creating the lesson with students
			}

			_, err := s.bookingCreator.CreateBooking(ctx, bookingReq)
			if err != nil {
				return nil, fmt.Errorf("failed to enroll student %s: %w", studentID, err)
			}

			log.Info().
				Str("lesson_id", lesson.ID.String()).
				Str("student_id", studentID.String()).
				Msg("Student enrolled on lesson creation")
		}

		// Send Telegram notifications to enrolled students (non-blocking)
		if s.telegramService != nil && len(req.StudentIDs) > 0 {
			studentNames := make([]string, 0, len(req.StudentIDs))
			for _, studentID := range req.StudentIDs {
				student, err := s.userRepo.GetByID(ctx, studentID)
				if err == nil {
					studentNames = append(studentNames, student.GetFullName())
				} else {
					studentNames = append(studentNames, "Студент")
				}
			}

			go func(lessonCtx context.Context) {
				notifCtx, cancel := context.WithTimeout(lessonCtx, 10*time.Second)
				defer cancel()

				for i, studentID := range req.StudentIDs {
					// Проверяем отмену контекста перед отправкой каждого уведомления
					select {
					case <-notifCtx.Done():
						log.Warn().
							Str("lesson_id", lesson.ID.String()).
							Msg("Lesson booking notification goroutine cancelled")
						return
					default:
					}

					studentName := studentNames[i]
					if err := s.telegramService.NotifyLessonBooking(notifCtx, lesson, studentName, []uuid.UUID{studentID}); err != nil {
						log.Warn().
							Str("lesson_id", lesson.ID.String()).
							Str("student_id", studentID.String()).
							Err(err).
							Msg("Failed to send booking notification to user")
					}
				}
			}(ctx)
		}
	}

	return lesson, nil
}

// GetLesson получает урок по ID
func (s *LessonService) GetLesson(ctx context.Context, lessonID uuid.UUID) (*models.Lesson, error) {
	return s.lessonRepo.GetByID(ctx, lessonID)
}

// GetLessonWithTeacher получает урок с информацией о преподавателе
func (s *LessonService) GetLessonWithTeacher(ctx context.Context, lessonID uuid.UUID) (*models.LessonWithTeacher, error) {
	return s.lessonRepo.GetWithTeacher(ctx, lessonID)
}

// UpdateLesson обновляет урок
func (s *LessonService) UpdateLesson(ctx context.Context, lessonID uuid.UUID, req *models.UpdateLessonRequest) (*models.Lesson, error) {
	// Проверяем валидность запроса
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate color and subject
	if err := s.lessonValidator.ValidateUpdateLessonRequest(req); err != nil {
		return nil, err
	}

	// Get current lesson to validate type changes
	currentLesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("lesson not found: %w", err)
	}

	// Save old start time for reschedule notification
	oldStartTime := currentLesson.StartTime

	// Validate lesson type change: cannot change to individual if multiple students enrolled
	if req.LessonType != nil && *req.LessonType == models.LessonTypeIndividual {
		if currentLesson.CurrentStudents > 1 {
			return nil, models.ErrCannotChangeToIndividual
		}
	}

	updates := make(map[string]interface{})

	if req.TeacherID != nil {
		// Проверяем существование нового преподавателя
		teacher, err := s.userRepo.GetByID(ctx, *req.TeacherID)
		if err != nil {
			return nil, fmt.Errorf("teacher not found: %w", err)
		}

		// Проверяем что пользователь может быть назначен преподавателем
		// Допустимые роли: teacher, admin, methodologist
		if !teacher.CanBeAssignedAsTeacher() {
			return nil, fmt.Errorf("user cannot be assigned as teacher (role: %s): %w", teacher.Role, models.ErrInvalidTeacherID)
		}

		updates["teacher_id"] = *req.TeacherID
	}

	// Handle automatic end_time calculation when start_time is updated
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime

		// Auto-calculate end_time if not explicitly provided
		if req.EndTime == nil {
			autoEndTime := req.StartTime.Add(2 * time.Hour)
			updates["end_time"] = autoEndTime
		}
	}

	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}

	if req.MaxStudents != nil {
		updates["max_students"] = *req.MaxStudents
	}

	if req.CreditsCost != nil {
		updates["credits_cost"] = *req.CreditsCost
	}

	if req.Color != nil {
		updates["color"] = *req.Color
	}

	if req.Subject != nil {
		if *req.Subject == "" {
			// Empty string means clear the subject
			updates["subject"] = nil
		} else {
			updates["subject"] = *req.Subject
		}
	}

	if req.HomeworkText != nil {
		if *req.HomeworkText == "" {
			// Empty string means clear the homework_text
			updates["homework_text"] = nil
		} else {
			updates["homework_text"] = *req.HomeworkText
		}
	}

	if req.Link != nil {
		if *req.Link == "" {
			// Empty string means clear the link
			updates["link"] = nil
		} else {
			updates["link"] = *req.Link
		}
	}

	if err := s.lessonRepo.Update(ctx, lessonID, updates); err != nil {
		return nil, fmt.Errorf("failed to update lesson: %w", err)
	}

	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated lesson: %w", err)
	}

	// Check if start time changed and send reschedule notifications (non-blocking)
	if s.telegramService != nil && !lesson.StartTime.Equal(oldStartTime) {
		bookings, err := s.lessonRepo.GetLessonBookings(ctx, lessonID)
		if err == nil && len(bookings) > 0 {
			studentIDs := make([]uuid.UUID, 0, len(bookings))
			for _, b := range bookings {
				studentIDs = append(studentIDs, b.StudentID)
			}

			go func(lessonCtx context.Context) {
				notifCtx, cancel := context.WithTimeout(lessonCtx, 10*time.Second)
				defer cancel()

				if err := s.telegramService.NotifyLessonReschedule(notifCtx, lesson, oldStartTime, lesson.StartTime, studentIDs); err != nil {
					log.Warn().
						Str("lesson_id", lessonID.String()).
						Err(err).
						Msg("Failed to send reschedule notification for lesson")
				}
			}(ctx)
		}
	}

	return lesson, nil
}

// DeleteLesson выполняет мягкое удаление урока
func (s *LessonService) DeleteLesson(ctx context.Context, lessonID uuid.UUID) error {
	return s.lessonRepo.Delete(ctx, lessonID)
}

// ListLessons получает список уроков с фильтрами
func (s *LessonService) ListLessons(ctx context.Context, filter *models.ListLessonsFilter) ([]*models.LessonWithTeacher, error) {
	return s.lessonRepo.List(ctx, filter)
}

// GetVisibleLessons получает уроки, видимые для пользователя, с учетом роли
func (s *LessonService) GetVisibleLessons(ctx context.Context, userID uuid.UUID, userRole string, filter *models.ListLessonsFilter) ([]*models.LessonWithTeacher, error) {
	return s.lessonRepo.GetVisibleLessons(ctx, userID, userRole, filter)
}

// GetLessonsByIDs получает несколько уроков по их ID в одном запросе (решение N+1 проблемы)
func (s *LessonService) GetLessonsByIDs(ctx context.Context, lessonIDs []uuid.UUID) ([]*models.LessonWithTeacher, error) {
	if len(lessonIDs) == 0 {
		return []*models.LessonWithTeacher{}, nil
	}
	return s.lessonRepo.GetByIDs(ctx, lessonIDs)
}

// SyncStudentCounts синхронизирует счетчики студентов для всех занятий на основе активных бронирований
// Этот метод вызывается администратором после изменения расписания или удаления студентов
func (s *LessonService) SyncStudentCounts(ctx context.Context) error {
	return s.lessonRepo.SyncStudentCounts(ctx)
}

// GetLessonBookings получает информацию о студентах, записанных на занятие
func (s *LessonService) GetLessonBookings(ctx context.Context, lessonID uuid.UUID) ([]models.BookingInfo, error) {
	return s.lessonRepo.GetLessonBookings(ctx, lessonID)
}

// GetLessonBookingsForLessons получает информацию о студентах для ВСЕХ занятий в одном batch запросе
func (s *LessonService) GetLessonBookingsForLessons(ctx context.Context, lessonIDs []uuid.UUID) (map[uuid.UUID][]models.BookingInfo, error) {
	return s.lessonRepo.GetLessonBookingsForLessons(ctx, lessonIDs)
}

// CreateRecurringLessons создает повторяющиеся занятия на несколько недель
func (s *LessonService) CreateRecurringLessons(ctx context.Context, req *models.CreateLessonRequest) ([]*models.Lesson, error) {
	if !req.IsRecurring {
		return s.createSingleLesson(ctx, req)
	}

	weeks := config.DefaultRecurringWeeks
	if req.RecurringWeeks != nil {
		if *req.RecurringWeeks > config.MaxRecurringWeeks {
			return nil, fmt.Errorf("max recurring weeks is %d", config.MaxRecurringWeeks)
		}
		if *req.RecurringWeeks <= 0 {
			return nil, fmt.Errorf("recurring weeks must be positive")
		}
		weeks = *req.RecurringWeeks
	}

	groupID := uuid.New()
	var lessons []*models.Lesson

	for i := 0; i < weeks; i++ {
		weekReq := *req
		weekReq.StartTime = req.StartTime.AddDate(0, 0, i*7)
		weekReq.EndTime = req.EndTime.AddDate(0, 0, i*7)
		weekReq.IsRecurring = true
		weekReq.RecurringWeeks = nil
		weekReq.RecurringEndDate = req.RecurringEndDate

		lesson, err := s.createLessonWithGroup(ctx, &weekReq, &groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to create week %d lesson: %w", i+1, err)
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (s *LessonService) createSingleLesson(ctx context.Context, req *models.CreateLessonRequest) ([]*models.Lesson, error) {
	lesson, err := s.CreateLesson(ctx, req)
	if err != nil {
		return nil, err
	}
	return []*models.Lesson{lesson}, nil
}

func (s *LessonService) createLessonWithGroup(ctx context.Context, req *models.CreateLessonRequest, groupID *uuid.UUID) (*models.Lesson, error) {
	req.ApplyDefaults()

	if err := req.Validate(); err != nil {
		return nil, err
	}

	if err := s.lessonValidator.ValidateCreateLessonRequest(req); err != nil {
		return nil, err
	}

	teacher, err := s.userRepo.GetByID(ctx, req.TeacherID)
	if err != nil {
		return nil, fmt.Errorf("teacher not found: %w", err)
	}

	if !teacher.CanBeAssignedAsTeacher() {
		return nil, fmt.Errorf("user cannot be assigned as teacher (role: %s): %w", teacher.Role, models.ErrInvalidTeacherID)
	}

	lesson := &models.Lesson{
		TeacherID:         req.TeacherID,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		MaxStudents:       req.MaxStudents,
		CreditsCost:       req.CreditsCost,
		Color:             req.Color,
		IsRecurring:       req.IsRecurring,
		RecurringGroupID:  groupID,
	}

	if req.Subject != nil && *req.Subject != "" {
		lesson.Subject.String = *req.Subject
		lesson.Subject.Valid = true
	}

	if req.HomeworkText != nil && *req.HomeworkText != "" {
		lesson.HomeworkText.String = *req.HomeworkText
		lesson.HomeworkText.Valid = true
	}

	if req.Link != nil && *req.Link != "" {
		lesson.Link.String = *req.Link
		lesson.Link.Valid = true
	}

	if req.RecurringEndDate != nil {
		lesson.RecurringEndDate.Time = *req.RecurringEndDate
		lesson.RecurringEndDate.Valid = true
	}

	if err := s.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	if len(req.StudentIDs) > 0 && s.bookingCreator != nil {
		if len(req.StudentIDs) > req.MaxStudents {
			return nil, fmt.Errorf("number of students (%d) exceeds max_students (%d)", len(req.StudentIDs), req.MaxStudents)
		}

		seenStudents := make(map[uuid.UUID]bool)
		for _, studentID := range req.StudentIDs {
			if seenStudents[studentID] {
				return nil, fmt.Errorf("duplicate student_id: %s", studentID)
			}
			seenStudents[studentID] = true
		}

		for _, studentID := range req.StudentIDs {
			student, err := s.userRepo.GetByID(ctx, studentID)
			if err != nil {
				return nil, fmt.Errorf("student not found: %s", studentID)
			}
			if !student.IsStudent() {
				return nil, fmt.Errorf("user %s is not a student", studentID)
			}
		}

		for _, studentID := range req.StudentIDs {
			bookingReq := &models.CreateBookingRequest{
				StudentID: studentID,
				LessonID:  lesson.ID,
				IsAdmin:   true,
			}

			_, err := s.bookingCreator.CreateBooking(ctx, bookingReq)
			if err != nil {
				return nil, fmt.Errorf("failed to enroll student %s: %w", studentID, err)
			}

			log.Info().
				Str("lesson_id", lesson.ID.String()).
				Str("student_id", studentID.String()).
				Msg("Student enrolled on lesson creation")
		}

		if s.telegramService != nil && len(req.StudentIDs) > 0 {
			studentNames := make([]string, 0, len(req.StudentIDs))
			for _, studentID := range req.StudentIDs {
				student, err := s.userRepo.GetByID(ctx, studentID)
				if err == nil {
					studentNames = append(studentNames, student.GetFullName())
				} else {
					studentNames = append(studentNames, "Студент")
				}
			}

			go func() {
				notifCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				for i, studentID := range req.StudentIDs {
					select {
					case <-notifCtx.Done():
						log.Warn().
							Str("lesson_id", lesson.ID.String()).
							Msg("Lesson booking notification goroutine cancelled")
						return
					default:
					}

					studentName := studentNames[i]
					if err := s.telegramService.NotifyLessonBooking(notifCtx, lesson, studentName, []uuid.UUID{studentID}); err != nil {
						log.Warn().
							Str("lesson_id", lesson.ID.String()).
							Str("student_id", studentID.String()).
							Err(err).
							Msg("Failed to send booking notification to user")
					}
				}
			}()
		}
	}

	return lesson, nil
}
