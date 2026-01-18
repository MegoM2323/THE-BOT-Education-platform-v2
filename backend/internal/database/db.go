package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL драйвер
	"github.com/jmoiron/sqlx"
	"log"
	"time"
	"tutoring-platform/internal/config"
)

// DB оборачивает пулы соединений с базой данных
type DB struct {
	Pool  *pgxpool.Pool // Для нативных операций pgx
	Sqlx  *sqlx.DB      // Для операций sqlx
	Close func() error  // Функция закрытия всех соединений
}

// New создает новое подключение к базе данных
func New(cfg *config.DatabaseConfig) (*DB, error) {
	ctx := context.Background()

	// Создаем конфигурацию пула pgx
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Настраиваем параметры пула
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Создаем пул соединений pgx
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Тестируем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// Создаем sqlx соединение для удобных методов
	sqlxDB, err := sqlx.Connect("pgx", cfg.GetDSN())
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to create sqlx connection: %w", err)
	}

	// Настраиваем параметры пула sqlx
	sqlxDB.SetMaxOpenConns(25)
	sqlxDB.SetMaxIdleConns(5)
	sqlxDB.SetConnMaxLifetime(time.Hour)
	sqlxDB.SetConnMaxIdleTime(30 * time.Minute)

	log.Println("log")

	return &DB{
		Pool: pool,
		Sqlx: sqlxDB,
		Close: func() error {
			pool.Close()
			return sqlxDB.Close()
		},
	}, nil
}

// HealthCheck проверяет работоспособность базы данных
func (db *DB) HealthCheck(ctx context.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("pgx pool health check failed: %w", err)
	}
	if err := db.Sqlx.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlx health check failed: %w", err)
	}
	return nil
}

// Stats возвращает статистику пула соединений
func (db *DB) Stats() map[string]interface{} {
	stats := db.Pool.Stat()
	return map[string]interface{}{
		"acquired_conns":   stats.AcquiredConns(),
		"idle_conns":       stats.IdleConns(),
		"total_conns":      stats.TotalConns(),
		"max_conns":        stats.MaxConns(),
		"acquire_count":    stats.AcquireCount(),
		"acquire_duration": stats.AcquireDuration(),
		"empty_acquire":    stats.EmptyAcquireCount(),
		"canceled_acquire": stats.CanceledAcquireCount(),
	}
}

// BeginTx начинает новую транзакцию с указанным уровнем изоляции
func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (pgx.Tx, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Устанавливаем уровень изоляции, если указан
	if opts != nil && opts.IsolationLevel != "" {
		_, err = tx.Exec(ctx, fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", opts.IsolationLevel))
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("failed to set isolation level: %w", err)
		}
	}

	return tx, nil
}

// TxOptions представляет опции транзакции
type TxOptions struct {
	IsolationLevel string // "READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"
}

// CleanupExpiredSessions удаляет просроченные сессии из базы данных
func (db *DB) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`

	result, err := db.Pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsDeleted := result.RowsAffected()
	if rowsDeleted > 0 {
		log.Println("log")
	}

	return rowsDeleted, nil
}

// ToSqlxPool converts a pgxpool.Pool to an sqlx.DB by creating a new connection
// This is used in tests when repositories need sqlx.DB but we have pgxpool.Pool
func ToSqlxPool(pool *pgxpool.Pool) *sqlx.DB {
	// Get connection string from pool and create sqlx.DB
	// This is a test-only utility - in production, use the DB.Sqlx field
	config := pool.Config()
	connString := config.ConnString()

	db := sqlx.MustConnect("pgx", connString)
	return db
}
