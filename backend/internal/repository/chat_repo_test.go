package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Setup Helpers ====================

// setupChatTestDB returns the shared test database pool.
// Uses database.GetTestPool to avoid connection exhaustion.
// Note: НЕ очищает таблицы здесь, т.к. это вызывает race conditions при параллельном запуске тестов.
func setupChatTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := database.GetTestPool(t)

	return pool
}

// cleanupChatTestDB no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные и не конфликтуют.
func cleanupChatTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
}

// createChatTestUser creates a test user in the database.
// Uses unique email based on UUID to prevent conflicts in parallel tests.
func createChatTestUser(t *testing.T, pool *pgxpool.Pool, role models.UserRole, email, fullName string) *models.User {
	t.Helper()

	ctx := context.Background()

	userID := uuid.New()
	uniqueEmail := fmt.Sprintf("%s-%s", userID.String()[:8], email)

	// Split fullName into first and last name
	firstName := fullName
	lastName := ""
	if idx := len(fullName) / 2; idx > 0 {
		parts := []rune(fullName)
		for i, r := range parts {
			if r == ' ' {
				firstName = string(parts[:i])
				lastName = string(parts[i+1:])
				break
			}
		}
	}

	user := &models.User{
		ID:           userID,
		Email:        uniqueEmail,
		PasswordHash: "test-hash",
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	require.NoError(t, err, "Failed to create test user")

	return user
}

// createChatTestRoom creates a test chat room
func createChatTestRoom(t *testing.T, pool *pgxpool.Pool, teacherID, studentID uuid.UUID) *models.ChatRoom {
	t.Helper()

	ctx := context.Background()

	roomID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO chat_rooms (id, teacher_id, student_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := pool.Exec(ctx, query, roomID, teacherID, studentID, now, now)
	require.NoError(t, err, "Failed to create test chat room")

	return &models.ChatRoom{
		ID:        roomID,
		TeacherID: teacherID,
		StudentID: studentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// createSqlxFromPool returns the shared sqlx.DB for testing.
// Uses database.GetTestSqlxDB to avoid connection exhaustion.
func createSqlxFromPool(t *testing.T, pool *pgxpool.Pool) *sqlx.DB {
	t.Helper()

	return database.GetTestSqlxDB(t)
}

// ==================== ListRoomsByTeacher Tests ====================

// TestListRoomsByTeacher_WithParticipantInfo проверяет что репозиторий возвращает информацию о студентах
func TestListRoomsByTeacher_WithParticipantInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create test users
	teacher := createChatTestUser(t, pool, models.RoleTeacher, "teacher@test.com", "Иван Преподаватель")
	student1 := createChatTestUser(t, pool, models.RoleStudent, "student1@test.com", "Петр Студент")
	student2 := createChatTestUser(t, pool, models.RoleStudent, "student2@test.com", "Мария Студентова")

	// Create chat rooms
	room1 := createChatTestRoom(t, pool, teacher.ID, student1.ID)
	room2 := createChatTestRoom(t, pool, teacher.ID, student2.ID)

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Get rooms by teacher
	rooms, err := repo.ListRoomsByTeacher(ctx, teacher.ID)
	require.NoError(t, err)
	require.NotNil(t, rooms)
	require.Len(t, rooms, 2, "Should have 2 rooms")

	// Verify participant info
	found := false
	for _, room := range rooms {
		if room.ID == room1.ID {
			assert.Equal(t, student1.ID, room.ParticipantID, "Should see student1 as participant")
			assert.Equal(t, student1.GetFullName(), room.ParticipantName, "Should have student1 name")
			assert.Equal(t, string(models.RoleStudent), room.ParticipantRole, "Should have student role")
			found = true
			break
		}
	}
	assert.True(t, found, "Room 1 should be in the list with correct participant info")

	// Verify second room
	found = false
	for _, room := range rooms {
		if room.ID == room2.ID {
			assert.Equal(t, student2.ID, room.ParticipantID, "Should see student2 as participant")
			assert.Equal(t, student2.GetFullName(), room.ParticipantName, "Should have student2 name")
			assert.Equal(t, string(models.RoleStudent), room.ParticipantRole, "Should have student role")
			found = true
			break
		}
	}
	assert.True(t, found, "Room 2 should be in the list with correct participant info")

	t.Log("✓ Teacher sees all students with correct participant info")
}

// TestListRoomsByTeacher_Empty проверяет пустой результат для преподавателя без чатов
func TestListRoomsByTeacher_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create teacher with no rooms
	teacher := createChatTestUser(t, pool, models.RoleTeacher, "teacher@test.com", "Иван Преподаватель")

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Get rooms for teacher with no chats
	rooms, err := repo.ListRoomsByTeacher(ctx, teacher.ID)
	require.NoError(t, err)
	assert.NotNil(t, rooms, "Should return empty slice, not nil")
	assert.Len(t, rooms, 0, "Should have 0 rooms")

	t.Log("✓ Teacher with no chats returns empty list")
}

// TestListRoomsByTeacher_OrderByLastMessage проверяет сортировку по последнему сообщению
func TestListRoomsByTeacher_OrderByLastMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create test users
	teacher := createChatTestUser(t, pool, models.RoleTeacher, "teacher@test.com", "Иван Преподаватель")
	student1 := createChatTestUser(t, pool, models.RoleStudent, "student1@test.com", "Петр Студент")
	student2 := createChatTestUser(t, pool, models.RoleStudent, "student2@test.com", "Мария Студентова")

	// Create chat rooms
	room1 := createChatTestRoom(t, pool, teacher.ID, student1.ID)
	room2 := createChatTestRoom(t, pool, teacher.ID, student2.ID)

	// Update last_message_at for room1 (should be first)
	now := time.Now()
	oldTime := now.Add(-1 * time.Hour)

	_, err := pool.Exec(ctx,
		"UPDATE chat_rooms SET last_message_at = $1 WHERE id = $2",
		now, room1.ID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		"UPDATE chat_rooms SET last_message_at = $1 WHERE id = $2",
		oldTime, room2.ID)
	require.NoError(t, err)

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Rooms should be ordered by last_message_at DESC
	rooms, err := repo.ListRoomsByTeacher(ctx, teacher.ID)
	require.NoError(t, err)
	require.Len(t, rooms, 2)

	// First room should have the newer last_message_at
	assert.Equal(t, room1.ID, rooms[0].ID, "Room with newer message should be first")
	assert.Equal(t, room2.ID, rooms[1].ID, "Room with older message should be second")

	t.Log("✓ Rooms ordered by last_message_at DESC")
}

// ==================== ListRoomsByStudent Tests ====================

// TestListRoomsByStudent_WithParticipantInfo проверяет что студент видит информацию о преподавателях
func TestListRoomsByStudent_WithParticipantInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create test users
	teacher1 := createChatTestUser(t, pool, models.RoleTeacher, "teacher1@test.com", "Иван Преподаватель")
	teacher2 := createChatTestUser(t, pool, models.RoleTeacher, "teacher2@test.com", "Сергей Учитель")
	student := createChatTestUser(t, pool, models.RoleStudent, "student@test.com", "Петр Студент")

	// Create chat rooms
	room1 := createChatTestRoom(t, pool, teacher1.ID, student.ID)
	room2 := createChatTestRoom(t, pool, teacher2.ID, student.ID)

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Get rooms by student
	rooms, err := repo.ListRoomsByStudent(ctx, student.ID)
	require.NoError(t, err)
	require.NotNil(t, rooms)
	require.Len(t, rooms, 2, "Should have 2 rooms")

	// Verify participant info for first teacher
	found := false
	for _, room := range rooms {
		if room.ID == room1.ID {
			assert.Equal(t, teacher1.ID, room.ParticipantID, "Should see teacher1 as participant")
			assert.Equal(t, teacher1.GetFullName(), room.ParticipantName, "Should have teacher1 name")
			assert.Equal(t, string(models.RoleTeacher), room.ParticipantRole, "Should have teacher role")
			found = true
			break
		}
	}
	assert.True(t, found, "Room 1 should be in the list with correct participant info")

	// Verify participant info for second teacher
	found = false
	for _, room := range rooms {
		if room.ID == room2.ID {
			assert.Equal(t, teacher2.ID, room.ParticipantID, "Should see teacher2 as participant")
			assert.Equal(t, teacher2.GetFullName(), room.ParticipantName, "Should have teacher2 name")
			assert.Equal(t, string(models.RoleTeacher), room.ParticipantRole, "Should have teacher role")
			found = true
			break
		}
	}
	assert.True(t, found, "Room 2 should be in the list with correct participant info")

	t.Log("✓ Student sees all teachers with correct participant info")
}

