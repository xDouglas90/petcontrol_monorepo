package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/require"
)

func TestInternalChatHubRegisterAndUnregister(t *testing.T) {
	hub := NewInternalChatHub()

	first := InternalChatConnection{
		ID:                "conn-1",
		CompanyID:         "company-1",
		UserID:            "user-1",
		CounterpartUserID: "user-2",
		UserRole:          "admin",
		ConnectedAt:       time.Now(),
	}
	second := InternalChatConnection{
		ID:                "conn-2",
		CompanyID:         "company-1",
		UserID:            "user-1",
		CounterpartUserID: "user-2",
		UserRole:          "admin",
		ConnectedAt:       time.Now(),
	}

	hub.Register(first)
	hub.Register(second)

	require.Equal(t, 2, hub.TotalConnections())
	require.Equal(t, 2, hub.ConnectionCount("company-1", "user-1"))

	hub.Unregister("conn-1")

	require.Equal(t, 1, hub.TotalConnections())
	require.Equal(t, 1, hub.ConnectionCount("company-1", "user-1"))

	hub.Unregister("conn-2")

	require.Equal(t, 0, hub.TotalConnections())
	require.Equal(t, 0, hub.ConnectionCount("company-1", "user-1"))
}

func TestInternalChatHubBroadcastConversationEvent(t *testing.T) {
	hub := NewInternalChatHub()
	serverConnected := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols:   []string{"petcontrol.internal-chat.v1"},
			OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
		})
		require.NoError(t, err)

		hub.Register(InternalChatConnection{
			ID:                "conn-1",
			CompanyID:         "company-1",
			UserID:            "admin-1",
			CounterpartUserID: "system-1",
			UserRole:          "admin",
			ConnectedAt:       time.Now(),
			Socket:            conn,
		})
		close(serverConnected)

		ctx := conn.CloseRead(r.Context())
		<-ctx.Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+server.URL[len("http"):], &websocket.DialOptions{
		Subprotocols: []string{"petcontrol.internal-chat.v1"},
	})
	require.NoError(t, err)
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}()

	<-serverConnected

	hub.BroadcastConversationEvent(ctx, "company-1", "admin-1", "system-1", func(connection InternalChatConnection) map[string]any {
		return map[string]any{
			"type":                "chat.message.created",
			"company_id":          connection.CompanyID,
			"counterpart_user_id": connection.CounterpartUserID,
			"emitted_at":          time.Now().UTC().Format(time.RFC3339),
			"message": map[string]any{
				"id": "message-1",
			},
		}
	})

	var event map[string]any
	require.NoError(t, wsjson.Read(ctx, conn, &event))
	require.Equal(t, "chat.message.created", event["type"])
	require.Equal(t, "company-1", event["company_id"])
	require.Equal(t, "system-1", event["counterpart_user_id"])
}
