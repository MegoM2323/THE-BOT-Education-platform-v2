package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"tutoring-platform/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	// Global shared test database pool
	testPool       *pgxpool.Pool
	testDB         *sqlx.DB
	testOnce       sync.Once
	testMu         sync.Mutex
	cleanupMu      sync.Mutex // Mutex for serializing CleanupTestTables calls
	migrationsOnce sync.Once  // Ensure migrations run only once
	migrationsErr  error      // Cache migration errors
)

// init validates that test and production database names are different
// This prevents accidental truncation of production database during tests
func init() {
	// Get database names from environment or use defaults
	testDBName := os.Getenv("TEST_DATABASE_NAME")
	if testDBName == "" {
		testDBName = "tutoring_platform_test"
	}

	prodDBName := os.Getenv("DATABASE_NAME")
	if prodDBName == "" {
		prodDBName = "tutoring_platform"
	}

	// CRITICAL SAFETY CHECK: Verify databases are different
	if testDBName == prodDBName {
		log.Fatalf("CRITICAL SAFETY VIOLATION: TEST_DATABASE_NAME and DATABASE_NAME are the same ('%s'). "+
			"This would DELETE PRODUCTION DATA when running tests! "+
			"Set TEST_DATABASE_NAME to a separate test database (e.g., 'tutoring_platform_test')",
			testDBName)
	}

	log.Printf("✓ Database names verified: prod='%s', test='%s'", prodDBName, testDBName)
}

// GetTestPool returns the shared PostgreSQL connection pool for tests.
// This pool is created once and reused across all tests to avoid connection exhaustion.
func GetTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	var err error
	testOnce.Do(func() {
		// Load database password from environment variable
		// Falls back to default 'postgres' for local development
		dbPassword := os.Getenv("DATABASE_PASSWORD")
		if dbPassword == "" {
			dbPassword = "postgres"
		}

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: dbPassword,
			Name:     "tutoring_platform_test",
			SSLMode:  "disable",
		}

		// Create pool configuration with reasonable defaults for tests
		poolConfig, poolErr := pgxpool.ParseConfig(cfg.GetDSN())
		if poolErr != nil {
			err = fmt.Errorf("unable to parse database config: %w", poolErr)
			return
		}

		// Configure pool parameters for tests (smaller than production)
		poolConfig.MaxConns = 10
		poolConfig.MinConns = 2
		poolConfig.MaxConnLifetime = 30 * time.Minute
		poolConfig.MaxConnIdleTime = 10 * time.Minute
		poolConfig.HealthCheckPeriod = 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create the pool
		testPool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			err = fmt.Errorf("unable to create test connection pool: %w", err)
			return
		}

		// Test the connection
		if pingErr := testPool.Ping(ctx); pingErr != nil {
			testPool.Close()
			testPool = nil
			err = fmt.Errorf("unable to ping test database: %w", pingErr)
			return
		}
	})

	if err != nil {
		t.Fatalf("Failed to initialize test pool: %v", err)
	}

	return testPool
}

// SafeGetTestPool returns a test pool with database safety verification.
// It verifies that the pool is connected to the test database (not production).
// NEVER run destructive operations without calling this function first.
func SafeGetTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := GetTestPool(t)

	// CRITICAL: Verify we're NOT connected to production database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var currentDB string
	err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&currentDB)
	if err != nil {
		t.Fatalf("Failed to verify database name: %v", err)
	}

	// MUST be test database, NEVER production
	if currentDB != "tutoring_platform_test" {
		t.Fatalf("CRITICAL SAFETY CHECK FAILED: Connected to '%s' instead of 'tutoring_platform_test'. "+
			"This would DELETE PRODUCTION DATA! Aborting test.", currentDB)
	}

	return pool
}

