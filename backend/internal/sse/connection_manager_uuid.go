package sse

import (
	"sync"

	"github.com/google/uuid"
)

type EventUUID struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ChatParticipantsProviderUUID func(chatID uuid.UUID) []uuid.UUID

type ConnectionManagerUUID struct {
	mu          sync.RWMutex
	connections map[uuid.UUID][]chan EventUUID
	chatUsers   ChatParticipantsProviderUUID
}

func NewConnectionManagerUUID(chatUsersProvider ChatParticipantsProviderUUID) *ConnectionManagerUUID {
	return &ConnectionManagerUUID{
		connections: make(map[uuid.UUID][]chan EventUUID),
		chatUsers:   chatUsersProvider,
	}
}

func (cm *ConnectionManagerUUID) AddConnection(userID uuid.UUID, eventChan chan EventUUID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connections[userID] = append(cm.connections[userID], eventChan)
}

func (cm *ConnectionManagerUUID) RemoveConnection(userID uuid.UUID, eventChan chan EventUUID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	channels, exists := cm.connections[userID]
	if !exists {
		return
	}

	for i, ch := range channels {
		if ch == eventChan {
			cm.connections[userID] = append(channels[:i], channels[i+1:]...)
			close(eventChan)
			break
		}
	}

	if len(cm.connections[userID]) == 0 {
		delete(cm.connections, userID)
	}
}

func (cm *ConnectionManagerUUID) SendToUser(userID uuid.UUID, event EventUUID) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels, exists := cm.connections[userID]
	if !exists || len(channels) == 0 {
		return false
	}

	sent := false
	for _, ch := range channels {
		select {
		case ch <- event:
			sent = true
		default:
		}
	}

	return sent
}

func (cm *ConnectionManagerUUID) SendToChat(chatID uuid.UUID, event EventUUID, excludeUserID uuid.UUID) {
	if cm.chatUsers == nil {
		return
	}

	userIDs := cm.chatUsers(chatID)

	for _, userID := range userIDs {
		if userID == excludeUserID {
			continue
		}
		cm.SendToUser(userID, event)
	}
}

func CreateEventChannelUUID() chan EventUUID {
	return make(chan EventUUID, EventChannelBufferSize)
}

func (cm *ConnectionManagerUUID) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	count := 0
	for _, channels := range cm.connections {
		count += len(channels)
	}
	return count
}

func (cm *ConnectionManagerUUID) GetUserConnectionCount(userID uuid.UUID) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.connections[userID])
}

func (cm *ConnectionManagerUUID) Broadcast(event EventUUID) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, channels := range cm.connections {
		for _, ch := range channels {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (cm *ConnectionManagerUUID) IsUserOnline(userID uuid.UUID) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels, exists := cm.connections[userID]
	return exists && len(channels) > 0
}
