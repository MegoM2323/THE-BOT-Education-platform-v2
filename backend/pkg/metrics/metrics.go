package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	// Счетчик всех HTTP запросов с метками метода, пути и статуса
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Гистограмма времени обработки HTTP запросов (для расчета перцентилей)
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		},
		[]string{"method", "path"},
	)

	// Business metrics
	// Счетчик созданных бронирований
	BookingsCreated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bookings_created_total",
			Help: "Total number of bookings created",
		},
	)

	// Счетчик отмененных бронирований
	BookingsCancelled = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bookings_cancelled_total",
			Help: "Total number of bookings cancelled",
		},
	)

	// Счетчик добавленных кредитов (сумма)
	CreditsAdded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "credits_added_total",
			Help: "Total credits added to users",
		},
	)

	// Счетчик списанных кредитов (сумма)
	CreditsDeducted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "credits_deducted_total",
			Help: "Total credits deducted from users",
		},
	)

	// Счетчик возвращенных кредитов (сумма)
	CreditsRefunded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "credits_refunded_total",
			Help: "Total credits refunded to users",
		},
	)

	// Счетчик платежей по статусу
	PaymentsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_processed_total",
			Help: "Total payments processed",
		},
		[]string{"status"}, // "succeeded", "pending", "cancelled"
	)

	// Сумма платежей в рублях
	PaymentAmountTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "payment_amount_rub_total",
			Help: "Total payment amount in RUB",
		},
	)

	// Database metrics
	// Gauge для активных подключений к базе данных
	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// Gauge для idle подключений к базе данных
	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// Счетчик ошибок базы данных
	DBErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors",
		},
	)

	// Template metrics
	// Счетчик применений шаблонов
	TemplatesApplied = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "templates_applied_total",
			Help: "Total number of template applications",
		},
	)

	// Счетчик откатов шаблонов
	TemplatesRolledBack = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "templates_rolled_back_total",
			Help: "Total number of template rollbacks",
		},
	)

	// Telegram metrics
	// Счетчик отправленных Telegram сообщений
	TelegramMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_messages_sent_total",
			Help: "Total number of Telegram messages sent",
		},
		[]string{"type"}, // "notification", "broadcast"
	)

	// Счетчик ошибок Telegram
	TelegramErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "telegram_errors_total",
			Help: "Total number of Telegram API errors",
		},
	)
)