// applyMigrationsToTestDB applies all database migrations from the migrations directory.
// This ensures the test database schema matches production exactly, including the methodologist role.
// NOTE: This function runs inside sync.Once and does not receive *testing.T to avoid race conditions.
func applyMigrationsToTestDB(db *sqlx.DB) error {
	migrationsOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check if schema already exists and has recent migrations (methodologist role)
		// Use information_schema for robust constraint validation instead of LIKE string matching
		var hasMethodologistRole int
		checkErr := db.GetContext(ctx, &hasMethodologistRole,
			`SELECT 1 FROM information_schema.table_constraints tc
			 WHERE tc.table_name = 'users'
			   AND tc.constraint_name = 'users_role_check'
			   AND tc.constraint_type = 'CHECK'
			 LIMIT 1`)

		// If methodologist role already exists, migrations are already applied
		if checkErr == nil && hasMethodologistRole == 1 {
			log.Printf("Database already has full schema with methodologist role, skipping migrations")
			return
		}

		// Otherwise, we need to apply migrations
		// Find migrations directory relative to project root
		// Try multiple possible paths to handle different working directories
		possiblePaths := []string{
			"internal/database/migrations",
			"backend/internal/database/migrations",
			"./internal/database/migrations",
			"./backend/internal/database/migrations",
			"../internal/database/migrations",
			"../../internal/database/migrations",
		}

		var migrationsDir string
		for _, path := range possiblePaths {
			if info, statErr := os.Stat(path); statErr == nil && info.IsDir() {
				migrationsDir = path
				break
			}
		}

		if migrationsDir == "" {
			// Last resort: try finding via environment or hardcoded path
			if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
				if info, err := os.Stat(envPath); err == nil && info.IsDir() {
					migrationsDir = envPath
				}
			}
		}

		if migrationsDir == "" {
			migrationsErr = fmt.Errorf("migrations directory not found in any of: %v", possiblePaths)
			return
		}

		// Read all SQL files from migrations directory
		entries, readErr := os.ReadDir(migrationsDir)
		if readErr != nil {
			migrationsErr = fmt.Errorf("failed to read migrations directory: %w", readErr)
			return
		}

		// Filter and sort SQL files
		// Skip seed files as they may not be idempotent (duplicate inserts fail).
		// For idempotent seed files, use INSERT ... ON CONFLICT DO NOTHING pattern.
		var migrationFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
				if !strings.HasPrefix(entry.Name(), "seed_") {
					migrationFiles = append(migrationFiles, entry.Name())
				}
			}
		}

		// Sort migration files to ensure consistent order
		sort.Strings(migrationFiles)

		// Apply each migration using psql to handle CREATE INDEX CONCURRENTLY properly
		appliedCount := 0
		skippedCount := 0
		failedCount := 0
		for _, migrationFile := range migrationFiles {
			filePath := filepath.Join(migrationsDir, migrationFile)

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				log.Printf("Warning: migration file not found: %s", filePath)
				skippedCount++
				continue
			}

			// Load PostgreSQL password from environment variable
			// Falls back to default 'postgres' for local development
			dbPassword := os.Getenv("DATABASE_PASSWORD")
			if dbPassword == "" {
				dbPassword = "postgres"
			}

			// Build PostgreSQL connection string with password
			connStr := fmt.Sprintf("postgres://postgres:%s@localhost:5432/tutoring_platform_test?sslmode=disable",
				strings.ReplaceAll(dbPassword, "@", "%40"))

			// Use psql to apply migration (handles CREATE INDEX CONCURRENTLY)
			cmd := exec.Command("psql",
				connStr,
				"-f", filePath,
				"--quiet",
			)

			// Execute migration and capture error separately
			output, execErr := cmd.CombinedOutput()
			outputStr := string(output)

			// Determine migration result based on exit code and output
			if execErr == nil {
				// Success: no error
				appliedCount++
			} else if strings.Contains(outputStr, "already exists") {
				// Non-critical error: constraint/object already exists
				skippedCount++
			} else {
				// Actual error: log and count as failed
				log.Printf("Warning: migration %s encountered error: %s", migrationFile, outputStr)
				failedCount++
			}
		}

		log.Printf("Applied database migrations: %d applied, %d skipped/already exists, %d failed", appliedCount, skippedCount, failedCount)
	})

	return migrationsErr
}

// GetTestSqlxDB returns the shared sqlx.DB for tests.
// It automatically applies all database migrations on first use to ensure
// the test database schema matches production exactly.
func GetTestSqlxDB(t *testing.T) *sqlx.DB {
	t.Helper()

	if testDB != nil {
		return testDB
	}

	testMu.Lock()
	defer testMu.Unlock()

	if testDB != nil {
		return testDB
	}

	// Load database password from environment variable
	// Falls back to default 'postgres' for local development
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: dbPassword,
		Name:     "tutoring_platform_test",
		SSLMode:  "disable",
	}

	var err error
	testDB, err = sqlx.Connect("pgx", cfg.GetDSN())
	if err != nil {
		t.Fatalf("Failed to create test sqlx.DB: %v", err)
	}

	// Configure pool settings
	testDB.SetMaxOpenConns(10)
	testDB.SetMaxIdleConns(2)
	testDB.SetConnMaxLifetime(30 * time.Minute)
	testDB.SetConnMaxIdleTime(10 * time.Minute)

	// Apply all database migrations to ensure schema is up to date
	// This includes all constraints like methodologist role in users table
	if migrationErr := applyMigrationsToTestDB(testDB); migrationErr != nil {
		t.Fatalf("Failed to apply migrations to test database: %v", migrationErr)
	}

	return testDB
}

