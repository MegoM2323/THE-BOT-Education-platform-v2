package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// HomeworkHandler обрабатывает HTTP запросы для домашних заданий
type HomeworkHandler struct {
	homeworkService *service.HomeworkService
}

// NewHomeworkHandler создает новый HomeworkHandler
func NewHomeworkHandler(homeworkService *service.HomeworkService) *HomeworkHandler {
	return &HomeworkHandler{
		homeworkService: homeworkService,
	}
}

// validateFilePath проверяет что путь безопасен для чтения
// - должен находиться в разрешенной директории
// - не должен содержать path traversal символы
// - не должен быть абсолютным путем вне разрешенной директории
func validateFilePath(filePath string, allowedDir string) (bool, string) {
	// Проверка пустого пути
	if filePath == "" {
		return false, "empty file path"
	}

	// Проверка абсолютного пути (не должно начинаться с /)
	if strings.HasPrefix(filePath, "/") {
		return false, "absolute paths not allowed"
	}

	// Проверка на явные path traversal попытки
	if strings.Contains(filePath, "..") {
		return false, "path traversal detected"
	}

	// Проверка на null bytes
	if strings.Contains(filePath, "\x00") {
		return false, "null bytes detected"
	}

	// Нормализуем пути для сравнения
	normalizedPath, err := filepath.Abs(filePath)
	if err != nil {
		return false, "invalid file path"
	}

	normalizedAllowed, err := filepath.Abs(allowedDir)
	if err != nil {
		return false, "invalid allowed directory"
	}

	// Проверяем что путь находится в разрешенной директории
	// добавляем separator чтобы "uploads/home_test" не прошел когда разрешена "uploads/home"
	rel, err := filepath.Rel(normalizedAllowed, normalizedPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false, "path outside allowed directory"
	}

	return true, ""
}

// sanitizeFileName проверяет что имя файла содержит только разрешенные символы
// разрешены: буквы, цифры, точки, дефисы, подчеркивания, пробелы
func sanitizeFileName(fileName string) (bool, string) {
	if fileName == "" {
		return false, "empty file name"
	}

	// Разрешенные символы: буквы (a-z, A-Z, кириллица), цифры, точки, дефисы, подчеркивания, пробелы
	// Используем более строгую regex для дополнительной безопасности
	allowedPattern := regexp.MustCompile(`^[a-zA-Z0-9а-яА-Я\s._-]+$`)

	if !allowedPattern.MatchString(fileName) {
		return false, "file name contains disallowed characters"
	}

	return true, ""
}

// UploadHomework загружает файл домашнего задания для урока
// POST /api/v1/lessons/:id/homework
// Требует авторизации: admin или methodologist урока
// Multipart form data с полем "file"
// Максимальный размер файла: 10MB
func (h *HomeworkHandler) UploadHomework(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	// Получаем ID урока из URL параметра
	lessonID := chi.URLParam(r, "id")
	parsedLessonID, err := uuid.Parse(lessonID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Парсим multipart form (макс 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Failed to parse form: file size may exceed 10MB")
		return
	}

	// Получаем файл из формы
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Failed to read file from form")
		return
	}
	defer file.Close()

	// Проверка размера файла (10MB = 10485760 bytes)
	if fileHeader.Size <= 0 || fileHeader.Size > 10485760 {
		response.BadRequest(w, response.ErrCodeValidationFailed, "File size must be between 1 byte and 10MB")
		return
	}

	// Получаем MIME тип
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Missing Content-Type header")
		return
	}

	// Создаем запрос для сервиса
	req := &models.CreateHomeworkRequest{
		LessonID: parsedLessonID,
		FileName: fileHeader.Filename,
		FileSize: fileHeader.Size,
		MimeType: mimeType,
	}

	// Вызываем сервис для создания домашнего задания
	homework, err := h.homeworkService.CreateHomework(ctx, session.UserID, file, req)
	if err != nil {
		// Обработка специфичных ошибок
		switch err {
		case models.ErrInvalidLessonID:
			response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid lesson ID")
		case models.ErrInvalidFileName, models.ErrFileNameTooLong:
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
		case models.ErrInvalidFileSize:
			response.BadRequest(w, response.ErrCodeValidationFailed, "File size must be between 1 byte and 10MB")
		case models.ErrInvalidMimeType, models.ErrMimeTypeNotAllowed:
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
		case models.ErrFileStorageFailed:
			response.BadRequest(w, response.ErrCodeInternalError, "Failed to save file")
		default:
			// Проверка на ошибки из repository
			if errors.Is(err, repository.ErrLessonNotFound) {
				response.NotFound(w, "Lesson not found")
			} else if errors.Is(err, repository.ErrUnauthorized) {
				response.Forbidden(w, "You don't have permission to upload homework for this lesson")
			} else {
				response.InternalError(w, "Failed to upload homework")
			}
		}
		return
	}

	response.Success(w, http.StatusCreated, homework)
}

