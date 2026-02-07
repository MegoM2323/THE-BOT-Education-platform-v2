package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
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
	telegramService *service.TelegramService
}

// NewLessonHandler создает новый LessonHandler
func NewLessonHandler(lessonService *service.LessonService, bookingService *service.BookingService, bulkEditService *service.BulkEditService, telegramService *service.TelegramService) *LessonHandler {
	return &LessonHandler{
		lessonService:   lessonService,
		bookingService:  bookingService,
		bulkEditService: bulkEditService,
		telegramService: telegramService,
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
	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	var req models.CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Методист (учитель) может назначить только себя преподавателем
	// Админ может назначить любого преподавателя
	if user.IsTeacher() && !user.IsAdmin() {
		if req.TeacherID != user.ID {
			response.Forbidden(w, "Вы можете назначить только себя преподавателем занятия")
			return
		}
	}

	// Проверяем, нужно ли создать повторяющиеся занятия
	if req.IsRecurring {
		lessons, err := h.lessonService.CreateRecurringLessons(r.Context(), &req)
		if err != nil {
			response.BadRequest(w, response.ErrCodeValidationFailed, fmt.Sprintf("Failed to create recurring lessons: %s", err.Error()))
			return
		}
		lessonResponses := make([]*models.LessonResponse, len(lessons))
		for i, lesson := range lessons {
			lessonResponses[i] = lesson.ToResponseWithoutTeacher()
		}
		response.Created(w, map[string]interface{}{
			"message": fmt.Sprintf("Создано %d повторяющихся занятий", len(lessons)),
			"lessons": lessonResponses,
			"count":   len(lessons),
		})
		return
	}

	// Создаем обычное занятие
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

	// Методист (учитель) может назначить только себя преподавателем
	// Админ может назначить любого преподавателя
	if user.IsTeacher() && !user.IsAdmin() && req.TeacherID != nil {
		if *req.TeacherID != user.ID {
			response.Forbidden(w, "Вы можете назначить только себя преподавателем занятия")
			return
		}
	}

	// Проверяем права доступа:
	// - Admin может редактировать все поля всех занятий
	// - Teacher может редактировать только homework_text и report_text своих занятий
	// - Student не может редактировать
	isTeacherOwnLesson := user.IsTeacher() && lesson.TeacherID == user.ID
	isTextFieldsOnlyUpdate := (req.HomeworkText != nil || req.ReportText != nil) &&
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
		Bool("is_text_fields_only", isTextFieldsOnlyUpdate).
		Bool("has_homework_text", req.HomeworkText != nil).
		Bool("has_report_text", req.ReportText != nil).
		Msg("UpdateLesson authorization check")

	if !user.IsAdmin() && !user.IsTeacher() {
		// Если преподаватель пытается обновить только homework_text или report_text своего урока
		if isTeacherOwnLesson && isTextFieldsOnlyUpdate {
			// Teacher может редактировать homework_text и report_text своих уроков
			log.Debug().
				Str("teacher_id", user.ID.String()).
				Str("lesson_id", lessonID.String()).
				Msg("Teacher updating text fields for own lesson")
		} else if user.IsTeacher() {
			// Преподаватель пытается обновить что-то другое или не свой урок
			if isTextFieldsOnlyUpdate && !isTeacherOwnLesson {
				// Teacher пытается обновить текстовые поля, но урок не его
				log.Warn().
					Str("user_id", user.ID.String()).
					Str("lesson_teacher_id", lesson.TeacherID.String()).
					Str("lesson_id", lessonID.String()).
					Msg("Teacher tried to update text fields for another teacher's lesson")
				response.Forbidden(w, "Вы можете редактировать только свои занятия")
				return
			} else if isTeacherOwnLesson && !isTextFieldsOnlyUpdate {
				// Teacher пытается обновить не только текстовые поля своего урока
				log.Warn().
					Str("user_id", user.ID.String()).
					Str("lesson_id", lessonID.String()).
					Msg("Teacher tried to update non-text fields for own lesson")
				response.Forbidden(w, "Преподаватели могут редактировать только текстовые поля (домашнее задание и отчет)")
				return
			} else {
				// Другие случаи для преподавателя
				response.Forbidden(w, "Вы можете редактировать только текстовые поля своих занятий")
				return
			}
		} else {
			// Студент или другой пользователь
			response.Forbidden(w, "Только администраторы и методисты могут редактировать занятия")
			return
		}
	}

	// Проверяем право на редактирование отчета: только после начала занятия
	isPastLesson := lesson.StartTime.Before(time.Now())
	if req.ReportText != nil && !isPastLesson && !user.IsAdmin() {
		response.Forbidden(w, "Отчет можно редактировать только после начала занятия")
		return
	}

	// Логируем редактирование прошлых занятий админом или методистом для аудита
	if isPastLesson && (user.IsAdmin() || user.IsTeacher()) {
		log.Warn().
			Str("user_id", user.ID.String()).
			Str("user_email", user.Email).
			Str("user_role", string(user.Role)).
			Str("lesson_id", lessonID.String()).
			Str("lesson_start_time", lesson.StartTime.Format("2006-01-02 15:04")).
			Msg("Admin/Teacher is editing a past lesson")
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
	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	// Проверяем query параметр delete_series
	deleteSeries := r.URL.Query().Get("delete_series") == "true"

	if deleteSeries {
		// Сначала получаем занятие, чтобы узнать recurring_group_id
		lesson, err := h.lessonService.GetLesson(r.Context(), lessonID)
		if err != nil {
			h.handleLessonError(w, err)
			return
		}

		// Проверяем, что это повторяющееся занятие
		if lesson.RecurringGroupID == nil {
			response.BadRequest(w, response.ErrCodeInvalidInput, "Lesson is not part of a recurring series")
			return
		}

		// Удаляем всю серию
		deletedCount, err := h.lessonService.DeleteRecurringSeries(r.Context(), *lesson.RecurringGroupID)
		if err != nil {
			h.handleLessonError(w, err)
			return
		}

		response.OK(w, map[string]interface{}{
			"message":       "Recurring series deleted successfully",
			"deleted_count": deletedCount,
		})
		return
	}

	// Удаляем одно занятие (мягкое удаление)
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

	if user.IsTeacher() {
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
// @Description  List students enrolled in a lesson (teacher, admin and teacher)
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
	if !user.IsAdmin() && !user.IsTeacher() {
		if !user.IsTeacher() || lesson.TeacherID != user.ID {
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

	if !admin.IsAdmin() && !admin.IsTeacher() {
		response.Forbidden(w, "Only admins and teachers can apply bulk edits")
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

// SendReportToParents отправляет отчет о занятии родителям студентов
func (h *LessonHandler) SendReportToParents(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Only admins and teachers can send reports to parents")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	lesson, err := h.lessonService.GetLesson(r.Context(), lessonID)
	if err != nil {
		response.NotFound(w, "Lesson not found")
		return
	}

	if !lesson.ReportText.Valid || lesson.ReportText.String == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Отчет о занятии пустой")
		return
	}

	bookings, err := h.lessonService.GetLessonBookings(r.Context(), lessonID)
	if err != nil {
		log.Error().Err(err).Str("lesson_id", lessonID.String()).Msg("Failed to get lesson bookings")
		response.InternalError(w, "Failed to get lesson bookings")
		return
	}

	result, err := h.telegramService.SendLessonReportToParents(r.Context(), lesson, lesson.ReportText.String, bookings)
	if err != nil {
		log.Error().Err(err).Str("lesson_id", lessonID.String()).Msg("Failed to send report to parents")
		response.InternalError(w, "Failed to send report to parents")
		return
	}
	response.Created(w, result)
}

// CreateRecurringSeriesFromLesson создаёт серию повторяющихся занятий на основе существующего
func (h *LessonHandler) CreateRecurringSeriesFromLesson(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID")
		return
	}

	var req models.CreateRecurringSeriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	result, err := h.lessonService.CreateRecurringSeriesFromLesson(r.Context(), lessonID, user.ID)
	if err != nil {
		log.Error().Err(err).Str("lesson_id", lessonID.String()).Msg("Failed to create recurring series")
		response.InternalError(w, "Failed to create recurring series")
		return
	}

	response.Created(w, map[string]interface{}{
		"recurring_group_id": result.RecurringGroupID,
		"lessons":            result.Lessons,
		"count":              len(result.Lessons),
		"message":            fmt.Sprintf("Создано %d повторяющихся занятий", len(result.Lessons)),
	})

	response.OK(w, result)
}
