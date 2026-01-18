package handlers

import (
	"encoding/json"
	"testing"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

func TestSendBroadcastRequest_Validate(t *testing.T) {
	listID := uuid.New()

	tests := []struct {
		name    string
		req     models.SendBroadcastRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with list_id",
			req: models.SendBroadcastRequest{
				ListID:  &listID,
				Message: "Test message",
			},
			wantErr: false,
		},
		{
			name: "valid with user_ids",
			req: models.SendBroadcastRequest{
				UserIDs: []uuid.UUID{uuid.New(), uuid.New()},
				Message: "Test message",
			},
			wantErr: false,
		},
		{
			name: "invalid - both list_id and user_ids",
			req: models.SendBroadcastRequest{
				ListID:  &listID,
				UserIDs: []uuid.UUID{uuid.New()},
				Message: "Test message",
			},
			wantErr: true,
			errMsg:  "cannot specify both list_id and user_ids",
		},
		{
			name: "invalid - neither list_id nor user_ids",
			req: models.SendBroadcastRequest{
				Message: "Test message",
			},
			wantErr: true,
			errMsg:  "either list_id or user_ids must be provided",
		},
		{
			name: "invalid - empty message",
			req: models.SendBroadcastRequest{
				ListID:  &listID,
				Message: "",
			},
			wantErr: true,
		},
		{
			name: "invalid - message too long",
			req: models.SendBroadcastRequest{
				ListID:  &listID,
				Message: string(make([]byte, 4097)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSendBroadcastRequest_JSONMarshaling(t *testing.T) {
	t.Run("marshal with list_id", func(t *testing.T) {
		listID := uuid.New()
		req := models.SendBroadcastRequest{
			ListID:  &listID,
			Message: "Test message",
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded models.SendBroadcastRequest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if decoded.ListID == nil || *decoded.ListID != listID {
			t.Errorf("ListID mismatch: got %v, want %v", decoded.ListID, listID)
		}
		if len(decoded.UserIDs) != 0 {
			t.Errorf("UserIDs should be empty, got %v", decoded.UserIDs)
		}
	})

	t.Run("marshal with user_ids", func(t *testing.T) {
		userIDs := []uuid.UUID{uuid.New(), uuid.New()}
		req := models.SendBroadcastRequest{
			UserIDs: userIDs,
			Message: "Test message",
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var decoded models.SendBroadcastRequest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if decoded.ListID != nil {
			t.Errorf("ListID should be nil, got %v", decoded.ListID)
		}
		if len(decoded.UserIDs) != 2 {
			t.Errorf("UserIDs length mismatch: got %d, want 2", len(decoded.UserIDs))
		}
	})

	t.Run("unmarshal missing list_id and user_ids", func(t *testing.T) {
		jsonData := `{"message": "Test message"}`

		var req models.SendBroadcastRequest
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if req.ListID != nil {
			t.Errorf("ListID should be nil, got %v", req.ListID)
		}
		if len(req.UserIDs) != 0 {
			t.Errorf("UserIDs should be empty, got %v", req.UserIDs)
		}

		// Should fail validation
		if err := req.Validate(); err == nil {
			t.Error("Validate() expected error for missing both list_id and user_ids")
		}
	})
}

func TestSendBroadcastRequest_BackwardCompatibility(t *testing.T) {
	// Тест проверяет что старый формат с только list_id все еще работает
	t.Run("old format with list_id", func(t *testing.T) {
		listID := uuid.New()
		jsonData := map[string]interface{}{
			"list_id": listID.String(),
			"message": "Test broadcast",
		}

		data, err := json.Marshal(jsonData)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}

		var req models.SendBroadcastRequest
		if err := json.Unmarshal(data, &req); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}

		if req.ListID == nil {
			t.Error("ListID should not be nil for old format")
		}
		if *req.ListID != listID {
			t.Errorf("ListID mismatch: got %v, want %v", *req.ListID, listID)
		}

		// Should pass validation
		if err := req.Validate(); err != nil {
			t.Errorf("Validate() unexpected error: %v", err)
		}
	})
}