// TestListRoomsByStudent_Empty проверяет пустой результат для студента без чатов
func TestListRoomsByStudent_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create student with no rooms
	student := createChatTestUser(t, pool, models.RoleStudent, "student@test.com", "Петр Студент")

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Get rooms for student with no chats
	rooms, err := repo.ListRoomsByStudent(ctx, student.ID)
	require.NoError(t, err)
	assert.NotNil(t, rooms, "Should return empty slice, not nil")
	assert.Len(t, rooms, 0, "Should have 0 rooms")

	t.Log("✓ Student with no chats returns empty list")
}

// TestListRoomsByStudent_OnlyOwnRooms проверяет что студент видит только свои чаты
func TestListRoomsByStudent_OnlyOwnRooms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create test users
	teacher := createChatTestUser(t, pool, models.RoleTeacher, "teacher@test.com", "Иван Преподаватель")
	student1 := createChatTestUser(t, pool, models.RoleStudent, "student1@test.com", "Петр Студент")
	student2 := createChatTestUser(t, pool, models.RoleStudent, "student2@test.com", "Мария Студентова")

	// Create rooms: teacher-student1 and teacher-student2
	_ = createChatTestRoom(t, pool, teacher.ID, student1.ID)
	_ = createChatTestRoom(t, pool, teacher.ID, student2.ID)

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: student1 should only see their own room
	rooms, err := repo.ListRoomsByStudent(ctx, student1.ID)
	require.NoError(t, err)
	require.Len(t, rooms, 1, "Student1 should see only 1 room")

	assert.Equal(t, student1.ID, rooms[0].StudentID, "Room should be for student1")
	assert.Equal(t, teacher.ID, rooms[0].ParticipantID, "Participant should be teacher")

	t.Log("✓ Student sees only their own chat rooms")
}

