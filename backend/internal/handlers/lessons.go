package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// LessonHandler обрабатывает эндпоинты занятий
type LessonHandler struct {
	lessonService   *service.LessonService
	bookingService  *service.BookingService
	bulkEditService *service.BulkEditService
}

// NewLessonHandler создает новый LessonHandler
func NewLessonHandler(lessonService *service.LessonService, bookingService *service.BookingService, bulkEditService *service.BulkEditService) *LessonHandler {
	return &LessonHandler{
		lessonService:   lessonService,
		bookingService:  bookingService,
		bulkEditService: bulkEditService,
	}
}

// GetLessons обрабатывает GET /api/v1/lessons
// Возвращает занятия с учетом правил видимости:
// - Admin: все занятия
// - Teacher: только свои занятия
// - Student: групповые занятия + индивидуальные, на которые они записаны
func (h *LessonHandler) GetLessons(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	filter := &models.ListLessonsFilter{}

	// Парсим параметры запроса
	if teacherIDStr := r.URL.Query().Get("teacher_id"); teacherIDStr != "" {
		teacherID, err := uuid.Parse(teacherIDStr)
		if err == nil {
			filter.TeacherID = &teacherID
		}
	}

	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			// Устанавливаем начало дня
			startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			filter.StartDate = &startOfDay
			// Устанавливаем конец дня
			endOfDay := startOfDay.Add(24 * time.Hour)
			filter.EndDate = &endOfDay
		}
	}

	if availableStr := r.URL.Query().Get("available"); availableStr == "true" {
		available := true
		filter.Available = &available
	}

	// Use visibility-aware query based on user role
	lessons, err := h.lessonService.GetVisibleLessons(r.Context(), user.ID, string(user.Role), filter)
	if err != nil {
		response.InternalError(w, "Failed to retrieve lessons")
		return
	}

	// Преобразуем lessons в response с is_past полем
	lessonResponses := make([]*models.LessonResponse, len(lessons))
	for i, lesson := range lessons {
		lessonResponses[i] = lesson.ToResponse()
	}

	// Загружаем информацию о студентах для ВСЕх занятий в одном batch запросе (избегаем N+1 queries)
	if len(lessons) > 0 {
		lessonIDs := make([]uuid.UUID, len(lessons))
		for i, lesson := range lessons {
			lessonIDs[i] = lesson.ID
		}

		bookingsMap, err := h.lessonService.GetLessonBookingsForLessons(r.Context(), lessonIDs)
		if err != nil {
			// Логируем ошибку, но не прерываем обработку
			log.Warn().Err(err).Msg("failed to get bookings for lessons batch")
			bookingsMap = make(map[uuid.UUID][]models.BookingInfo)
		}
		// Заполняем bookings из map для каждого lesson
		for i, lesson := range lessons {
			if bookings, ok := bookingsMap[lesson.ID]; ok {
				lessonResponses[i].Bookings = bookings
			} else {
				lessonResponses[i].Bookings = []models.BookingInfo{}
			}
		}
	}

	response.OK(w, map[string]interface{}{
		"lessons": lessonResponses,
		"count":   len(lessonResponses),
	})
}

// CreateLesson обрабатывает POST /api/v1/lessons
func (h *LessonHandler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы и методисты могут создавать занятия
	if !user.IsAdmin() && !user.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	var req models.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Создаем занятие
	lesson, err := h.lessonService.CreateLesson(r.Context(), &req)
	if err != nil {
		h.handleLessonError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"lesson": lesson.ToResponseWithoutTeacher(),
	})
}

// GetLesson обрабатывает GET /api/v1/lessons/:id
func (h *LessonHandler) GetLesson(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	lesson, err := h.lessonService.GetLessonWithTeacher(r.Context(), lessonID)
	if err != nil {
		response.NotFound(w, "Lesson not found")
		return
	}

	response.OK(w, map[string]interface{}{
		"lesson": lesson.ToResponse(),
	})
}

