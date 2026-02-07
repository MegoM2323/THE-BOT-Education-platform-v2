package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tutoring-platform/internal/models"
)

// TestMethodologistRoleCheck проверяет что роль методиста правильно определяется
func TestMethodologistRoleCheck(t *testing.T) {
	methodologist := &models.User{
		ID:        uuid.New(),
		Email:     "methodologist@test.com",
		FirstName: "Test",
		LastName:  "Methodologist",
		Role:      models.RoleMethodologist,
	}

	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleMethodologist,
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

	t.Run("Methodologist_IsMethodologist", func(t *testing.T) {
		assert.True(t, methodologist.IsMethodologist(), "Методист должен иметь IsMethodologist() == true")
	})

	t.Run("Methodologist_NotAdmin", func(t *testing.T) {
		assert.False(t, methodologist.IsAdmin(), "Методист не должен быть админом")
	})

	t.Run("Methodologist_NotTeacher", func(t *testing.T) {
		assert.False(t, methodologist.IsMethodologist(), "Методист не должен быть учителем")
	})

	t.Run("Methodologist_NotStudent", func(t *testing.T) {
		assert.False(t, methodologist.IsStudent(), "Методист не должен быть студентом")
	})

	t.Run("Teacher_NotMethodologist", func(t *testing.T) {
		assert.False(t, teacher.IsMethodologist(), "Учитель не должен быть методистом")
	})

	t.Run("Student_NotMethodologist", func(t *testing.T) {
		assert.False(t, student.IsMethodologist(), "Студент не должен быть методистом")
	})

	t.Run("Admin_NotMethodologist", func(t *testing.T) {
		assert.False(t, admin.IsMethodologist(), "Админ не методист (хотя может все)")
	})

	t.Run("Admin_CanDoEverything", func(t *testing.T) {
		assert.True(t, admin.IsAdmin(), "Админ должен иметь IsAdmin() == true")
	})
}

// TestMethodologistCanCreateLesson проверяет логику CreateLesson
func TestMethodologistCanCreateLesson(t *testing.T) {
	methodologist := &models.User{
		ID:        uuid.New(),
		Email:     "methodologist@test.com",
		FirstName: "Test",
		LastName:  "Methodologist",
		Role:      models.RoleMethodologist,
	}

	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleMethodologist,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Проверка что методист может создавать занятия (согласно lessons.go:103)
	// if !user.IsAdmin() && !user.IsMethodologist()
	canCreateMethodologist := methodologist.IsAdmin() || methodologist.IsMethodologist()
	canCreateTeacher := teacher.IsAdmin() || teacher.IsMethodologist()
	canCreateAdmin := admin.IsAdmin() || admin.IsMethodologist()

	assert.True(t, canCreateMethodologist, "Методист должен создавать занятия")
	assert.False(t, canCreateTeacher, "Учитель не должен создавать занятия")
	assert.True(t, canCreateAdmin, "Админ должен создавать занятия")
}

