package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/pkg/errmessages"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
	"tutoring-platform/pkg/pagination"
	"tutoring-platform/pkg/response"
)

// BookingHandler обрабатывает эндпоинты бронирований
type BookingHandler struct {
	bookingService *service.BookingService
}

// NewBookingHandler создает новый BookingHandler
func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

// CreateBooking обрабатывает POST /api/v1/bookings
// @Summary      Create booking
// @Description  Book a lesson (students book for themselves, admins can book for others)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        payload  body      models.CreateBookingRequest  true  "Booking details"
// @Success      201  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /bookings [post]
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	var req models.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Студенты создают бронирования для себя
	// Администраторы и методисты могут создавать бронирования для любого студента
	if user.IsStudent() {
		req.StudentID = user.ID
		req.IsAdmin = false
	} else if user.IsAdmin() || user.IsMethodologist() {
		// Админ/методист должен указать student_id в запросе
		if req.StudentID == uuid.Nil {
			response.BadRequest(w, response.ErrCodeInvalidInput, "student_id is required for admin or methodologist")
			return
		}
		req.IsAdmin = true
		req.AdminID = user.ID // Track admin/methodologist who created the booking
	} else {
		response.Forbidden(w, "Only students, admins and methodologists can create bookings")
		return
	}

	// Создаем бронирование
	// Если происходит ошибка в транзакции, она будет автоматически отменена
	booking, err := h.bookingService.CreateBooking(r.Context(), &req)
	if err != nil {
		h.handleBookingError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"booking": booking,
	})
}

// GetBooking обрабатывает GET /api/v1/bookings/:id
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	idParam := chi.URLParam(r, "id")

	// Handle special case for /bookings/my
	if idParam == "my" {
		h.GetMyBookings(w, r)
		return
	}

	bookingID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid booking ID")
		return
	}

	booking, err := h.bookingService.GetBooking(r.Context(), bookingID)
	if err != nil {
		response.NotFound(w, "Booking not found")
		return
	}

	// Проверка доступа
	// Студенты видят только свои бронирования
	if user.IsStudent() && booking.StudentID != user.ID {
		response.Forbidden(w, "Unauthorized access to booking")
		return
	}

	// Преподаватели видят только бронирования на свои занятия
	if user.IsTeacher() && booking.TeacherID != user.ID {
		response.Forbidden(w, "Unauthorized access to booking")
		return
	}

	response.OK(w, map[string]interface{}{
		"booking": booking,
	})
}

// GetBookingStatus обрабатывает GET /api/v1/bookings/:id/status
// Lightweight endpoint to check booking status before operations
// @Summary      Get booking status
// @Description  Check booking status (lightweight endpoint for pre-operation validation)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Booking ID"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /bookings/{id}/status [get]
func (h *BookingHandler) GetBookingStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid booking ID")
		return
	}

	booking, err := h.bookingService.GetBooking(r.Context(), bookingID)
	if err != nil {
		response.NotFound(w, "Booking not found")
		return
	}

	// Проверка доступа
	// Студенты видят только свои бронирования
	if user.IsStudent() && booking.StudentID != user.ID {
		response.Forbidden(w, "Unauthorized access to booking")
		return
	}

	// Преподаватели видят только бронирования на свои занятия
	if user.IsTeacher() && booking.TeacherID != user.ID {
		response.Forbidden(w, "Unauthorized access to booking")
		return
	}

	// Return only the status for minimal response
	response.OK(w, map[string]interface{}{
		"status": booking.Status,
		"id":     booking.ID,
	})
}

// GetMyBookings обрабатывает GET /api/v1/bookings/my
func (h *BookingHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только студенты могут получать свои бронирования через этот endpoint
	if !user.IsStudent() {
		response.Forbidden(w, "Only students can use this endpoint")
		return
	}

	activeStatus := models.BookingStatusActive
	filter := &models.ListBookingsFilter{
		StudentID: &user.ID,
		Status:    &activeStatus,
	}

	bookings, err := h.bookingService.ListBookings(r.Context(), filter)
	if err != nil {
		// Логируем детальную ошибку для администраторов, но возвращаем generic сообщение пользователю
		log.Error().
			Str("student_id", user.ID.String()).
			Err(err).
			Msg("Failed to retrieve bookings for student")
		response.InternalError(w, "Failed to retrieve bookings. Please try again later.")
		return
	}

	// Если бронирований нет, возвращаем пустой список вместо ошибки
	if bookings == nil {
		bookings = []*models.BookingWithDetails{}
	}

	response.OK(w, map[string]interface{}{
		"bookings": bookings,
		"count":    len(bookings),
	})
}

