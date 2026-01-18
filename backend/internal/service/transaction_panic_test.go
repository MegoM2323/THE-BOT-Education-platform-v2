//go:build test
// +build test

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"tutoring-platform/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionRollbackOnPanic проверяет что транзакции откатываются при panic
func TestTransactionRollbackOnPanic(t *testing.T) {
	pool := database.GetTestPool(t)

	ctx := context.Background()

	t.Run("Panic during transaction is caught and rollback happens", func(t *testing.T) {
		uniqueEmail := fmt.Sprintf("panic_test_%s@example.com", uuid.New().String()[:8])

		err := performTransactionWithPanic(ctx, pool, uniqueEmail)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovered")

		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", uniqueEmail).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Transaction should have been rolled back - no data in DB")
	})

	t.Run("Normal error during transaction triggers rollback", func(t *testing.T) {
		uniqueEmail := fmt.Sprintf("error_test_%s@example.com", uuid.New().String()[:8])

		err := performTransactionWithError(ctx, pool, uniqueEmail)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated error")

		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", uniqueEmail).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "Transaction should have been rolled back on error")
	})

	t.Run("Successful transaction commits properly", func(t *testing.T) {
		uniqueEmail := fmt.Sprintf("success_test_%s@example.com", uuid.New().String()[:8])

		err := performSuccessfulTransaction(ctx, pool, uniqueEmail)
		require.NoError(t, err)

		var count int
		err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", uniqueEmail).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Transaction should have been committed")
	})
}

// performTransactionWithPanic симулирует транзакцию с panic внутри
func performTransactionWithPanic(ctx context.Context, pool *pgxpool.Pool, email string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("panic recovered: transaction should be rolled back")
		}
	}()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'hash', 'Panic Test', 'student', NOW(), NOW())
	`, email)
	if err != nil {
		return err
	}

	panic("simulated panic during transaction")
}

// performTransactionWithError симулирует транзакцию с ошибкой (не panic)
func performTransactionWithError(ctx context.Context, pool *pgxpool.Pool, email string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'hash', 'Error Test', 'student', NOW(), NOW())
	`, email)
	if err != nil {
		return err
	}

	return errors.New("simulated error during transaction")
}

// performSuccessfulTransaction симулирует успешную транзакцию
func performSuccessfulTransaction(ctx context.Context, pool *pgxpool.Pool, email string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'hash', 'Success Test', 'student', NOW(), NOW())
	`, email)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
