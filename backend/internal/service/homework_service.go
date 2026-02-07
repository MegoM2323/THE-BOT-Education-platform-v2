package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
)

// HomeworkService обрабатывает бизнес-логику для домашних заданий
type HomeworkService struct {
	homeworkRepo homeworkServiceRepository
	lessonRepo   homeworkServiceLessonRepository
	bookingRepo  homeworkServiceBookingRepository
	userRepo     homeworkServiceUserRepository
}

// homeworkServiceRepository - интерфейс для dependency injection в тестах
type homeworkServiceRepository interface {
	CreateHomework(ctx context.Context, homework *models.LessonHomework) (*models.LessonHomework, error)
	GetHomeworkByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonHomework, error)
	GetHomeworkByID(ctx context.Context, homeworkID uuid.UUID) (*models.LessonHomework, error)
	DeleteHomework(ctx context.Context, homeworkID uuid.UUID) error
	UpdateHomework(ctx context.Context, homeworkID uuid.UUID, textContent string) error
}

// homeworkServiceLessonRepository - интерфейс для работы с уроками
type homeworkServiceLessonRepository interface {
	GetByID(ctx context.Context, lessonID uuid.UUID) (*models.Lesson, error)
}

// homeworkServiceBookingRepository - интерфейс для работы с бронированиями
type homeworkServiceBookingRepository interface {
	GetActiveBookingByStudentAndLesson(ctx context.Context, studentID, lessonID uuid.UUID) (*models.Booking, error)
}

// homeworkServiceUserRepository - интерфейс для работы с пользователями
type homeworkServiceUserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// NewHomeworkService создает новый HomeworkService
func NewHomeworkService(
	homeworkRepo homeworkServiceRepository,
	lessonRepo homeworkServiceLessonRepository,
	bookingRepo homeworkServiceBookingRepository,
	userRepo homeworkServiceUserRepository,
) *HomeworkService {
	return &HomeworkService{
		homeworkRepo: homeworkRepo,
		lessonRepo:   lessonRepo,
		bookingRepo:  bookingRepo,
		userRepo:     userRepo,
	}
}

// CreateHomework сохраняет файл и создает запись домашнего задания
// Проверяет:
// - Существование урока
// - Права пользователя (admin или teacher этого урока)
// - Валидность файла
func (s *HomeworkService) CreateHomework(ctx context.Context, userID uuid.UUID, file io.Reader, req *models.CreateHomeworkRequest) (*models.LessonHomework, error) {
	// Валидация запроса
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверка существования урока
	lesson, err := s.lessonRepo.GetByID(ctx, req.LessonID)
	if err != nil {
		return nil, repository.ErrLessonNotFound
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка прав: admin, методист или teacher этого урока
	if !user.IsAdmin() && !user.IsTeacher() && lesson.TeacherID != userID {
		return nil, repository.ErrUnauthorized
	}

	// Создаем директорию для файлов если не существует
	uploadsDir := "uploads/homework"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, models.ErrFileStorageFailed
	}

	// Генерируем уникальное имя файла: UUID + расширение
	// Используем filepath.Base для безопасности (против path traversal)
	safeFileName := filepath.Base(req.FileName)
	ext := filepath.Ext(safeFileName)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(uploadsDir, filename)

	// Сохраняем файл
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, models.ErrFileStorageFailed
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		// Удаляем файл если копирование не удалось
		os.Remove(filePath)
		return nil, models.ErrFileStorageFailed
	}

	// Создаем запись в БД
	homework := &models.LessonHomework{
		LessonID:  req.LessonID,
		FileName:  safeFileName,
		FilePath:  filePath,
		FileSize:  req.FileSize,
		MimeType:  req.MimeType,
		CreatedBy: userID,
	}

	created, err := s.homeworkRepo.CreateHomework(ctx, homework)
	if err != nil {
		// Удаляем файл если запись в БД не удалась
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create homework record: %w", err)
	}

	return created, nil
}

// GetHomeworkByLesson получает список файлов ДЗ с проверкой доступа
// Правила доступа:
// - Admin: может видеть все ДЗ всех уроков
// - Teacher: может видеть ДЗ только своего урока
// - Student: может видеть ДЗ только если записан на урок или это групповой урок
func (s *HomeworkService) GetHomeworkByLesson(ctx context.Context, userID uuid.UUID, lessonID uuid.UUID) ([]*models.LessonHomework, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем урок
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, repository.ErrLessonNotFound
	}

	// Проверка доступа в зависимости от роли
	hasAccess := false

	switch user.Role {
	case models.RoleAdmin:
		// Admin видит все ДЗ всех уроков
		hasAccess = true

	case models.RoleTeacher:
		// Методист видит все ДЗ всех уроков (как админ)
		hasAccess = true

	case models.RoleStudent:
		// Student видит ДЗ:
		// 1. Групповых уроков (max_students > 1)
		// 2. Индивидуальных уроков (max_students == 1), на которые записан
		if lesson.MaxStudents > 1 {
			hasAccess = true
		} else {
			// Проверяем запись на индивидуальный урок
			_, err := s.bookingRepo.GetActiveBookingByStudentAndLesson(ctx, userID, lessonID)
			if err == nil {
				hasAccess = true
			}
		}
	}

	if !hasAccess {
		return nil, repository.ErrUnauthorized
	}

	// Получаем список домашних заданий
	homeworks, err := s.homeworkRepo.GetHomeworkByLesson(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get homework by lesson: %w", err)
	}

	return homeworks, nil
}