// UpdateLesson обрабатывает PUT /api/v1/lessons/:id
func (h *LessonHandler) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Получаем занятие для проверки времени и прав
	lesson, err := h.lessonService.GetLesson(r.Context(), lessonID)
	if err != nil {
		response.NotFound(w, "Lesson not found")
		return
	}

	var req models.UpdateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Проверяем права доступа:
	// - Admin может редактировать все поля всех занятий
	// - Teacher может редактировать только homework_text своих занятий
	// - Student не может редактировать
	isTeacherOwnLesson := user.IsMethodologist() && lesson.TeacherID == user.ID
	isHomeworkTextOnlyUpdate := req.HomeworkText != nil &&
		req.TeacherID == nil &&
		req.StartTime == nil &&
		req.EndTime == nil &&
		req.LessonType == nil &&
		req.MaxStudents == nil &&
		req.Subject == nil &&
		req.Color == nil

	// Логирование для диагностики
	log.Debug().
		Str("user_id", user.ID.String()).
		Str("user_role", string(user.Role)).
		Str("lesson_id", lessonID.String()).
		Str("lesson_teacher_id", lesson.TeacherID.String()).
		Bool("is_teacher_own_lesson", isTeacherOwnLesson).
		Bool("is_homework_text_only", isHomeworkTextOnlyUpdate).
		Bool("has_homework_text", req.HomeworkText != nil).
		Msg("UpdateLesson authorization check")

	if !user.IsAdmin() && !user.IsMethodologist() {
		// Если преподаватель пытается обновить только homework_text своего урока
		if isTeacherOwnLesson && isHomeworkTextOnlyUpdate {
			// Methodologist может редактировать homework_text своих уроков
			log.Debug().
				Str("teacher_id", user.ID.String()).
				Str("lesson_id", lessonID.String()).
				Msg("Methodologist updating homework_text for own lesson")
		} else if user.IsMethodologist() {
			// Преподаватель пытается обновить что-то другое или не свой урок
			if isHomeworkTextOnlyUpdate && !isTeacherOwnLesson {
				// Methodologist пытается обновить homework_text, но урок не его
				log.Warn().
					Str("user_id", user.ID.String()).
					Str("lesson_teacher_id", lesson.TeacherID.String()).
					Str("lesson_id", lessonID.String()).
					Msg("Methodologist tried to update homework_text for another methodologist's lesson")
				response.Forbidden(w, "Вы можете редактировать только свои занятия")
				return
			} else if isTeacherOwnLesson && !isHomeworkTextOnlyUpdate {
				// Methodologist пытается обновить не только homework_text своего урока
				log.Warn().
					Str("user_id", user.ID.String()).
					Str("lesson_id", lessonID.String()).
					Msg("Methodologist tried to update non-homework_text fields for own lesson")
				response.Forbidden(w, "Преподаватели могут редактировать только описание домашнего задания")
				return
			} else {
				// Другие случаи для преподавателя
				response.Forbidden(w, "Вы можете редактировать только описание домашнего задания своих занятий")
				return
			}
		} else {
			// Студент или другой пользователь
			response.Forbidden(w, "Только администраторы и методисты могут редактировать занятия")
			return
		}
	}

	// Проверяем право на редактирование прошлых занятий
	// Только админ и методист могут редактировать занятия которые уже начались (кроме homework_text)
	isPastLesson := lesson.StartTime.Before(time.Now())
	if isPastLesson && !user.IsAdmin() && !user.IsMethodologist() && !isHomeworkTextOnlyUpdate {
		response.Forbidden(w, "Вы не можете редактировать занятия которые уже прошли")
		return
	}

	// Логируем редактирование прошлых занятий админом или методистом для аудита
	if isPastLesson && (user.IsAdmin() || user.IsMethodologist()) {
		log.Warn().
			Str("user_id", user.ID.String()).
			Str("user_email", user.Email).
			Str("user_role", string(user.Role)).
			Str("lesson_id", lessonID.String()).
			Str("lesson_start_time", lesson.StartTime.Format("2006-01-02 15:04")).
			Msg("Admin/Methodologist is editing a past lesson")
	}

	// Обновляем занятие
	updatedLesson, err := h.lessonService.UpdateLesson(r.Context(), lessonID, &req)
	if err != nil {
		h.handleLessonError(w, err)
		return
	}

	// Формируем ответ с предупреждением для прошлых занятий
	responseData := map[string]interface{}{
		"lesson": updatedLesson.ToResponseWithoutTeacher(),
	}

	if isPastLesson && user.IsAdmin() {
		responseData["warning"] = "Это занятие уже началось. Изменения могут повлиять на записанных студентов."
	}

	response.OK(w, responseData)
}

