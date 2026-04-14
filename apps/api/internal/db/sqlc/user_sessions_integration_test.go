package sqlc_test

import (

	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_UserSessions_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	// Setup: Create user and login history
	email := uniqueEmail("sessions-integ")
	user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
		Email: email,
		Role:  sqlc.UserRoleTypeAdmin,
	})
	require.NoError(t, err)

	var loginHistID pgtype.UUID
	err = pool.QueryRow(ctx, 
		"INSERT INTO login_history(user_id, ip_address, user_agent, result) VALUES ($1, $2, $3, $4) RETURNING id",
		user.ID, "127.0.0.1", "TestAgent", sqlc.LoginResultSuccess,
	).Scan(&loginHistID)
	require.NoError(t, err)

	t.Run("InsertAndGetSessionByToken", func(t *testing.T) {
		token := "session-token-123"
		expiresAt := time.Now().Add(24 * time.Hour).Round(time.Microsecond)
		
		_, err := queries.InsertUserSession(ctx, sqlc.InsertUserSessionParams{
			UserID:         user.ID,
			LoginHistoryID: loginHistID,
			SessionToken:   token,
			IPAddress:      netip.MustParseAddr("127.0.0.1"),
			UserAgent:      "TestAgent",
			LastActivityAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: expiresAt, Valid: true},
		})
		require.NoError(t, err)

		session, err := queries.GetUserSessionByToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, token, session.SessionToken)
		require.Equal(t, user.ID, session.UserID)
		require.WithinDuration(t, expiresAt, session.ExpiresAt.Time, time.Second)
	})

	t.Run("GetActiveUserSessionsByUserID", func(t *testing.T) {
		tokenActive := "active-token"
		tokenExpired := "expired-token"
		tokenLoggedOut := "logged-out-token"

		now := time.Now()
		
		// Active
		_, _ = queries.InsertUserSession(ctx, sqlc.InsertUserSessionParams{
			UserID: user.ID, LoginHistoryID: loginHistID, SessionToken: tokenActive,
			IPAddress:      netip.MustParseAddr("127.0.0.1"),
			UserAgent:      "TestAgent",
			LastActivityAt: pgtype.Timestamptz{Time: now, Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: now.Add(1 * time.Hour), Valid: true},
		})
		// Expired
		_, _ = queries.InsertUserSession(ctx, sqlc.InsertUserSessionParams{
			UserID: user.ID, LoginHistoryID: loginHistID, SessionToken: tokenExpired,
			IPAddress:      netip.MustParseAddr("127.0.0.1"),
			UserAgent:      "TestAgent",
			LastActivityAt: pgtype.Timestamptz{Time: now, Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: now.Add(-1 * time.Hour), Valid: true},
		})
		// Logged Out
		_, _ = queries.InsertUserSession(ctx, sqlc.InsertUserSessionParams{
			UserID: user.ID, LoginHistoryID: loginHistID, SessionToken: tokenLoggedOut,
			IPAddress:      netip.MustParseAddr("127.0.0.1"),
			UserAgent:      "TestAgent",
			LastActivityAt: pgtype.Timestamptz{Time: now, Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: now.Add(1 * time.Hour), Valid: true},
			LoggedOutAt:    pgtype.Timestamptz{Time: now, Valid: true},
		})

		activeSessions, err := queries.GetActiveUserSessionsByUserID(ctx, user.ID)
		require.NoError(t, err)
		
		// Should find tokenActive and the one from the previous subtest if it's still alive
		foundActive := false
		for _, s := range activeSessions {
			if s.SessionToken == tokenActive {
				foundActive = true
			}
			require.False(t, s.LoggedOutAt.Valid)
			require.True(t, s.ExpiresAt.Time.After(now))
		}
		require.True(t, foundActive)
	})

	t.Run("UpdateUserSession_Logout", func(t *testing.T) {
		token := "logout-test-token"
		_, err = queries.InsertUserSession(ctx, sqlc.InsertUserSessionParams{
			UserID: user.ID, LoginHistoryID: loginHistID, SessionToken: token,
			IPAddress:      netip.MustParseAddr("127.0.0.1"),
			UserAgent:      "TestAgent",
			LastActivityAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
		})
		require.NoError(t, err)

		logoutTime := time.Now().Round(time.Microsecond)
		rows, err := queries.UpdateUserSession(ctx, sqlc.UpdateUserSessionParams{
			SessionToken: token,
			LoggedOutAt:  pgtype.Timestamptz{Time: logoutTime, Valid: true},
			LogoutReason: sqlc.NullLogoutReason{LogoutReason: sqlc.LogoutReasonUserInitiated, Valid: true},
		})
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		updated, err := queries.GetUserSessionByToken(ctx, token)
		require.NoError(t, err)
		require.True(t, updated.LoggedOutAt.Valid)
		require.WithinDuration(t, logoutTime, updated.LoggedOutAt.Time, time.Second)
		require.Equal(t, sqlc.LogoutReasonUserInitiated, updated.LogoutReason.LogoutReason)
	})
}
