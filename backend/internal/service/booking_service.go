package service

import (
	"context"
	"fmt"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/utils"
	"tutoring-platform/internal/validator"
	"tutoring-platform/pkg/metrics"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// BookingService обрабатывает бизнес-логику для бронирований
type BookingService struct {
	pool                 *pgxpool.Pool
	bookingRepo          *repository.BookingRepository
	lessonRepo           *repository.LessonRepository
	creditRepo           *repository.CreditRepository
	cancelledBookingRepo *repository.CancelledBookingRepository
	bookingValidator     *validator.BookingValidator
}

// NewBookingService создает новый BookingService
func NewBookingService(
	pool *pgxpool.Pool,
	bookingRepo *repository.BookingRepository,
	lessonRepo *repository.LessonRepository,
	creditRepo *repository.CreditRepository,
	cancelledBookingRepo *repository.CancelledBookingRepository,
	bookingValidator *validator.BookingValidator,
) *BookingService {
	return &BookingService{
		pool:                 pool,
		bookingRepo:          bookingRepo,
		lessonRepo:           lessonRepo,
		creditRepo:           creditRepo,
		cancelledBookingRepo: cancelledBookingRepo,
		bookingValidator:     bookingValidator,
	}
}

// CreateBooking создает новое бронирование (атомарная операция)
func (s *BookingService) CreateBooking(ctx context.Context, req *models.CreateBookingRequest) (*models.Booking, error) {
	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверяем возможность бронирования
	if err := s.bookingValidator.ValidateBooking(ctx, req.StudentID, req.LessonID, req.IsAdmin); err != nil {
		return nil, err
	}

	// Проверяем, что студент не отписался ранее от этого урока
	// Админ может повторно записать ученика после отписки - в этом случае удаляем запись из cancelled_bookings
	hasCancelled, err := s.cancelledBookingRepo.HasCancelledBooking(ctx, req.StudentID, req.LessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to check cancelled bookings: %w", err)
	}
	if hasCancelled {
		if req.IsAdmin {
			if err := s.cancelledBookingRepo.DeleteByStudentAndLesson(ctx, req.StudentID, req.LessonID); err != nil {
				log.Error().
					Str("student_id", req.StudentID.String()).
					Str("lesson_id", req.LessonID.String()).
					Err(err).
					Msg("Failed to delete cancelled booking record for admin re-booking")
				return nil, fmt.Errorf("failed to clear cancelled booking record: %w", err)
			}
		} else {
			return nil, repository.ErrLessonPreviouslyCancelled
		}
	}

	// Проверяем, что студент еще не забронировал этот урок
	// ВАЖНО: Эта проверка может быть обогнана concurrent request'ом
	// Финальная проверка будет в транзакции через UNIQUE constraint (см. миграции)
	// Для админа пропускаем эту проверку - он может перезаписывать студентов
	// Финальная защита через UNIQUE constraint в базе
	if !req.IsAdmin {
		existingBooking, err := s.bookingRepo.GetActiveBookingByStudentAndLesson(ctx, req.StudentID, req.LessonID)
		if err == nil && existingBooking != nil {
			return nil, repository.ErrAlreadyBooked
		}
		// Игнорируем ошибку ErrBookingNotFound - это нормально, значит бронирования нет
		if err != nil && err != repository.ErrBookingNotFound {
			return nil, fmt.Errorf("failed to check existing booking: %w", err)
		}
	}

	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Rollback дефер гарантирует откат при ошибке (успешный commit закроет транзакцию)
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		// Игнорируем ошибку "tx is closed" - это нормально после успешного commit
		if rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			log.Warn().Err(rollbackErr).Msg("Failed to rollback transaction in CreateBooking")
		}
	}()

	// Получаем урок для определения стоимости в кредитах (нужно для обоих путей)
	lesson, err := s.lessonRepo.GetByIDForUpdate(ctx, tx, req.LessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	// Verify lesson was found (safety check)
	if lesson == nil {
		return nil, fmt.Errorf("lesson not found")
	}

	creditsCost := lesson.CreditsCost
	if creditsCost < 0 {
		creditsCost = 0 // Исправляем отрицательные значения на 0
	}

	// Для админа: попробовать реактивировать существующее cancelled бронирование
	// ВАЖНО: этот путь изолирован с помощью раннего return на строке 183
	// Это гарантирует что кредит списывается ровно 1 раз и путь 2 не выполнится
	if req.IsAdmin {
		// Check capacity before reactivation (lesson already locked via SELECT FOR UPDATE at line 114)
		// This ensures we don't reactivate if lesson is already at capacity
		if lesson.IsFull() {
			return nil, repository.ErrLessonFull
		}

		reactivated, reactivateErr := s.bookingRepo.ReactivateBooking(ctx, tx, req.StudentID, req.LessonID)
		if reactivateErr == nil && reactivated != nil {
			log.Info().
				Str("booking_id", reactivated.ID.String()).
				Str("student_id", utils.MaskUserID(req.StudentID)).
				Str("lesson_id", req.LessonID.String()).
				Msg("Capacity check passed - admin reactivating cancelled booking")

			// Списываем кредиты только если урок платный (creditsCost > 0)
			if creditsCost > 0 {
				// Блокируем и получаем баланс кредитов для списания
				credit, creditErr := s.creditRepo.GetBalanceForUpdate(ctx, tx, req.StudentID)
				if creditErr != nil {
					return nil, fmt.Errorf("failed to get credit balance: %w", creditErr)
				}

				if !credit.HasSufficientBalance(creditsCost) {
					return nil, repository.ErrInsufficientCredits
				}

				// Списываем кредиты
				newBalance := credit.Balance - creditsCost
				if err := s.creditRepo.UpdateBalance(ctx, tx, req.StudentID, newBalance); err != nil {
					return nil, fmt.Errorf("failed to update credit balance: %w", err)
				}

				log.Info().
					Str("student_id", utils.MaskUserID(req.StudentID)).
					Str("balance", utils.MaskAmount(newBalance)).
					Bool("is_reactivation", true).
					Msg("Credit deducted for reactivated booking")

				// Записываем транзакцию кредитов
				reason := "Admin re-booking (reactivation)"
				var performedBy uuid.NullUUID
				if req.AdminID != uuid.Nil {
					performedBy = uuid.NullUUID{UUID: req.AdminID, Valid: true}
				}

				transaction := &models.CreditTransaction{
					UserID:        req.StudentID,
					Amount:        -creditsCost,
					OperationType: models.OperationTypeDeduct,
					Reason:        reason,
					PerformedBy:   performedBy,
					BookingID:     uuid.NullUUID{UUID: reactivated.ID, Valid: true},
					BalanceBefore: credit.Balance,
					BalanceAfter:  newBalance,
				}
				if err := s.creditRepo.CreateTransaction(ctx, tx, transaction); err != nil {
					return nil, fmt.Errorf("failed to create credit transaction: %w", err)
				}
			}

			// Увеличиваем счётчик студентов в уроке
			if err := s.lessonRepo.IncrementStudents(ctx, tx, req.LessonID); err != nil {
				log.Error().
					Str("lesson_id", req.LessonID.String()).
					Err(err).
					Msg("Failed to increment students for reactivated booking")
				return nil, fmt.Errorf("failed to increment students: %w", err)
			}

			// Фиксируем транзакцию
			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("failed to commit transaction: %w", err)
			}

			metrics.BookingsCreated.Inc()
			metrics.CreditsDeducted.Inc()

			// РАННИЙ ВЫХОД: гарантирует что путь 2 (создание нового бронирования ниже) не выполнится
			// Это предотвращает двойное списание кредитов
			return reactivated, nil
		}
		// Если ErrBookingNotFound - продолжаем создание нового бронирования (путь 2)
		if reactivateErr != nil && reactivateErr != repository.ErrBookingNotFound {
			log.Warn().Err(reactivateErr).Msg("Unexpected error during reactivation attempt, proceeding with new booking")
		}
	}

	// Путь 2: Создание нового бронирования (выполняется если путь 1 не сработал)
	// Списываем кредиты только если урок платный (creditsCost > 0)
	var credit *models.Credit
	var newBalance int
	if creditsCost > 0 {
		// Блокируем и получаем баланс кредитов
		var err error
		credit, err = s.creditRepo.GetBalanceForUpdate(ctx, tx, req.StudentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get credit balance: %w", err)
		}

		// Calculate final balance after deduction (safety check after SELECT FOR UPDATE)
		// This is the ONLY balance check - performed after acquiring lock via SELECT FOR UPDATE
		// Database trigger provides additional safety net to prevent negative balance
		newBalance = credit.Balance - creditsCost
		if newBalance < 0 {
			return nil, repository.ErrInsufficientCredits
		}
		if err := s.creditRepo.UpdateBalance(ctx, tx, req.StudentID, newBalance); err != nil {
			return nil, fmt.Errorf("failed to update credit balance: %w", err)
		}

		// Log credit deduction
		log.Info().
			Str("student_id", utils.MaskUserID(req.StudentID)).
			Str("balance", utils.MaskAmount(newBalance)).
			Bool("is_admin_booking", req.IsAdmin).
			Msg("Credit deducted for booking")
	}

	// Проверяем доступность урока (после блокировки - это авторитетная проверка)
	if lesson.IsFull() {
		log.Warn().
			Str("lesson_id", lesson.ID.String()).
			Int("current_students", lesson.CurrentStudents).
			Int("max_students", lesson.MaxStudents).
			Msg("Lesson is full - race condition detected")
		return nil, repository.ErrLessonFull
	}

	// Создаем бронирование
	// Если при INSERT-е произойдет нарушение UNIQUE constraint (concurrent booking),
	// это будет перехвачено и преобразовано в ErrAlreadyBooked в Create методе
	booking := &models.Booking{
		StudentID: req.StudentID,
		LessonID:  req.LessonID,
	}
	if err := s.bookingRepo.Create(ctx, tx, booking); err != nil {
		// ErrDuplicateBooking уже содержит пользовательское сообщение об ошибке
		// которое возвращается из репозитория при нарушении UNIQUE constraint
		if err == repository.ErrDuplicateBooking {
			return nil, repository.ErrDuplicateBooking
		}
		// Для совместимости, также проверяем IsUniqueViolationError (на случай других UNIQUE constraints)
		if repository.IsUniqueViolationError(err) {
			return nil, repository.ErrAlreadyBooked
		}
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Записываем транзакцию кредитов только если урок платный (creditsCost > 0)
	if creditsCost > 0 && credit != nil {
		// Determine transaction reason and performed_by based on who created the booking
		reason := "Booking lesson"
		var performedBy uuid.NullUUID
		if req.IsAdmin && req.AdminID != uuid.Nil {
			reason = "Admin booking for student"
			performedBy = uuid.NullUUID{UUID: req.AdminID, Valid: true}
		}

		transaction := &models.CreditTransaction{
			UserID:        req.StudentID,
			Amount:        -creditsCost,
			OperationType: models.OperationTypeDeduct,
			Reason:        reason,
			PerformedBy:   performedBy,
			BookingID:     uuid.NullUUID{UUID: booking.ID, Valid: true},
			BalanceBefore: credit.Balance,
			BalanceAfter:  newBalance,
		}
		if err := s.creditRepo.CreateTransaction(ctx, tx, transaction); err != nil {
			return nil, fmt.Errorf("failed to create credit transaction: %w", err)
		}

		log.Info().
			Str("student_id", req.StudentID.String()).
			Str("reason", reason).
			Str("performed_by", performedBy.UUID.String()).
			Msg("Credit transaction created")
	}

	// Увеличиваем счетчик студентов в уроке
	// IncrementStudents has additional safety: WHERE current_students < max_students
	// This is a final defense against race condition (should never fail if IsFull check passed)
	if err := s.lessonRepo.IncrementStudents(ctx, tx, req.LessonID); err != nil {
		log.Error().
			Str("lesson_id", req.LessonID.String()).
			Err(err).
			Msg("Failed to increment students - race condition detected at DB level")
		return nil, fmt.Errorf("failed to increment students: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Обновляем метрики успешного бронирования
	metrics.BookingsCreated.Inc()
	metrics.CreditsDeducted.Inc()

	return booking, nil
}

// CancelBooking отменяет бронирование и возвращает кредит (атомарная операция)
// Возвращает CancelBookingResult с информацией о статусе операции
// Status: "success" если бронирование было активно и успешно отменено
// Status: "already_cancelled" если бронирование было уже отменено (идемпотентное поведение)
func (s *BookingService) CancelBooking(ctx context.Context, req *models.CancelBookingRequest) (*models.CancelBookingResult, error) {
	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Начинаем транзакцию
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Rollback дефер гарантирует откат при ошибке (успешный commit закроет транзакцию)
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		// Игнорируем ошибку "tx is closed" - это нормально после успешного commit
		if rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			log.Warn().Err(rollbackErr).Msg("Failed to rollback transaction in CancelBooking")
		}
	}()

	// Блокируем и получаем бронирование (внутри транзакции для избежания race condition)
	booking, err := s.bookingRepo.GetByIDForUpdate(ctx, tx, req.BookingID)
	if err != nil {
		return nil, err
	}

	// Проверяем владельца (внутри транзакции)
	// Студенты могут отменять только свои бронирования, админы - любые
	if !req.IsAdmin && booking.StudentID != req.StudentID {
		return nil, repository.ErrUnauthorized
	}

	// Если бронирование уже отменено, возвращаем успешный результат (идемпотентность)
	// Это позволяет клиенту переповторно отменять одно и то же бронирование без ошибок
	if booking.Status == models.BookingStatusCancelled {
		log.Info().
			Str("booking_id", booking.ID.String()).
			Str("student_id", utils.MaskUserID(booking.StudentID)).
			Msg("Duplicate cancellation attempt - booking already cancelled")
		return &models.CancelBookingResult{
			Status:  models.CancelResultAlreadyCancelled,
			Message: "Booking was already cancelled",
		}, nil
	}

	// Проверяем возможность отмены
	// ValidateCancellation will skip 24-hour check for admins
	if err := s.bookingValidator.ValidateCancellation(ctx, booking, req.IsAdmin); err != nil {
		return nil, err
	}

	// Отменяем бронирование
	if err := s.bookingRepo.Cancel(ctx, tx, req.BookingID); err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Используем ID студента из бронирования для возврата кредитов (особенно важно для админов)
	studentID := booking.StudentID

	// Возвращаем кредиты ТОЛЬКО если бронирование было в статусе Active
	// Это гарантирует что кредиты были списаны при создании и нужно их вернуть
	// Для других статусов (Completed, Cancelled и т.д.) кредиты уже были обработаны корректно
	if booking.Status == models.BookingStatusActive {
		// Получаем урок для определения стоимости в кредитах
		lesson, err := s.lessonRepo.GetByIDForUpdate(ctx, tx, booking.LessonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get lesson: %w", err)
		}

		creditsCost := lesson.CreditsCost
		if creditsCost < 0 {
			creditsCost = 0 // Исправляем отрицательные значения на 0
		}

		// Блокируем и получаем баланс кредитов
		credit, err := s.creditRepo.GetBalanceForUpdate(ctx, tx, studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get credit balance: %w", err)
		}

		// Возвращаем кредиты
		newBalance := credit.Balance + creditsCost
		if err := s.creditRepo.UpdateBalance(ctx, tx, studentID, newBalance); err != nil {
			return nil, fmt.Errorf("failed to update credit balance: %w", err)
		}

		// Записываем транзакцию кредитов
		transaction := &models.CreditTransaction{
			UserID:        studentID,
			Amount:        creditsCost,
			OperationType: models.OperationTypeRefund,
			Reason:        "Booking cancelled",
			BookingID:     uuid.NullUUID{UUID: booking.ID, Valid: true},
			BalanceBefore: credit.Balance,
			BalanceAfter:  newBalance,
		}
		if err := s.creditRepo.CreateTransaction(ctx, tx, transaction); err != nil {
			return nil, fmt.Errorf("failed to create credit transaction: %w", err)
		}
	}

	// Уменьшаем счетчик студентов в уроке
	if err := s.lessonRepo.DecrementStudents(ctx, tx, booking.LessonID); err != nil {
		return nil, fmt.Errorf("failed to decrement students: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// После успешной отписки создать запись о отмене для предотвращения повторной записи
	// Это НЕ критичная операция - если не сработает, не блокируем отмену
	cancelledBooking := &models.CancelledBooking{
		ID:          uuid.New(),
		BookingID:   booking.ID,
		StudentID:   booking.StudentID,
		LessonID:    booking.LessonID,
		CancelledAt: time.Now(),
	}
	if err := s.cancelledBookingRepo.CreateCancelledBooking(ctx, cancelledBooking); err != nil {
		// Логируем ошибку но не fail операцию (отмена более важна)
		log.Warn().
			Str("booking_id", booking.ID.String()).
			Err(err).
			Msg("Failed to create cancelled booking record")
	}

	// Обновляем метрики успешной отмены бронирования
	metrics.BookingsCancelled.Inc()

	// Метрика возврата кредитов обновляется только если кредиты действительно были возвращены
	if booking.Status == models.BookingStatusActive {
		metrics.CreditsRefunded.Inc()
	}

	return &models.CancelBookingResult{
		Status:  models.CancelResultSuccess,
		Message: "Booking cancelled successfully",
	}, nil
}

// GetBooking получает бронирование с деталями
func (s *BookingService) GetBooking(ctx context.Context, bookingID uuid.UUID) (*models.BookingWithDetails, error) {
	return s.bookingRepo.GetWithDetails(ctx, bookingID)
}

// ListBookings получает список бронирований с фильтрами
func (s *BookingService) ListBookings(ctx context.Context, filter *models.ListBookingsFilter) ([]*models.BookingWithDetails, error) {
	bookings, err := s.bookingRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Валидация возвращаемых данных: убедимся что все бронирования корректны
	validBookings := make([]*models.BookingWithDetails, 0, len(bookings))
	for _, booking := range bookings {
		// Проверяем что lesson_id и student_id существуют и не нулевые
		if booking.LessonID == uuid.Nil || booking.StudentID == uuid.Nil {
			// Логируем некорректную запись, но не возвращаем ошибку
			log.Warn().
				Str("booking_id", booking.ID.String()).
				Str("lesson_id", booking.LessonID.String()).
				Str("student_id", booking.StudentID.String()).
				Msg("Skipping booking with invalid IDs")
			continue
		}

		// Проверяем что статус корректен
		if booking.Status != models.BookingStatusActive && booking.Status != models.BookingStatusCancelled {
			log.Warn().
				Str("booking_id", booking.ID.String()).
				Str("status", string(booking.Status)).
				Msg("Skipping booking with invalid status")
			continue
		}

		validBookings = append(validBookings, booking)
	}

	return validBookings, nil
}

// ListBookingsWithPagination получает список бронирований с пагинацией
func (s *BookingService) ListBookingsWithPagination(ctx context.Context, filter *models.ListBookingsFilter, offset, limit int) ([]*models.BookingWithDetails, int, error) {
	bookings, total, err := s.bookingRepo.ListWithPagination(ctx, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Валидация возвращаемых данных: убедимся что все бронирования корректны
	validBookings := make([]*models.BookingWithDetails, 0, len(bookings))
	for _, booking := range bookings {
		// Проверяем что lesson_id и student_id существуют и не нулевые
		if booking.LessonID == uuid.Nil || booking.StudentID == uuid.Nil {
			log.Warn().
				Str("booking_id", booking.ID.String()).
				Str("lesson_id", booking.LessonID.String()).
				Str("student_id", booking.StudentID.String()).
				Msg("Skipping booking with invalid IDs in paginated list")
			continue
		}

		// Проверяем что статус корректен
		if booking.Status != models.BookingStatusActive && booking.Status != models.BookingStatusCancelled {
			log.Warn().
				Str("booking_id", booking.ID.String()).
				Str("status", string(booking.Status)).
				Msg("Skipping booking with invalid status in paginated list")
			continue
		}

		validBookings = append(validBookings, booking)
	}

	return validBookings, total, nil
}

// GetStudentLessons получает все занятия студента с информацией о преподавателях в ОДНОМ JOIN запросе.
// Эта функция исправляет N+1 проблему: вместо 1 запроса на бронирования + N запросов на занятия,
// делается 1 JOIN запрос через INNER JOIN bookings с lessons и LEFT JOIN с users для учителя.
// Возвращает только активные бронирования, упорядоченные по времени начала (новые сначала).
func (s *BookingService) GetStudentLessons(ctx context.Context, studentID uuid.UUID) ([]*models.LessonWithTeacher, error) {
	log.Debug().Str("student_id", utils.MaskUserID(studentID)).Msg("Getting lessons for student")

	lessons, err := s.bookingRepo.GetBookingsWithLessons(ctx, studentID)
	if err != nil {
		log.Error().
			Str("student_id", studentID.String()).
			Err(err).
			Msg("Failed to get bookings with lessons")
		return nil, fmt.Errorf("failed to get student lessons: %w", err)
	}

	log.Debug().
		Str("student_id", studentID.String()).
		Int("lessons_count", len(lessons)).
		Msg("Successfully retrieved lessons for student")

	return lessons, nil
}

// GetCancelledLessonIDs возвращает список ID занятий, от которых студент отписался
// Используется для блокировки повторной записи на занятие
func (s *BookingService) GetCancelledLessonIDs(ctx context.Context, studentID uuid.UUID) ([]uuid.UUID, error) {
	log.Debug().Str("student_id", utils.MaskUserID(studentID)).Msg("Getting cancelled lesson IDs for student")

	cancelled, err := s.cancelledBookingRepo.ListCancelledByStudent(ctx, studentID)
	if err != nil {
		log.Error().
			Str("student_id", studentID.String()).
			Err(err).
			Msg("Failed to get cancelled bookings")
		return nil, fmt.Errorf("failed to get cancelled lessons: %w", err)
	}

	lessonIDs := make([]uuid.UUID, 0, len(cancelled))
	for _, c := range cancelled {
		lessonIDs = append(lessonIDs, c.LessonID)
	}

	log.Debug().
		Str("student_id", studentID.String()).
		Int("cancelled_count", len(lessonIDs)).
		Msg("Successfully retrieved cancelled lesson IDs")

	return lessonIDs, nil
}
