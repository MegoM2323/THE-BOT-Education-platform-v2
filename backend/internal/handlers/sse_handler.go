package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/sse"
	"tutoring-platform/pkg/response"
)

const (
	heartbeatInterval = 30 * time.Second
)

type SSEHandler struct {
	connManager *sse.ConnectionManagerUUID
	chatRepo    *repository.ChatRepository
}

func NewSSEHandler(connManager *sse.ConnectionManagerUUID, chatRepo *repository.ChatRepository) *SSEHandler {
	return &SSEHandler{
		connManager: connManager,
		chatRepo:    chatRepo,
	}
}

func (h *SSEHandler) HandleChatEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, ok := middleware.GetSessionFromContext(ctx)
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	userID := session.UserID

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error().Msg("SSE: ResponseWriter does not support Flusher interface")
		response.InternalError(w, "Streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	eventChan := sse.CreateEventChannelUUID()
	h.connManager.AddConnection(userID, eventChan)

	defer func() {
		h.connManager.RemoveConnection(userID, eventChan)
		log.Debug().
			Str("user_id", userID.String()).
			Msg("SSE: Connection closed")
	}()

	log.Debug().
		Str("user_id", userID.String()).
		Msg("SSE: Connection established")

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return
			}

			if err := h.writeEvent(w, flusher, event); err != nil {
				log.Error().Err(err).
					Str("user_id", userID.String()).
					Str("event_type", event.Type).
					Msg("SSE: Failed to write event")
				return
			}

		case <-ticker.C:
			if err := h.writeHeartbeat(w, flusher); err != nil {
				log.Debug().Err(err).
					Str("user_id", userID.String()).
					Msg("SSE: Heartbeat failed, closing connection")
				return
			}

		case <-ctx.Done():
			log.Debug().
				Str("user_id", userID.String()).
				Msg("SSE: Context cancelled")
			return
		}
	}
}

func (h *SSEHandler) writeEvent(w http.ResponseWriter, flusher http.Flusher, event sse.EventUUID) error {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(dataBytes))
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	flusher.Flush()
	return nil
}

func (h *SSEHandler) writeHeartbeat(w http.ResponseWriter, flusher http.Flusher) error {
	_, err := fmt.Fprint(w, ": heartbeat\n\n")
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func (h *SSEHandler) GetChatParticipants(roomID uuid.UUID) ([]uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	room, err := h.chatRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return []uuid.UUID{room.TeacherID, room.StudentID}, nil
}
