package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserRequest_ValidateTeacherRole(t *testing.T) {
	req := &CreateUserRequest{
		Email:    "teacher@example.com",
		Password: "SecurePassword123",
		FirstName: "Teacher Name", LastName: "Lastname",
		Role:     RoleTeacher,
	}

	err := req.Validate()
	assert.NoError(t, err, "Should allow creating user with teacher role")
	assert.Equal(t, RoleTeacher, req.Role)
}

func TestUser_IsTeacher(t *testing.T) {
	teacherUser := &User{
		ID:       uuid.New(),
		Email:    "teacher@example.com",
		FirstName: "Teacher", LastName: "Lastname",
		Role:     RoleTeacher,
	}

	assert.True(t, teacherUser.IsTeacher(), "User with teacher role should be detected as teacher")

	studentUser := &User{
		ID:       uuid.New(),
		Email:    "student@example.com",
		FirstName: "Student", LastName: "Lastname",
		Role:     RoleStudent,
	}

	assert.False(t, studentUser.IsTeacher(), "Student should not be teacher")
}

func TestUserRole_ValidRoles(t *testing.T) {
	validRoles := []UserRole{
		RoleStudent,
		RoleTeacher,
		RoleAdmin,
	}

	for _, role := range validRoles {
		t.Run(string(role), func(t *testing.T) {
			req := &CreateUserRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123",
				FirstName: "Test User", LastName: "Lastname",
				Role:     role,
			}

			err := req.Validate()
			assert.NoError(t, err, "Role %s should be valid", role)
		})
	}
}

func TestUserRole_InvalidRole(t *testing.T) {
	req := &CreateUserRequest{
		Email:    "test@example.com",
		Password: "SecurePassword123",
		FirstName: "Test User", LastName: "Lastname",
		Role:     UserRole("invalid_role"),
	}

	err := req.Validate()
	assert.Error(t, err, "Should reject invalid role")
	assert.Equal(t, ErrInvalidRole, err)
}

func TestUpdateUserRequest_TeacherRole(t *testing.T) {
	teacherRole := RoleTeacher
	req := &UpdateUserRequest{
		Role: &teacherRole,
	}

	assert.NotNil(t, req.Role)
	assert.Equal(t, RoleTeacher, *req.Role)
}

func TestUser_MarshalJSON_TeacherRole(t *testing.T) {
	user := &User{
		ID:             uuid.New(),
		Email:          "teacher@example.com",
		FirstName: "Teacher User", LastName: "Lastname",
		Role:           RoleTeacher,
		PaymentEnabled: false,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "role")
	assert.Equal(t, string(RoleTeacher), result["role"])
}
