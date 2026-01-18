package service

import (
	"context"
	"fmt"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/validator"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// SwapService обрабатывает бизнес-логику для обменов уроков
type SwapService struct {
	pool          *pgxpool.Pool
	swapRepo      *repository.SwapRepository
	bookingRepo   *repository.BookingRepository
	lessonRepo    *repository.LessonRepository
	swapValidator *validator.SwapValidator
}

// NewSwapService создает новый SwapService
func NewSwapService(
	pool *pgxpool.Pool,
	swapRepo *repository.SwapRepository,
	bookingRepo *repository.BookingRepository,
	lessonRepo *repository.LessonRepository,
	swapValidator *validator.SwapValidator,
) *SwapService {
	return &SwapService{
		pool:          pool,
		swapRepo:      swapRepo,
		bookingRepo:   bookingRepo,
		lessonRepo:    lessonRepo,
		swapValidator: swapValidator,
	}
}

// PerformSwap выполняет обмен урока с изоляцией SERIALIZABLE (атомарная операция)
// Гарантирует, что проверка мест и регистрация студента выполняются атомарно
// используя SELECT FOR UPDATE и единую транзакцию.
func (s *SwapService) PerformSwap(ctx context.Context, req *models.PerformSwapRequest) (*models.Swap, error) {
	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверяем возможность обмена
	if err := s.swapValidator.ValidateSwap(ctx, req.StudentID, req.OldLessonID, req.NewLessonID); err != nil {
		return nil, err
	}

	// Начинаем транзакцию с уровнем изоляции SERIALIZABLE
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Гарантируем откат при выходе из функции
	txClosed := false
	defer func() {
		if !txClosed {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr.Error() != "tx is closed" {
				// Логируем ошибку откката с полным контекстом
				log.Error().
					Err(rollbackErr).
					Str("studentID", req.StudentID.String()).
					Str("oldLessonID", req.OldLessonID.String()).
					Str("newLessonID", req.NewLessonID.String()).
					Msg("failed to rollback swap transaction")
			}
		}
	}()

	// Устанавливаем уровень изоляции SERIALIZABLE для максимальной согласованности
	if _, err := tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE"); err != nil {
		return nil, fmt.Errorf("failed to set isolation level: %w", err)
	}

	// Для предотвращения deadlock блокируем уроки в детерминированном порядке (сортируем по UUID)
	// Если два студента одновременно меняются местами (A->B и B->A), без сортировки возможен deadlock
	var oldLesson, newLesson *models.Lesson
	if req.OldLessonID.String() < req.NewLessonID.String() {
		// Блокируем в порядке: old -> new
		oldLesson, err = s.lessonRepo.GetByIDForUpdate(ctx, tx, req.OldLessonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get old lesson: %w", err)
		}
		newLesson, err = s.lessonRepo.GetByIDForUpdate(ctx, tx, req.NewLessonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get new lesson: %w", err)
		}
	} else {
		// Блокируем в порядке: new -> old
		newLesson, err = s.lessonRepo.GetByIDForUpdate(ctx, tx, req.NewLessonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get new lesson: %w", err)
		}
		oldLesson, err = s.lessonRepo.GetByIDForUpdate(ctx, tx, req.OldLessonID)
		if err != nil {
			return nil, fmt.Errorf("failed to get old lesson: %w", err)
		}
	}

	// Проверяем, есть ли места в новом уроке (этот чек теперь под блокировкой SELECT FOR UPDATE)
	// Дополнительно IncrementStudents проверит это при выполнении UPDATE с условием current_students < max_students
	if newLesson.IsFull() {
		return nil, repository.ErrLessonFull
	}

	// Используем oldLesson для проверки (избегаем unused variable)
	_ = oldLesson

	// Получаем активное бронирование для старого урока
	oldBooking, err := s.bookingRepo.GetActiveBookingByStudentAndLesson(ctx, req.StudentID, req.OldLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get old booking: %w", err)
	}

	// Отменяем старое бронирование
	if err := s.bookingRepo.Cancel(ctx, tx, oldBooking.ID); err != nil {
		return nil, fmt.Errorf("failed to cancel old booking: %w", err)
	}

	// Уменьшаем счетчик студентов в старом уроке
	if err := s.lessonRepo.DecrementStudents(ctx, tx, req.OldLessonID); err != nil {
		return nil, fmt.Errorf("failed to decrement old lesson students: %w", err)
	}

	// Создаем новое бронирование
	newBooking := &models.Booking{
		StudentID: req.StudentID,
		LessonID:  req.NewLessonID,
	}
	if err := s.bookingRepo.Create(ctx, tx, newBooking); err != nil {
		return nil, fmt.Errorf("failed to create new booking: %w", err)
	}

	// Увеличиваем счетчик студентов в новом уроке
	// АТОМАРНАЯ ПРОВЕРКА: Если current_students >= max_students, это вернет ErrLessonFull
	// Это гарантирует, что запись произойдет ТОЛЬКО если есть место
	if err := s.lessonRepo.IncrementStudents(ctx, tx, req.NewLessonID); err != nil {
		// Если места нет, откатываем и возвращаем ошибку
		// Бронирование будет отменено откатом транзакции
		return nil, fmt.Errorf("failed to increment new lesson students: %w", err)
	}

	// Создаем запись обмена
	swap := &models.Swap{
		StudentID:    req.StudentID,
		OldLessonID:  req.OldLessonID,
		NewLessonID:  req.NewLessonID,
		OldBookingID: oldBooking.ID,
		NewBookingID: newBooking.ID,
	}
	if err := s.swapRepo.Create(ctx, tx, swap); err != nil {
		return nil, fmt.Errorf("failed to create swap record: %w", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	txClosed = true

	return swap, nil
}

// ValidateSwap проверяет возможность выполнения обмена без его выполнения
func (s *SwapService) ValidateSwap(ctx context.Context, req *models.ValidateSwapRequest) (*models.ValidateSwapResponse, error) {
	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return &models.ValidateSwapResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	// Проверяем возможность обмена
	if err := s.swapValidator.ValidateSwap(ctx, req.StudentID, req.OldLessonID, req.NewLessonID); err != nil {
		return &models.ValidateSwapResponse{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	// Получаем детали урока для ответа
	oldLesson, err := s.lessonRepo.GetWithTeacher(ctx, req.OldLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get old lesson: %w", err)
	}

	newLesson, err := s.lessonRepo.GetWithTeacher(ctx, req.NewLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get new lesson: %w", err)
	}

	response := &models.ValidateSwapResponse{
		Valid:  true,
		Errors: nil,
	}
	response.Details.OldLesson = oldLesson
	response.Details.NewLesson = newLesson

	return response, nil
}

// GetSwapHistory получает историю обменов с фильтрами
func (s *SwapService) GetSwapHistory(ctx context.Context, filter *models.GetSwapHistoryFilter) ([]*models.SwapWithDetails, error) {
	return s.swapRepo.List(ctx, filter)
}

// GetSwapByID получает обмен по ID с деталями
func (s *SwapService) GetSwapByID(ctx context.Context, swapID uuid.UUID) (*models.SwapWithDetails, error) {
	return s.swapRepo.GetWithDetails(ctx, swapID)
}