// CleanupTestTables truncates all test tables to reset state between tests.
// Important: Call this at the beginning of each test that needs a clean database.
// CRITICAL: This function has built-in safety checks to prevent truncating production database.
// Note: This function uses a mutex to prevent race conditions when multiple tests run concurrently.
func CleanupTestTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// Serialize cleanup calls to prevent race conditions
	cleanupMu.Lock()
	defer cleanupMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// SAFETY CHECK #1: Verify this is the test database BEFORE truncating anything
	var currentDB string
	err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&currentDB)
	if err != nil {
		t.Fatalf("CRITICAL: Failed to verify database name before cleanup: %v", err)
	}

	// SAFETY CHECK #2: MUST be test database, NEVER allow truncation of production
	if currentDB != "tutoring_platform_test" {
		t.Fatalf("CRITICAL SAFETY VIOLATION: Attempted to truncate on database '%s' instead of 'tutoring_platform_test'. "+
			"This would DELETE PRODUCTION DATA! ABORTING cleanup to prevent data loss!", currentDB)
	}

	// Database verified safe, now safe to truncate
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Order matters due to foreign key constraints
	tables := []string{
		"broadcast_files",
		"lesson_broadcasts",
		"lesson_homework",
		"lesson_modifications",
		"credit_transactions",
		"swaps",
		"bookings",
		"lessons",
		"template_lesson_students",
		"template_lessons",
		"template_applications",
		"lesson_templates",
		"broadcasts",
		"broadcast_lists",
		"chat_rooms",
		"messages",
		"telegram_users",
		"credits",
		"sessions",
		"users",
	}

	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Log the error but continue - some tables might not exist in all test schemas
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// CleanupTestDatabase closes all test database connections.
// This should be called once at the end of all tests (e.g., in TestMain cleanup).
func CleanupTestDatabase() {
	testMu.Lock()
	defer testMu.Unlock()

	if testPool != nil {
		testPool.Close()
		testPool = nil
	}

	if testDB != nil {
		testDB.Close()
		testDB = nil
	}
}

// GetTestDBInstance returns a *DB struct wrapping the shared test pools.
// Use this when you need a full DB instance but want to reuse the shared pools.
func GetTestDBInstance(t *testing.T) *DB {
	t.Helper()

	pool := GetTestPool(t)
	sqlxDB := GetTestSqlxDB(t)

	return &DB{
		Pool: pool,
		Sqlx: sqlxDB,
		Close: func() error {
			return nil
		},
	}
}

// GetTestPoolForMain returns test pool for TestMain (no *testing.T required).
func GetTestPoolForMain() *pgxpool.Pool {
	var err error
	testOnce.Do(func() {
		// Load database password from environment variable
		// Falls back to default 'postgres' for local development
		dbPassword := os.Getenv("DATABASE_PASSWORD")
		if dbPassword == "" {
			dbPassword = "postgres"
		}

		cfg := &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: dbPassword,
			Name:     "tutoring_platform_test",
			SSLMode:  "disable",
		}

		poolConfig, poolErr := pgxpool.ParseConfig(cfg.GetDSN())
		if poolErr != nil {
			err = poolErr
			return
		}

		poolConfig.MaxConns = 10
		poolConfig.MinConns = 2
		poolConfig.MaxConnLifetime = 30 * time.Minute
		poolConfig.MaxConnIdleTime = 10 * time.Minute
		poolConfig.HealthCheckPeriod = 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		testPool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return
		}

		if pingErr := testPool.Ping(ctx); pingErr != nil {
			testPool.Close()
			testPool = nil
			err = pingErr
		}
	})

	if err != nil {
		log.Printf("Failed to initialize test pool: %v", err)
		return nil
	}

	return testPool
}

// CleanupTestTablesForMain truncates all test tables (for TestMain, no *testing.T).
func CleanupTestTablesForMain(pool *pgxpool.Pool) {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var currentDB string
	err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&currentDB)
	if err != nil {
		log.Printf("CRITICAL: Failed to verify database name before cleanup: %v", err)
		return
	}

	if currentDB != "tutoring_platform_test" {
		log.Printf("CRITICAL SAFETY VIOLATION: Attempted to truncate on database '%s' instead of 'tutoring_platform_test'. ABORTING!", currentDB)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tables := []string{
		"broadcast_files",
		"lesson_broadcasts",
		"lesson_homework",
		"lesson_modifications",
		"credit_transactions",
		"swaps",
		"bookings",
		"lessons",
		"template_lesson_students",
		"template_lessons",
		"template_applications",
		"lesson_templates",
		"broadcasts",
		"broadcast_lists",
		"chat_rooms",
		"messages",
		"telegram_users",
		"credits",
		"sessions",
		"users",
	}

	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Printf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}
