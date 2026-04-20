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

type InternalChatHub struct {
	mu                sync.RWMutex
	connections       map[string]InternalChatConnection
	participantCounts map[string]int
}

const internalChatSocketWriteTimeout = 10 * time.Second

func NewInternalChatHub() *InternalChatHub {
	return &InternalChatHub{
		connections:       make(map[string]InternalChatConnection),
		participantCounts: make(map[string]int),
	}
}

func (h *InternalChatHub) Register(connection InternalChatConnection) {
	if connection.ID == "" || connection.Socket == nil {
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

func (h *InternalChatHub) BroadcastConversationEvent(
	ctx context.Context,
	companyID string,
	firstUserID string,
	secondUserID string,
	payloadForConnection func(connection InternalChatConnection) map[string]any,
) {
	targets := h.matchingConversationConnections(companyID, firstUserID, secondUserID)

	for _, connection := range targets {
		payload := payloadForConnection(connection)
		if payload == nil {
			continue
		}

		writeCtx, cancel := context.WithTimeout(ctx, internalChatSocketWriteTimeout)
		err := wsjson.Write(writeCtx, connection.Socket, payload)
		cancel()
		if err != nil {
			_ = connection.Socket.Close(websocket.StatusInternalError, "broadcast failed")
			h.Unregister(connection.ID)
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
