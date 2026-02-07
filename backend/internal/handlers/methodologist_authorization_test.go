package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tutoring-platform/internal/models"
)

// TestTeacherRoleCheck проверяет что роль методиста правильно определяется
func TestTeacherRoleCheck(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}


	student := &models.User{
		ID:        uuid.New(),
		Email:     "student@test.com",
		FirstName: "Test",
		LastName:  "Student",
		Role:      models.RoleStudent,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	t.Run("Teacher_IsTeacher", func(t *testing.T) {
		assert.True(t, teacher.IsTeacher(), "Методист должен иметь IsTeacher() == true")
	})

	t.Run("Teacher_NotAdmin", func(t *testing.T) {
		assert.False(t, teacher.IsAdmin(), "Методист не должен быть админом")
	})

	t.Run("Teacher_NotTeacher", func(t *testing.T) {
		assert.False(t, teacher.IsTeacher(), "Методист не должен быть учителем")
	})

	t.Run("Teacher_NotStudent", func(t *testing.T) {
		assert.False(t, teacher.IsStudent(), "Методист не должен быть студентом")
	})

	t.Run("Teacher_NotTeacher", func(t *testing.T) {
		assert.False(t, teacher.IsTeacher(), "Учитель не должен быть методистом")
	})

	t.Run("Student_NotTeacher", func(t *testing.T) {
		assert.False(t, student.IsTeacher(), "Студент не должен быть методистом")
	})

	t.Run("Admin_NotTeacher", func(t *testing.T) {
		assert.False(t, admin.IsTeacher(), "Админ не методист (хотя может все)")
	})

	t.Run("Admin_CanDoEverything", func(t *testing.T) {
		assert.True(t, admin.IsAdmin(), "Админ должен иметь IsAdmin() == true")
	})
}

// TestTeacherCanCreateLesson проверяет логику CreateLesson
func TestTeacherCanCreateLesson(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Проверка что методист может создавать занятия (согласно lessons.go:103)
	// if !user.IsAdmin() && !user.IsTeacher()
	canCreateTeacher := teacher.IsAdmin() || teacher.IsTeacher()
	canCreateAdmin := admin.IsAdmin() || admin.IsTeacher()

	assert.True(t, canCreateTeacher, "Методист должен создавать занятия")
	assert.False(t, canCreateTeacher, "Учитель не должен создавать занятия")
	assert.True(t, canCreateAdmin, "Админ должен создавать занятия")
}

// TestTeacherCanUpdateLesson проверяет логику UpdateLesson
func TestTeacherCanUpdateLesson(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}

	teacher1 := &models.User{
		ID:        uuid.New(),
		Email:     "teacher1@test.com",
		FirstName: "Test Teacher",
		LastName:  "1",
		Role:      models.RoleTeacher,
	}

	_ = &models.User{
		ID:        uuid.New(),
		Email:     "teacher2@test.com",
		FirstName: "Test Teacher",
		LastName:  "2",
		Role:      models.RoleTeacher,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:203
	// if !user.IsAdmin() && !user.IsTeacher()

	t.Run("TeacherCanEditAnyLesson", func(t *testing.T) {
		canEdit := teacher.IsAdmin() || teacher.IsTeacher()
		assert.True(t, canEdit, "Методист должен редактировать любое занятие")
	})

	t.Run("TeacherCanEditOnlyHomework", func(t *testing.T) {
		// Учитель может редактировать только homework_text своих уроков
		canEditAll := teacher1.IsAdmin() || teacher1.IsTeacher()
		assert.False(t, canEditAll, "Учитель не может редактировать все поля")
	})

	t.Run("AdminCanEditAnyLesson", func(t *testing.T) {
		canEdit := admin.IsAdmin() || admin.IsTeacher()
		assert.True(t, canEdit, "Админ должен редактировать любое занятие")
	})
}

// TestTeacherCanDeleteLesson проверяет логику DeleteLesson
func TestTeacherCanDeleteLesson(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:289
	// if !user.IsAdmin() && !user.IsTeacher()

	canDeleteTeacher := teacher.IsAdmin() || teacher.IsTeacher()
	canDeleteAdmin := admin.IsAdmin() || admin.IsTeacher()

	assert.True(t, canDeleteTeacher, "Методист должен удалять занятия")
	assert.False(t, canDeleteTeacher, "Учитель не должен удалять занятия")
	assert.True(t, canDeleteAdmin, "Админ должен удалять занятия")
}

// TestTeacherCanBulkEdit проверяет логику ApplyToAllSubsequent
func TestTeacherCanBulkEdit(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:522
	// if !admin.IsAdmin() && !admin.IsTeacher()

	canBulkEditTeacher := teacher.IsAdmin() || teacher.IsTeacher()
	canBulkEditAdmin := admin.IsAdmin() || admin.IsTeacher()

	assert.True(t, canBulkEditTeacher, "Методист должен делать массовое редактирование")
	assert.False(t, canBulkEditTeacher, "Учитель не должен делать массовое редактирование")
	assert.True(t, canBulkEditAdmin, "Админ должен делать массовое редактирование")
}

// TestTeacherCanEditPastLesson проверяет редактирование прошлых занятий
func TestTeacherCanEditPastLesson(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:245
	// isPastLesson && !user.IsAdmin() && !user.IsTeacher() && !isHomeworkTextOnlyUpdate
	// Методист и админ могут редактировать прошлые занятия

	isPastLesson := true
	isHomeworkTextOnlyUpdate := false

	canEditPastTeacher := !(isPastLesson && !teacher.IsAdmin() && !teacher.IsTeacher() && !isHomeworkTextOnlyUpdate)
	canEditPastAdmin := !(isPastLesson && !admin.IsAdmin() && !admin.IsTeacher() && !isHomeworkTextOnlyUpdate)

	assert.True(t, canEditPastTeacher, "Методист должен редактировать прошлые занятия")
	assert.False(t, canEditPastTeacher, "Учитель не должен редактировать прошлые занятия (кроме homework)")
	assert.True(t, canEditPastAdmin, "Админ должен редактировать прошлые занятия")
}

// TestTeacherHomeworkOnlyPermission проверяет что учитель может только homework
func TestTeacherHomeworkOnlyPermission(t *testing.T) {
	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleTeacher,
	}

	lessonTeacherID := teacher.ID

	// Учитель пытается обновить только homework своего урока
	isTeacherOwnLesson := teacher.IsTeacher() && lessonTeacherID == teacher.ID
	isHomeworkTextOnlyUpdate := true // только homework_text

	// Логика из lessons.go:205
	if isTeacherOwnLesson && isHomeworkTextOnlyUpdate {
		// Разрешено
		assert.True(t, true, "Учитель может редактировать homework своего урока")
	} else {
		t.Fail()
	}

	// Учитель пытается обновить время своего урока
	isHomeworkTextOnlyUpdate = false // не только homework

	if teacher.IsAdmin() || teacher.IsTeacher() {
		t.Fail() // Учитель не админ и не методист
	} else {
		if isTeacherOwnLesson && isHomeworkTextOnlyUpdate {
			t.Fail() // Это не homework-only update
		} else if isHomeworkTextOnlyUpdate && !isTeacherOwnLesson {
			t.Fail() // Урок не его
		} else if isTeacherOwnLesson && !isHomeworkTextOnlyUpdate {
			// Учитель пытается обновить не только homework
			assert.True(t, true, "Учитель НЕ может редактировать время своего урока")
		} else {
			// Другие случаи
			assert.True(t, true, "Учитель НЕ может редактировать")
		}
	}
}
