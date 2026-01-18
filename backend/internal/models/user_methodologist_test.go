package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUserRequest_ValidateMethodologistRole(t *testing.T) {
	req := &CreateUserRequest{
		Email:    "methodologist@example.com",
		Password: "SecurePassword123",
		FullName: "Methodologist Name",
		Role:     RoleMethodologist,
	}

	err := req.Validate()
	assert.NoError(t, err, "Should allow creating user with methodologist role")
	assert.Equal(t, RoleMethodologist, req.Role)
}

func TestUser_IsMethodologist(t *testing.T) {
	methodologistUser := &User{
		ID:       uuid.New(),
		Email:    "methodologist@example.com",
		FullName: "Methodologist",
		Role:     RoleMethodologist,
	}

	assert.True(t, methodologistUser.IsMethodologist(), "User with methodologist role should be detected as methodologist")

	studentUser := &User{
		ID:       uuid.New(),
		Email:    "student@example.com",
		FullName: "Student",
		Role:     RoleStudent,
	}

	assert.False(t, studentUser.IsMethodologist(), "Student should not be methodologist")
}

func TestUserRole_ValidRoles(t *testing.T) {
	validRoles := []UserRole{
		RoleStudent,
		RoleMethodologist,
		RoleAdmin,
	}

	for _, role := range validRoles {
		t.Run(string(role), func(t *testing.T) {
			req := &CreateUserRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123",
				FullName: "Test User",
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
		FullName: "Test User",
		Role:     UserRole("invalid_role"),
	}

	err := req.Validate()
	assert.Error(t, err, "Should reject invalid role")
	assert.Equal(t, ErrInvalidRole, err)
}

func TestUpdateUserRequest_MethodologistRole(t *testing.T) {
	methodologistRole := RoleMethodologist
	req := &UpdateUserRequest{
		Role: &methodologistRole,
	}

	assert.NotNil(t, req.Role)
	assert.Equal(t, RoleMethodologist, *req.Role)
}

func TestUser_MarshalJSON_MethodologistRole(t *testing.T) {
	user := &User{
		ID:             uuid.New(),
		Email:          "methodologist@example.com",
		FullName:       "Methodologist User",
		Role:           RoleMethodologist,
		PaymentEnabled: false,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Contains(t, result, "role")
	assert.Equal(t, string(RoleMethodologist), result["role"])
}
