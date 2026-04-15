package sqlc_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_EmailVerificationTokens_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	// Setup user
	user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
		Email: uniqueEmail("email-verif"),
		Role:  sqlc.UserRoleTypeAdmin,
	})
	require.NoError(t, err)

	t.Run("Lifecycle", func(t *testing.T) {
		hash := "email-secret-hash"
		ip := netip.MustParseAddr("10.0.0.1")

		_, err := queries.InsertEmailVerificationToken(ctx, sqlc.InsertEmailVerificationTokenParams{
			UserID:           user.ID,
			TokenHash:        hash,
			Email:            user.Email,
			ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
			RequestIP:        &ip,
			RequestUserAgent: pgtype.Text{String: "Mozilla", Valid: true},
		})
		require.NoError(t, err)

		// Get by hash
		token, err := queries.GetEmailVerificationTokenByHash(ctx, hash)
		require.NoError(t, err)
		require.Equal(t, hash, token.TokenHash)
		require.Equal(t, "Mozilla", token.RequestUserAgent.String)

		// Consuming the token
		consumeIP := netip.MustParseAddr("10.0.0.2")
		rows, err := queries.MarkEmailVerificationTokenAsUsed(ctx, sqlc.MarkEmailVerificationTokenAsUsedParams{
			TokenHash:         hash,
			UsedAt:            pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ConsumedIP:        &consumeIP,
			ConsumedUserAgent: pgtype.Text{String: "Chrome", Valid: true},
		})
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Verify consumed data
		token, _ = queries.GetEmailVerificationTokenByHash(ctx, hash)
		require.True(t, token.UsedAt.Valid)
		require.Equal(t, "Chrome", token.ConsumedUserAgent.String)
	})
}