// GetHomework получает список домашних заданий для урока
// GET /api/v1/lessons/:id/homework
// Требует авторизации
// Доступ:
// - Admin: видит все ДЗ всех уроков
// - Methodologist: видит ДЗ только своих уроков
// - Student: видит ДЗ групповых уроков и индивидуальных уроков, на которые записан
func (h *HomeworkHandler) GetHomework(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	// Получаем ID урока из URL параметра
	lessonID := chi.URLParam(r, "id")
	parsedLessonID, err := uuid.Parse(lessonID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Вызываем сервис для получения списка домашних заданий
	homeworks, err := h.homeworkService.GetHomeworkByLesson(ctx, session.UserID, parsedLessonID)
	if err != nil {
		// Обработка специфичных ошибок
		if errors.Is(err, repository.ErrLessonNotFound) {
			response.NotFound(w, "Lesson not found")
		} else if errors.Is(err, repository.ErrUnauthorized) {
			response.Forbidden(w, "You don't have permission to view homework for this lesson")
		} else {
			response.InternalError(w, "Failed to get homework")
		}
		return
	}

	// Если список пуст, возвращаем пустой массив
	if homeworks == nil {
		homeworks = []*models.LessonHomework{}
	}

	response.Success(w, http.StatusOK, homeworks)
}

// DeleteHomework удаляет файл домашнего задания
// DELETE /api/v1/lessons/:id/homework/:file_id
// Требует авторизации: admin, creator файла, или methodologist урока
// Удаляет файл из файловой системы и запись из БД
func (h *HomeworkHandler) DeleteHomework(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	// Получаем ID урока из URL параметра
	lessonID := chi.URLParam(r, "id")
	_, err := uuid.Parse(lessonID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Получаем ID файла домашнего задания из URL параметра
	fileID := chi.URLParam(r, "file_id")
	parsedFileID, err := uuid.Parse(fileID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid file ID")
		return
	}

	// Вызываем сервис для удаления домашнего задания
	if err := h.homeworkService.DeleteHomework(ctx, session.UserID, parsedFileID); err != nil {
		// Обработка специфичных ошибок
		if errors.Is(err, repository.ErrHomeworkNotFound) {
			response.NotFound(w, "Homework file not found")
		} else if errors.Is(err, repository.ErrUnauthorized) {
			response.Forbidden(w, "You don't have permission to delete this homework file")
		} else {
			response.InternalError(w, "Failed to delete homework")
		}
		return
	}

	// Успешное удаление - возвращаем пустой success response
	response.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Homework file deleted successfully",
	})
}

// DownloadHomework скачивает файл домашнего задания
// GET /api/v1/lessons/:id/homework/:file_id/download
// Требует авторизации
// Доступ: те же правила что и для GetHomework
func (h *HomeworkHandler) DownloadHomework(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	// Получаем ID урока из URL параметра
	lessonID := chi.URLParam(r, "id")
	_, err := uuid.Parse(lessonID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Получаем ID файла домашнего задания из URL параметра
	fileID := chi.URLParam(r, "file_id")
	parsedFileID, err := uuid.Parse(fileID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid file ID")
		return
	}

	// Вызываем сервис для получения домашнего задания с проверкой доступа
	homework, err := h.homeworkService.GetHomeworkByIDWithAccess(ctx, session.UserID, parsedFileID)
	if err != nil {
		// Обработка специфичных ошибок
		if errors.Is(err, repository.ErrHomeworkNotFound) {
			response.NotFound(w, "Homework file not found")
		} else if errors.Is(err, repository.ErrUnauthorized) {
			response.Forbidden(w, "You don't have permission to download this homework file")
		} else {
			response.InternalError(w, "Failed to get homework")
		}
		return
	}

	// SECURITY: Валидируем путь до файла (защита от path traversal)
	// даже если FilePath скомпрометирован в БД или через SQL injection
	allowedDir := "uploads/homework"
	isValid, errMsg := validateFilePath(homework.FilePath, allowedDir)
	if !isValid {
		response.BadRequest(w, response.ErrCodeValidationFailed, fmt.Sprintf("Invalid file path: %s", errMsg))
		return
	}

	// SECURITY: Валидируем имя файла
	isValidFileName, errMsg := sanitizeFileName(homework.FileName)
	if !isValidFileName {
		response.BadRequest(w, response.ErrCodeValidationFailed, fmt.Sprintf("Invalid file name: %s", errMsg))
		return
	}

	// Проверяем что файл существует в файловой системе
	if _, err := os.Stat(homework.FilePath); err != nil {
		response.NotFound(w, "File not found on server")
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", homework.FileName))
	w.Header().Set("Content-Type", homework.MimeType)

	// Отправляем файл
	http.ServeFile(w, r, homework.FilePath)
}

// UpdateHomework обновляет описание домашнего задания
// PATCH /api/v1/lessons/:id/homework/:file_id
// Требует авторизации: admin, creator файла, или methodologist урока
// Принимает JSON с полем text_content
func (h *HomeworkHandler) UpdateHomework(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	// Получаем ID урока из URL параметра
	lessonID := chi.URLParam(r, "id")
	_, err := uuid.Parse(lessonID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Получаем ID файла домашнего задания из URL параметра
	fileID := chi.URLParam(r, "file_id")
	parsedFileID, err := uuid.Parse(fileID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid file ID")
		return
	}

	// Парсим JSON запрос
	var req models.UpdateHomeworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON format")
		return
	}
	defer r.Body.Close()

	// Вызываем сервис для обновления домашнего задания
	if err := h.homeworkService.UpdateHomework(ctx, session.UserID, parsedFileID, req.TextContent); err != nil {
		// Обработка специфичных ошибок
		if errors.Is(err, repository.ErrHomeworkNotFound) {
			response.NotFound(w, "Homework file not found")
		} else if errors.Is(err, repository.ErrUnauthorized) {
			response.Forbidden(w, "You don't have permission to update this homework file")
		} else if strings.Contains(err.Error(), "validation") {
			response.BadRequest(w, response.ErrCodeValidationFailed, "Text content is too long (max 10000 characters)")
		} else {
			response.InternalError(w, "Failed to update homework")
		}
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"message": "Homework description updated successfully",
	})
}
