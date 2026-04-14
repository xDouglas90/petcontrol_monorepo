package sqlc_test

import (

	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Auth_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	// Setup: Create a user and its auth record
	email := uniqueEmail("auth-integ")
	user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
		Email:    email,
		Role:     sqlc.UserRoleTypeAdmin,
		IsActive: true,
	})
	require.NoError(t, err)

	err = queries.InsertUserAuth(ctx, sqlc.InsertUserAuthParams{
		UserID:             user.ID,
		PasswordHash:       "hashed-password",
		MustChangePassword: pgtype.Bool{Bool: false, Valid: true},
	})
	require.NoError(t, err)

	t.Run("GetUserAuthByUserID", func(t *testing.T) {
		auth, err := queries.GetUserAuthByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, user.ID, auth.UserID)
		require.Equal(t, "hashed-password", auth.PasswordHash)
		require.Equal(t, int16(0), auth.LoginAttempts)
	})

	t.Run("LoginAttempts_Management", func(t *testing.T) {
		// Increment attempts
		err := queries.IncrementUserAuthLoginAttempts(ctx, user.ID)
		require.NoError(t, err)

		auth, err := queries.GetUserAuthByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, int16(1), auth.LoginAttempts)

		// Set locked until
		until := time.Now().Add(1 * time.Hour).Round(time.Microsecond)
		err = queries.SetUserAuthLockedUntil(ctx, sqlc.SetUserAuthLockedUntilParams{
			UserID:      user.ID,
			LockedUntil: pgtype.Timestamptz{Time: until, Valid: true},
		})
		require.NoError(t, err)

		auth, err = queries.GetUserAuthByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.True(t, auth.LockedUntil.Valid)
		// PostgreSQL might have slight precision diffs depending on type, but Round helps.
		require.WithinDuration(t, until, auth.LockedUntil.Time, time.Second)

		// Reset attempts
		err = queries.ResetUserAuthLoginAttempts(ctx, user.ID)
		require.NoError(t, err)

		auth, err = queries.GetUserAuthByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, int16(0), auth.LoginAttempts)
		require.False(t, auth.LockedUntil.Valid)
		require.True(t, auth.LastLoginAt.Valid)
	})

	t.Run("InsertLoginHistory", func(t *testing.T) {
		arg := sqlc.InsertLoginHistoryParams{
			UserID:    user.ID,
			IPAddress: netip.MustParseAddr("192.168.1.1"),
			UserAgent: "TestAgent",
			Result:    sqlc.LoginResultSuccess,
		}

		err := queries.InsertLoginHistory(ctx, arg)
		require.NoError(t, err)

		// We don't have a GetLoginHistoryByUserID in auth.sql.go yet, 
		// but we can verify it doesn't fail. 
		// In a real scenario, we'd check the DB directly if needed.
	})
}
