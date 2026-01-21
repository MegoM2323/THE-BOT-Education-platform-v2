package handlers

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// ChatHandler handles HTTP requests for chat
type ChatHandler struct {
	chatService *service.ChatService
	uploadDir   string
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(chatService *service.ChatService, uploadDir string) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		uploadDir:   uploadDir,
	}
}

// GetMyRooms retrieves list of chat rooms for the current user
func (h *ChatHandler) GetMyRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	rooms, err := h.chatService.GetUserChats(ctx, session.UserID, string(session.UserRole))
	if err != nil {
		// Логируем полную ошибку для диагностики
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("method", "GetUserChats").
			Msg("Failed to retrieve user chat rooms")
		// Возвращаем клиенту только обобщённое сообщение
		response.BadRequest(w, response.ErrCodeInternalError, "Unable to retrieve chat rooms")
		return
	}

	if rooms == nil {
		rooms = []*models.ChatRoom{}
	}

	response.Success(w, http.StatusOK, rooms)
}

// GetOrCreateRoom creates or retrieves an existing chat room
func (h *ChatHandler) GetOrCreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	var req struct {
		ParticipantID string `json:"participant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Debug().Err(err).Msg("Failed to decode request body for GetOrCreateRoom")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	participantID, err := uuid.Parse(req.ParticipantID)
	if err != nil {
		log.Debug().Err(err).Str("participant_id", req.ParticipantID).Msg("Invalid participant ID format")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid participant_id format")
		return
	}

	room, err := h.chatService.GetOrCreateRoom(ctx, session.UserID, participantID)
	if err != nil {
		// Логируем полную ошибку для диагностики
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("participant_id", participantID.String()).
			Str("method", "GetOrCreateRoom").
			Msg("Failed to get or create chat room")
		// Возвращаем клиенту только обобщённое сообщение
		response.BadRequest(w, response.ErrCodeChatRoomNotFound, "Unable to create or retrieve chat room")
		return
	}

	otherUserID := room.GetOtherParticipant(session.UserID)
	participant, err := h.chatService.GetUserByID(ctx, otherUserID)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("other_user_id", otherUserID.String()).
			Str("method", "GetOrCreateRoom:GetUserByID").
			Msg("Failed to get participant data")
		response.BadRequest(w, response.ErrCodeInternalError, "Unable to retrieve participant information")
		return
	}

	room.ParticipantID = participant.ID
	room.ParticipantName = participant.FullName
	room.ParticipantRole = string(participant.Role)

	response.Success(w, http.StatusOK, room)
}

// SendMessage sends a message to a room
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	roomID := chi.URLParam(r, "roomId")
	parsedRoomID, err := uuid.Parse(roomID)
	if err != nil {
		log.Debug().Err(err).Str("room_id", roomID).Msg("Invalid room ID format")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid room ID format")
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Debug().Err(err).Str("user_id", session.UserID.String()).Msg("Failed to parse multipart form")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid form data")
		return
	}

	messageText := r.FormValue("message")
	if messageText == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message cannot be empty")
		return
	}

	sendReq := &models.SendMessageRequest{
		RoomID:      parsedRoomID,
		MessageText: messageText,
	}

	message, err := h.chatService.SendMessage(ctx, session.UserID, sendReq)
	if err != nil {
		// Логируем полную ошибку для диагностики
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("room_id", parsedRoomID.String()).
			Str("method", "SendMessage").
			Msg("Failed to send message")
		// Возвращаем клиенту только обобщённое сообщение
		response.BadRequest(w, response.ErrCodeInternalError, "Unable to send message")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) > 0 {
		uploadsDir := h.uploadDir
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			// Логируем полную ошибку для диагностики
			log.Error().Err(err).
				Str("directory", uploadsDir).
				Str("method", "SendMessage:MkdirAll").
				Msg("Failed to create uploads directory")
			// Возвращаем клиенту только обобщённое сообщение
			response.BadRequest(w, response.ErrCodeChatFileUploadFailed, "Unable to process file uploads")
			return
		}

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				// Логируем полную ошибку для диагностики
				log.Error().Err(err).
					Str("filename", fileHeader.Filename).
					Str("method", "SendMessage:fileHeader.Open").
					Msg("Failed to open uploaded file")
				// Возвращаем клиенту только обобщённое сообщение
				response.BadRequest(w, response.ErrCodeChatFileUploadFailed, "Unable to process file upload")
				return
			}
			defer file.Close()

			filename := uuid.New().String() + filepath.Ext(fileHeader.Filename)
			filePath := filepath.Join(uploadsDir, filename)
			dst, err := os.Create(filePath)
			if err != nil {
				// Логируем полную ошибку для диагностики
				log.Error().Err(err).
					Str("file_path", filePath).
					Str("method", "SendMessage:os.Create").
					Msg("Failed to create file on disk")
				// Возвращаем клиенту только обобщённое сообщение
				response.BadRequest(w, response.ErrCodeChatFileUploadFailed, "Unable to save file")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				// Логируем полную ошибку для диагностики
				log.Error().Err(err).
					Str("file_path", filePath).
					Str("method", "SendMessage:io.Copy").
					Msg("Failed to write file content")
				// Возвращаем клиенту только обобщённое сообщение
				response.BadRequest(w, response.ErrCodeChatFileUploadFailed, "Unable to save file")
				return
			}

			attachment := &models.FileAttachment{
				ID:        uuid.New(),
				MessageID: message.ID,
				FileName:  fileHeader.Filename,
				FilePath:  filePath,
				FileSize:  fileHeader.Size,
				MimeType:  fileHeader.Header.Get("Content-Type"),
			}

			if err := h.chatService.CreateAttachment(ctx, session.UserID, attachment); err != nil {
				// Логируем полную ошибку для диагностики
				log.Error().Err(err).
					Str("message_id", message.ID.String()).
					Str("filename", fileHeader.Filename).
					Str("method", "SendMessage:CreateAttachment").
					Msg("Failed to save attachment metadata")
				// Возвращаем клиенту только обобщённое сообщение
				response.BadRequest(w, response.ErrCodeChatFileUploadFailed, "Unable to attach file")
				return
			}
		}
	}

	// Загружаем полное сообщение с вложениями для ответа
	fullMessage, err := h.chatService.GetMessageWithAttachments(ctx, session.UserID, message.ID)
	if err != nil {
		// Логируем полную ошибку для диагностики (не критичная, сообщение уже создано)
		log.Warn().Err(err).
			Str("message_id", message.ID.String()).
			Str("method", "SendMessage:GetMessageWithAttachments").
			Msg("Failed to load message with attachments, returning base message")
		// Если не удалось загрузить полное сообщение, возвращаем оригинальное
		// (сообщение уже создано, это не критичная ошибка)
		response.Success(w, http.StatusOK, message)
		return
	}

	response.Success(w, http.StatusOK, fullMessage)
}

// GetMessages retrieves messages from a room with pagination
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	roomID := chi.URLParam(r, "roomId")
	parsedRoomID, err := uuid.Parse(roomID)
	if err != nil {
		log.Debug().Err(err).Str("room_id", roomID).Msg("Invalid room ID format")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid room ID format")
		return
	}

	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	getReq := &models.GetMessagesRequest{
		RoomID: parsedRoomID,
		Limit:  limit,
		Offset: offset,
	}

	messages, err := h.chatService.GetChatHistory(ctx, session.UserID, string(session.UserRole), getReq)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("room_id", parsedRoomID.String()).
			Str("role", string(session.UserRole)).
			Int("limit", limit).
			Int("offset", offset).
			Str("method", "GetChatHistory").
			Msg("Failed to retrieve chat history")
		response.BadRequest(w, response.ErrCodeInternalError, "Unable to retrieve messages")
		return
	}

	if messages == nil {
		messages = []*models.Message{}
	}

	response.Success(w, http.StatusOK, messages)
}

// DownloadFile downloads a file from a message
func (h *ChatHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	roomID := chi.URLParam(r, "roomId")
	fileID := chi.URLParam(r, "fileId")

	parsedRoomID, err := uuid.Parse(roomID)
	if err != nil {
		log.Debug().Err(err).Str("room_id", roomID).Msg("Invalid room ID format")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid room ID format")
		return
	}

	parsedFileID, err := uuid.Parse(fileID)
	if err != nil {
		log.Debug().Err(err).Str("file_id", fileID).Msg("Invalid file ID format")
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid file ID format")
		return
	}

	attachment, err := h.chatService.GetAttachmentByID(ctx, session.UserID, parsedRoomID, parsedFileID)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("file_id", parsedFileID.String()).
			Str("room_id", parsedRoomID.String()).
			Str("method", "GetAttachmentByID").
			Msg("Failed to retrieve attachment metadata")
		response.BadRequest(w, response.ErrCodeChatFileNotFound, "File not found")
		return
	}

	// Проверяем существование файла в файловой системе
	// Если путь в БД относительный, используем uploadDir как базовую директорию
	fullPath := attachment.FilePath
	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(h.uploadDir, filepath.Base(attachment.FilePath))
	}

	if _, err := os.Stat(fullPath); err != nil {
		log.Error().Err(err).
			Str("file_path", fullPath).
			Str("attachment_path", attachment.FilePath).
			Str("method", "DownloadFile:os.Stat").
			Msg("File not found on disk")
		response.BadRequest(w, response.ErrCodeChatFileNotFound, "File not available")
		return
	}

	contentDisposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": attachment.FileName,
	})
	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", attachment.MimeType)
	http.ServeFile(w, r, fullPath)
}

// ListAllChats returns all chats for admin panel
func (h *ChatHandler) ListAllChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	if session.UserRole != models.RoleAdmin {
		log.Warn().
			Str("user_id", session.UserID.String()).
			Str("role", string(session.UserRole)).
			Msg("Non-admin user attempted to access ListAllChats")
		response.Forbidden(w, "Admin access required")
		return
	}

	chats, err := h.chatService.GetAllChats(ctx)
	if err != nil {
		log.Error().Err(err).
			Str("user_id", session.UserID.String()).
			Str("method", "ListAllChats").
			Msg("Failed to retrieve all chats")
		response.BadRequest(w, response.ErrCodeInternalError, "Unable to retrieve chats")
		return
	}

	if chats == nil {
		chats = []repository.ChatRoomWithDetails{}
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"chats": chats,
	})
}
