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
	"tutoring-platform/internal/pkg/errmessages"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/pagination"
	"tutoring-platform/pkg/response"
)

// CreditHandler обрабатывает эндпоинты кредитов
type CreditHandler struct {
	creditService *service.CreditService
	userService   *service.UserService
}

// NewCreditHandler создает новый CreditHandler
func NewCreditHandler(creditService *service.CreditService, userService *service.UserService) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
		userService:   userService,
	}
}

// GetMyCredits обрабатывает GET /api/v1/credits
// Возвращает собственный баланс для студентов, все балансы для админов
// @Summary      Get credit balance
// @Description  Get current credit balance (student: own balance, admin: all students)
// @Tags         credits
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      401  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /credits [get]
func (h *CreditHandler) GetMyCredits(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if user.IsAdmin() {
		// Админ может видеть все кредиты - получаем список пользователей и их балансы
		users, err := h.userService.ListUsers(r.Context(), ptrUserRole(models.RoleStudent))
		if err != nil {
			response.InternalError(w, "Failed to retrieve credit balances")
			return
		}

		type UserBalance struct {
			UserID   uuid.UUID `json:"user_id"`
			Email    string    `json:"email"`
			FullName string    `json:"full_name"`
			Balance  int       `json:"balance"`
		}

		balances := make([]UserBalance, 0, len(users))
		for _, u := range users {
			credit, err := h.creditService.GetBalance(r.Context(), u.ID)
			if err != nil {
				// Если записи о кредитах нет, по умолчанию 0
				balances = append(balances, UserBalance{
					UserID:   u.ID,
					Email:    u.Email,
					FullName: u.GetFullName(),
					Balance:  0,
				})
			} else {
				balances = append(balances, UserBalance{
					UserID:   u.ID,
					Email:    u.Email,
					FullName: u.GetFullName(),
					Balance:  credit.Balance,
				})
			}
		}

		response.OK(w, map[string]interface{}{
			"balances": balances,
			"count":    len(balances),
		})
		return
	}

	// Студенты и преподаватели видят только свой баланс
	credit, err := h.creditService.GetBalance(r.Context(), user.ID)
	if err != nil {
		// Возвращаем 0, если записи о кредитах нет
		response.OK(w, map[string]interface{}{
			"balance": 0,
			"user_id": user.ID,
		})
		return
	}

	response.OK(w, map[string]interface{}{
		"balance": credit.Balance,
		"user_id": credit.UserID,
	})
}

// GetMyHistory обрабатывает GET /api/v1/credits/history
// Возвращает собственную историю для студентов, всю историю для админов с пагинацией
// ПАГИНАЦИЯ: limit (по умолчанию 50, максимум 500), offset (по умолчанию 0)
// @Summary      Get credit transaction history
// @Description  Get credit transaction history with optional filtering and pagination
// @Tags         credits
// @Accept       json
// @Produce      json
// @Param        user_id  query  string  false  "Filter by user ID (admin only)"
// @Param        operation_type  query  string  false  "Filter by operation (add, deduct, refund)"
// @Param        start_date  query  string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date  query  string  false  "End date (YYYY-MM-DD)"
// @Param        limit  query  int  false  "Number of transactions per page (default 50, max 500)"
// @Param        offset  query  int  false  "Pagination offset (default 0)"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      401  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /credits/history [get]
func (h *CreditHandler) GetMyHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	filter := &models.GetCreditHistoryFilter{}

	// Студенты видят только свою историю
	if user.IsStudent() {
		filter.UserID = &user.ID
	} else if user.IsAdmin() || user.IsTeacher() {
		// Админы и методисты могут фильтровать по user_id, если указан
		if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				filter.UserID = &userID
			}
		}
	} else {
		// Преподаватели видят только свою историю
		filter.UserID = &user.ID
	}

	// Парсим опциональные фильтры
	if opTypeStr := r.URL.Query().Get("operation_type"); opTypeStr != "" {
		opType := models.OperationType(opTypeStr)
		if opType == models.OperationTypeAdd || opType == models.OperationTypeDeduct || opType == models.OperationTypeRefund {
			filter.OperationType = &opType
		}
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			// Добавляем 1 день, чтобы включить конечную дату
			endDate = endDate.Add(24 * time.Hour)
			filter.EndDate = &endDate
		}
	}

	// Парсим параметры пагинации
	// limit: количество на страницу (по умолчанию 50, максимум 500)
	// offset: смещение от начала (по умолчанию 0)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := parsePositiveInt(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := parseNonNegativeInt(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	transactions, err := h.creditService.GetTransactionHistory(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to retrieve credit history")
		return
	}

	// Определяем реальный лимит и смещение для метаданных
	actualLimit := models.DefaultTransactionLimit
	if filter.Limit > 0 {
		actualLimit = filter.Limit
	}
	if actualLimit > models.MaxTransactionLimit {
		actualLimit = models.MaxTransactionLimit
	}

	actualOffset := 0
	if filter.Offset > 0 {
		actualOffset = filter.Offset
	}

	response.OK(w, map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
		"limit":        actualLimit,
		"offset":       actualOffset,
		"pagination": map[string]interface{}{
			"limit":  actualLimit,
			"offset": actualOffset,
			"total":  len(transactions),
		},
	})
}

