package realtime

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type InternalChatConnection struct {
	ID                string
	CompanyID         string
	UserID            string
	CounterpartUserID string
	UserRole          string
	ConnectedAt       time.Time
	Socket            *websocket.Conn
}

type InternalChatPresence struct {
	UserID        string
	Status        string
	Connections   int
	LastChangedAt time.Time
}

type InternalChatHub struct {
	mu                     sync.RWMutex
	connections            map[string]InternalChatConnection
	participantCounts      map[string]int
	participantLastChanged map[string]time.Time
}

const internalChatSocketWriteTimeout = 10 * time.Second

func NewInternalChatHub() *InternalChatHub {
	return &InternalChatHub{
		connections:            make(map[string]InternalChatConnection),
		participantCounts:      make(map[string]int),
		participantLastChanged: make(map[string]time.Time),
	}
}

func (h *InternalChatHub) Register(connection InternalChatConnection) (InternalChatPresence, bool) {
	if connection.ID == "" || connection.Socket == nil {
		return InternalChatPresence{}, false
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now().UTC()
	key := h.participantKey(connection.CompanyID, connection.UserID)
	previousCount := h.participantCounts[key]

	h.connections[connection.ID] = connection
	h.participantCounts[key] = previousCount + 1
	if previousCount == 0 {
		h.participantLastChanged[key] = now
	}

	return InternalChatPresence{
		UserID:        connection.UserID,
		Status:        "online",
		Connections:   h.participantCounts[key],
		LastChangedAt: h.participantLastChanged[key],
	}, previousCount == 0
}

func (h *InternalChatHub) Unregister(connectionID string) (InternalChatConnection, InternalChatPresence, bool) {
	if connectionID == "" {
		return InternalChatConnection{}, InternalChatPresence{}, false
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	connection, ok := h.connections[connectionID]
	if !ok {
		return InternalChatConnection{}, InternalChatPresence{}, false
	}

	delete(h.connections, connectionID)

	key := h.participantKey(connection.CompanyID, connection.UserID)
	previousCount := h.participantCounts[key]
	count := previousCount - 1
	if count <= 0 {
		now := time.Now().UTC()
		delete(h.participantCounts, key)
		h.participantLastChanged[key] = now
		return connection, InternalChatPresence{
			UserID:        connection.UserID,
			Status:        "offline",
			Connections:   0,
			LastChangedAt: now,
		}, previousCount > 0
	}

	h.participantCounts[key] = count
	return connection, InternalChatPresence{
		UserID:        connection.UserID,
		Status:        "online",
		Connections:   count,
		LastChangedAt: h.participantLastChanged[key],
	}, false
}

func (h *InternalChatHub) Presence(companyID string, userID string) InternalChatPresence {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.presenceLocked(companyID, userID)
}

func (h *InternalChatHub) ConversationSnapshot(
	companyID string,
	firstUserID string,
	secondUserID string,
) []InternalChatPresence {
	h.mu.Lock()
	defer h.mu.Unlock()

	return []InternalChatPresence{
		h.presenceLocked(companyID, firstUserID),
		h.presenceLocked(companyID, secondUserID),
	}
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

func (h *InternalChatHub) BroadcastConversationEvent(
	ctx context.Context,
	companyID string,
	firstUserID string,
	secondUserID string,
	skipConnectionID string,
	payloadForConnection func(connection InternalChatConnection) map[string]any,
) {
	targets := h.matchingConversationConnections(companyID, firstUserID, secondUserID)

	for _, connection := range targets {
		if skipConnectionID != "" && connection.ID == skipConnectionID {
			continue
		}

		payload := payloadForConnection(connection)
		if payload == nil {
			continue
		}

		writeCtx, cancel := context.WithTimeout(ctx, internalChatSocketWriteTimeout)
		err := wsjson.Write(writeCtx, connection.Socket, payload)
		cancel()
		if err != nil {
			_ = connection.Socket.Close(websocket.StatusInternalError, "broadcast failed")
			_, _, _ = h.Unregister(connection.ID)
		}
	}
}

func (h *InternalChatHub) participantKey(companyID string, userID string) string {
	return companyID + ":" + userID
}

func (h *InternalChatHub) matchingConversationConnections(
	companyID string,
	firstUserID string,
	secondUserID string,
) []InternalChatConnection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	items := make([]InternalChatConnection, 0)
	for _, connection := range h.connections {
		if connection.CompanyID != companyID {
			continue
		}

		if (connection.UserID == firstUserID && connection.CounterpartUserID == secondUserID) ||
			(connection.UserID == secondUserID && connection.CounterpartUserID == firstUserID) {
			items = append(items, connection)
		}
	}

	return items
}

func (h *InternalChatHub) presenceLocked(companyID string, userID string) InternalChatPresence {
	key := h.participantKey(companyID, userID)
	lastChangedAt, ok := h.participantLastChanged[key]
	if !ok {
		lastChangedAt = time.Now().UTC()
		h.participantLastChanged[key] = lastChangedAt
	}

	count := h.participantCounts[key]
	if count > 0 {
		return InternalChatPresence{
			UserID:        userID,
			Status:        "online",
			Connections:   count,
			LastChangedAt: lastChangedAt,
		}
	}

	return InternalChatPresence{
		UserID:        userID,
		Status:        "offline",
		Connections:   0,
		LastChangedAt: lastChangedAt,
	}
}
