package realtime

import (
	"testing"
	"time"

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