// DeleteLesson обрабатывает DELETE /api/v1/lessons/:id
func (h *LessonHandler) DeleteLesson(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы и методисты могут удалять занятия
	if !user.IsAdmin() && !user.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Удаляем занятие (мягкое удаление)
	if err := h.lessonService.DeleteLesson(r.Context(), lessonID); err != nil {
		h.handleLessonError(w, err)
		return
	}

	response.OK(w, map[string]string{
		"message": "Lesson deleted successfully",
	})
}

// GetMyLessons обрабатывает GET /api/v1/lessons/my
// Возвращает:
// - для студентов: занятия, на которые они забронированы (с информацией о преподавателях)
// - для преподавателей: занятия, которые они ведут
// - для админов: все занятия
func (h *LessonHandler) GetMyLessons(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Студенты видят занятия, на которые они записались
	if user.IsStudent() {
		lessons, err := h.bookingService.GetStudentLessons(r.Context(), user.ID)
		if err != nil {
			log.Error().
				Str("student_id", user.ID.String()).
				Err(err).
				Msg("Failed to get student lessons")
			response.InternalError(w, "Failed to retrieve your lessons")
			return
		}

		log.Debug().
			Str("student_id", user.ID.String()).
			Int("lessons_count", len(lessons)).
			Msg("Retrieved lessons for student")

		response.OK(w, map[string]interface{}{
			"lessons": lessons,
			"count":   len(lessons),
		})
		return
	}

	// Преподаватели и админы видят уроки через фильтр
	filter := &models.ListLessonsFilter{}

	if user.IsMethodologist() {
		// Преподаватели видят только свои занятия
		filter.TeacherID = &user.ID
		log.Debug().Str("teacher_id", user.ID.String()).Msg("Loading lessons for teacher")
	} else {
		// Админы видят все занятия (без фильтра)
		log.Debug().Str("admin_id", user.ID.String()).Msg("Loading all lessons for admin")
	}

	lessons, err := h.lessonService.ListLessons(r.Context(), filter)
	if err != nil {
		log.Error().
			Str("user_id", user.ID.String()).
			Err(err).
			Msg("Failed to list lessons")
		response.InternalError(w, "Failed to retrieve lessons")
		return
	}

	log.Debug().
		Str("user_id", user.ID.String()).
		Int("lessons_count", len(lessons)).
		Msg("Retrieved lessons")

	response.OK(w, map[string]interface{}{
		"lessons": lessons,
		"count":   len(lessons),
	})
}

// GetLessonStudents обрабатывает GET /api/v1/lessons/:id/students
// @Summary      Get lesson students
// @Description  List students enrolled in a lesson (methodologist, admin and methodologist)
// @Tags         lessons
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Lesson ID"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /lessons/{id}/students [get]
func (h *LessonHandler) GetLessonStudents(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Проверяем, что занятие существует
	lesson, err := h.lessonService.GetLesson(r.Context(), lessonID)
	if err != nil {
		response.NotFound(w, "Lesson not found")
		return
	}

	// Проверка доступа: админы и методисты видят все занятия, преподаватели видят только свои
	if !user.IsAdmin() && !user.IsMethodologist() {
		if !user.IsMethodologist() || lesson.TeacherID != user.ID {
			response.Forbidden(w, "Access denied")
			return
		}
	}

	// Студенты не могут просматривать список студентов занятия
	if user.IsStudent() {
		response.Forbidden(w, "Access denied")
		return
	}

	// Получаем бронирования для этого занятия с полными данными студентов
	filter := &models.ListBookingsFilter{
		LessonID: &lessonID,
		Status:   ptrBookingStatus(models.BookingStatusActive),
	}

	bookings, err := h.bookingService.ListBookings(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to retrieve lesson students")
		return
	}

	// Преобразуем BookingWithDetails в формат для фронтенда с полными именами студентов
	students := make([]map[string]interface{}, len(bookings))
	for i, booking := range bookings {
		students[i] = map[string]interface{}{
			"student_id":        booking.StudentID,
			"student_full_name": booking.StudentFullName,
			"student_email":     booking.StudentEmail,
			"booked_at":         booking.BookedAt,
		}
	}

	response.OK(w, map[string]interface{}{
		"students": students,
		"count":    len(students),
	})
}

