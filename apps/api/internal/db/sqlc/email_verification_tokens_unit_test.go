package sqlc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_EmailVerificationTokens_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("GetEmailVerificationTokenByHash", func(t *testing.T) {
		hash := "email-hash"
		rows := pgxmock.NewRows([]string{"id", "user_id", "token_hash", "email", "expires_at", "used_at", "revoked_at", "request_ip", "request_user_agent", "consumed_ip", "consumed_user_agent"}).
			AddRow(uuidValue(), uuidValue(), hash, "test@example.com", pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{}, pgtype.Timestamptz{}, nil, pgtype.Text{}, nil, pgtype.Text{})

		mock.ExpectQuery(`(?s)name: GetEmailVerificationTokenByHash`).
			WithArgs(hash).
			WillReturnRows(rows)

		res, err := queries.GetEmailVerificationTokenByHash(ctx, hash)
		require.NoError(t, err)
		require.Equal(t, hash, res.TokenHash)

		// Failure
		mock.ExpectQuery(`(?s)name: GetEmailVerificationTokenByHash`).
			WithArgs(hash).
			WillReturnError(errExpected)

		_, err = queries.GetEmailVerificationTokenByHash(ctx, hash)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("InsertEmailVerificationToken", func(t *testing.T) {
		arg := sqlc.InsertEmailVerificationTokenParams{
			UserID:    uuidValue(),
			TokenHash: "hash",
			Email:     "test@example.com",
			ExpiresAt: pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: InsertEmailVerificationToken`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		rows, err := queries.InsertEmailVerificationToken(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Failure
		mock.ExpectExec(`(?s)name: InsertEmailVerificationToken`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.InsertEmailVerificationToken(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("ListActiveEmailVerificationTokensByUserID", func(t *testing.T) {
		userID := uuidValue()
		mock.ExpectQuery(`(?s)name: ListActiveEmailVerificationTokensByUserID`).
			WithArgs(userID, int32(0), int32(10)).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "token_hash", "email", "expires_at", "used_at", "revoked_at", "request_ip", "request_user_agent", "consumed_ip", "consumed_user_agent"}).
				AddRow(uuidValue(), userID, "hash", "test@example.com", pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}, pgtype.Timestamptz{}, pgtype.Timestamptz{}, nil, pgtype.Text{}, nil, pgtype.Text{}))

		res, err := queries.ListActiveEmailVerificationTokensByUserID(ctx, sqlc.ListActiveEmailVerificationTokensByUserIDParams{
			UserID: userID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, res, 1)

		// Failure
		mock.ExpectQuery(`(?s)name: ListActiveEmailVerificationTokensByUserID`).
			WithArgs(userID, int32(0), int32(10)).
			WillReturnError(errExpected)

		_, err = queries.ListActiveEmailVerificationTokensByUserID(ctx, sqlc.ListActiveEmailVerificationTokensByUserIDParams{
			UserID: userID,
			Limit:  10,
			Offset: 0,
		})
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("MarkEmailVerificationTokenAsUsed", func(t *testing.T) {
		arg := sqlc.MarkEmailVerificationTokenAsUsedParams{
			TokenHash: "hash",
			UsedAt:    pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: MarkEmailVerificationTokenAsUsed`).
			WithArgs(arg.UsedAt, arg.ConsumedIP, arg.ConsumedUserAgent, arg.TokenHash).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.MarkEmailVerificationTokenAsUsed(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Failure
		mock.ExpectExec(`(?s)name: MarkEmailVerificationTokenAsUsed`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.MarkEmailVerificationTokenAsUsed(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("RevokeEmailVerificationToken", func(t *testing.T) {
		arg := sqlc.RevokeEmailVerificationTokenParams{
			TokenHash: "hash",
			RevokedAt: pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: RevokeEmailVerificationToken`).
			WithArgs(arg.RevokedAt, arg.TokenHash).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		_, err := queries.RevokeEmailVerificationToken(ctx, arg)
		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