// AddCredits обрабатывает POST /api/v1/users/:id/credits
// Эндпоинт только для админов для добавления или списания кредитов
// @Summary      Add credits to user
// @Description  Add or deduct credits for a user (admin only)
// @Tags         credits
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Param        payload  body      map[string]interface{}  true  "Amount and reason"
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /users/{id}/credits [post]
func (h *CreditHandler) AddCredits(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут добавлять/списывать кредиты
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	var reqBody struct {
		Amount int    `json:"amount"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Проверяем, что сумма не равна 0
	if reqBody.Amount == 0 {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Amount cannot be zero")
		return
	}

	// Валидация диапазона: amount должен быть от -100 до 100 (не включая 0)
	// Положительные значения = добавление, отрицательные = списание
	if reqBody.Amount < -100 || reqBody.Amount > 100 {
		response.BadRequest(w, response.ErrCodeInvalidInput, errmessages.ErrMsgInvalidAmount)
		return
	}

	if reqBody.Reason == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, errmessages.ErrMsgInvalidReason)
		return
	}

	// Проверка минимальной длины причины (хотя бы 3 символа)
	if len(reqBody.Reason) < 3 {
		response.BadRequest(w, response.ErrCodeInvalidInput, errmessages.ErrMsgInvalidReason)
		return
	}

	// Проверяем, что целевой пользователь существует
	_, err = h.userService.GetUser(r.Context(), userID)
	if err != nil {
		response.NotFound(w, errmessages.ErrMsgUserNotFound)
		return
	}

	// Обрабатываем в зависимости от знака суммы
	if reqBody.Amount > 0 {
		// Добавляем кредиты
		req := &models.AddCreditsRequest{
			UserID:      userID,
			Amount:      reqBody.Amount,
			Reason:      reqBody.Reason,
			PerformedBy: user.ID,
		}

		if err := h.creditService.AddCredits(r.Context(), req); err != nil {
			h.handleCreditError(w, err)
			return
		}

		// Получаем новый баланс для стандартизированного ответа
		newBalance := 0
		if credit, err := h.creditService.GetBalance(r.Context(), userID); err == nil {
			newBalance = credit.Balance
		}

		response.OK(w, map[string]interface{}{
			"message": "Credits added successfully",
			"user_id": userID,
			"amount":  reqBody.Amount,
			"balance": newBalance,
		})
	} else {
		// Списываем кредиты (сумма отрицательная)
		req := &models.DeductCreditsRequest{
			UserID:      userID,
			Amount:      -reqBody.Amount, // Конвертируем в положительную для списания
			Reason:      reqBody.Reason,
			PerformedBy: user.ID, // Сохраняем ID администратора, выполнившего операцию
		}

		if err := h.creditService.DeductCredits(r.Context(), req); err != nil {
			h.handleCreditError(w, err)
			return
		}

		// Получаем новый баланс для стандартизированного ответа
		newBalance := 0
		if credit, err := h.creditService.GetBalance(r.Context(), userID); err == nil {
			newBalance = credit.Balance
		}

		response.OK(w, map[string]interface{}{
			"message": "Credits deducted successfully",
			"user_id": userID,
			"amount":  reqBody.Amount,
			"balance": newBalance,
		})
	}
}

// handleCreditError обрабатывает специфичные для кредитов ошибки
func (h *CreditHandler) handleCreditError(w http.ResponseWriter, err error) {
	// Проверяем ошибки repository
	if errors.Is(err, repository.ErrInsufficientCredits) {
		response.Conflict(w, response.ErrCodeInsufficientCredits, errmessages.ErrMsgInsufficientCredits)
		return
	}

	// Проверяем превышение максимального баланса
	if errors.Is(err, models.ErrBalanceExceeded) {
		response.Conflict(w, response.ErrCodeInsufficientCredits, errmessages.ErrMsgBalanceExceeded)
		return
	}

	// Критичные ошибки инициализации кредитов (не должны быть в нормальном потоке)
	if errors.Is(err, repository.ErrCreditNotFound) {
		log.Error().Err(err).Msg("Credit account not found for user - account may not be initialized")
		response.InternalError(w, errmessages.ErrMsgCreditNotInitialized)
		return
	}

	if errors.Is(err, repository.ErrDuplicateCredit) {
		log.Error().Err(err).Msg("Duplicate credit account detected")
		response.InternalError(w, errmessages.ErrMsgCreditDuplicate)
		return
	}

	// Проверяем ошибки models (валидация)
	if errors.Is(err, models.ErrInvalidUserID) {
		response.BadRequest(w, response.ErrCodeValidationFailed, errmessages.ErrMsgUserNotFound)
		return
	}
	if errors.Is(err, models.ErrInvalidCreditAmount) {
		response.BadRequest(w, response.ErrCodeValidationFailed, errmessages.ErrMsgInvalidAmount)
		return
	}
	if errors.Is(err, models.ErrInvalidReason) {
		response.BadRequest(w, response.ErrCodeValidationFailed, errmessages.ErrMsgInvalidReason)
		return
	}
	if errors.Is(err, models.ErrInvalidPerformedBy) {
		response.BadRequest(w, response.ErrCodeValidationFailed, errmessages.ErrMsgUserNotFound)
		return
	}

	// Неизвестная ошибка - логируем её для отладки
	log.Error().Err(err).Msg("Unhandled error in credit operation")
	response.InternalError(w, errmessages.ErrMsgOperationFailed)
}

// GetUserCredits обрабатывает GET /api/v1/credits/user/{id}
// Только для админов и методистов - получить баланс конкретного пользователя по ID
// @Summary      Get user credits
// @Description  Get credit balance for a specific user (admin and teacher only)
// @Tags         credits
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /credits/user/{id} [get]
func (h *CreditHandler) GetUserCredits(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы и методисты могут видеть баланс других пользователей
	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	// Проверяем, что целевой пользователь существует
	targetUser, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Получаем баланс кредитов
	credit, err := h.creditService.GetBalance(r.Context(), userID)
	if err != nil {
		// Если записи о кредитах нет, возвращаем 0
		response.OK(w, map[string]interface{}{
			"user_id":   userID,
			"balance":   0,
			"email":     targetUser.Email,
			"full_name": targetUser.GetFullName(),
		})
		return
	}

	response.OK(w, map[string]interface{}{
		"user_id":   credit.UserID,
		"balance":   credit.Balance,
		"email":     targetUser.Email,
		"full_name": targetUser.GetFullName(),
	})
}

// GetBalance обрабатывает GET /api/v1/credits/balance
// Оптимизированный endpoint для быстрого получения баланса кредитов (для sidebar)
// Возвращает только необходимые поля: balance, user_id, currency
// Результат кэшируется на 5 секунд на клиенте для уменьшения нагрузки на БД
// Оптимизация: SELECT только balance (не full Credit object)
// @Summary      Get current credit balance (optimized for sidebar)
// @Description  Get credit balance for current user - optimized for frequent polling from sidebar
// @Tags         credits
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{} "balance, user_id, currency"
// @Failure      401  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /credits/balance [get]
func (h *CreditHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Добавляем Cache-Control header для кэширования на клиенте (5 секунд)
	// private = не кэшируется прокси (только клиент)
	// max-age=5 = кэш валидный 5 секунд
	// Это уменьшит нагрузку на backend при частых запросах от sidebar
	w.Header().Set("Cache-Control", "private, max-age=5")

	// Получаем баланс кредитов используя оптимизированный query
	// SELECT только balance, не полный Credit object
	balance, err := h.creditService.GetBalanceOptimized(r.Context(), user.ID)
	if err != nil {
		// Если ошибка при получении - возвращаем 0, не ошибку
		// GetBalanceOptimized уже обрабатывает ErrNoRows и возвращает 0
		response.OK(w, map[string]interface{}{
			"balance":  0,
			"user_id":  user.ID,
			"currency": "credits",
		})
		return
	}

	// Возвращаем только необходимые поля для sidebar (минимум данных)
	response.OK(w, map[string]interface{}{
		"balance":  balance,
		"user_id":  user.ID,
		"currency": "credits",
	})
}

// GetAllCredits обрабатывает GET /api/v1/credits/all
// Для админов и методистов - получить все балансы кредитов студентов
// @Summary      Get all credits
// @Description  Get all student credit balances (admin and teacher)
// @Tags         credits
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /credits/all [get]
func (h *CreditHandler) GetAllCredits(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы и методисты могут получать все кредиты
	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Получить ВСЕ кредиты студентов без пагинации для админ-панели
	studentCredits, err := h.creditService.GetAllStudentCreditsNoPagination(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch student credits")
		return
	}

	// Преобразуем результаты в нужный формат
	type StudentBalance struct {
		UserID   uuid.UUID `json:"user_id"`
		Email    string    `json:"email"`
		FullName string    `json:"full_name"`
		Balance  int       `json:"balance"`
	}

	balances := make([]StudentBalance, 0, len(studentCredits))
	for _, record := range studentCredits {
		balance := StudentBalance{
			Email:    record["email"].(string),
			FullName: record["full_name"].(string),
			Balance:  0,
		}

		// Parse user_id
		userIDVal := record["user_id"]
		if uid, ok := userIDVal.(uuid.UUID); ok {
			balance.UserID = uid
		} else if uidStr, ok := userIDVal.(string); ok {
			if parsedID, err := uuid.Parse(uidStr); err == nil {
				balance.UserID = parsedID
			}
		}

		// Parse balance
		if balVal, ok := record["balance"].(int64); ok {
			balance.Balance = int(balVal)
		} else if balVal, ok := record["balance"].(int); ok {
			balance.Balance = balVal
		}

		balances = append(balances, balance)
	}

	if balances == nil {
		balances = []StudentBalance{}
	}

	// Возвращать в формате пагинации но с total=1 (нет реальной пагинации)
	response.OK(w, pagination.NewResponse(balances, 1, len(studentCredits), len(studentCredits)))
}

// ptrUserRole вспомогательная функция для создания указателя на UserRole
func ptrUserRole(role models.UserRole) *models.UserRole {
	return &role
}

// parsePositiveInt парсит строку в положительное целое число
// Возвращает значение > 0 или ошибку
// Используется для параметра limit, который должен быть > 0
func parsePositiveInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return 0, err
	}
	if val <= 0 {
		return 0, fmt.Errorf("value must be positive")
	}
	return val, nil
}

// parseNonNegativeInt парсит строку в неотрицательное целое число
// Возвращает значение >= 0 или ошибку
// Используется для параметра offset, который может быть >= 0
func parseNonNegativeInt(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return 0, err
	}
	if val < 0 {
		return 0, fmt.Errorf("value must be non-negative")
	}
	return val, nil
}
