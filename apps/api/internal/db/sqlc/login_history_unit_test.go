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

func TestQueries_LoginHistory_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("GetLoginHistoryByID", func(t *testing.T) {
		id := uuidValue()
		rows := pgxmock.NewRows([]string{"id", "user_id", "ip_address", "user_agent", "result", "failure_detail", "attempted_at"}).
			AddRow(id, uuidValue(), netip.MustParseAddr("127.0.0.1"), "agent", sqlc.LoginResultSuccess, pgtype.Text{}, pgtype.Timestamptz{Valid: true})

		mock.ExpectQuery(`(?s)name: GetLoginHistoryByID`).
			WithArgs(id).
			WillReturnRows(rows)

		res, err := queries.GetLoginHistoryByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, res.ID)

		mock.ExpectQuery(`(?s)name: GetLoginHistoryByID`).
			WithArgs(id).
			WillReturnError(errExpected)

		_, err = queries.GetLoginHistoryByID(ctx, id)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("ListLoginHistoryByResult", func(t *testing.T) {
		arg := sqlc.ListLoginHistoryByResultParams{
			Result: sqlc.LoginResultSuccess,
			Limit:  10,
			Offset: 0,
		}

		mock.ExpectQuery(`(?s)name: ListLoginHistoryByResult`).
			WithArgs(arg.Result, arg.Offset, arg.Limit).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "ip_address", "user_agent", "result", "failure_detail", "attempted_at"}).
				AddRow(uuidValue(), uuidValue(), netip.MustParseAddr("127.0.0.1"), "agent", arg.Result, pgtype.Text{}, pgtype.Timestamptz{}))

		res, err := queries.ListLoginHistoryByResult(ctx, arg)
		require.NoError(t, err)
		require.Len(t, res, 1)

		mock.ExpectQuery(`(?s)name: ListLoginHistoryByResult`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.ListLoginHistoryByResult(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("ListLoginHistoryByUserID", func(t *testing.T) {
		arg := sqlc.ListLoginHistoryByUserIDParams{
			UserID: uuidValue(),
			Limit:  5,
			Offset: 0,
		}

		mock.ExpectQuery(`(?s)name: ListLoginHistoryByUserID`).
			WithArgs(arg.UserID, arg.Offset, arg.Limit).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "ip_address", "user_agent", "result", "failure_detail", "attempted_at"}).
				AddRow(uuidValue(), arg.UserID, netip.MustParseAddr("127.0.0.1"), "agent", sqlc.LoginResultSuccess, pgtype.Text{}, pgtype.Timestamptz{}))

		res, err := queries.ListLoginHistoryByUserID(ctx, arg)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