// ==================== JoinQuery Tests ====================

// TestChatRepository_JoinQueryCorrectness проверяет что JOIN правильно связывает таблицы
func TestChatRepository_JoinQueryCorrectness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create test users
	teacher := createChatTestUser(t, pool, models.RoleTeacher, "teacher@test.com", "Иван Преподаватель")
	student := createChatTestUser(t, pool, models.RoleStudent, "student@test.com", "Петр Студент")

	// Create room
	room := createChatTestRoom(t, pool, teacher.ID, student.ID)

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Verify JOIN selects correct user columns
	rooms, err := repo.ListRoomsByTeacher(ctx, teacher.ID)
	require.NoError(t, err)
	require.Len(t, rooms, 1)

	returnedRoom := rooms[0]

	// Verify room fields
	assert.Equal(t, room.ID, returnedRoom.ID)
	assert.Equal(t, teacher.ID, returnedRoom.TeacherID)
	assert.Equal(t, student.ID, returnedRoom.StudentID)

	// Verify JOIN returned correct participant (student)
	assert.Equal(t, student.ID, returnedRoom.ParticipantID)
	assert.NotEqual(t, teacher.ID, returnedRoom.ParticipantID)

	// Verify user data is populated
	assert.NotEmpty(t, returnedRoom.ParticipantName)
	assert.Equal(t, student.GetFullName(), returnedRoom.ParticipantName)

	t.Log("✓ JOIN query correctly links chat_rooms and users tables")
}

// TestChatRepository_MultipleRoomsWithSameTeacher проверяет что JOINs работает для нескольких строк
func TestChatRepository_MultipleRoomsWithSameTeacher(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupChatTestDB(t)
	defer cleanupChatTestDB(t, pool)

	ctx := context.Background()

	// Create test users
	teacher := createChatTestUser(t, pool, models.RoleTeacher, "teacher@test.com", "Иван Преподаватель")
	students := make([]*models.User, 3)
	for i := 0; i < 3; i++ {
		students[i] = createChatTestUser(t, pool, models.RoleStudent, "student"+string(rune('1'+i))+"@test.com", "Студент "+string(rune('A'+i)))
	}

	// Create multiple rooms
	for i := 0; i < 3; i++ {
		_ = createChatTestRoom(t, pool, teacher.ID, students[i].ID)
	}

	// Create repository
	db := createSqlxFromPool(t, pool)
	repo := NewChatRepository(db)

	// Test: Get all rooms for teacher
	rooms, err := repo.ListRoomsByTeacher(ctx, teacher.ID)
	require.NoError(t, err)
	require.Len(t, rooms, 3, "Should have 3 rooms")

	// Verify each room has correct participant info
	for _, room := range rooms {
		// ParticipantID should be one of the students
		found := false
		for i := 0; i < 3; i++ {
			if room.ParticipantID == students[i].ID {
				assert.Equal(t, students[i].GetFullName(), room.ParticipantName)
				assert.Equal(t, string(models.RoleStudent), room.ParticipantRole)
				found = true
				break
			}
		}
		assert.True(t, found, "Room should have one of the students as participant")
	}

	t.Log("✓ Multiple rooms with JOIN queries work correctly")
}
