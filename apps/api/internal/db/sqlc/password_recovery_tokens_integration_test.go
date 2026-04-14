package sqlc_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_PasswordRecoveryTokens_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	// Setup user
	user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
		Email: uniqueEmail("pwd-rec"),
		Role:  sqlc.UserRoleTypeAdmin,
	})
	require.NoError(t, err)

	t.Run("Lifecycle", func(t *testing.T) {
		hash := "secret-hash-123"
		expiresAt := time.Now().Add(1 * time.Hour).Round(time.Microsecond)

		_, err := queries.InsertPasswordRecoveryToken(ctx, sqlc.InsertPasswordRecoveryTokenParams{
			UserID:         user.ID,
			TokenHash:      hash,
			RequestedEmail: user.Email,
			ExpiresAt:      pgtype.Timestamptz{Time: expiresAt, Valid: true},
		})
		require.NoError(t, err)

		// Get by hash
		token, err := queries.GetPasswordRecoveryTokenByHash(ctx, hash)
		require.NoError(t, err)
		require.Equal(t, hash, token.TokenHash)
		require.False(t, token.UsedAt.Valid)

		// List active
		active, err := queries.ListActivePasswordRecoveryTokensByUserID(ctx, sqlc.ListActivePasswordRecoveryTokensByUserIDParams{
			UserID: user.ID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, active, 1)

		// Revoke
		_, err = queries.RevokePasswordRecoveryToken(ctx, sqlc.RevokePasswordRecoveryTokenParams{
			TokenHash: hash,
			RevokedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})
		require.NoError(t, err)

		// Should no longer be active
		active, err = queries.ListActivePasswordRecoveryTokensByUserID(ctx, sqlc.ListActivePasswordRecoveryTokensByUserIDParams{
			UserID: user.ID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, active, 0)
	})

	t.Run("UseToken", func(t *testing.T) {
		hash := "use-me-hash"
		_, _ = queries.InsertPasswordRecoveryToken(ctx, sqlc.InsertPasswordRecoveryTokenParams{
			UserID:         user.ID,
			TokenHash:      hash,
			RequestedEmail: user.Email,
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
		})

		_, err := queries.MarkPasswordRecoveryTokenAsUsed(ctx, sqlc.MarkPasswordRecoveryTokenAsUsedParams{
			TokenHash: hash,
			UsedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})
		require.NoError(t, err)

		// Verify it's used
		token, _ := queries.GetPasswordRecoveryTokenByHash(ctx, hash)
		require.True(t, token.UsedAt.Valid)
	})
}