// ListBookings обрабатывает GET /api/v1/bookings
// @Summary      List bookings
// @Description  Get filtered list of bookings (students see own, teachers and admins see by filters)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        student_id  query  string  false  "Filter by student ID (teachers/admins only)"
// @Param        lesson_id  query  string  false  "Filter by lesson ID"
// @Param        status  query  string  false  "Filter by status (active, cancelled)"
// @Param        page  query  int  false  "Page number (default: 1)"
// @Param        per_page  query  int  false  "Items per page (default: 20, max: 100)"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      401  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /bookings [get]
func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	filter := &models.ListBookingsFilter{}

	// Студенты видят только свои бронирования
	if user.IsStudent() {
		filter.StudentID = &user.ID
	} else {
		// Преподаватели и админы могут фильтровать по ID студента
		if studentIDStr := r.URL.Query().Get("student_id"); studentIDStr != "" {
			studentID, err := uuid.Parse(studentIDStr)
			if err == nil {
				filter.StudentID = &studentID
			}
		}
	}

	// Опциональные фильтры
	if lessonIDStr := r.URL.Query().Get("lesson_id"); lessonIDStr != "" {
		lessonID, err := uuid.Parse(lessonIDStr)
		if err == nil {
			filter.LessonID = &lessonID
		}
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		status := models.BookingStatus(statusStr)
		filter.Status = &status
	}

	// Парсим параметры пагинации
	params := pagination.ParseParams(r)

	bookings, total, err := h.bookingService.ListBookingsWithPagination(r.Context(), filter, params.Offset, params.PerPage)
	if err != nil {
		response.InternalError(w, "Failed to retrieve bookings")
		return
	}

	if bookings == nil {
		bookings = []*models.BookingWithDetails{}
	}

	response.OK(w, pagination.NewResponse(map[string]interface{}{
		"bookings": bookings,
	}, params.Page, params.PerPage, total))
}

// CancelBooking обрабатывает DELETE /api/v1/bookings/:id
// @Summary      Cancel booking
// @Description  Cancel a booking and refund credits if >24h before lesson. Идемпотентная операция - повторные запросы на отмену уже отменённого бронирования вернут 200 OK с status="already_cancelled"
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Booking ID"
// @Success      200  {object}  response.SuccessResponse{data=models.CancelBookingResult}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /bookings/{id} [delete]
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid booking ID")
		return
	}

	req := &models.CancelBookingRequest{
		BookingID: bookingID,
		StudentID: user.ID,
		IsAdmin:   user.IsAdmin() || user.IsMethodologist(),
	}

	// Отменяем бронирование и получаем результат с информацией о статусе операции
	result, err := h.bookingService.CancelBooking(r.Context(), req)
	if err != nil {
		h.handleBookingError(w, err)
		return
	}

	// Возвращаем 200 OK в обоих случаях (success и already_cancelled)
	// Это обеспечивает идемпотентное поведение API
	response.OK(w, result)
}

// GetCancelledLessons обрабатывает GET /api/v1/bookings/cancelled-lessons
// @Summary      Get cancelled lessons
// @Description  Get list of lesson IDs that the student has cancelled (for blocking re-booking)
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      401  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /bookings/cancelled-lessons [get]
func (h *BookingHandler) GetCancelledLessons(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только студенты могут получать свои отменённые занятия
	if !user.IsStudent() {
		response.Forbidden(w, "Only students can access this endpoint")
		return
	}

	lessonIDs, err := h.bookingService.GetCancelledLessonIDs(r.Context(), user.ID)
	if err != nil {
		log.Error().
			Str("student_id", user.ID.String()).
			Err(err).
			Msg("Failed to get cancelled lessons")
		response.InternalError(w, "Failed to retrieve cancelled lessons")
		return
	}

	// Конвертируем UUID в строки для JSON
	lessonIDStrings := make([]string, 0, len(lessonIDs))
	for _, id := range lessonIDs {
		lessonIDStrings = append(lessonIDStrings, id.String())
	}

	response.OK(w, map[string]interface{}{
		"lesson_ids": lessonIDStrings,
	})
}

