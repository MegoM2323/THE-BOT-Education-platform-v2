package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
	"tutoring-platform/pkg/response"
)

// MethodologistHandler обрабатывает эндпоинты для методистов
type MethodologistHandler struct {
	lessonService          *service.LessonService
	bookingService         *service.BookingService
	lessonBroadcastService *service.LessonBroadcastService
	lessonRepo             *repository.LessonRepository
	timeValidator          *validator.TimeValidator
}

// NewMethodologistHandler создает новый MethodologistHandler
func NewMethodologistHandler(
	lessonService *service.LessonService,
	bookingService *service.BookingService,
	lessonBroadcastService *service.LessonBroadcastService,
	lessonRepo *repository.LessonRepository,
) *MethodologistHandler {
	return &MethodologistHandler{
		lessonService:          lessonService,
		bookingService:         bookingService,
		lessonBroadcastService: lessonBroadcastService,
		lessonRepo:             lessonRepo,
		timeValidator:          validator.NewTimeValidator(),
	}
}

// SendLessonBroadcast обрабатывает POST /api/v1/methodologist/lessons/:id/broadcast
// Отправляет рассылку всем студентам занятия через Telegram
func (h *MethodologistHandler) SendLessonBroadcast(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя (методист) из контекста
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем что пользователь не удален/деактивирован
	if user.IsDeleted() {
		response.Unauthorized(w, "User account is deleted or deactivated")
		return
	}

	// Проверяем что пользователь - методист или админ
	if !user.IsMethodologist() && !user.IsAdmin() {
		response.Forbidden(w, "Only methodologists and admins can send lesson broadcasts")
		return
	}

	// Извлекаем lesson_id из URL
	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Получаем занятие из БД
	lesson, err := h.lessonService.GetLesson(r.Context(), lessonID)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			response.NotFound(w, "Lesson not found")
			return
		}
		log.Printf("[ERROR] Failed to retrieve lesson %s: %v\n", lessonID, err)
		response.InternalError(w, "Failed to retrieve lesson")
		return
	}

	// Проверяем что текущий методист может отправлять рассылку для этого занятия
	// (методист может отправлять для всех занятий, или ограничить по факультету/группе в будущем)
	if !user.IsAdmin() && lesson.TeacherID != user.ID {
		response.Forbidden(w, "You can only send broadcasts for your own lessons")
		return
	}

	// Декодируем request body
	var req models.TeacherLessonBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация сообщения
	if err := req.Validate(); err != nil {
		h.handleMethodologistBroadcastError(w, err)
		return
	}

	// Создаем рассылку урока с использованием LessonBroadcastService
	// Этот сервис сам получит enrolled студентов и отправит им рассылку
	broadcast, err := h.lessonBroadcastService.CreateLessonBroadcast(
		r.Context(),
		user.ID,  // sender - текущий методист
		lessonID, // lesson ID
		req.Message,
		nil, // files - пока не поддерживаем через этот endpoint
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create lesson broadcast for lesson %s: %v\n", lessonID, err)
		h.handleMethodologistBroadcastError(w, err)
		return
	}

	log.Printf("[INFO] Lesson broadcast %s created and sending started for lesson %s\n", broadcast.ID, lessonID)

	// Возвращаем успешный ответ с информацией о запущенной рассылке
	response.Created(w, map[string]interface{}{
		"broadcast_id": broadcast.ID,
		"message":      "Broadcast started successfully via Telegram",
		"status":       broadcast.Status,
	})
}

// handleMethodologistBroadcastError обрабатывает ошибки рассылки методиста
func (h *MethodologistHandler) handleMethodologistBroadcastError(w http.ResponseWriter, err error) {
	// Проверяем ошибки lesson broadcast service
	if errors.Is(err, service.ErrInvalidMessage) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast message must be between 1 and 4096 characters")
		return
	}
	if errors.Is(err, service.ErrTooManyFiles) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Too many files (maximum 10)")
		return
	}

	// Проверяем специфичные ошибки models (валидация)
	if errors.Is(err, models.ErrInvalidBroadcastMessage) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast message is required")
		return
	}
	if errors.Is(err, models.ErrBroadcastMessageTooLong) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast message must not exceed 4096 characters")
		return
	}

	// Проверяем repository ошибки
	if errors.Is(err, repository.ErrLessonNotFound) {
		response.NotFound(w, "Lesson not found")
		return
	}
	if errors.Is(err, repository.ErrUnauthorized) {
		response.Forbidden(w, "You are not authorized to broadcast for this lesson")
		return
	}
	if errors.Is(err, repository.ErrUserNotFound) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "User not found")
		return
	}

	// Неизвестная ошибка
	log.Printf("[ERROR] Unhandled methodologist broadcast error: %v\n", err)
	response.InternalError(w, "An error occurred processing your request")
}

// ptrMethodologistBookingStatus вспомогательная функция для создания указателя на BookingStatus
func ptrMethodologistBookingStatus(status models.BookingStatus) *models.BookingStatus {
	return &status
}