// handleLessonError обрабатывает специфичные для занятий ошибки
func (h *LessonHandler) handleLessonError(w http.ResponseWriter, err error) {
	// Проверяем ошибки repository
	if errors.Is(err, repository.ErrLessonNotFound) {
		response.NotFound(w, "Lesson not found")
		return
	}

	// Проверяем ошибку конфликта расписания (EXCLUDE constraint нарушение)
	if errors.Is(err, repository.ErrLessonOverlapConflict) {
		response.Conflict(w, response.ErrCodeConflict, "Teacher has overlapping lessons at this time")
		return
	}

	// Ошибки удаления занятия (FK constraints)
	if errors.Is(err, repository.ErrLessonHasActiveBookings) {
		response.Conflict(w, response.ErrCodeConflict, "Cannot delete lesson with active bookings: cancel bookings and refund credits first")
		return
	}
	if errors.Is(err, repository.ErrLessonHasHomework) {
		response.Conflict(w, response.ErrCodeConflict, "Cannot delete lesson with homework: delete homework first")
		return
	}

	// Проверяем ошибки models (валидация)
	if errors.Is(err, models.ErrInvalidTeacherID) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid teacher ID")
		return
	}
	if errors.Is(err, models.ErrInvalidLessonTime) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid lesson time")
		return
	}
	if errors.Is(err, models.ErrInvalidMaxStudents) {
		response.BadRequest(w, response.ErrCodeInvalidLessonTypeUpdate, "Invalid max students")
		return
	}

	// Неизвестная ошибка - логируем для отладки
	log.Error().Err(err).Msg("Unhandled lesson error")
	response.InternalError(w, "An error occurred processing your request")
}

// ptrBookingStatus вспомогательная функция для создания указателя на BookingStatus
func ptrBookingStatus(status models.BookingStatus) *models.BookingStatus {
	return &status
}

// ApplyToAllSubsequent handles POST /api/v1/lessons/:id/apply-to-all - Bulk edit (Admin only)
// @Summary      Apply modification to all subsequent lessons
// @Description  Apply lesson modification to all future matching lessons (bulk edit, admin only)
// @Tags         lessons
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Lesson ID"
// @Param        payload  body      models.ApplyToAllSubsequentRequest  true  "Modification details"
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /lessons/{id}/apply-to-all [post]
func (h *LessonHandler) ApplyToAllSubsequent(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Only admins and methodologists can apply bulk edits")
		return
	}

	lessonID := chi.URLParam(r, "id")
	if lessonID == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Lesson ID is required")
		return
	}

	// Validate UUID format
	lessonUUID, err := uuid.Parse(lessonID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID format")
		return
	}

	var req models.ApplyToAllSubsequentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	// Set lesson ID from URL parameter
	req.LessonID = lessonUUID

	// Validate request
	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
		return
	}

	modification, err := h.bulkEditService.ApplyToAllSubsequent(r.Context(), admin.ID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrLessonNotFound) {
			response.NotFound(w, "Lesson not found")
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			response.BadRequest(w, response.ErrCodeValidationFailed, "Referenced user not found")
			return
		}
		if errors.Is(err, repository.ErrBookingNotFound) {
			response.BadRequest(w, response.ErrCodeValidationFailed, "Student booking not found")
			return
		}
		response.InternalError(w, "Failed to apply bulk edit: "+err.Error())
		return
	}

	response.OK(w, modification)
}
