package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/telegram"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	// ErrTooManyFiles возвращается когда превышено максимальное количество файлов
	ErrTooManyFiles = errors.New("превышено максимальное количество файлов (максимум 10)")
	// ErrInvalidMessage возвращается когда сообщение невалидно
	ErrInvalidMessage = errors.New("сообщение должно быть от 1 до 4096 символов")
	// ErrTelegramSendFailed возвращается когда не удалось отправить через Telegram
	ErrTelegramSendFailed = errors.New("не удалось отправить сообщение через Telegram")
)

// LessonRepositoryInterface определяет методы для работы с уроками
type LessonRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error)
}

// LessonBroadcastService управляет рассылками по урокам
type LessonBroadcastService struct {
	db               *sqlx.DB
	broadcastRepo    repository.LessonBroadcastRepository
	lessonRepo       LessonRepositoryInterface
	userRepo         repository.UserRepository
	telegramUserRepo repository.TelegramUserRepository
	telegramClient   *telegram.Client
	uploadDir        string
	mu               sync.Mutex
}

// NewLessonBroadcastService создает новый LessonBroadcastService
func NewLessonBroadcastService(
	db *sqlx.DB,
	broadcastRepo repository.LessonBroadcastRepository,
	lessonRepo LessonRepositoryInterface,
	userRepo repository.UserRepository,
	telegramUserRepo repository.TelegramUserRepository,
	telegramClient *telegram.Client,
	uploadDir string,
) *LessonBroadcastService {
	// Создаём директорию для загрузок если её нет
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Failed to create upload directory %s: %v", uploadDir, err)
	}

	return &LessonBroadcastService{
		db:               db,
		broadcastRepo:    broadcastRepo,
		lessonRepo:       lessonRepo,
		userRepo:         userRepo,
		telegramUserRepo: telegramUserRepo,
		telegramClient:   telegramClient,
		uploadDir:        uploadDir,
	}
}

// CreateLessonBroadcast создает и отправляет рассылку по уроку
func (s *LessonBroadcastService) CreateLessonBroadcast(
	ctx context.Context,
	userID uuid.UUID,
	lessonID uuid.UUID,
	message string,
	files []*multipart.FileHeader,
) (*models.LessonBroadcast, error) {
	// Проверка наличия Telegram клиента
	if s.telegramClient == nil {
		log.Printf("Warning: Telegram client not configured, broadcast will be created but not sent")
	}

	// Валидация сообщения
	if len(message) < models.MinBroadcastMessageLen || len(message) > models.MaxBroadcastMessageLen {
		return nil, ErrInvalidMessage
	}

	// Валидация количества файлов
	if len(files) > models.MaxBroadcastFiles {
		return nil, ErrTooManyFiles
	}

	// Проверяем существование урока
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			return nil, repository.ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	// Проверяем права доступа: admin или teacher этого урока
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка прав: admin, методист или teacher этого урока
	if !user.IsAdmin() && !user.IsTeacher() && !(user.IsTeacher() && lesson.TeacherID == userID) {
		return nil, repository.ErrUnauthorized
	}

	// Создаём запись рассылки
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: userID,
		Message:  message,
		Status:   models.LessonBroadcastStatusPending,
	}

	createdBroadcast, err := s.broadcastRepo.CreateBroadcast(ctx, broadcast)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	// Сохраняем файлы
	if len(files) > 0 {
		if err := s.saveFiles(ctx, createdBroadcast.ID, files); err != nil {
			log.Printf("Failed to save files for broadcast %s: %v", createdBroadcast.ID, err)
			// Не прерываем процесс, рассылка всё равно будет отправлена
		}
	}

	// Загружаем файлы в объект для возврата
	savedFiles, err := s.broadcastRepo.GetBroadcastFiles(ctx, createdBroadcast.ID)
	if err == nil {
		createdBroadcast.Files = savedFiles
	}

	// Запускаем отправку в горутине (не ждём завершения)
	// Создаём отделённый контекст (не привязан к request) с timeout для долгоживущей операции
	// WithoutCancel отделяет от request context, но preserves values и позволяет явно отменить
	bgCtx := context.WithoutCancel(ctx)
	asyncCtx, cancel := context.WithTimeout(bgCtx, 30*time.Minute)
	go func() {
		defer cancel()
		s.SendBroadcastAsync(asyncCtx, createdBroadcast.ID)
	}()

	log.Printf("Lesson broadcast created: %s for lesson %s", createdBroadcast.ID, lessonID)
	return createdBroadcast, nil
}

