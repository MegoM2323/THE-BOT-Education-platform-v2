package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"tutoring-platform/internal/config"
	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/pkg/response"

	"github.com/google/uuid"
)

// PaymentServiceInterface определяет интерфейс для работы с платежами
type PaymentServiceInterface interface {
	CreatePayment(ctx context.Context, userID uuid.UUID, req *models.CreatePaymentRequest) (*models.PaymentResponse, error)
	GetPaymentHistory(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error)
	ProcessPaymentSuccess(ctx context.Context, paymentID string) error
	ProcessPaymentCancellation(ctx context.Context, paymentID string) error
}

// PaymentHandler обрабатывает запросы связанные с платежами
type PaymentHandler struct {
	paymentService PaymentServiceInterface
	cfg            *config.Config
}

// NewPaymentHandler создает новый PaymentHandler
func NewPaymentHandler(paymentService PaymentServiceInterface, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		cfg:            cfg,
	}
}

// CreatePayment создает новый платеж
// POST /api/v1/payments/create
// @Summary      Create payment
// @Description  Create a new payment and return YooKassa confirmation URL
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payload  body      models.CreatePaymentRequest  true  "Payment details"
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /payments/create [post]
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем user_id из контекста
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}
	userID := user.ID

	// Парсим request body
	var req models.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Создаем платеж
	payment, err := h.paymentService.CreatePayment(ctx, userID, &req)
	if err != nil {
		log.Printf("ERROR: Failed to create payment for user %s: %v", userID, err)
		response.InternalError(w, "Failed to create payment")
		return
	}

	response.Success(w, http.StatusOK, payment)
}

// GetHistory возвращает историю платежей пользователя
// GET /api/v1/payments/history
// @Summary      Get payment history
// @Description  Get current user's payment history
// @Tags         payments
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      401  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /payments/history [get]
func (h *PaymentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Извлекаем user_id из контекста
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}
	userID := user.ID

	// Получаем историю платежей
	payments, err := h.paymentService.GetPaymentHistory(ctx, userID)
	if err != nil {
		log.Printf("ERROR: Failed to get payment history for user %s: %v", userID, err)
		response.InternalError(w, "Failed to get payment history")
		return
	}

	response.Success(w, http.StatusOK, payments)
}

// verifyYooKassaSignature проверяет HMAC-SHA256 подпись webhook от YooKassa
// Подпись вычисляется как: HMAC-SHA256(body, secretKey)
// Подпись передается в заголовке X-Yookassa-Shop-Api-Signature-SHA256
func (h *PaymentHandler) verifyYooKassaSignature(r *http.Request, body []byte) error {
	// Получаем подпись из заголовка запроса
	signatureHeader := r.Header.Get("X-Yookassa-Shop-Api-Signature-SHA256")
	if signatureHeader == "" {
		return fmt.Errorf("отсутствует заголовок X-Yookassa-Shop-Api-Signature-SHA256")
	}

	// Получаем секретный ключ YooKassa из конфигурации
	secretKey := h.cfg.YooKassa.SecretKey
	if secretKey == "" {
		log.Printf("ERROR: YOOKASSA_SECRET_KEY не установлен в переменных окружения")
		return fmt.Errorf("конфигурация YooKassa не полностью инициализирована")
	}

	// Вычисляем ожидаемую подпись
	// YooKassa использует формат: HMAC-SHA256(body, secretKey)
	h256 := hmac.New(sha256.New, []byte(secretKey))
	h256.Write(body)
	expectedSignature := hex.EncodeToString(h256.Sum(nil))

	// Сравниваем подписи с использованием constant-time comparison
	if !hmac.Equal([]byte(signatureHeader), []byte(expectedSignature)) {
		log.Printf("SECURITY: Invalid YooKassa webhook signature. Expected: %s, Got: %s",
			expectedSignature, signatureHeader)
		return fmt.Errorf("неверная подпись webhook")
	}

	return nil
}

// YooKassaWebhook обрабатывает webhook от YooKassa
// POST /api/v1/payments/webhook
// @Summary      YooKassa webhook
// @Description  Handle YooKassa payment confirmation webhook (public endpoint)
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        payload  body      models.YooKassaWebhookRequest  true  "Webhook payload"
// @Success      200  {object}  response.SuccessResponse{data=map[string]string}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /payments/webhook [post]
func (h *PaymentHandler) YooKassaWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Читаем тело запроса для проверки подписи
	// Необходимо прочитать тело до его парсинга, так как io.Reader может быть прочитан только один раз
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read webhook body: %v", err)
		response.BadRequest(w, response.ErrCodeInvalidInput, "Failed to read request body")
		return
	}

	// Парсим webhook payload ПЕРВОЙ (до проверки подписи)
	// Это позволяет возвращать 400 для невалидного JSON
	var webhook models.YooKassaWebhookRequest
	parseErr := json.Unmarshal(body, &webhook)

	// Специальная обработка для пустого body - проверяем подпись
	// так как парсинг пустого JSON вернет ошибку
	if len(body) == 0 {
		// Проверяем подпись для пустого body
		if err := h.verifyYooKassaSignature(r, body); err != nil {
			log.Printf("SECURITY: Webhook signature verification failed: %v", err)
			response.Error(w, http.StatusUnauthorized, response.ErrCodeInvalidSignature, "Webhook signature verification failed")
			return
		}
		// Если подпись валидна, но JSON не парсится, это ошибка payload
		log.Printf("ERROR: Failed to parse webhook payload: %v", parseErr)
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid webhook payload")
		return
	}

	// Если есть ошибка парсинга JSON (для не-пустого body), возвращаем 400 INVALID_INPUT
	if parseErr != nil {
		log.Printf("ERROR: Failed to parse webhook payload: %v", parseErr)
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid webhook payload")
		return
	}

	// Проверяем HMAC подпись webhook ТОЛЬКО если JSON валидный
	// Это критично для безопасности - подпись вычисляется на основе сырых bytes
	if err := h.verifyYooKassaSignature(r, body); err != nil {
		log.Printf("SECURITY: Webhook signature verification failed: %v", err)
		response.Error(w, http.StatusUnauthorized, response.ErrCodeInvalidSignature, "Webhook signature verification failed")
		return
	}

	// Обрабатываем событие в зависимости от типа
	switch webhook.Event {
	case "payment.succeeded":
		// Платеж успешно оплачен
		if h.paymentService != nil {
			if err := h.paymentService.ProcessPaymentSuccess(ctx, webhook.Object.ID); err != nil {
				log.Printf("ERROR: Failed to process payment success for payment ID %s: %v", webhook.Object.ID, err)
				response.InternalError(w, "Failed to process payment")
				return
			}
		}
		log.Printf("INFO: Payment %s processed successfully", webhook.Object.ID)

	case "payment.canceled":
		// Платеж отменен
		if h.paymentService != nil {
			if err := h.paymentService.ProcessPaymentCancellation(ctx, webhook.Object.ID); err != nil {
				log.Printf("ERROR: Failed to process payment cancellation for payment ID %s: %v", webhook.Object.ID, err)
				response.InternalError(w, "Failed to process cancellation")
				return
			}
		}
		log.Printf("INFO: Payment %s cancellation processed", webhook.Object.ID)

	default:
		// Неизвестное событие, логируем и возвращаем 200
		log.Printf("INFO: Received unknown webhook event type: %s", webhook.Event)
	}

	// Возвращаем 200 OK для подтверждения получения webhook
	w.WriteHeader(http.StatusOK)
}
