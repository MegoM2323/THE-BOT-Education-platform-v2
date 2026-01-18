package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"tutoring-platform/internal/config"
	"tutoring-platform/internal/database"
	"tutoring-platform/internal/handlers"
	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
	"tutoring-platform/pkg/auth"
	"tutoring-platform/pkg/logger"
	"tutoring-platform/pkg/metrics"
	"tutoring-platform/pkg/telegram"
)

// loadEnvFile загружает переменные окружения из .env файла
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// Если файл не существует, это не критическая ошибка - используем переменные окружения системы
		if os.IsNotExist(err) {
			log.Warn().Str("file", filename).Msg(".env file not found, using system environment variables")
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Пропускаем пустые строки и комментарии
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Разбираем строку вида KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Не перезаписываем переменные окружения, которые уже установлены
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func main() {
	// Load environment variables from .env file
	if err := loadEnvFile(".env"); err != nil {
		log.Warn().Err(err).Msg("Failed to load .env file")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Setup structured logging based on environment
	logger.Setup(cfg.Server.Env)

	log.Info().Str("env", cfg.Server.Env).Str("port", cfg.Server.Port).Str("config", cfg.String()).Msg("Starting Tutoring Platform API Server")

	// Connect to database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	// NOTE: Do NOT defer db.Close() here - database must be closed AFTER all goroutines stop
	// See graceful shutdown sequence at end of main() (Phase 3)

	// initializeApp handles all remaining initialization with proper error collection and cleanup
	// On error, it will clean up resources before returning
	if err := initializeApp(cfg, db); err != nil {
		// Log the initialization error with context
		log.Error().Err(err).Msg("Application initialization failed, cleaning up resources")
		// Close database before exiting to prevent resource leaks
		if closeErr := db.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing database during error cleanup")
		}
		log.Fatal().Err(err).Msg("Fatal initialization error")
	}
}

// initializeApp handles all application initialization after database connection.
// It collects all errors and ensures proper cleanup on failure.
// Returns error if any critical initialization step fails.
func initializeApp(cfg *config.Config, db *database.DB) error {
	log.Info().Msg("Database connected successfully")

	// Create context for graceful shutdown of health check goroutine
	healthCheckCtx, cancelHealthCheck := context.WithCancel(context.Background())
	// NOTE: Do NOT defer cancelHealthCheck() - it will be called explicitly in shutdown sequence
	_ = cancelHealthCheck // mark as used to avoid lint warnings

	// Start periodic database health check and metrics collection
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		failureCount := 0
		const (
			healthCheckTimeout = 5 * time.Second
			slowHealthCheckMs  = 1000 // Log if health check takes longer than 1 second
		)

		for {
			select {
			case <-healthCheckCtx.Done():
				log.Debug().Msg("Health check goroutine shutting down")
				return
			case <-ticker.C:
				// Create context with timeout, derived from healthCheckCtx for proper cancellation propagation
				ctx, cancel := context.WithTimeout(healthCheckCtx, healthCheckTimeout)

				// Measure health check duration
				startTime := time.Now()
				err := db.Pool.Ping(ctx)
				duration := time.Since(startTime)
				cancel()

				// Check if context was cancelled (shutdown signal)
				if healthCheckCtx.Err() != nil {
					log.Debug().Msg("Health check interrupted by shutdown signal")
					return
				}

				// Log slow health checks for monitoring
				if duration.Milliseconds() > int64(slowHealthCheckMs) {
					log.Warn().Int64("duration_ms", duration.Milliseconds()).Msg("Slow database health check")
				}

				if err != nil {
					failureCount++
					log.Warn().Err(err).Int("failure_count", failureCount).Int("max_failures", 3).Msg("Database health check failed")
					metrics.DBErrorsTotal.Inc()

					if failureCount >= 3 {
						log.Fatal().Msg("Database connection lost after 3 consecutive failures, shutting down")
					}
				} else {
					if failureCount > 0 {
						log.Info().Int("previous_failures", failureCount).Msg("Database health check recovered")
					}
					failureCount = 0
				}

				// Обновляем метрики подключений к БД
				stats := db.Pool.Stat()
				metrics.DBConnectionsActive.Set(float64(stats.AcquiredConns()))
				metrics.DBConnectionsIdle.Set(float64(stats.IdleConns()))
			}
		}
	}()

	// Start periodic expired sessions cleanup (every hour)
	sessionCleanupCtx, cancelSessionCleanup := context.WithCancel(context.Background())
	_ = cancelSessionCleanup // mark as used
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-sessionCleanupCtx.Done():
				log.Debug().Msg("Session cleanup goroutine shutting down")
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(sessionCleanupCtx, 30*time.Second)
				rowsDeleted, err := db.CleanupExpiredSessions(ctx)
				cancel()

				if sessionCleanupCtx.Err() != nil {
					return
				}

				if err != nil {
					log.Warn().Err(err).Msg("Failed to cleanup expired sessions")
				} else if rowsDeleted > 0 {
					log.Info().Int64("deleted", rowsDeleted).Msg("Cleaned up expired sessions")
				}
			}
		}
	}()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Sqlx)
	lessonRepo := repository.NewLessonRepository(db.Sqlx)
	bookingRepo := repository.NewBookingRepository(db.Sqlx)
	creditRepo := repository.NewCreditRepository(db.Sqlx)
	swapRepo := repository.NewSwapRepository(db.Sqlx)
	sessionRepo := repository.NewSessionRepository(db.Sqlx)
	trialRequestRepo := repository.NewTrialRequestRepository(db.Sqlx)
	telegramUserRepo := repository.NewTelegramUserRepository(db.Sqlx)
	telegramTokenRepo := repository.NewTelegramTokenRepository(db.Sqlx)
	broadcastRepo := repository.NewBroadcastRepository(db.Sqlx)
	broadcastListRepo := repository.NewBroadcastListRepository(db.Sqlx)
	lessonTemplateRepo := repository.NewLessonTemplateRepository(db.Sqlx)
	templateLessonRepo := repository.NewTemplateLessonRepository(db.Sqlx)
	templateApplicationRepo := repository.NewTemplateApplicationRepository(db.Sqlx)
	lessonModificationRepo := repository.NewLessonModificationRepository(db.Sqlx)
	paymentRepo := repository.NewPaymentRepository(db.Sqlx)
	chatRepo := repository.NewChatRepository(db.Sqlx)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(db.Sqlx)
	homeworkRepo := repository.NewHomeworkRepository(db.Sqlx)
	lessonBroadcastRepo := repository.NewLessonBroadcastRepository(db.Sqlx)
	subjectRepo := repository.NewSubjectRepository(db.Sqlx)

	// Initialize validators
	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)
	swapValidator := validator.NewSwapValidator(lessonRepo, bookingRepo)
	trialRequestValidator := validator.NewTrialRequestValidator()

	// Initialize session manager
	sessionMgr := auth.NewSessionManager(cfg.Session.Secret)

	// Initialize Telegram client (if token is configured)
	var telegramClient *telegram.Client
	var telegramService *service.TelegramService
	var broadcastService *service.BroadcastService
	var telegramHandler *handlers.TelegramHandler
	var broadcastHandler *handlers.BroadcastHandler
	var adminTelegramHandler *handlers.AdminTelegramHandler
	var botUsername string

	if cfg.Telegram.BotToken != "" {
		telegramClient = telegram.NewClientWithProxy(cfg.Telegram.BotToken, cfg.Telegram.ProxyURL)

		// Validate bot token by calling getMe
		botInfo, err := telegramClient.GetMe()
		if err != nil {
			// Return error instead of fatal - allows cleanup on error
			return fmt.Errorf("invalid Telegram bot token: %w", err)
		}
		botUsername = botInfo.Username
		log.Info().Str("bot_username", botUsername).Msg("Telegram bot connected")

		// Setup webhook if in production mode
		if cfg.IsProduction() {
			webhookURL := cfg.GetWebhookURL()
			if webhookURL == "" {
				// Return error instead of fatal - allows cleanup on error
				return fmt.Errorf("PRODUCTION_DOMAIN is required for Telegram webhook")
			}

			if err := telegramClient.SetWebhook(webhookURL, cfg.Telegram.WebhookSecret); err != nil {
				// Log warning but don't fail startup - webhook will be retried in background
				log.Warn().Err(err).Str("webhook_url", webhookURL).Msg("Failed to register Telegram webhook on startup, will retry in background")

				// Background task to retry webhook setup after nginx is ready
				go func() {
					maxRetries := 30
					for attempt := 0; attempt < maxRetries; attempt++ {
						time.Sleep(2 * time.Second)
						if err := telegramClient.SetWebhook(webhookURL, cfg.Telegram.WebhookSecret); err == nil {
							log.Info().Str("webhook_url", webhookURL).Msg("Telegram webhook registered successfully (background retry)")
							return
						}
						if attempt%5 == 0 {
							log.Debug().Err(err).Int("attempt", attempt+1).Msg("Retrying Telegram webhook registration")
						}
					}
					log.Error().Str("webhook_url", webhookURL).Msg("Failed to register Telegram webhook after maximum retries")
				}()
			} else {
				log.Info().Str("webhook_url", webhookURL).Msg("Telegram webhook registered")
			}
		} else {
			// In development, remove webhook (use polling mode)
			if err := telegramClient.DeleteWebhook(); err != nil {
				log.Warn().Err(err).Msg("Failed to delete Telegram webhook")
			}
			log.Info().Msg("Telegram polling mode enabled (development)")
		}

		// Validate admin Telegram ID if configured
		if cfg.Telegram.AdminTelegramID <= 0 {
			log.Warn().Msg("ADMIN_TELEGRAM_ID not configured, admin notifications will be skipped")
		}
	} else {
		log.Warn().Msg("Telegram is not configured (TELEGRAM_BOT_TOKEN not set). Telegram features will be disabled")
	}

	// Initialize YooKassa client (if configured)
	var yookassaClient *service.YooKassaClient
	var paymentService *service.PaymentService
	if cfg.YooKassa.ShopID != "" && cfg.YooKassa.SecretKey != "" {
		yookassaClient = service.NewYooKassaClient(cfg.YooKassa.ShopID, cfg.YooKassa.SecretKey)
		log.Info().Msg("YooKassa payment gateway configured")
	} else {
		log.Warn().Msg("YooKassa is not configured (YOOKASSA_SHOP_ID or YOOKASSA_SECRET_KEY not set). Payment features will be disabled")
	}

	// Initialize services
	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, cfg.Session.MaxAge)
	userService := service.NewUserService(userRepo, creditRepo)
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(db.Pool, bookingRepo, lessonRepo, creditRepo, cancelledBookingRepo, bookingValidator)

	// Wire up booking creator for lesson service (enables enrolling students on lesson creation)
	lessonService.SetBookingCreator(bookingService)
	creditService := service.NewCreditService(db.Pool, creditRepo)
	swapService := service.NewSwapService(db.Pool, swapRepo, bookingRepo, lessonRepo, swapValidator)
	trialRequestService := service.NewTrialRequestService(trialRequestRepo, trialRequestValidator, nil) // TelegramService will be created below
	templateService := service.NewTemplateService(db.Sqlx, lessonTemplateRepo, templateLessonRepo, templateApplicationRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	bulkEditService := service.NewBulkEditService(db.Pool, lessonRepo, lessonModificationRepo, userRepo, creditRepo)

	// Initialize chat service (moderation will be handled by the service internally)
	chatService := service.NewChatService(chatRepo, userRepo, nil)

	// Initialize homework service
	homeworkService := service.NewHomeworkService(homeworkRepo, lessonRepo, bookingRepo, userRepo)

	// Initialize broadcast service always (for list management), even without Telegram bot
	// Pass nil telegramClient if not configured - sending will fail gracefully but CRUD operations work
	broadcastService = service.NewBroadcastService(broadcastRepo, broadcastListRepo, telegramUserRepo, userRepo, telegramClient)

	// Initialize lesson broadcast service (for lesson-specific broadcasts to enrolled students)
	// Pass nil telegramClient if not configured - sending will fail gracefully
	uploadDir := "./uploads/lesson_broadcasts"
	lessonBroadcastService := service.NewLessonBroadcastService(db.Sqlx, lessonBroadcastRepo, lessonRepo, userRepo, telegramUserRepo, telegramClient, uploadDir)

	// Initialize payment service only if YooKassa is configured
	if yookassaClient != nil {
		paymentService = service.NewPaymentService(db.Pool, paymentRepo, creditService, yookassaClient, userRepo, cfg.YooKassa.ReturnURL)
	}

	// Initialize payment settings service
	paymentSettingsService := service.NewPaymentSettingsService(userRepo)

	// Initialize Telegram-related services only if token is configured
	if telegramClient != nil {
		telegramService = service.NewTelegramService(telegramUserRepo, telegramTokenRepo, userRepo, telegramClient, cfg.Telegram.AdminTelegramID)

		// Now update TrialRequestService with TelegramService
		trialRequestService.SetTelegramService(telegramService)

		// Start polling in development mode (production uses webhook)
		if cfg.IsDevelopment() {
			// Создаём обработчик, который передаёт обновления в TelegramService
			updateHandler := func(ctx context.Context, update *telegram.Update) error {
				return telegramService.HandleWebhook(ctx, update)
			}

			if err := telegramClient.StartPolling(updateHandler); err != nil {
				log.Warn().Err(err).Msg("Failed to start Telegram polling")
			}
		}
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddlewareWithSessionMaxAge(authService, cfg.IsProduction(), cfg.Session.SameSite, int(cfg.Session.MaxAge.Seconds()))
	corsConfig := middleware.DefaultCORSConfig()

	// Add production domain to CORS if configured
	if cfg.Server.ProductionDomain != "" {
		corsConfig.AllowedOrigins = append(
			corsConfig.AllowedOrigins,
			"https://"+cfg.Server.ProductionDomain,
		)
		log.Info().Str("domain", cfg.Server.ProductionDomain).Msg("Added production domain to CORS allowed origins")
	}

	// Initialize CSRF store для защиты от CSRF атак
	csrfStore := middleware.NewCSRFTokenStore()

	// Initialize rate limiters для защиты от brute-force атак
	loginRateLimiter := middleware.LoginRateLimiter()
	trialRequestRateLimiter := middleware.TrialRequestRateLimiter()
	paymentRateLimiter := middleware.PaymentRateLimiter() // 10 requests/min per user

	// Initialize body limit config для защиты от DoS атак через большие payload'ы
	bodyLimitConfig := middleware.DefaultBodyLimitConfig()

	// Initialize handlers
	authHandler := handlers.NewAuthHandlerWithSameSite(authService, creditService, int(cfg.Session.MaxAge.Seconds()), cfg.IsProduction(), cfg.Session.SameSite)
	authHandler.SetUserService(userService)
	authHandler.SetCSRFStore(csrfStore)
	userHandler := handlers.NewUserHandler(userService)
	lessonHandler := handlers.NewLessonHandler(lessonService, bookingService, bulkEditService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	creditHandler := handlers.NewCreditHandler(creditService, userService)
	swapHandler := handlers.NewSwapHandler(swapService)
	trialRequestHandler := handlers.NewTrialRequestHandler(trialRequestService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	teacherHandler := handlers.NewTeacherHandler(lessonService, bookingService, lessonBroadcastService, lessonRepo)
	chatHandler := handlers.NewChatHandler(chatService)
	homeworkHandler := handlers.NewHomeworkHandler(homeworkService)
	lessonBroadcastHandler := handlers.NewLessonBroadcastHandler(lessonBroadcastService, uploadDir)
	subjectsHandler := handlers.NewSubjectsHandler(subjectRepo)

	// Initialize broadcast handler always (for list management CRUD)
	// telegramService can be nil - SendBroadcast will fail gracefully, but list CRUD works
	broadcastHandler = handlers.NewBroadcastHandler(broadcastService, telegramService)

	// Initialize Telegram handlers only if service is available
	if telegramService != nil {
		telegramHandler = handlers.NewTelegramHandler(telegramService, botUsername, cfg.Telegram.WebhookSecret)
		adminTelegramHandler = handlers.NewAdminTelegramHandler(telegramService, userService, telegramUserRepo)
	}

	// Initialize payment handler only if service is available
	var paymentHandler *handlers.PaymentHandler
	if paymentService != nil {
		paymentHandler = handlers.NewPaymentHandler(paymentService, cfg)
	}

	// Initialize payment settings handler
	paymentSettingsHandler := handlers.NewPaymentSettingsHandler(paymentSettingsService)

	// Initialize health handler
	healthHandler := handlers.NewHealthHandler(db.Pool)

	// Setup router
	r := chi.NewRouter()

	// Global middleware (порядок важен!)
	r.Use(chiMiddleware.RequestID)                         // 1. Генерируем Request ID для трекинга
	r.Use(chiMiddleware.RealIP)                            // 2. Определяем реальный IP клиента
	r.Use(middleware.LoggingMiddleware)                    // 3. Логируем все запросы с метриками
	r.Use(middleware.MetricsMiddleware)                    // 4. Собираем Prometheus метрики
	r.Use(middleware.BodyLimitMiddleware(bodyLimitConfig)) // 5. Ограничиваем размер тела запроса (защита от DoS)

	// Debug logging только в development режиме
	if cfg.IsDevelopment() {
		// 		// Debug logging removed (zerolog dependency removed)
	}

	r.Use(chiMiddleware.Recoverer)               // 6. Обработка паник
	r.Use(middleware.CORSMiddleware(corsConfig)) // 7. CORS headers

	// Health check endpoint with database connectivity verification
	r.Get("/health", healthHandler.HealthCheck)

	// Prometheus metrics endpoint (без auth для Prometheus scraper)
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			// Auth routes с rate limiting для защиты от brute-force
			r.With(middleware.RateLimitMiddleware(loginRateLimiter)).Post("/auth/register", authHandler.Register)
			r.With(middleware.RateLimitMiddleware(loginRateLimiter)).Post("/auth/login", authHandler.Login)
			r.With(middleware.RateLimitMiddleware(loginRateLimiter)).Post("/auth/register-telegram", authHandler.RegisterViaTelegram)
			// Trial requests с rate limiting для защиты от спама
			r.With(middleware.RateLimitMiddleware(trialRequestRateLimiter)).Post("/trial-requests", trialRequestHandler.CreateTrialRequest)
			// Public subjects list (no authentication required for browsing subjects)
			r.Get("/subjects", subjectsHandler.GetSubjects)
			// Telegram webhook - public endpoint (only if Telegram is configured)
			if telegramHandler != nil {
				r.Post("/telegram/webhook", telegramHandler.HandleWebhook)
			}
			// YooKassa webhook - public endpoint (only if payments are configured)
			if paymentHandler != nil {
				r.Post("/payments/webhook", paymentHandler.YooKassaWebhook)
			}
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// GET /csrf-token endpoint (доступен без CSRF токена)
			r.Get("/csrf-token", authHandler.GetCSRFToken)

			// Auth routes (без CSRF для GET, с CSRF для POST/PUT)
			r.Route("/auth", func(r chi.Router) {
				r.Get("/me", authHandler.GetMe)
				// State-changing endpoints с CSRF protection
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/logout", authHandler.Logout)
				r.With(middleware.CSRFMiddleware(csrfStore)).Put("/profile", authHandler.UpdateProfile)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/change-password", authHandler.ChangePassword)
			})

			// User routes
			r.Route("/users", func(r chi.Router) {
				// GET /users with role filter accessible to all authenticated users (for chat)
				// GET /users/{id} and all write operations require admin
				r.Get("/", userHandler.GetUsers)

				// Admin-only routes (GET specific user + all write operations)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdmin)
					r.Get("/{id}", userHandler.GetUser)
					// State-changing endpoints с CSRF protection
					r.With(middleware.CSRFMiddleware(csrfStore)).Post("/", userHandler.CreateUser)
					r.With(middleware.CSRFMiddleware(csrfStore)).Put("/{id}", userHandler.UpdateUser)
					r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{id}", userHandler.DeleteUser)
					r.With(middleware.CSRFMiddleware(csrfStore)).Post("/{id}/credits", creditHandler.AddCredits)

					// Admin Telegram management routes (only if Telegram is configured)
					if adminTelegramHandler != nil {
						r.Get("/telegram", adminTelegramHandler.ListUsersWithTelegram)
						r.Get("/{id}/telegram", adminTelegramHandler.GetUserTelegramInfo)
						r.With(middleware.CSRFMiddleware(csrfStore)).Put("/{id}/telegram", adminTelegramHandler.SetUserTelegram)
						r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{id}/telegram", adminTelegramHandler.UnlinkUserTelegram)
						r.With(middleware.CSRFMiddleware(csrfStore)).Post("/{id}/telegram/message", adminTelegramHandler.SendMessageToUser)
					}
				})
			})

			// Subject routes
			r.Route("/subjects", func(r chi.Router) {
				// GET single subject endpoint (accessible to authenticated users)
				r.Get("/{id}", subjectsHandler.GetSubject)

				// Admin or Methodologist routes for subject management
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdminOrMethodologist)
					// State-changing endpoints с CSRF protection
					r.With(middleware.CSRFMiddleware(csrfStore)).Post("/", subjectsHandler.CreateSubject)
					r.With(middleware.CSRFMiddleware(csrfStore)).Put("/{id}", subjectsHandler.UpdateSubject)
					r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{id}", subjectsHandler.DeleteSubject)
				})
			})

			// Teacher subject routes
			r.Get("/my-subjects", subjectsHandler.GetMySubjects)
			r.Route("/teachers/{id}/subjects", func(r chi.Router) {
				r.Get("/", subjectsHandler.GetTeacherSubjects)
				// Admin or Methodologist routes for assigning subjects to teachers
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdminOrMethodologist)
					r.With(middleware.CSRFMiddleware(csrfStore)).Post("/", subjectsHandler.AssignSubjectToTeacher)
					r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{subjectId}", subjectsHandler.RemoveSubjectFromTeacher)
				})
			})

			// Lesson routes
			r.Route("/lessons", func(r chi.Router) {
				// Public lesson endpoints (GET only)
				r.Get("/", lessonHandler.GetLessons)
				r.Get("/my", lessonHandler.GetMyLessons)
				r.Get("/{id}", lessonHandler.GetLesson)
				r.Get("/{id}/students", lessonHandler.GetLessonStudents)

				// Homework routes
				r.Route("/{id}/homework", func(r chi.Router) {
					// GET homework list - доступно всем авторизованным пользователям
					r.Get("/", homeworkHandler.GetHomework)
					// Upload homework - только admin или teacher урока (CSRF protected, with 10MB body limit for file uploads)
					r.With(middleware.BodyLimitMiddlewareForFileUpload(bodyLimitConfig), middleware.CSRFMiddleware(csrfStore)).Post("/", homeworkHandler.UploadHomework)
					// Delete homework file - только creator или admin (CSRF protected)
					r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{file_id}", homeworkHandler.DeleteHomework)
					// Update homework description - только admin, creator или teacher урока (CSRF protected)
					r.With(middleware.CSRFMiddleware(csrfStore)).Patch("/{file_id}", homeworkHandler.UpdateHomework)
					// Download homework file - доступно всем авторизованным пользователям
					r.Get("/{file_id}/download", homeworkHandler.DownloadHomework)
				})

				// Lesson broadcast routes - доступно admin или teacher урока
				r.Route("/{id}/broadcasts", func(r chi.Router) {
					// GET broadcasts - доступно всем авторизованным пользователям
					r.Get("/", lessonBroadcastHandler.ListBroadcasts)
					r.Get("/{broadcast_id}", lessonBroadcastHandler.GetBroadcast)
					// Create broadcast - только admin или teacher урока (CSRF protected, with 10MB body limit for file uploads)
					r.With(middleware.BodyLimitMiddlewareForFileUpload(bodyLimitConfig), middleware.CSRFMiddleware(csrfStore)).Post("/", lessonBroadcastHandler.CreateBroadcast)
					// Download broadcast file - доступно всем авторизованным пользователям с проверкой прав
					r.Get("/{broadcast_id}/files/{file_id}/download", lessonBroadcastHandler.DownloadBroadcastFile)
				})

				// Admin or Methodologist lesson endpoints с CSRF protection
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdminOrMethodologist)
					r.Use(middleware.CSRFMiddleware(csrfStore))
					r.Post("/", lessonHandler.CreateLesson)
					r.Delete("/{id}", lessonHandler.DeleteLesson)
					r.Post("/{id}/apply-to-all", lessonHandler.ApplyToAllSubsequent)
				})

				// Update lesson - доступно admin и teacher (проверка прав внутри обработчика)
				r.With(middleware.CSRFMiddleware(csrfStore)).Put("/{id}", lessonHandler.UpdateLesson)
			})

			// Template routes (Admin and Methodologist)
			r.Route("/templates", func(r chi.Router) {
				r.Use(middleware.RequireAdminOrMethodologist)
				r.Get("/", templateHandler.GetTemplates)
				r.Get("/{id}", templateHandler.GetTemplate)
				// State-changing endpoints с CSRF protection
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/", templateHandler.CreateTemplate)
				r.With(middleware.CSRFMiddleware(csrfStore)).Put("/{id}", templateHandler.UpdateTemplate)
				r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{id}", templateHandler.DeleteTemplate)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/{id}/apply", templateHandler.ApplyTemplate)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/{id}/rollback", templateHandler.RollbackTemplate)

				// Template lesson CRUD endpoints с CSRF protection
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/{id}/lessons", templateHandler.CreateTemplateLesson)
				r.With(middleware.CSRFMiddleware(csrfStore)).Put("/{id}/lessons/{lesson_id}", templateHandler.UpdateTemplateLesson)
				r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{id}/lessons/{lesson_id}", templateHandler.DeleteTemplateLesson)
			})

			// Booking routes
			r.Route("/bookings", func(r chi.Router) {
				r.Get("/", bookingHandler.ListBookings)
				r.Get("/cancelled-lessons", bookingHandler.GetCancelledLessons)
				r.Get("/{id}/status", bookingHandler.GetBookingStatus)
				r.Get("/{id}", bookingHandler.GetBooking)
				// State-changing endpoints с CSRF protection
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/", bookingHandler.CreateBooking)
				r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/{id}", bookingHandler.CancelBooking)
			})

			// Credit routes
			r.Route("/credits", func(r chi.Router) {
				r.Get("/", creditHandler.GetMyCredits)
				r.Get("/balance", creditHandler.GetBalance) // Optimized endpoint for sidebar polling
				r.Get("/history", creditHandler.GetMyHistory)
				// Admin and methodologist routes (GET only - CSRF не нужен)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdminOrMethodologist)
					r.Get("/all", creditHandler.GetAllCredits)
					r.Get("/user/{id}", creditHandler.GetUserCredits)
				})
			})

			// Swap routes с CSRF protection
			r.Route("/swaps", func(r chi.Router) {
				r.Get("/history", swapHandler.GetSwapHistory)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/", swapHandler.PerformSwap)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/validate", swapHandler.ValidateSwap)
			})

			// Payment routes (only if YooKassa is configured)
			if paymentHandler != nil {
				r.Route("/payments", func(r chi.Router) {
					r.Get("/history", paymentHandler.GetHistory)
					r.With(middleware.UserRateLimitMiddleware(paymentRateLimiter), middleware.CSRFMiddleware(csrfStore)).Post("/create", paymentHandler.CreatePayment)
				})
			}

			// Trial requests - GET list (admin or methodologist)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdminOrMethodologist)
				r.Get("/trial-requests", trialRequestHandler.GetTrialRequests)
			})

			// Telegram routes - authenticated users (only if Telegram is configured)
			if telegramHandler != nil {
				r.Route("/telegram", func(r chi.Router) {
					r.Get("/me", telegramHandler.GetMyTelegramLink)
					r.Get("/link-token", telegramHandler.GenerateLinkToken)
					r.With(middleware.CSRFMiddleware(csrfStore)).Post("/subscribe", telegramHandler.Subscribe)
					r.With(middleware.CSRFMiddleware(csrfStore)).Post("/unsubscribe", telegramHandler.Unsubscribe)
					r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/link", telegramHandler.UnlinkTelegram)
				})
			}

			// Telegram broadcast routes - admin and methodologist (always available, even without Telegram)
			// List management (CRUD) works without bot, but sending requires bot configured
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdminOrMethodologist)

				// Telegram users management (GET only)
				r.Get("/admin/telegram/users", broadcastHandler.GetLinkedUsers)

				// Broadcast lists management (GET + CSRF protected state-changing)
				r.Get("/admin/telegram/lists", broadcastHandler.GetBroadcastLists)
				r.Get("/admin/telegram/lists/{id}", broadcastHandler.GetBroadcastListByID)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/admin/telegram/lists", broadcastHandler.CreateBroadcastList)
				r.With(middleware.CSRFMiddleware(csrfStore)).Put("/admin/telegram/lists/{id}", broadcastHandler.UpdateBroadcastList)
				r.With(middleware.CSRFMiddleware(csrfStore)).Delete("/admin/telegram/lists/{id}", broadcastHandler.DeleteBroadcastList)

				// Broadcasts (GET + CSRF protected sending)
				r.Get("/admin/telegram/broadcasts", broadcastHandler.GetBroadcasts)
				r.Get("/admin/telegram/broadcasts/{id}", broadcastHandler.GetBroadcastDetails)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/admin/telegram/broadcast", broadcastHandler.SendBroadcast)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/admin/telegram/broadcasts/{id}/cancel", broadcastHandler.CancelBroadcast)
			})

			// Payment settings management - admin only (GET + CSRF protected state-changing)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdminOnly)

				r.Get("/admin/payment-settings", paymentSettingsHandler.ListStudentsPaymentStatus)
				r.With(middleware.CSRFMiddleware(csrfStore)).Put("/admin/users/{id}/payment-settings", paymentSettingsHandler.UpdatePaymentStatus)
			})

			// Chat routes (authenticated users - students and teachers)
			r.Route("/chat", func(r chi.Router) {
				// Get user's chat rooms
				r.Get("/rooms", chatHandler.GetMyRooms)
				// Create/get room with participant
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/rooms", chatHandler.GetOrCreateRoom)
				// Messages in a room
				r.Route("/rooms/{roomId}", func(r chi.Router) {
					// Get messages with pagination
					r.Get("/messages", chatHandler.GetMessages)
					// Send message (with optional file attachments, with 5MB body limit)
					r.With(middleware.BodyLimitMiddlewareWithLimit(middleware.BodyLimitLarge), middleware.CSRFMiddleware(csrfStore)).Post("/messages", chatHandler.SendMessage)
					// Download file attachment
					r.Get("/files/{fileId}", chatHandler.DownloadFile)
				})
			})

			// Teacher routes (teacher-only endpoints)
			r.Route("/teacher", func(r chi.Router) {
				r.Use(middleware.RequireTeacher)

				// Teacher schedule - calendar view with lessons and enrolled students (GET only)
				r.Get("/schedule", teacherHandler.GetTeacherSchedule)

				// Lesson broadcasts - send message to all students in a lesson (CSRF protected)
				r.With(middleware.CSRFMiddleware(csrfStore)).Post("/lessons/{id}/broadcast", teacherHandler.SendLessonBroadcast)
			})

		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	// Use a channel to capture server startup errors
	serverErrChan := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("Server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	// Give server a brief moment to start, checking for immediate errors
	select {
	case err := <-serverErrChan:
		// Server failed to start - return error for cleanup
		cancelHealthCheck() // Cancel health check before returning
		return fmt.Errorf("server failed to start: %w", err)
	case <-time.After(100 * time.Millisecond):
		// Server started successfully, continue
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Server is shutting down")

	// GRACEFUL SHUTDOWN SEQUENCE (CRITICAL - ORDER MATTERS)
	// Purpose: Prevent panics from goroutines accessing database after Close()
	//         Ensure all dependencies are cleaned before DB connection closes
	//
	// Order of operations:
	// Phase 1: Shutdown HTTP server (stops accepting new requests)
	// Phase 2: Stop all background goroutines in reverse order of creation
	//          - Health check goroutine (primary consumer of DB)
	//          - Rate limiters (cleanup goroutines)
	//          - Telegram service (token cleanup goroutines, polling)
	//          - Broadcast service (broadcast sending goroutines)
	// Phase 3: Wait brief grace period for goroutines to exit gracefully
	// Phase 4: Close database connection (after all goroutines have stopped)
	//
	// This ensures no goroutine tries to use DB after it's closed

	// PHASE 1: Shutdown HTTP server (stops accepting new requests)
	log.Debug().Msg("Phase 1: Shutting down HTTP server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}
	log.Debug().Msg("Phase 1: HTTP server shutdown complete")

	// PHASE 2: Stop all background goroutines that use database
	// These must complete before database is closed
	log.Debug().Msg("Phase 2: Stopping background goroutines")

	// 2a. Stop health check goroutine
	cancelHealthCheck()
	log.Debug().Msg("  - Health check goroutine cancelled")

	// 2a2. Stop session cleanup goroutine
	cancelSessionCleanup()
	log.Debug().Msg("  - Session cleanup goroutine cancelled")

	// 2b. Stop rate limiter cleanup goroutines (global)
	loginRateLimiter.Stop()
	trialRequestRateLimiter.Stop()
	paymentRateLimiter.Stop()
	log.Debug().Msg("  - Rate limiter cleanup stopped")

	// 2c. Shutdown Telegram service (if it was initialized)
	// This stops any background token cleanup goroutines and message sending
	if telegramService != nil {
		telegramService.Shutdown()
		log.Debug().Msg("  - Telegram service shutdown complete")
	}

	// 2d. Shutdown Broadcast service (if it was initialized)
	// This stops the internal rate limiter and cancels all active broadcast goroutines
	if broadcastService != nil {
		broadcastService.Shutdown()
		log.Debug().Msg("  - Broadcast service shutdown complete")
	}

	// 2e. Stop Telegram polling if it was started (development mode)
	// Must be done after TelegramService.Shutdown() to avoid race conditions
	if telegramClient != nil && cfg.IsDevelopment() {
		telegramClient.StopPolling()
		log.Debug().Msg("  - Telegram polling stopped")
	}

	// PHASE 3: Wait for background goroutines to exit
	// This gives background tasks time to notice context cancellation and cleanup
	// Goroutines must exit before database is closed to prevent:
	// - Panic from accessing closed connections
	// - Race conditions in cleanup code
	// - Orphaned database handles
	shutdownGracePeriod := time.Duration(200) * time.Millisecond
	log.Debug().Dur("grace_period", shutdownGracePeriod).Msg("Waiting for background goroutines to exit")
	time.Sleep(shutdownGracePeriod)

	// PHASE 4: Close database connection
	// At this point, all goroutines that use the database should have stopped:
	// - Health check goroutine (cancelled via healthCheckCtx)
	// - Broadcast goroutines (cancelled via broadcastService.Shutdown())
	// - Telegram token cleanup (cancelled via telegramService.Shutdown())
	// - Rate limiter cleanup (stopped via Stop() calls)
	// - Any in-flight HTTP requests (HTTP server already shutdown in Phase 1)
	log.Debug().Msg("Phase 4: Closing database connection")
	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing database")
	}
	log.Debug().Msg("Phase 4: Database connection closed")

	log.Info().Msg("Server shutdown complete")
	return nil
}