// GetMethodologistSchedule обрабатывает GET /api/v1/methodologist/schedule
// Возвращает расписание методиста в формате календаря с дополнительной информацией
func (h *MethodologistHandler) GetMethodologistSchedule(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем что пользователь не удален/деактивирован
	if user.IsDeleted() {
		response.Unauthorized(w, "User account is deleted or deactivated")
		return
	}

	// Проверяем что пользователь - методист или админ
	if !user.IsMethodologist() && !user.IsAdmin() {
		response.Forbidden(w, "Only methodologists and admins can access methodologist schedule")
		return
	}

	// Парсим query параметры
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	// Если start_date не указан - начало текущей недели (понедельник)
	if startDateStr == "" {
		now := time.Now()
		// Получаем день недели (0 = Sunday, 1 = Monday)
		weekday := now.Weekday()
		// Количество дней до понедельника
		daysToMonday := int(weekday) - 1
		if weekday == time.Sunday {
			daysToMonday = 6 // Воскресенье - идем на 6 дней назад
		}
		startDate = time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, now.Location())
	} else {
		// Валидируем формат даты строго: YYYY-MM-DD и проверяем диапазон 2020-2030
		startDate, err = h.timeValidator.ValidateDateString(startDateStr)
		if err != nil {
			if errors.Is(err, validator.ErrInvalidDateFormat) {
				response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid start_date format (use YYYY-MM-DD)")
			} else if errors.Is(err, validator.ErrDateOutOfRange) {
				response.BadRequest(w, response.ErrCodeInvalidInput, "start_date must be between 2020 and 2030")
			} else {
				response.BadRequest(w, response.ErrCodeInvalidInput, err.Error())
			}
			return
		}
	}

	// Если end_date не указан - start_date + 7 дней
	if endDateStr == "" {
		endDate = startDate.AddDate(0, 0, 7)
	} else {
		// Валидируем формат даты строго: YYYY-MM-DD и проверяем диапазон 2020-2030
		endDate, err = h.timeValidator.ValidateDateString(endDateStr)
		if err != nil {
			if errors.Is(err, validator.ErrInvalidDateFormat) {
				response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid end_date format (use YYYY-MM-DD)")
			} else if errors.Is(err, validator.ErrDateOutOfRange) {
				response.BadRequest(w, response.ErrCodeInvalidInput, "end_date must be between 2020 and 2030")
			} else {
				response.BadRequest(w, response.ErrCodeInvalidInput, err.Error())
			}
			return
		}
	}

	// Валидируем диапазон дат: end_date должен быть >= start_date и диапазон не более 365 дней
	if err := h.timeValidator.ValidateDateRange(startDate, endDate, 365); err != nil {
		if errors.Is(err, validator.ErrInvalidTimeRange) {
			response.BadRequest(w, response.ErrCodeInvalidInput, "end_date must be after or equal to start_date")
		} else {
			response.BadRequest(w, response.ErrCodeInvalidInput, "Date range cannot exceed 365 days")
		}
		return
	}

	// Добавляем 1 день, чтобы включить конечную дату целиком
	// (аналогично credits.go:177 и swaps.go:158)
	endDate = endDate.Add(24 * time.Hour)

	// Определяем методистID - для методиста это его ID, для админа можно фильтровать
	var lessons []*models.TeacherScheduleLesson

	if user.IsAdmin() {
		// Админ может запросить расписание любого методиста через query param
		methodologistIDParam := r.URL.Query().Get("methodologist_id")
		if methodologistIDParam != "" {
			methodologistID, err := uuid.Parse(methodologistIDParam)
			if err != nil {
				response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid methodologist_id format")
				return
			}
			// Получаем расписание конкретного методиста
			lessons, err = h.lessonRepo.GetTeacherSchedule(r.Context(), methodologistID, startDate, endDate)
			if err != nil {
				log.Printf("[ERROR] Failed to get methodologist schedule for methodologist %s: %v\n", methodologistID, err)
				response.InternalError(w, "Failed to retrieve methodologist schedule")
				return
			}
		} else {
			// Админ без methodologist_id получает расписание ВСЕХ методистов
			lessons, err = h.lessonRepo.GetAllTeachersSchedule(r.Context(), startDate, endDate)
			if err != nil {
				log.Printf("[ERROR] Failed to get all methodologists schedule: %v\n", err)
				response.InternalError(w, "Failed to retrieve schedule")
				return
			}
		}
	} else {
		// Методист видит только свои занятия
		lessons, err = h.lessonRepo.GetTeacherSchedule(r.Context(), user.ID, startDate, endDate)
		if err != nil {
			log.Printf("[ERROR] Failed to get methodologist schedule for methodologist %s: %v\n", user.ID, err)
			response.InternalError(w, "Failed to retrieve methodologist schedule")
			return
		}
	}

	// Преобразуем в response format с is_past полем
	lessonResponses := make([]map[string]interface{}, len(lessons))
	for i, lesson := range lessons {
		lessonResponses[i] = lesson.ToResponse()
	}

	response.OK(w, map[string]interface{}{
		"lessons": lessonResponses,
		"count":   len(lessonResponses),
	})
}
