package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// LessonBroadcastServiceInterface определяет интерфейс для работы с рассылками по занятиям
type LessonBroadcastServiceInterface interface {
	CreateLessonBroadcast(ctx context.Context, userID uuid.UUID, lessonID uuid.UUID, message string, files []*multipart.FileHeader) (*models.LessonBroadcast, error)
	ListLessonBroadcasts(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonBroadcast, error)
	GetLessonBroadcast(ctx context.Context, broadcastID uuid.UUID) (*models.LessonBroadcast, error)
	GetBroadcastFileWithAccess(ctx context.Context, userID uuid.UUID, fileID uuid.UUID) (*models.BroadcastFile, error)
}

// LessonBroadcastHandler обрабатывает HTTP запросы для рассылок по занятиям
type LessonBroadcastHandler struct {
	broadcastService LessonBroadcastServiceInterface
	uploadDir        string
}

// NewLessonBroadcastHandler создает новый LessonBroadcastHandler
func NewLessonBroadcastHandler(broadcastService LessonBroadcastServiceInterface, uploadDir string) *LessonBroadcastHandler {
	return &LessonBroadcastHandler{
		broadcastService: broadcastService,
		uploadDir:        uploadDir,
	}
}

// CreateBroadcast обрабатывает POST /api/v1/lessons/:id/broadcasts
// Создает и отправляет рассылку для студентов урока
func (h *LessonBroadcastHandler) CreateBroadcast(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя из контекста
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Парсим lesson ID из URL
	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Парсим multipart form (максимум 110MB для 10 файлов по 10MB + текст + метаданные)
	if err := r.ParseMultipartForm(110 * 1024 * 1024); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Failed to parse multipart form")
		return
	}

	// Получаем текстовое сообщение
	message := r.FormValue("message")
	if message == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message is required")
		return
	}

	// Валидация длины сообщения
	if len(message) < models.MinBroadcastMessageLen {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message is required")
		return
	}
	if len(message) > models.MaxBroadcastMessageLen {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message must not exceed 4096 characters")
		return
	}

	// Получаем файлы (опционально)
	var files []*multipart.FileHeader
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if fileHeaders, ok := r.MultipartForm.File["files"]; ok {
			files = fileHeaders
		}
	}

	// Валидация количества файлов
	if len(files) > models.MaxBroadcastFiles {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Maximum 10 files allowed")
		return
	}

	// Валидация размера каждого файла
	for _, fileHeader := range files {
		if fileHeader.Size > models.MaxBroadcastFileSize {
			response.BadRequest(w, response.ErrCodeValidationFailed, "File "+fileHeader.Filename+" exceeds maximum size of 10MB")
			return
		}
	}

	// Создаем рассылку через сервис
	broadcast, err := h.broadcastService.CreateLessonBroadcast(
		r.Context(),
		user.ID,
		lessonID,
		message,
		files,
	)
	if err != nil {
		h.handleBroadcastError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"broadcast": broadcast,
		"message":   "Broadcast created and sending started",
	})
}