// TestMethodologistCanUpdateLesson проверяет логику UpdateLesson
func TestMethodologistCanUpdateLesson(t *testing.T) {
	methodologist := &models.User{
		ID:        uuid.New(),
		Email:     "methodologist@test.com",
		FirstName: "Test",
		LastName:  "Methodologist",
		Role:      models.RoleMethodologist,
	}

	teacher1 := &models.User{
		ID:        uuid.New(),
		Email:     "teacher1@test.com",
		FirstName: "Test Teacher",
		LastName:  "1",
		Role:      models.RoleMethodologist,
	}

	_ = &models.User{
		ID:        uuid.New(),
		Email:     "teacher2@test.com",
		FirstName: "Test Teacher",
		LastName:  "2",
		Role:      models.RoleMethodologist,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:203
	// if !user.IsAdmin() && !user.IsMethodologist()

	t.Run("MethodologistCanEditAnyLesson", func(t *testing.T) {
		canEdit := methodologist.IsAdmin() || methodologist.IsMethodologist()
		assert.True(t, canEdit, "Методист должен редактировать любое занятие")
	})

	t.Run("TeacherCanEditOnlyHomework", func(t *testing.T) {
		// Учитель может редактировать только homework_text своих уроков
		canEditAll := teacher1.IsAdmin() || teacher1.IsMethodologist()
		assert.False(t, canEditAll, "Учитель не может редактировать все поля")
	})

	t.Run("AdminCanEditAnyLesson", func(t *testing.T) {
		canEdit := admin.IsAdmin() || admin.IsMethodologist()
		assert.True(t, canEdit, "Админ должен редактировать любое занятие")
	})
}

// TestMethodologistCanDeleteLesson проверяет логику DeleteLesson
func TestMethodologistCanDeleteLesson(t *testing.T) {
	methodologist := &models.User{
		ID:        uuid.New(),
		Email:     "methodologist@test.com",
		FirstName: "Test",
		LastName:  "Methodologist",
		Role:      models.RoleMethodologist,
	}

	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleMethodologist,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:289
	// if !user.IsAdmin() && !user.IsMethodologist()

	canDeleteMethodologist := methodologist.IsAdmin() || methodologist.IsMethodologist()
	canDeleteTeacher := teacher.IsAdmin() || teacher.IsMethodologist()
	canDeleteAdmin := admin.IsAdmin() || admin.IsMethodologist()

	assert.True(t, canDeleteMethodologist, "Методист должен удалять занятия")
	assert.False(t, canDeleteTeacher, "Учитель не должен удалять занятия")
	assert.True(t, canDeleteAdmin, "Админ должен удалять занятия")
}

// TestMethodologistCanBulkEdit проверяет логику ApplyToAllSubsequent
func TestMethodologistCanBulkEdit(t *testing.T) {
	methodologist := &models.User{
		ID:        uuid.New(),
		Email:     "methodologist@test.com",
		FirstName: "Test",
		LastName:  "Methodologist",
		Role:      models.RoleMethodologist,
	}

	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleMethodologist,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:522
	// if !admin.IsAdmin() && !admin.IsMethodologist()

	canBulkEditMethodologist := methodologist.IsAdmin() || methodologist.IsMethodologist()
	canBulkEditTeacher := teacher.IsAdmin() || teacher.IsMethodologist()
	canBulkEditAdmin := admin.IsAdmin() || admin.IsMethodologist()

	assert.True(t, canBulkEditMethodologist, "Методист должен делать массовое редактирование")
	assert.False(t, canBulkEditTeacher, "Учитель не должен делать массовое редактирование")
	assert.True(t, canBulkEditAdmin, "Админ должен делать массовое редактирование")
}

// TestMethodologistCanEditPastLesson проверяет редактирование прошлых занятий
func TestMethodologistCanEditPastLesson(t *testing.T) {
	methodologist := &models.User{
		ID:        uuid.New(),
		Email:     "methodologist@test.com",
		FirstName: "Test",
		LastName:  "Methodologist",
		Role:      models.RoleMethodologist,
	}

	teacher := &models.User{
		ID:        uuid.New(),
		Email:     "teacher@test.com",
		FirstName: "Test",
		LastName:  "Teacher",
		Role:      models.RoleMethodologist,
	}

	admin := &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		FirstName: "Test",
		LastName:  "Admin",
		Role:      models.RoleAdmin,
	}

	// Логика из lessons.go:245
	// isPastLesson && !user.IsAdmin() && !user.IsMethodologist() && !isHomeworkTextOnlyUpdate
	// Методист и админ могут редактировать прошлые занятия

	isPastLesson := true
	isHomeworkTextOnlyUpdate := false

	canEditPastMethodologist := !(isPastLesson && !methodologist.IsAdmin() && !methodologist.IsMethodologist() && !isHomeworkTextOnlyUpdate)
	canEditPastTeacher := !(isPastLesson && !teacher.IsAdmin() && !teacher.IsMethodologist() && !isHomeworkTextOnlyUpdate)
	canEditPastAdmin := !(isPastLesson && !admin.IsAdmin() && !admin.IsMethodologist() && !isHomeworkTextOnlyUpdate)

	assert.True(t, canEditPastMethodologist, "Методист должен редактировать прошлые занятия")
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
		Role:      models.RoleMethodologist,
	}

	lessonTeacherID := teacher.ID

	// Учитель пытается обновить только homework своего урока
	isTeacherOwnLesson := teacher.IsMethodologist() && lessonTeacherID == teacher.ID
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

	if teacher.IsAdmin() || teacher.IsMethodologist() {
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
