package models

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLessonBroadcast_Validate(t *testing.T) {
	tests := []struct {
		name      string
		broadcast *LessonBroadcast
		wantErr   error
	}{
		{
			name: "valid broadcast",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  "Важное объявление для студентов урока",
				Status:   LessonBroadcastStatusPending,
			},
			wantErr: nil,
		},
		{
			name: "invalid lesson_id (nil)",
			broadcast: &LessonBroadcast{
				LessonID: uuid.Nil,
				SenderID: uuid.New(),
				Message:  "Test message",
				Status:   LessonBroadcastStatusPending,
			},
			wantErr: ErrInvalidLessonID,
		},
		{
			name: "invalid sender_id (nil)",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.Nil,
				Message:  "Test message",
				Status:   LessonBroadcastStatusPending,
			},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "empty message",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  "",
				Status:   LessonBroadcastStatusPending,
			},
			wantErr: ErrInvalidBroadcastMessage,
		},
		{
			name: "message too long (>4096 chars)",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  strings.Repeat("a", 4097),
				Status:   LessonBroadcastStatusPending,
			},
			wantErr: ErrBroadcastMessageTooLong,
		},
		{
			name: "invalid status",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  "Test message",
				Status:   "invalid_status",
			},
			wantErr: ErrInvalidBroadcastStatus,
		},
		{
			name: "too many files (>10)",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  "Test message",
				Status:   LessonBroadcastStatusPending,
				Files: []*BroadcastFile{
					{BroadcastID: uuid.New(), FileName: "file1.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file2.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file3.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file4.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file5.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file6.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file7.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file8.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file9.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file10.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file11.pdf", FileSize: 1024, MimeType: "application/pdf"},
				},
			},
			wantErr: ErrTooManyFiles,
		},
		{
			name: "max message length (4096 chars)",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  strings.Repeat("a", 4096),
				Status:   LessonBroadcastStatusPending,
			},
			wantErr: nil,
		},
		{
			name: "exactly 10 files (max allowed)",
			broadcast: &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  "Test message",
				Status:   LessonBroadcastStatusPending,
				Files: []*BroadcastFile{
					{BroadcastID: uuid.New(), FileName: "file1.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file2.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file3.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file4.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file5.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file6.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file7.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file8.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file9.pdf", FileSize: 1024, MimeType: "application/pdf"},
					{BroadcastID: uuid.New(), FileName: "file10.pdf", FileSize: 1024, MimeType: "application/pdf"},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.broadcast.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBroadcastFile_Validate(t *testing.T) {
	tests := []struct {
		name    string
		file    *BroadcastFile
		wantErr error
	}{
		{
			name: "valid file",
			file: &BroadcastFile{
				BroadcastID: uuid.New(),
				FileName:    "document.pdf",
				FilePath:    "/uploads/abc123.pdf",
				FileSize:    2048000,
				MimeType:    "application/pdf",
			},
			wantErr: nil,
		},
		{
			name: "invalid broadcast_id (nil)",
			file: &BroadcastFile{
				BroadcastID: uuid.Nil,
				FileName:    "document.pdf",
				FileSize:    2048000,
				MimeType:    "application/pdf",
			},
			wantErr: ErrInvalidBroadcastID,
		},
		{
			name: "empty file_name",
			file: &BroadcastFile{
				BroadcastID: uuid.New(),
				FileName:    "",
				FileSize:    2048000,
				MimeType:    "application/pdf",
			},
			wantErr: ErrInvalidFileName,
		},
		{
			name: "file size zero",
			file: &BroadcastFile{
				BroadcastID: uuid.New(),
				FileName:    "document.pdf",
				FileSize:    0,
				MimeType:    "application/pdf",
			},
			wantErr: ErrInvalidFileSize,
		},
		{
			name: "file size too large (>10MB)",
			file: &BroadcastFile{
				BroadcastID: uuid.New(),
				FileName:    "large.pdf",
				FileSize:    10485761, // 10MB + 1 byte
				MimeType:    "application/pdf",
			},
			wantErr: ErrInvalidFileSize,
		},
		{
			name: "empty mime_type",
			file: &BroadcastFile{
				BroadcastID: uuid.New(),
				FileName:    "document.pdf",
				FileSize:    2048000,
				MimeType:    "",
			},
			wantErr: ErrInvalidMimeType,
		},
		{
			name: "exactly 10MB (max allowed)",
			file: &BroadcastFile{
				BroadcastID: uuid.New(),
				FileName:    "max_size.pdf",
				FileSize:    10485760, // Exactly 10MB
				MimeType:    "application/pdf",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.file.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateLessonBroadcastRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *CreateLessonBroadcastRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: &CreateLessonBroadcastRequest{
				LessonID: uuid.New(),
				Message:  "Важное сообщение для студентов",
			},
			wantErr: nil,
		},
		{
			name: "invalid lesson_id (nil)",
			request: &CreateLessonBroadcastRequest{
				LessonID: uuid.Nil,
				Message:  "Test message",
			},
			wantErr: ErrInvalidLessonID,
		},
		{
			name: "empty message",
			request: &CreateLessonBroadcastRequest{
				LessonID: uuid.New(),
				Message:  "",
			},
			wantErr: ErrInvalidBroadcastMessage,
		},
		{
			name: "message too long",
			request: &CreateLessonBroadcastRequest{
				LessonID: uuid.New(),
				Message:  strings.Repeat("a", 4097),
			},
			wantErr: ErrBroadcastMessageTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLessonBroadcast_StatusChecks(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		isPending  bool
		isSending  bool
		isComplete bool
		isFailed   bool
		isFinal    bool
	}{
		{
			name:       "pending status",
			status:     LessonBroadcastStatusPending,
			isPending:  true,
			isSending:  false,
			isComplete: false,
			isFailed:   false,
			isFinal:    false,
		},
		{
			name:       "sending status",
			status:     LessonBroadcastStatusSending,
			isPending:  false,
			isSending:  true,
			isComplete: false,
			isFailed:   false,
			isFinal:    false,
		},
		{
			name:       "completed status",
			status:     LessonBroadcastStatusCompleted,
			isPending:  false,
			isSending:  false,
			isComplete: true,
			isFailed:   false,
			isFinal:    true,
		},
		{
			name:       "failed status",
			status:     LessonBroadcastStatusFailed,
			isPending:  false,
			isSending:  false,
			isComplete: false,
			isFailed:   true,
			isFinal:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broadcast := &LessonBroadcast{
				LessonID: uuid.New(),
				SenderID: uuid.New(),
				Message:  "Test",
				Status:   tt.status,
			}

			assert.Equal(t, tt.isPending, broadcast.IsPending())
			assert.Equal(t, tt.isSending, broadcast.IsSending())
			assert.Equal(t, tt.isComplete, broadcast.IsCompleted())
			assert.Equal(t, tt.isFailed, broadcast.IsFailed())
			assert.Equal(t, tt.isFinal, broadcast.IsFinalStatus())
		})
	}
}

func TestIsValidBroadcastStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"pending", LessonBroadcastStatusPending, true},
		{"sending", LessonBroadcastStatusSending, true},
		{"completed", LessonBroadcastStatusCompleted, true},
		{"failed", LessonBroadcastStatusFailed, true},
		{"invalid", "invalid_status", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidBroadcastStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}