// ListLessonBroadcasts получает все рассылки для урока
func (s *LessonBroadcastService) ListLessonBroadcasts(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonBroadcast, error) {
	// Проверяем существование урока
	_, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			return nil, repository.ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	broadcasts, err := s.broadcastRepo.ListBroadcastsByLesson(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts: %w", err)
	}

	return broadcasts, nil
}

// GetLessonBroadcast получает рассылку по ID с файлами
func (s *LessonBroadcastService) GetLessonBroadcast(ctx context.Context, broadcastID uuid.UUID) (*models.LessonBroadcast, error) {
	broadcast, err := s.broadcastRepo.GetBroadcast(ctx, broadcastID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonBroadcastNotFound) {
			return nil, repository.ErrLessonBroadcastNotFound
		}
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}

	return broadcast, nil
}

// GetBroadcastFileWithAccess получает файл рассылки с проверкой доступа
// Доступ имеют: admin, teacher урока, enrolled студенты
func (s *LessonBroadcastService) GetBroadcastFileWithAccess(
	ctx context.Context,
	userID uuid.UUID,
	fileID uuid.UUID,
) (*models.BroadcastFile, error) {
	// Получаем файл
	file, err := s.broadcastRepo.GetBroadcastFile(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("broadcast file not found")
	}

	// Получаем рассылку для проверки доступа
	broadcast, err := s.broadcastRepo.GetBroadcast(ctx, file.BroadcastID)
	if err != nil {
		return nil, fmt.Errorf("broadcast not found")
	}

	// Получаем урок для проверки прав доступа
	lesson, err := s.lessonRepo.GetByID(ctx, broadcast.LessonID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			return nil, repository.ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	// Получаем информацию о пользователе
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка прав доступа
	hasAccess := false

	// Admin имеет полный доступ
	if user.IsAdmin() {
		hasAccess = true
	}

	// Teacher урока имеет доступ
	if !hasAccess && user.IsTeacher() && lesson.TeacherID == userID {
		hasAccess = true
	}

	// Студент имеет доступ, если enrolled в урок
	if !hasAccess && user.IsStudent() {
		enrolledStudents, err := s.getEnrolledStudents(ctx, lesson.ID)
		if err == nil {
			for _, student := range enrolledStudents {
				if student.ID == userID {
					hasAccess = true
					break
				}
			}
		}
	}

	if !hasAccess {
		return nil, repository.ErrUnauthorized
	}

	return file, nil
}

// SendBroadcastAsync отправляет рассылку асинхронно (в горутине)
func (s *LessonBroadcastService) SendBroadcastAsync(ctx context.Context, broadcastID uuid.UUID) {
	// Защита от panic в goroutine
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in SendBroadcastAsync for broadcast %s: %v", broadcastID, r)
		}
	}()

	// Проверка что Telegram настроен
	if s.telegramClient == nil {
		log.Printf("Telegram client not configured, cannot send broadcast %s", broadcastID)
		return
	}

	// Получаем рассылку
	broadcast, err := s.broadcastRepo.GetBroadcast(ctx, broadcastID)
	if err != nil {
		log.Printf("Failed to get broadcast %s: %v", broadcastID, err)
		return
	}

	// Проверяем статус (должен быть pending)
	if !broadcast.IsPending() {
		log.Printf("Broadcast %s is not pending (status: %s)", broadcastID, broadcast.Status)
		return
	}

	// Обновляем статус на sending
	if err := s.broadcastRepo.UpdateBroadcastStatus(ctx, broadcastID, models.LessonBroadcastStatusSending, 0, 0); err != nil {
		log.Printf("Failed to update broadcast status to sending: %v", err)
		// Продолжаем отправку даже если не удалось обновить статус
	}

	// Получаем enrolled студентов из bookings
	students, err := s.getEnrolledStudents(ctx, broadcast.LessonID)
	if err != nil {
		log.Printf("Failed to get enrolled students for lesson %s: %v", broadcast.LessonID, err)
		s.finalizeBroadcast(ctx, broadcastID, 0, 0, models.LessonBroadcastStatusFailed)
		return
	}

	if len(students) == 0 {
		log.Printf("No enrolled students for lesson %s", broadcast.LessonID)
		s.finalizeBroadcast(ctx, broadcastID, 0, 0, models.LessonBroadcastStatusCompleted)
		return
	}

	// Получаем информацию о уроке для сообщения
	lesson, err := s.lessonRepo.GetByID(ctx, broadcast.LessonID)
	if err != nil {
		log.Printf("Failed to get lesson %s: %v", broadcast.LessonID, err)
		s.finalizeBroadcast(ctx, broadcastID, 0, 0, models.LessonBroadcastStatusFailed)
		return
	}

	// Формируем текст сообщения
	messageText := fmt.Sprintf("📢 Сообщение от преподавателя по уроку %s:\n\n%s",
		lesson.StartTime.Format("02.01.2006 15:04"),
		broadcast.Message)

	var sentCount, failedCount, skippedCount int
	var mu sync.Mutex

	// Отправляем каждому студенту
	for _, student := range students {
		// Получаем Telegram привязку
		telegramUser, err := s.telegramUserRepo.GetByUserID(ctx, student.ID)
		if err != nil {
			mu.Lock()
			skippedCount++
			mu.Unlock()
			log.Printf("Warning: Student %s (%s) not linked to Telegram - skipping broadcast", student.ID, student.Email)
			continue
		}

		// Проверяем подписку на уведомления
		if !telegramUser.Subscribed {
			mu.Lock()
			skippedCount++
			mu.Unlock()
			log.Printf("Warning: Student %s (%s) unsubscribed from notifications - skipping broadcast", student.ID, student.Email)
			continue
		}

		// Отправляем текстовое сообщение
		if err := s.sendMessage(telegramUser.ChatID, messageText); err != nil {
			mu.Lock()
			failedCount++
			mu.Unlock()
			log.Printf("Failed to send message to student %s: %v", student.ID, err)
			continue
		}

		// Отправляем файлы (если есть)
		if len(broadcast.Files) > 0 {
			log.Printf("Sending %d files to student %s", len(broadcast.Files), student.ID)

			// Если файлов несколько (2-10), используем SendMediaGroup для оптимизации
			if len(broadcast.Files) > 1 {
				mediaFiles := make([]telegram.MediaFile, len(broadcast.Files))
				for i, file := range broadcast.Files {
					filePath := filepath.Join(s.uploadDir, file.FilePath)
					mediaFiles[i] = telegram.MediaFile{
						FileName: file.FileName,
						FilePath: filePath,
					}
				}

				// Отправляем группу файлов
				sentCount, err := s.sendMediaGroup(telegramUser.ChatID, mediaFiles)
				if err != nil {
					log.Printf("Failed to send media group to student %s: %v (sent %d/%d files)",
						student.ID, err, sentCount, len(mediaFiles))
					// Не критичная ошибка - сообщение уже доставлено
				} else {
					log.Printf("Media group sent to student %s: %d files delivered", student.ID, sentCount)
				}
			} else {
				// Если файл один, отправляем обычным способом
				for _, file := range broadcast.Files {
					if err := s.sendFile(telegramUser.ChatID, file); err != nil {
						log.Printf("Failed to send file %s to student %s: %v", file.FileName, student.ID, err)
						// Не считаем критичной ошибкой если не отправился файл, главное - сообщение доставлено
					} else {
						log.Printf("File %s sent to student %s", file.FileName, student.ID)
					}
				}
			}
		}

		mu.Lock()
		sentCount++
		mu.Unlock()
		log.Printf("Broadcast sent to student %s (%s)", student.ID, student.Email)
	}

	// Завершаем рассылку
	status := models.LessonBroadcastStatusCompleted
	if failedCount > 0 && sentCount == 0 {
		status = models.LessonBroadcastStatusFailed
	}

	s.finalizeBroadcast(ctx, broadcastID, sentCount, failedCount, status)
	log.Printf("Broadcast %s completed: sent=%d, failed=%d, skipped=%d (total students=%d)",
		broadcastID, sentCount, failedCount, skippedCount, len(students))
}

