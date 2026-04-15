package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_PasswordRecoveryTokens_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("GetPasswordRecoveryTokenByHash", func(t *testing.T) {
		hash := "token-hash"
		rows := pgxmock.NewRows([]string{"id", "user_id", "token_hash", "requested_email", "expires_at", "used_at", "revoked_at", "request_ip", "request_user_agent", "triggered_by_user_id"}).
			AddRow(uuidValue(), uuidValue(), hash, "test@example.com", pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{}, pgtype.Timestamptz{}, nil, pgtype.Text{}, uuidValue())

		mock.ExpectQuery(`(?s)name: GetPasswordRecoveryTokenByHash`).
			WithArgs(hash).
			WillReturnRows(rows)

		res, err := queries.GetPasswordRecoveryTokenByHash(ctx, hash)
		require.NoError(t, err)
		require.Equal(t, hash, res.TokenHash)
	})

	t.Run("InsertPasswordRecoveryToken", func(t *testing.T) {
		arg := sqlc.InsertPasswordRecoveryTokenParams{
			UserID:         uuidValue(),
			TokenHash:      "hash",
			RequestedEmail: "test@example.com",
			ExpiresAt:      pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: InsertPasswordRecoveryToken`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		rows, err := queries.InsertPasswordRecoveryToken(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})

	t.Run("ListActivePasswordRecoveryTokensByUserID", func(t *testing.T) {
		userID := uuidValue()
		rows := pgxmock.NewRows([]string{"id", "user_id", "token_hash", "requested_email", "expires_at", "used_at", "revoked_at", "request_ip", "request_user_agent", "triggered_by_user_id"}).
			AddRow(uuidValue(), userID, "hash", "email", pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{}, pgtype.Timestamptz{}, nil, pgtype.Text{}, uuidValue())

		mock.ExpectQuery(`(?s)name: ListActivePasswordRecoveryTokensByUserID`).
			WithArgs(userID, int32(0), int32(10)).
			WillReturnRows(rows)

		res, err := queries.ListActivePasswordRecoveryTokensByUserID(ctx, sqlc.ListActivePasswordRecoveryTokensByUserIDParams{
			UserID: userID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("MarkPasswordRecoveryTokenAsUsed", func(t *testing.T) {
		arg := sqlc.MarkPasswordRecoveryTokenAsUsedParams{
			TokenHash: "hash",
			UsedAt:    pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: MarkPasswordRecoveryTokenAsUsed`).
			WithArgs(arg.UsedAt, arg.TokenHash).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.MarkPasswordRecoveryTokenAsUsed(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})

	t.Run("RevokePasswordRecoveryToken", func(t *testing.T) {
		arg := sqlc.RevokePasswordRecoveryTokenParams{
			TokenHash: "hash",
			RevokedAt: pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: RevokePasswordRecoveryToken`).
			WithArgs(arg.RevokedAt, arg.TokenHash).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		_, err := queries.RevokePasswordRecoveryToken(ctx, arg)
		require.NoError(t, err)

		mock.ExpectExec(`(?s)name: RevokePasswordRecoveryToken`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.RevokePasswordRecoveryToken(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