// ListBroadcasts обрабатывает GET /api/v1/lessons/:id/broadcasts
// Возвращает список всех рассылок для урока с фильтрацией
func (h *LessonBroadcastHandler) ListBroadcasts(w http.ResponseWriter, r *http.Request) {
	// Проверяем аутентификацию
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Парсим lesson ID из URL
	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Парсим query параметры для пагинации
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		} else {
			response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid limit parameter")
			return
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		} else {
			response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid offset parameter")
			return
		}
	}

	// Парсим опциональный фильтр по статусу
	statusFilter := r.URL.Query().Get("status")
	if statusFilter != "" {
		// Валидируем статус если указан
		validStatuses := map[string]bool{
			models.LessonBroadcastStatusPending:   true,
			models.LessonBroadcastStatusSending:   true,
			models.LessonBroadcastStatusCompleted: true,
			models.LessonBroadcastStatusFailed:    true,
		}
		if !validStatuses[statusFilter] {
			response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid status filter")
			return
		}
	}

	// Получаем список рассылок
	broadcasts, err := h.broadcastService.ListLessonBroadcasts(r.Context(), lessonID)
	if err != nil {
		h.handleBroadcastError(w, err)
		return
	}

	// Фильтруем по статусу если указан
	var filteredBroadcasts []*models.LessonBroadcast
	if statusFilter != "" {
		for _, b := range broadcasts {
			if b.Status == statusFilter {
				filteredBroadcasts = append(filteredBroadcasts, b)
			}
		}
	} else {
		filteredBroadcasts = broadcasts
	}

	// Применяем пагинацию
	totalCount := len(filteredBroadcasts)
	start := offset
	end := offset + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	paginatedBroadcasts := filteredBroadcasts[start:end]

	response.OK(w, map[string]interface{}{
		"broadcasts": paginatedBroadcasts,
		"count":      len(paginatedBroadcasts),
		"total":      totalCount,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetBroadcast обрабатывает GET /api/v1/lessons/:id/broadcasts/:broadcast_id
// Возвращает детали конкретной рассылки
func (h *LessonBroadcastHandler) GetBroadcast(w http.ResponseWriter, r *http.Request) {
	// Проверяем аутентификацию
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Парсим lesson ID из URL
	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Парсим broadcast ID из URL
	broadcastID, err := uuid.Parse(chi.URLParam(r, "broadcast_id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid broadcast ID")
		return
	}

	// Получаем рассылку
	broadcast, err := h.broadcastService.GetLessonBroadcast(r.Context(), broadcastID)
	if err != nil {
		h.handleBroadcastError(w, err)
		return
	}

	// Проверяем что рассылка принадлежит указанному уроку
	if broadcast.LessonID != lessonID {
		response.NotFound(w, "Broadcast not found for this lesson")
		return
	}

	response.OK(w, map[string]interface{}{
		"broadcast": broadcast,
	})
}

// DownloadBroadcastFile обрабатывает GET /api/v1/lessons/:id/broadcasts/:broadcast_id/files/:file_id/download
// Скачивает файл рассылки с проверкой прав доступа
func (h *LessonBroadcastHandler) DownloadBroadcastFile(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя из контекста
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Парсим lesson ID из URL (для валидации маршрута)
	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Парсим broadcast ID из URL (для валидации маршрута)
	broadcastID, err := uuid.Parse(chi.URLParam(r, "broadcast_id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid broadcast ID")
		return
	}

	// Парсим file ID из URL
	fileID, err := uuid.Parse(chi.URLParam(r, "file_id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid file ID")
		return
	}

	// Получаем файл с проверкой прав доступа через сервис
	file, err := h.broadcastService.GetBroadcastFileWithAccess(r.Context(), user.ID, fileID)
	if err != nil {
		// Обработка специфичных ошибок
		if err.Error() == "broadcast file not found" {
			response.NotFound(w, "Broadcast file not found")
			return
		}
		if err.Error() == "broadcast not found" {
			response.NotFound(w, "Broadcast not found")
			return
		}
		if errors.Is(err, repository.ErrLessonNotFound) {
			response.NotFound(w, "Lesson not found")
			return
		}
		if errors.Is(err, repository.ErrUnauthorized) {
			response.Forbidden(w, "You don't have permission to download this file")
			return
		}
		response.InternalError(w, "Failed to get broadcast file")
		return
	}

	// Проверяем что файл принадлежит указанной рассылке (защита от path traversal)
	if file.BroadcastID != broadcastID {
		response.NotFound(w, "File not found in this broadcast")
		return
	}

	// Получаем рассылку для проверки lesson_id (защита от path traversal)
	broadcast, err := h.broadcastService.GetLessonBroadcast(r.Context(), broadcastID)
	if err != nil {
		response.NotFound(w, "Broadcast not found")
		return
	}
	if broadcast.LessonID != lessonID {
		response.NotFound(w, "Broadcast not found for this lesson")
		return
	}

	// Полный путь к файлу на диске
	fullPath := filepath.Join(h.uploadDir, file.FilePath)

	// Проверяем существование файла в файловой системе
	if _, err := os.Stat(fullPath); err != nil {
		log.Printf("File not found on disk: %s (error: %v)", fullPath, err)
		response.NotFound(w, "File not found on server")
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.FileName))
	w.Header().Set("Content-Type", file.MimeType)

	// Отправляем файл
	http.ServeFile(w, r, fullPath)
}

// handleBroadcastError обрабатывает специфичные ошибки рассылок
func (h *LessonBroadcastHandler) handleBroadcastError(w http.ResponseWriter, err error) {
	// Логируем ошибку
	log.Printf("Broadcast error: %v", err)

	// Проверяем ошибки repository
	if errors.Is(err, repository.ErrLessonNotFound) {
		response.NotFound(w, "Lesson not found")
		return
	}
	if errors.Is(err, repository.ErrLessonBroadcastNotFound) {
		response.NotFound(w, "Broadcast not found")
		return
	}
	if errors.Is(err, repository.ErrUnauthorized) {
		response.Forbidden(w, "You don't have permission to broadcast to this lesson")
		return
	}

	// Проверяем ошибки валидации
	if errors.Is(err, service.ErrInvalidMessage) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message must be between 1 and 4096 characters")
		return
	}
	if errors.Is(err, service.ErrTooManyFiles) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Maximum 10 files allowed")
		return
	}
	if errors.Is(err, models.ErrInvalidBroadcastMessage) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid broadcast message")
		return
	}
	if errors.Is(err, models.ErrBroadcastMessageTooLong) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message must not exceed 4096 characters")
		return
	}

	// Ошибки файлов
	if errors.Is(err, models.ErrTooManyFiles) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Maximum 10 files allowed")
		return
	}
	if errors.Is(err, models.ErrInvalidFileSize) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "File size must not exceed 10MB")
		return
	}

	// Неизвестная ошибка
	response.InternalError(w, "Failed to process broadcast request")
}
