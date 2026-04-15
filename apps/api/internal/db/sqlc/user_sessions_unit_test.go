package sqlc_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_UserSessions_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("GetActiveUserSessionsByUserID", func(t *testing.T) {
		userID := uuidValue()
		rows := pgxmock.NewRows([]string{"id", "user_id", "login_history_id", "session_token", "ip_address", "user_agent", "last_activity_at", "expires_at", "logged_out_at", "logout_reason"}).
			AddRow(uuidValue(), userID, uuidValue(), "token", netip.MustParseAddr("127.0.0.1"), "agent", pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{}, sqlc.NullLogoutReason{})

		mock.ExpectQuery(`(?s)name: GetActiveUserSessionsByUserID`).
			WithArgs(userID).
			WillReturnRows(rows)

		res, err := queries.GetActiveUserSessionsByUserID(ctx, userID)
		require.NoError(t, err)
		require.Len(t, res, 1)

		// Failure
		mock.ExpectQuery(`(?s)name: GetActiveUserSessionsByUserID`).
			WithArgs(userID).
			WillReturnError(errExpected)

		_, err = queries.GetActiveUserSessionsByUserID(ctx, userID)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("GetUserSessionByToken", func(t *testing.T) {
		token := "some-token"
		rows := pgxmock.NewRows([]string{"id", "user_id", "login_history_id", "session_token", "ip_address", "user_agent", "last_activity_at", "expires_at", "logged_out_at", "logout_reason"}).
			AddRow(uuidValue(), uuidValue(), uuidValue(), token, netip.MustParseAddr("127.0.0.1"), "agent", pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{}, sqlc.NullLogoutReason{})

		mock.ExpectQuery(`(?s)name: GetUserSessionByToken`).
			WithArgs(token).
			WillReturnRows(rows)

		res, err := queries.GetUserSessionByToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, token, res.SessionToken)

		// Failure
		mock.ExpectQuery(`(?s)name: GetUserSessionByToken`).
			WithArgs(token).
			WillReturnError(errExpected)

		_, err = queries.GetUserSessionByToken(ctx, token)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("InsertUserSession", func(t *testing.T) {
		arg := sqlc.InsertUserSessionParams{
			UserID:         uuidValue(),
			LoginHistoryID: uuidValue(),
			SessionToken:   "new-token",
			IPAddress:      netip.MustParseAddr("127.0.0.1"),
			UserAgent:      "agent",
			ExpiresAt:      pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: InsertUserSession`).
			WithArgs(arg.UserID, arg.LoginHistoryID, arg.SessionToken, arg.IPAddress, arg.UserAgent, arg.LastActivityAt, arg.ExpiresAt, arg.LoggedOutAt, arg.LogoutReason).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		rows, err := queries.InsertUserSession(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Failure
		mock.ExpectExec(`(?s)name: InsertUserSession`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.InsertUserSession(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("UpdateUserSession", func(t *testing.T) {
		arg := sqlc.UpdateUserSessionParams{
			SessionToken:   "token",
			LastActivityAt: pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: UpdateUserSession`).
			WithArgs(arg.LastActivityAt, arg.IPAddress, arg.UserAgent, arg.ExpiresAt, arg.LoggedOutAt, arg.LogoutReason, arg.SessionToken).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.UpdateUserSession(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		mock.ExpectExec(`(?s)name: UpdateUserSession`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.UpdateUserSession(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
