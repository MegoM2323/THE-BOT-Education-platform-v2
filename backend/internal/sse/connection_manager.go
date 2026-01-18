package sse

import (
	"log"
	"sync"
)

const (
	// EventChannelBufferSize is the buffer size for event channels
	EventChannelBufferSize = 10
)

// Event represents an SSE event to be sent to clients
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// ChatParticipantsProvider is a callback interface for getting chat participants
type ChatParticipantsProvider func(chatID int) []int

// ConnectionManager manages SSE connections for users
// Thread-safe implementation allowing multiple connections per user (multiple tabs)
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[int][]chan Event
	chatUsers   ChatParticipantsProvider
}

// NewConnectionManager creates a new ConnectionManager instance
func NewConnectionManager(chatUsersProvider ChatParticipantsProvider) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[int][]chan Event),
		chatUsers:   chatUsersProvider,
	}
}

// AddConnection registers a new SSE connection for a user
// Returns a buffered channel for receiving events
func (cm *ConnectionManager) AddConnection(userID int, eventChan chan Event) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connections[userID] = append(cm.connections[userID], eventChan)
}

// RemoveConnection removes an SSE connection for a user
// Safe to call even if connection doesn't exist
func (cm *ConnectionManager) RemoveConnection(userID int, eventChan chan Event) {
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

// SendToUser sends an event to all connections of a specific user
// Returns true if at least one connection received the event
func (cm *ConnectionManager) SendToUser(userID int, event Event) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels, exists := cm.connections[userID]
	if !exists || len(channels) == 0 {
		return false
	}

	sent := false
	for i, ch := range channels {
		select {
		case ch <- event:
			sent = true
		default:
			log.Printf("[WARNING] Message dropped for user %d (channel %d) - buffer full\n", userID, i)
		}
	}

	return sent
}

// SendToChat sends an event to all participants of a chat room
// Uses the ChatParticipantsProvider callback to get participant IDs
// excludeUserID can be used to skip sending to the original sender (pass 0 to send to all)
func (cm *ConnectionManager) SendToChat(chatID int, event Event, excludeUserID int) {
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

// CreateEventChannel creates a new buffered event channel
func CreateEventChannel() chan Event {
	return make(chan Event, EventChannelBufferSize)
}

// GetConnectionCount returns the total number of active connections
func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	count := 0
	for _, channels := range cm.connections {
		count += len(channels)
	}
	return count
}

// GetUserConnectionCount returns the number of connections for a specific user
func (cm *ConnectionManager) GetUserConnectionCount(userID int) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.connections[userID])
}

// Broadcast sends an event to all connected users
func (cm *ConnectionManager) Broadcast(event Event) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for userID, channels := range cm.connections {
		for i, ch := range channels {
			select {
			case ch <- event:
			default:
				log.Printf("[WARNING] Broadcast message dropped for user %d (channel %d) - buffer full\n", userID, i)
			}
		}
	}
}
