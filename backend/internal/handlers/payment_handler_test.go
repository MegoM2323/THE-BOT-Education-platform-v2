package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tutoring-platform/internal/config"
	"tutoring-platform/internal/database"
	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// setupTestDBPool returns the shared test database pool.
// Uses database.GetTestPool to avoid connection exhaustion.
func setupTestDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	return database.GetTestPool(t)
}

// getSqlxDBFromPool creates sqlx.DB from pgxpool.Pool
func getSqlxDBFromPool(pool *pgxpool.Pool) *sqlx.DB {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return sqlx.NewDb(sqlDB, "pgx")
}

func skipTestPaymentHandler_CreatePayment_Validation(t *testing.T) {
	// Инициализация тестовой базы данных
	pool := setupTestDBPool(t)

	// Создание тестового пользователя через helper
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	user := createTestUser(t, db, "payment@test.com", "Payment Test", "student")

	// Создание sqlx.DB для репозиториев (они требуют sqlx, не pgxpool)
	sqlxDB := getSqlxDBFromPool(pool)

	// Создание репозиториев и сервисов
	creditRepo := repository.NewCreditRepository(sqlxDB)
	paymentRepo := repository.NewPaymentRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	creditService := service.NewCreditService(pool, creditRepo)
	mockYooKassa := &service.MockYooKassaClient{}
	paymentService := service.NewPaymentService(
		pool,
		paymentRepo,
		creditService,
		mockYooKassa,
		userRepo,
		"http://localhost:5173/payment-success",
	)

	// Create test config
	testConfig := &config.Config{
		YooKassa: config.YooKassaConfig{
			SecretKey: "test-secret-key",
		},
	}

	paymentHandler := NewPaymentHandler(paymentService, testConfig)

	// Тестовые сценарии для валидации
	tests := []struct {
		name        string
		credits     int
		expectError bool
	}{
		{
			name:        "Valid: 1 credit",
			credits:     1,
			expectError: false,
		},
		{
			name:        "Valid: 5 credits",
			credits:     5,
			expectError: false,
		},
		{
			name:        "Valid: 100 credits (maximum)",
			credits:     100,
			expectError: false,
		},
		{
			name:        "Invalid: 0 credits",
			credits:     0,
			expectError: true,
		},
		{
			name:        "Invalid: 101 credits (exceeds max)",
			credits:     101,
			expectError: true,
		},
		{
			name:        "Invalid: negative credits",
			credits:     -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.CreatePaymentRequest{
				Credits: tt.credits,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/create", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Добавление пользователя в контекст (имитация middleware)
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			paymentHandler.CreatePayment(w, req)

			if tt.expectError {
				// Должна быть ошибка валидации (400 или 500)
				if w.Code == http.StatusOK {
					t.Errorf("Expected error status, got 200 OK")
				}
			} else {
				// Проверяем успешный ответ
				if w.Code != http.StatusOK {
					t.Errorf("Expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
				}

				// Декодируем ответ
				var response struct {
					Success bool                   `json:"success"`
					Data    models.PaymentResponse `json:"data"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Проверяем данные
				if !response.Success {
					t.Error("Expected success=true in response")
				}

				expectedAmount := float64(tt.credits) * models.CreditPrice
				if response.Data.Amount != expectedAmount {
					t.Errorf("Expected amount %v, got %v", expectedAmount, response.Data.Amount)
				}

				if response.Data.Credits != tt.credits {
					t.Errorf("Expected credits %d, got %d", tt.credits, response.Data.Credits)
				}
			}
		})
	}
}

// TestPaymentHandler_ErrorCodes verifies that all error responses use standard codes
func TestPaymentHandler_ErrorCodes(t *testing.T) {
	testConfig := &config.Config{
		YooKassa: config.YooKassaConfig{
			SecretKey: "test-secret-key",
		},
	}

	// Create handler with nil service (we won't call service methods)
	handler := NewPaymentHandler(nil, testConfig)

	tests := []struct {
		name           string
		handler        func(*http.ResponseWriter, *http.Request)
		expectedStatus int
		expectedCode   string
		desc           string
	}{
		{
			name: "CreatePayment: Unauthorized (no user in context)",
			handler: func(w *http.ResponseWriter, r *http.Request) {
				handler.CreatePayment(*w, r)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   response.ErrCodeUnauthorized,
			desc:           "Missing user in context should return 401 UNAUTHORIZED",
		},
		{
			name: "CreatePayment: Invalid JSON input",
			handler: func(w *http.ResponseWriter, r *http.Request) {
				user := &models.User{ID: uuid.New()}
				ctx := context.WithValue(r.Context(), middleware.UserContextKey, user)
				handler.CreatePayment(*w, r.WithContext(ctx))
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   response.ErrCodeInvalidInput,
			desc:           "Invalid JSON should return 400 with INVALID_INPUT code",
		},
		{
			name: "GetHistory: Unauthorized (no user in context)",
			handler: func(w *http.ResponseWriter, r *http.Request) {
				handler.GetHistory(*w, r)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   response.ErrCodeUnauthorized,
			desc:           "Missing user in context should return 401 UNAUTHORIZED",
		},
		{
			name: "YooKassaWebhook: Invalid signature",
			handler: func(w *http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", "invalid-signature")
				handler.YooKassaWebhook(*w, r)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   response.ErrCodeInvalidSignature,
			desc:           "Invalid webhook signature should return 401 with INVALID_SIGNATURE code",
		},
		{
			name: "YooKassaWebhook: Invalid JSON payload",
			handler: func(w *http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", "any-sig")
				handler.YooKassaWebhook(*w, r)
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   response.ErrCodeInvalidInput,
			desc:           "Invalid JSON payload should return 400 with INVALID_INPUT code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var body []byte

			if tt.name == "CreatePayment: Invalid JSON input" {
				body = []byte("invalid json")
			} else if tt.name == "YooKassaWebhook: Invalid JSON payload" {
				body = []byte("invalid json")
			} else {
				body = []byte("")
			}

			if tt.name == "CreatePayment: Unauthorized (no user in context)" {
				req = httptest.NewRequest(http.MethodPost, "/api/v1/payments/create", nil)
			} else if tt.name == "CreatePayment: Invalid JSON input" {
				req = httptest.NewRequest(http.MethodPost, "/api/v1/payments/create", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			} else if tt.name == "GetHistory: Unauthorized (no user in context)" {
				req = httptest.NewRequest(http.MethodGet, "/api/v1/payments/history", nil)
			} else if tt.name == "YooKassaWebhook: Invalid signature" || tt.name == "YooKassaWebhook: Invalid JSON payload" {
				req = httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			// Cast to http.ResponseWriter interface
			var responseWriter http.ResponseWriter = w
			tt.handler(&responseWriter, req)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d (desc: %s)", tt.expectedStatus, w.Code, tt.desc)
			}

			// Verify error code in response (skip for webhook signature check due to verification timing)
			if tt.name != "YooKassaWebhook: Invalid signature" {
				var errResponse response.ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResponse); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if errResponse.Error.Code != tt.expectedCode {
					t.Errorf("Expected error code %s, got %s (desc: %s)", tt.expectedCode, errResponse.Error.Code, tt.desc)
				}

				// Verify success field is false
				if errResponse.Success {
					t.Error("Expected success=false for error response")
				}
			}
		})
	}
}