// DeleteHomework удаляет файл и запись из БД
// Проверяет:
// - Существование домашнего задания
// - Права пользователя (admin, creator файла, или teacher урока)
func (s *HomeworkService) DeleteHomework(ctx context.Context, userID uuid.UUID, homeworkID uuid.UUID) error {
	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем домашнее задание
	homework, err := s.homeworkRepo.GetHomeworkByID(ctx, homeworkID)
	if err != nil {
		return err // ErrHomeworkNotFound уже обработан в repo
	}

	// Получаем урок для проверки teacher
	lesson, err := s.lessonRepo.GetByID(ctx, homework.LessonID)
	if err != nil {
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Проверка прав: admin, teacher, creator файла, или teacher урока
	isAdmin := user.IsAdmin()
	isTeacher := user.IsTeacher()
	isCreator := homework.CreatedBy == userID
	isLessonTeacher := lesson.TeacherID == userID

	if !isAdmin && !isTeacher && !isCreator && !isLessonTeacher {
		return repository.ErrUnauthorized
	}

	// Удаляем файл из файловой системы
	if err := os.Remove(homework.FilePath); err != nil {
		// Логируем ошибку но продолжаем (файл может быть уже удален вручную)
		fmt.Printf("[WARN] Failed to delete homework file %s: %v\n", homework.FilePath, err)
	}

	// Удаляем запись из БД
	if err := s.homeworkRepo.DeleteHomework(ctx, homeworkID); err != nil {
		return fmt.Errorf("failed to delete homework record: %w", err)
	}

	return nil
}

// GetHomeworkByIDWithAccess получает домашнее задание с проверкой доступа
// Используется для скачивания файлов
func (s *HomeworkService) GetHomeworkByIDWithAccess(ctx context.Context, userID uuid.UUID, homeworkID uuid.UUID) (*models.LessonHomework, error) {
	// Получаем домашнее задание
	homework, err := s.homeworkRepo.GetHomeworkByID(ctx, homeworkID)
	if err != nil {
		return nil, err
	}

	// Проверяем доступ к уроку через GetHomeworkByLesson
	homeworks, err := s.GetHomeworkByLesson(ctx, userID, homework.LessonID)
	if err != nil {
		return nil, err
	}

	// Проверяем что homework находится в списке доступных
	for _, hw := range homeworks {
		if hw.ID == homeworkID {
			return homework, nil
		}
	}

	return nil, repository.ErrUnauthorized
}

// UpdateHomework обновляет текстовое описание домашнего задания
// Проверяет:
// - Существование домашнего задания
// - Права пользователя (admin, creator файла, или teacher урока)
func (s *HomeworkService) UpdateHomework(ctx context.Context, userID uuid.UUID, homeworkID uuid.UUID, textContent string) error {
	// Валидация
	req := &models.UpdateHomeworkRequest{TextContent: textContent}
	if err := req.Validate(); err != nil {
		return err
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем домашнее задание
	homework, err := s.homeworkRepo.GetHomeworkByID(ctx, homeworkID)
	if err != nil {
		return err // ErrHomeworkNotFound уже обработан в repo
	}

	// Получаем урок для проверки teacher
	lesson, err := s.lessonRepo.GetByID(ctx, homework.LessonID)
	if err != nil {
		return fmt.Errorf("failed to get lesson: %w", err)
	}

	// Проверка прав: admin, teacher, creator файла, или teacher урока
	isAdmin := user.IsAdmin()
	isTeacher := user.IsTeacher()
	isCreator := homework.CreatedBy == userID
	isLessonTeacher := lesson.TeacherID == userID

	if !isAdmin && !isTeacher && !isCreator && !isLessonTeacher {
		return repository.ErrUnauthorized
	}

	// Обновляем текстовое описание в БД
	if err := s.homeworkRepo.UpdateHomework(ctx, homeworkID, textContent); err != nil {
		return fmt.Errorf("failed to update homework record: %w", err)
	}

	return nil
}
