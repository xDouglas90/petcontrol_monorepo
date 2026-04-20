package realtime

import (
	"sync"
	"time"
)

type InternalChatConnection struct {
	ID                string
	CompanyID         string
	UserID            string
	CounterpartUserID string
	UserRole          string
	ConnectedAt       time.Time
}

type InternalChatHub struct {
	mu                sync.RWMutex
	connections       map[string]InternalChatConnection
	participantCounts map[string]int
}

func NewInternalChatHub() *InternalChatHub {
	return &InternalChatHub{
		connections:       make(map[string]InternalChatConnection),
		participantCounts: make(map[string]int),
	}
}

func (h *InternalChatHub) Register(connection InternalChatConnection) {
	if connection.ID == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[connection.ID] = connection
	h.participantCounts[h.participantKey(connection.CompanyID, connection.UserID)]++
}

func (h *InternalChatHub) Unregister(connectionID string) {
	if connectionID == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	connection, ok := h.connections[connectionID]
	if !ok {
		return
	}

	delete(h.connections, connectionID)

	key := h.participantKey(connection.CompanyID, connection.UserID)
	count := h.participantCounts[key] - 1
	if count <= 0 {
		delete(h.participantCounts, key)
		return
	}

	h.participantCounts[key] = count
}

func (h *InternalChatHub) ConnectionCount(companyID string, userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.participantCounts[h.participantKey(companyID, userID)]
}

func (h *InternalChatHub) TotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.connections)
}

func (h *InternalChatHub) participantKey(companyID string, userID string) string {
	return companyID + ":" + userID
}