// getEnrolledStudents получает всех активных студентов урока через bookings
func (s *LessonBroadcastService) getEnrolledStudents(ctx context.Context, lessonID uuid.UUID) ([]*models.User, error) {
	// Проверка на nil db (для unit тестов)
	if s.db == nil {
		return []*models.User{}, nil
	}

	query := `
		SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.updated_at
		FROM users u
		INNER JOIN bookings b ON u.id = b.student_id
		WHERE b.lesson_id = $1
		  AND b.status = 'active'
		  AND u.deleted_at IS NULL
		ORDER BY u.first_name, u.last_name
	`

	var students []*models.User
	if err := s.db.SelectContext(ctx, &students, query, lessonID); err != nil {
		return nil, fmt.Errorf("failed to query enrolled students: %w", err)
	}

	return students, nil
}

// sendMessage отправляет текстовое сообщение через Telegram
func (s *LessonBroadcastService) sendMessage(chatID int64, message string) error {
	if s.telegramClient == nil {
		return fmt.Errorf("telegram client not configured")
	}

	// Retry логика для обработки rate limits
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := s.telegramClient.SendMessage(chatID, message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Проверяем тип ошибки
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			switch telegramErr.ErrorCode {
			case 403:
				// Forbidden - бот заблокирован, не ретраим
				return err
			case 400:
				// Bad Request - невалидные данные, не ретраим
				return err
			case 429:
				// Too Many Requests - exponential backoff
				if attempt < maxRetries-1 {
					backoff := time.Duration(1<<attempt) * time.Second
					log.Printf("Rate limit hit, backing off for %v", backoff)
					time.Sleep(backoff)
					continue
				}
			}
		}

		// Для других ошибок (network errors) - retry
		if attempt < maxRetries-1 {
			time.Sleep(time.Second)
			continue
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// sendFile отправляет файл через Telegram
func (s *LessonBroadcastService) sendFile(chatID int64, file *models.BroadcastFile) error {
	if s.telegramClient == nil {
		return fmt.Errorf("telegram client not configured")
	}

	// Полный путь к файлу на диске
	filePath := filepath.Join(s.uploadDir, file.FilePath)

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Отправляем документ через Telegram
	if err := s.telegramClient.SendDocument(chatID, file.FileName, filePath); err != nil {
		return fmt.Errorf("failed to send document: %w", err)
	}

	return nil
}

// sendMediaGroup отправляет группу файлов через Telegram (оптимизировано для 2-10 файлов)
func (s *LessonBroadcastService) sendMediaGroup(chatID int64, files []telegram.MediaFile) (int, error) {
	if s.telegramClient == nil {
		return 0, fmt.Errorf("telegram client not configured")
	}

	// Проверяем существование всех файлов перед отправкой
	for _, file := range files {
		if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", file.FilePath)
		}
	}

	// Отправляем медиа группу через Telegram
	// Возвращает количество успешно отправленных файлов
	sentCount, err := s.telegramClient.SendMediaGroup(chatID, files)
	if err != nil {
		return sentCount, fmt.Errorf("failed to send media group: %w", err)
	}

	return sentCount, nil
}

// saveFiles сохраняет файлы на диск и создаёт записи в БД
func (s *LessonBroadcastService) saveFiles(ctx context.Context, broadcastID uuid.UUID, files []*multipart.FileHeader) error {
	for _, fileHeader := range files {
		// Открываем файл
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
		}
		defer file.Close()

		// Проверяем размер файла
		if fileHeader.Size > models.MaxBroadcastFileSize {
			return fmt.Errorf("file %s exceeds maximum size of %d bytes", fileHeader.Filename, models.MaxBroadcastFileSize)
		}

		// Генерируем уникальное имя файла
		fileID := uuid.New()
		ext := filepath.Ext(fileHeader.Filename)
		savedFileName := fmt.Sprintf("%s%s", fileID.String(), ext)
		savedFilePath := filepath.Join(s.uploadDir, savedFileName)

		// Сохраняем файл на диск
		dst, err := os.Create(savedFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", savedFilePath, err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return fmt.Errorf("failed to save file %s: %w", savedFilePath, err)
		}

		// Создаём запись в БД
		broadcastFile := &models.BroadcastFile{
			BroadcastID: broadcastID,
			FileName:    fileHeader.Filename,
			FilePath:    savedFileName, // Относительный путь
			FileSize:    fileHeader.Size,
			MimeType:    fileHeader.Header.Get("Content-Type"),
		}

		if err := s.broadcastRepo.AddBroadcastFile(ctx, broadcastFile); err != nil {
			return fmt.Errorf("failed to add file record: %w", err)
		}

		log.Printf("File saved: %s -> %s", fileHeader.Filename, savedFileName)
	}

	return nil
}

// finalizeBroadcast завершает рассылку, обновляя счетчики и статус
func (s *LessonBroadcastService) finalizeBroadcast(
	ctx context.Context,
	broadcastID uuid.UUID,
	sentCount, failedCount int,
	status string,
) {
	if err := s.broadcastRepo.UpdateBroadcastStatus(ctx, broadcastID, status, sentCount, failedCount); err != nil {
		log.Printf("Failed to finalize broadcast %s: %v", broadcastID, err)
	}
}