// handleBookingError обрабатывает специфичные для бронирований ошибки
func (h *BookingHandler) handleBookingError(w http.ResponseWriter, err error) {
	// Проверяем ошибки repository
	if errors.Is(err, repository.ErrBookingNotFound) {
		response.NotFound(w, errmessages.ErrMsgBookingNotFound)
		return
	}
	if errors.Is(err, repository.ErrLessonNotFound) {
		response.NotFound(w, errmessages.ErrMsgLessonNotFound)
		return
	}
	if errors.Is(err, repository.ErrLessonFull) {
		response.Conflict(w, response.ErrCodeLessonFull, errmessages.ErrMsgLessonFull)
		return
	}
	if errors.Is(err, repository.ErrInsufficientCredits) {
		response.Conflict(w, response.ErrCodeInsufficientCredits, errmessages.ErrMsgInsufficientCredits)
		return
	}
	if errors.Is(err, repository.ErrCreditNotFound) {
		log.Warn().Err(err).Msg("Credit record not found for student")
		response.Conflict(w, response.ErrCodeInsufficientCredits, errmessages.ErrMsgCreditNotInitialized)
		return
	}
	if errors.Is(err, repository.ErrUnauthorized) {
		response.Forbidden(w, errmessages.ErrMsgUnauthorized)
		return
	}
	if errors.Is(err, repository.ErrAlreadyBooked) {
		response.Conflict(w, response.ErrCodeConflict, errmessages.ErrMsgAlreadyBooked)
		return
	}
	if errors.Is(err, repository.ErrDuplicateBooking) {
		response.Conflict(w, response.ErrCodeConflict, errmessages.ErrMsgAlreadyBooked)
		return
	}
	if errors.Is(err, repository.ErrLessonPreviouslyCancelled) {
		response.Error(w, http.StatusForbidden, response.ErrCodeLessonPreviouslyCancelled, errmessages.ErrMsgLessonPreviouslyCancelled)
		return
	}
	if errors.Is(err, repository.ErrBookingNotActive) {
		response.Conflict(w, response.ErrCodeConflict, errmessages.ErrMsgBookingNotActive)
		return
	}
	if errors.Is(err, repository.ErrUserNotFound) {
		response.NotFound(w, errmessages.ErrMsgUserNotFound)
		return
	}

	// Проверяем ошибки validator
	if errors.Is(err, validator.ErrScheduleConflict) {
		response.Conflict(w, response.ErrCodeScheduleConflict, errmessages.ErrMsgScheduleConflict)
		return
	}
	if errors.Is(err, validator.ErrCannotCancelWithin24Hours) {
		response.Conflict(w, response.ErrCodeCannotCancel, errmessages.ErrMsgCannotCancelWithin24Hours)
		return
	}
	if errors.Is(err, validator.ErrBookingNotActive) {
		response.Conflict(w, response.ErrCodeConflict, errmessages.ErrMsgBookingNotActive)
		return
	}
	if errors.Is(err, validator.ErrLessonNotAvailable) {
		response.Conflict(w, response.ErrCodeConflict, errmessages.ErrMsgLessonNotAvailable)
		return
	}
	if errors.Is(err, validator.ErrLessonInPast) {
		response.Conflict(w, response.ErrCodeConflict, errmessages.ErrMsgLessonInPast)
		return
	}

	// Логируем неизвестную ошибку для отладки
	log.Error().
		Err(err).
		Str("error_type", fmt.Sprintf("%T", err)).
		Msg("Unhandled booking error")

	response.InternalError(w, errmessages.ErrMsgOperationFailed)
}
