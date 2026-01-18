package handlers

import (
	"os"
	"testing"

	"tutoring-platform/internal/database"
)

func TestMain(m *testing.M) {
	pool := database.GetTestPoolForMain()
	if pool != nil {
		database.CleanupTestTablesForMain(pool)
	}

	code := m.Run()
	os.Exit(code)
}
