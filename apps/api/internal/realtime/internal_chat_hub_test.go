package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
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
		Socket:            &websocket.Conn{},
	}
	second := InternalChatConnection{
		ID:                "conn-2",
		CompanyID:         "company-1",
		UserID:            "user-1",
		CounterpartUserID: "user-2",
		UserRole:          "admin",
		ConnectedAt:       time.Now(),
		Socket:            &websocket.Conn{},
	}

	presence, becameOnline := hub.Register(first)
	require.True(t, becameOnline)
	require.Equal(t, "online", presence.Status)
	require.Equal(t, 1, presence.Connections)

	presence, becameOnline = hub.Register(second)
	require.False(t, becameOnline)
	require.Equal(t, "online", presence.Status)
	require.Equal(t, 2, presence.Connections)

	require.Equal(t, 2, hub.TotalConnections())
	require.Equal(t, 2, hub.ConnectionCount("company-1", "user-1"))

	_, presence, becameOffline := hub.Unregister("conn-1")
	require.False(t, becameOffline)
	require.Equal(t, "online", presence.Status)
	require.Equal(t, 1, presence.Connections)

	require.Equal(t, 1, hub.TotalConnections())
	require.Equal(t, 1, hub.ConnectionCount("company-1", "user-1"))

	_, presence, becameOffline = hub.Unregister("conn-2")
	require.True(t, becameOffline)
	require.Equal(t, "offline", presence.Status)
	require.Equal(t, 0, presence.Connections)

	require.Equal(t, 0, hub.TotalConnections())
	require.Equal(t, 0, hub.ConnectionCount("company-1", "user-1"))

	stats := hub.Stats()
	require.Equal(t, 0, stats.ActiveConnections)
	require.Equal(t, 0, stats.OnlineParticipants)
	require.EqualValues(t, 2, stats.TotalConnectionsOpened)
	require.EqualValues(t, 2, stats.TotalConnectionsClosed)
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

	hub.BroadcastConversationEvent(ctx, "company-1", "admin-1", "system-1", "", func(connection InternalChatConnection) map[string]any {
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

	stats := hub.Stats()
	require.EqualValues(t, 1, stats.TotalBroadcastEvents)
	require.EqualValues(t, 1, stats.TotalBroadcastDeliveries)
	require.EqualValues(t, 0, stats.TotalBroadcastFailures)
}

func TestInternalChatHubConversationSnapshot(t *testing.T) {
	hub := NewInternalChatHub()
	hub.Register(InternalChatConnection{
		ID:                "conn-1",
		CompanyID:         "company-1",
		UserID:            "admin-1",
		CounterpartUserID: "system-1",
		UserRole:          "admin",
		ConnectedAt:       time.Now(),
		Socket:            &websocket.Conn{},
	})

	snapshot := hub.ConversationSnapshot("company-1", "admin-1", "system-1")
	require.Len(t, snapshot, 2)
	require.Equal(t, "admin-1", snapshot[0].UserID)
	require.Equal(t, "online", snapshot[0].Status)
	require.Equal(t, 1, snapshot[0].Connections)
	require.Equal(t, "system-1", snapshot[1].UserID)
	require.Equal(t, "offline", snapshot[1].Status)
	require.Equal(t, 0, snapshot[1].Connections)
}

func TestInternalChatHubStatsCounters(t *testing.T) {
	hub := NewInternalChatHub()

	hub.RecordInvalidPayload()
	hub.RecordInvalidPayload()
	hub.RecordPingFailure()
	hub.RecordSocketError()

	stats := hub.Stats()
	require.EqualValues(t, 2, stats.InvalidPayloads)
	require.EqualValues(t, 1, stats.PingFailures)
	require.EqualValues(t, 1, stats.SocketErrors)
}

func TestInternalChatHubConcurrentRegisterBroadcastAndUnregister(t *testing.T) {
	hub := NewInternalChatHub()

	const totalConnections = 32
	serverConnected := make(chan struct{}, totalConnections)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols:   []string{"petcontrol.internal-chat.v1"},
			OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
		})
		require.NoError(t, err)

		id := r.URL.Query().Get("id")
		role := "admin"
		userID := "admin-1"
		counterpartID := "system-1"
		if r.URL.Query().Get("role") == "system" {
			role = "system"
			userID = "system-1"
			counterpartID = "admin-1"
		}

		hub.Register(InternalChatConnection{
			ID:                id,
			CompanyID:         "company-1",
			UserID:            userID,
			CounterpartUserID: counterpartID,
			UserRole:          role,
			ConnectedAt:       time.Now(),
			Socket:            conn,
		})
		serverConnected <- struct{}{}

		ctx := conn.CloseRead(r.Context())
		<-ctx.Done()
		_, _, _ = hub.Unregister(id)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connections := make([]*websocket.Conn, 0, totalConnections)
	for i := range totalConnections {
		role := "admin"
		if i >= totalConnections/2 {
			role = "system"
		}
		conn, _, err := websocket.Dial(ctx, "ws"+server.URL[len("http"):]+"?id=conn-"+strconv.Itoa(i)+"&role="+role, &websocket.DialOptions{
			Subprotocols: []string{"petcontrol.internal-chat.v1"},
		})
		require.NoError(t, err)
		connections = append(connections, conn)
	}

	for range totalConnections {
		<-serverConnected
	}

	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(connection *websocket.Conn) {
			defer wg.Done()
			var event map[string]any
			require.NoError(t, wsjson.Read(ctx, connection, &event))
			require.Equal(t, "chat.message.created", event["type"])
		}(conn)
	}

	hub.BroadcastConversationEvent(ctx, "company-1", "admin-1", "system-1", "", func(connection InternalChatConnection) map[string]any {
		return map[string]any{
			"type":                "chat.message.created",
			"company_id":          connection.CompanyID,
			"counterpart_user_id": connection.CounterpartUserID,
			"emitted_at":          time.Now().UTC().Format(time.RFC3339),
			"message": map[string]any{
				"id": "message-stress",
			},
		}
	})

	wg.Wait()

	for _, conn := range connections {
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}

	require.Eventually(t, func() bool {
		return hub.TotalConnections() == 0
	}, 5*time.Second, 20*time.Millisecond)

	stats := hub.Stats()
	require.EqualValues(t, totalConnections, stats.TotalConnectionsOpened)
	require.EqualValues(t, totalConnections, stats.TotalConnectionsClosed)
	require.EqualValues(t, 1, stats.TotalBroadcastEvents)
	require.EqualValues(t, totalConnections, stats.TotalBroadcastDeliveries)
}

