package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_MarshalJSON_PaymentEnabled(t *testing.T) {
	user := &User{
		ID:             uuid.New(),
		Email:          "test@example.com",
		FullName:       "Test User",
		Role:           RoleStudent,
		PaymentEnabled: true,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Проверяем что payment_enabled присутствует в JSON
	assert.Contains(t, result, "payment_enabled")
	assert.Equal(t, true, result["payment_enabled"])
}

func TestUser_MarshalJSON_PaymentDisabled(t *testing.T) {
	user := &User{
		ID:             uuid.New(),
		Email:          "test@example.com",
		FullName:       "Test User",
		Role:           RoleStudent,
		PaymentEnabled: false,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Проверяем что payment_enabled = false в JSON
	assert.Contains(t, result, "payment_enabled")
	assert.Equal(t, false, result["payment_enabled"])
}

func TestUpdateUserRequest_PaymentEnabled(t *testing.T) {
	// Тест с payment_enabled = true
	trueValue := true
	req := &UpdateUserRequest{
		PaymentEnabled: &trueValue,
	}

	assert.NotNil(t, req.PaymentEnabled)
	assert.Equal(t, true, *req.PaymentEnabled)

	// Тест с payment_enabled = false
	falseValue := false
	req2 := &UpdateUserRequest{
		PaymentEnabled: &falseValue,
	}

	assert.NotNil(t, req2.PaymentEnabled)
	assert.Equal(t, false, *req2.PaymentEnabled)

	// Тест без payment_enabled (nil)
	req3 := &UpdateUserRequest{}
	assert.Nil(t, req3.PaymentEnabled)
}
