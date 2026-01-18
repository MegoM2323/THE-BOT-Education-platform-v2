package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// SuccessResponse представляет успешный ответ API
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ErrorResponse представляет ответ API с ошибкой
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail содержит детали ошибки
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success отправляет успешный JSON ответ
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode success response: %v", err)
		// Note: Cannot send error response here as headers already written
	}
}

// Error отправляет JSON ответ с ошибкой
func Error(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode error response (code=%s): %v", code, err)
		// Note: Cannot send error response here as headers already written
	}
}

// Общие коды ошибок
const (
	// Ошибки аутентификации
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired     = "SESSION_EXPIRED"
	ErrCodeInvalidCSRF        = "INVALID_CSRF"
	ErrCodeInvalidSignature   = "INVALID_SIGNATURE"

	// Ошибки валидации
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_FIELD"

	// Ошибки ресурсов
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeConflict      = "CONFLICT"

	// Ошибки бизнес-логики
	ErrCodeInsufficientCredits       = "INSUFFICIENT_CREDITS"
	ErrCodeLessonFull                = "LESSON_FULL"
	ErrCodeBookingTooLate            = "BOOKING_TOO_LATE"
	ErrCodeScheduleConflict          = "SCHEDULE_CONFLICT"
	ErrCodeCannotCancel              = "CANNOT_CANCEL"
	ErrCodeInvalidSwap               = "INVALID_SWAP"
	ErrCodeCannotChangeToIndividual  = "CANNOT_CHANGE_TO_INDIVIDUAL"
	ErrCodeInvalidLessonTypeUpdate   = "INVALID_LESSON_TYPE_UPDATE"
	ErrCodeLessonPreviouslyCancelled = "LESSON_PREVIOUSLY_CANCELLED"

	// Ошибки сервера
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeDatabaseError      = "DATABASE_ERROR"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"

	// Ошибки чата
	ErrCodeChatRoomNotFound     = "CHAT_ROOM_NOT_FOUND"
	ErrCodeChatMessageNotFound  = "CHAT_MESSAGE_NOT_FOUND"
	ErrCodeChatFileNotFound     = "CHAT_FILE_NOT_FOUND"
	ErrCodeChatPermissionDenied = "CHAT_PERMISSION_DENIED"
	ErrCodeChatFileUploadFailed = "CHAT_FILE_UPLOAD_FAILED"
)

// Предопределенные ответы с ошибками для типичных сценариев

// BadRequest отправляет ответ 400 Bad Request
func BadRequest(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusBadRequest, code, message)
}

// Unauthorized отправляет ответ 401 Unauthorized
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden отправляет ответ 403 Forbidden
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound отправляет ответ 404 Not Found
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, ErrCodeNotFound, message)
}

// Conflict отправляет ответ 409 Conflict
func Conflict(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusConflict, code, message)
}

// InternalError отправляет ответ 500 Internal Server Error
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, ErrCodeInternalError, message)
}

// Created отправляет ответ 201 Created
func Created(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusCreated, data)
}

// OK отправляет ответ 200 OK
func OK(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusOK, data)
}

// NoContent отправляет ответ 204 No Content
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// TooManyRequests отправляет ответ 429 Too Many Requests
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, ErrCodeRateLimitExceeded, message)
}

// ServiceUnavailable отправляет ответ 503 Service Unavailable
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, ErrCodeServiceUnavailable, message)
}

// RequestTimeout отправляет ответ 408 Request Timeout
func RequestTimeout(w http.ResponseWriter, message string) {
	Error(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", message)
}

// ============================================================================
// STANDARDIZED RESPONSE WRAPPERS (для list endpoints с метаданными)
// ============================================================================

// PaginationMeta содержит информацию о пагинации
type PaginationMeta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalCount int `json:"total_count,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// ResponseMeta содержит метаданные ответа
type ResponseMeta struct {
	Pagination *PaginationMeta `json:"pagination,omitempty"`
	Count      int             `json:"count"`
	Timestamp  string          `json:"timestamp,omitempty"`
}

// StandardListResponse представляет стандартный ответ для list endpoints
type StandardListResponse struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data"`
	Meta    ResponseMeta `json:"meta"`
}

// StandardSingleResponse представляет стандартный ответ для single item endpoints
type StandardSingleResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// ListResponse отправляет стандартный ответ для list endpoints
// Используется для всех endpoints, возвращающих массивы данных
//
// Пример использования:
//
//	items := []MyItem{...}
//	response.ListResponse(w, http.StatusOK, items, &response.ResponseMeta{
//	    Count: len(items),
//	    Pagination: &response.PaginationMeta{
//	        Page: 1,
//	        PageSize: 20,
//	        TotalCount: 100,
//	        TotalPages: 5,
//	    },
//	})
func ListResponse(w http.ResponseWriter, statusCode int, data interface{}, meta *ResponseMeta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if meta == nil {
		meta = &ResponseMeta{}
	}

	// Count по умолчанию вычисляется из длины data если это slice
	if meta.Count == 0 {
		if slice, ok := data.([]interface{}); ok {
			meta.Count = len(slice)
		}
	}

	response := StandardListResponse{
		Success: true,
		Data:    data,
		Meta:    *meta,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode list response: %v", err)
	}
}

// SingleResponse отправляет стандартный ответ для single item endpoints
// Используется для endpoints, возвращающих один объект
//
// Пример использования:
//
//	item := MyItem{...}
//	response.SingleResponse(w, http.StatusOK, item)
func SingleResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := StandardSingleResponse{
		Success: true,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ERROR: Failed to encode single response: %v", err)
	}
}