func BenchmarkInternalChatHubBroadcastConversationEvent(b *testing.B) {
	hub := NewInternalChatHub()
	const totalConnections = 24
	serverConnected := make(chan struct{}, totalConnections)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			Subprotocols:   []string{"petcontrol.internal-chat.v1"},
			OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
		})
		require.NoError(b, err)

		id := r.URL.Query().Get("id")
		role := "admin"
		userID := "admin-1"
		counterpartID := "system-1"
		if r.URL.Query().Get("role") == "system" {
			role = "system"
			userID = "system-1"
			counterpartID = "admin-1"
		}

		hub.Register(InternalChatConnection{
			ID:                id,
			CompanyID:         "company-1",
			UserID:            userID,
			CounterpartUserID: counterpartID,
			UserRole:          role,
			ConnectedAt:       time.Now(),
			Socket:            conn,
		})
		serverConnected <- struct{}{}

		ctx := conn.CloseRead(r.Context())
		<-ctx.Done()
		_, _, _ = hub.Unregister(id)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connections := make([]*websocket.Conn, 0, totalConnections)
	for i := range totalConnections {
		role := "admin"
		if i >= totalConnections/2 {
			role = "system"
		}
		conn, _, err := websocket.Dial(ctx, "ws"+server.URL[len("http"):]+"?id=bench-"+strconv.Itoa(i)+"&role="+role, &websocket.DialOptions{
			Subprotocols: []string{"petcontrol.internal-chat.v1"},
		})
		require.NoError(b, err)
		connections = append(connections, conn)
	}

	for range totalConnections {
		<-serverConnected
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for _, conn := range connections {
			wg.Add(1)
			go func(connection *websocket.Conn) {
				defer wg.Done()
				var event map[string]any
				require.NoError(b, wsjson.Read(ctx, connection, &event))
			}(conn)
		}

		hub.BroadcastConversationEvent(ctx, "company-1", "admin-1", "system-1", "", func(connection InternalChatConnection) map[string]any {
			return map[string]any{
				"type":                "chat.message.created",
				"company_id":          connection.CompanyID,
				"counterpart_user_id": connection.CounterpartUserID,
				"emitted_at":          time.Now().UTC().Format(time.RFC3339),
				"message": map[string]any{
					"id": "bench-" + strconv.Itoa(i),
				},
			}
		})

		wg.Wait()
	}
	b.StopTimer()

	for _, conn := range connections {
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}
}
